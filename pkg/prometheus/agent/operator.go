// Copyright 2023 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheusagent

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	monitoringclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/prometheus-operator/prometheus-operator/pkg/informers"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/listwatch"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
	"github.com/prometheus-operator/prometheus-operator/pkg/webconfig"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncPeriod = 5 * time.Minute
)

// Operator manages life cycle of Prometheus agent deployments and
// monitoring configurations.
type Operator struct {
	kclient  kubernetes.Interface
	mclient  monitoringclient.Interface
	logger   log.Logger
	accessor *operator.Accessor

	nsPromInf cache.SharedIndexInformer
	nsMonInf  cache.SharedIndexInformer

	promInfs  *informers.ForResource
	smonInfs  *informers.ForResource
	pmonInfs  *informers.ForResource
	probeInfs *informers.ForResource
	cmapInfs  *informers.ForResource
	secrInfs  *informers.ForResource
	ssetInfs  *informers.ForResource

	rr *operator.ResourceReconciler

	metrics         *operator.Metrics
	reconciliations *operator.ReconciliationTracker

	host                   string
	config                 operator.Config
	endpointSliceSupported bool
}

// New creates a new controller.
func New(ctx context.Context, conf operator.Config, logger log.Logger, r prometheus.Registerer) (*Operator, error) {
	cfg, err := k8sutil.NewClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating cluster config failed")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	mclient, err := monitoringclient.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}

	if _, err := labels.Parse(conf.PromSelector); err != nil {
		return nil, errors.Wrap(err, "can not parse prometheus-agent selector value")
	}

	secretListWatchSelector, err := fields.ParseSelector(conf.SecretListWatchSelector)
	if err != nil {
		return nil, errors.Wrap(err, "can not parse secrets selector value")
	}

	// All the metrics exposed by the controller get the controller="prometheus-agent" label.
	r = prometheus.WrapRegistererWith(prometheus.Labels{"controller": "prometheus-agent"}, r)

	c := &Operator{
		kclient:         client,
		mclient:         mclient,
		logger:          logger,
		host:            cfg.Host,
		config:          conf,
		metrics:         operator.NewMetrics(r),
		reconciliations: &operator.ReconciliationTracker{},
	}
	c.metrics.MustRegister(
		c.reconciliations,
	)

	c.rr = operator.NewResourceReconciler(
		c.logger,
		c,
		c.metrics,
		monitoringv1alpha1.PrometheusAgentsKind,
		r,
	)

	c.promInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = c.config.PromSelector
			},
		),
		monitoringv1alpha1.SchemeGroupVersion.WithResource(monitoringv1alpha1.PrometheusAgentName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating prometheus-agent informers")
	}

	var promStores []cache.Store
	for _, informer := range c.promInfs.GetInformers() {
		promStores = append(promStores, informer.Informer().GetStore())
	}

	c.metrics.MustRegister(prompkg.NewCollectorForStores(promStores...))

	c.smonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ServiceMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating servicemonitor informers")
	}

	c.pmonInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.PodMonitorName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating podmonitor informers")
	}

	c.probeInfs, err = informers.NewInformersForResource(
		informers.NewMonitoringInformerFactories(
			c.config.Namespaces.AllowList,
			c.config.Namespaces.DenyList,
			mclient,
			resyncPeriod,
			nil,
		),
		monitoringv1.SchemeGroupVersion.WithResource(monitoringv1.ProbeName),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating probe informers")
	}

	c.cmapInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.LabelSelector = prompkg.LabelPrometheusName
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceConfigMaps)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating configmap informers")
	}

	c.secrInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			func(options *metav1.ListOptions) {
				options.FieldSelector = secretListWatchSelector.String()
			},
		),
		v1.SchemeGroupVersion.WithResource(string(v1.ResourceSecrets)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating secrets informers")
	}

	c.ssetInfs, err = informers.NewInformersForResource(
		informers.NewKubeInformerFactories(
			c.config.Namespaces.PrometheusAllowList,
			c.config.Namespaces.DenyList,
			c.kclient,
			resyncPeriod,
			nil,
		),
		appsv1.SchemeGroupVersion.WithResource("statefulsets"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating statefulset informers")
	}

	newNamespaceInformer := func(o *Operator, allowList map[string]struct{}) cache.SharedIndexInformer {
		// nsResyncPeriod is used to control how often the namespace informer
		// should resync. If the unprivileged ListerWatcher is used, then the
		// informer must resync more often because it cannot watch for
		// namespace changes.
		nsResyncPeriod := 15 * time.Second
		// If the only namespace is v1.NamespaceAll, then the client must be
		// privileged and a regular cache.ListWatch will be used. In this case
		// watching works and we do not need to resync so frequently.
		if listwatch.IsAllNamespaces(allowList) {
			nsResyncPeriod = resyncPeriod
		}
		nsInf := cache.NewSharedIndexInformer(
			o.metrics.NewInstrumentedListerWatcher(
				listwatch.NewUnprivilegedNamespaceListWatchFromClient(ctx, o.logger, o.kclient.CoreV1().RESTClient(), allowList, o.config.Namespaces.DenyList, fields.Everything()),
			),
			&v1.Namespace{}, nsResyncPeriod, cache.Indexers{},
		)

		return nsInf
	}
	c.nsMonInf = newNamespaceInformer(c, c.config.Namespaces.AllowList)
	if listwatch.IdenticalNamespaces(c.config.Namespaces.AllowList, c.config.Namespaces.PrometheusAllowList) {
		c.nsPromInf = c.nsMonInf
	} else {
		c.nsPromInf = newNamespaceInformer(c, c.config.Namespaces.PrometheusAllowList)
	}

	endpointSliceSupported, err := k8sutil.IsAPIGroupVersionResourceSupported(c.kclient.Discovery(), "discovery.k8s.io", "endpointslices")
	if err != nil {
		level.Warn(c.logger).Log("msg", "failed to check if the API supports the endpointslice resources", "err ", err)
	}
	level.Info(c.logger).Log("msg", "Kubernetes API capabilities", "endpointslices", endpointSliceSupported)
	c.endpointSliceSupported = endpointSliceSupported
	return c, nil
}

// Run the controller.
func (c *Operator) Run(ctx context.Context) error {
	crdInstalled, err := k8sutil.IsAPIGroupVersionResourceSupported(c.kclient.Discovery(), monitoringv1alpha1.SchemeGroupVersion.String(), monitoringv1alpha1.PrometheusAgentName)
	if err != nil {
		level.Warn(c.logger).Log("msg", "failed to check if the API supports the PrometheusAgent CRD", "err ", err)
		return nil
	}
	if !crdInstalled {
		level.Info(c.logger).Log("msg", "Prometheus agent controller disabled because the PrometheusAgent CRD isn't installed")
		return nil
	}

	missingPermissions, err := c.getMissingPermissions(ctx)
	if err != nil {
		return err
	}
	if len(missingPermissions) > 0 {
		level.Warn(c.logger).Log("msg", "Prometheus agent controller disabled because it lacks the required permissions on PrometheusAgent objects", "missingpermissions", fmt.Sprintf("%v", missingPermissions))
		return nil
	}

	errChan := make(chan error)
	go func() {
		v, err := c.kclient.Discovery().ServerVersion()
		if err != nil {
			errChan <- errors.Wrap(err, "communicating with server failed")
			return
		}
		level.Info(c.logger).Log("msg", "connection established", "cluster-version", v)
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		level.Info(c.logger).Log("msg", "CRD API endpoints ready")
	case <-ctx.Done():
		return nil
	}

	go c.rr.Run(ctx)
	defer c.rr.Stop()

	go c.promInfs.Start(ctx.Done())
	go c.smonInfs.Start(ctx.Done())
	go c.pmonInfs.Start(ctx.Done())
	go c.probeInfs.Start(ctx.Done())
	go c.cmapInfs.Start(ctx.Done())
	go c.secrInfs.Start(ctx.Done())
	go c.ssetInfs.Start(ctx.Done())
	go c.nsMonInf.Run(ctx.Done())
	if c.nsPromInf != c.nsMonInf {
		go c.nsPromInf.Run(ctx.Done())
	}
	if err := c.waitForCacheSync(ctx); err != nil {
		return err
	}

	// Refresh the status of the existing Prometheus agent objects.
	_ = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		c.rr.EnqueueForStatus(obj.(*monitoringv1alpha1.PrometheusAgent))
	})

	c.addHandlers()

	// Run a goroutine that refreshes regularly the Prometheus objects that
	// aren't fully available to keep the status up-to-date with the pod
	// conditions. In practice when a new version of the statefulset is rolled
	// out and the updated pod is crashlooping, the statefulset status won't
	// see any update because the number of ready/updated replicas doesn't
	// change. Without the periodic refresh, the Prometheus object's status
	// would report "containers with incomplete status: [init-config-reloader]"
	// forever.
	// TODO(simonpasquier): watch for Prometheus pods instead of polling.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := c.promInfs.ListAll(labels.Everything(), func(o interface{}) {
					p := o.(*monitoringv1alpha1.PrometheusAgent)
					for _, cond := range p.Status.Conditions {
						if cond.Type == monitoringv1.Available && cond.Status != monitoringv1.ConditionTrue {
							c.rr.EnqueueForStatus(p)
							break
						}
					}
				})
				if err != nil {
					level.Error(c.logger).Log("msg", "failed to list PrometheusAgent objects", "err", err)
				}
			}
		}
	}()

	c.metrics.Ready().Set(1)
	<-ctx.Done()
	return nil
}

// waitForCacheSync waits for the informers' caches to be synced.
func (c *Operator) waitForCacheSync(ctx context.Context) error {
	for _, infs := range []struct {
		name                 string
		informersForResource *informers.ForResource
	}{
		{"PrometheusAgent", c.promInfs},
		{"ServiceMonitor", c.smonInfs},
		{"PodMonitor", c.pmonInfs},
		{"Probe", c.probeInfs},
		{"ConfigMap", c.cmapInfs},
		{"Secret", c.secrInfs},
		{"StatefulSet", c.ssetInfs},
	} {
		for _, inf := range infs.informersForResource.GetInformers() {
			if !operator.WaitForNamedCacheSync(ctx, "prometheusagent", log.With(c.logger, "informer", infs.name), inf.Informer()) {
				return errors.Errorf("failed to sync cache for %s informer", infs.name)
			}
		}
	}

	for _, inf := range []struct {
		name     string
		informer cache.SharedIndexInformer
	}{
		{"PromNamespace", c.nsPromInf},
		{"MonNamespace", c.nsMonInf},
	} {
		if !operator.WaitForNamedCacheSync(ctx, "prometheusagent", log.With(c.logger, "informer", inf.name), inf.informer) {
			return errors.Errorf("failed to sync cache for %s informer", inf.name)
		}
	}

	level.Info(c.logger).Log("msg", "successfully synced all caches")
	return nil
}

// addHandlers adds the eventhandlers to the informers.
func (c *Operator) addHandlers() {
	c.promInfs.AddEventHandler(c.rr)

	c.ssetInfs.AddEventHandler(c.rr)

	c.smonInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSmonAdd,
		DeleteFunc: c.handleSmonDelete,
		UpdateFunc: c.handleSmonUpdate,
	})

	c.pmonInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePmonAdd,
		DeleteFunc: c.handlePmonDelete,
		UpdateFunc: c.handlePmonUpdate,
	})
	c.probeInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleBmonAdd,
		UpdateFunc: c.handleBmonUpdate,
		DeleteFunc: c.handleBmonDelete,
	})
	c.cmapInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleConfigMapAdd,
		DeleteFunc: c.handleConfigMapDelete,
		UpdateFunc: c.handleConfigMapUpdate,
	})
	c.secrInfs.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleSecretAdd,
		DeleteFunc: c.handleSecretDelete,
		UpdateFunc: c.handleSecretUpdate,
	})

	// The controller needs to watch the namespaces in which the service/pod
	// monitors and rules live because a label change on a namespace may
	// trigger a configuration change.
	// It doesn't need to watch on addition/deletion though because it's
	// already covered by the event handlers on service/pod monitors and rules.
	_, _ = c.nsMonInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.handleMonitorNamespaceUpdate,
	})
}

// Resolve implements the operator.Syncer interface.
func (c *Operator) Resolve(ss *appsv1.StatefulSet) metav1.Object {
	key, ok := c.accessor.MetaNamespaceKey(ss)
	if !ok {
		return nil
	}

	match, promKey := prompkg.StatefulSetKeyToPrometheusKey(key)
	if !match {
		level.Debug(c.logger).Log("msg", "StatefulSet key did not match a Prometheus key format", "key", key)
		return nil
	}

	p, err := c.promInfs.Get(promKey)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		level.Error(c.logger).Log("msg", "Prometheus lookup failed", "err", err)
		return nil
	}

	return p.(*monitoringv1alpha1.PrometheusAgent)
}

// Sync implements the operator.Syncer interface.
func (c *Operator) Sync(ctx context.Context, key string) error {
	err := c.sync(ctx, key)
	c.reconciliations.SetStatus(key, err)

	return err
}

func (c *Operator) sync(ctx context.Context, key string) error {
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		c.reconciliations.ForgetObject(key)
		// Dependent resources are cleaned up by K8s via OwnerReferences
		return nil
	}
	if err != nil {
		return err
	}

	p := pobj.(*monitoringv1alpha1.PrometheusAgent)
	p = p.DeepCopy()
	if err := k8sutil.AddTypeInformationToObject(p); err != nil {
		return errors.Wrap(err, "failed to set Prometheus type information")
	}

	logger := log.With(c.logger, "key", key)
	if p.Spec.Paused {
		level.Info(logger).Log("msg", "the resource is paused, not reconciling")
		return nil
	}

	level.Info(logger).Log("msg", "sync prometheus")

	cg, err := prompkg.NewConfigGenerator(c.logger, p, c.endpointSliceSupported)
	if err != nil {
		return err
	}

	assetStore := assets.NewStore(c.kclient.CoreV1(), c.kclient.CoreV1())
	if err := c.createOrUpdateConfigurationSecret(ctx, p, cg, assetStore); err != nil {
		return errors.Wrap(err, "creating config failed")
	}

	tlsAssets, err := c.createOrUpdateTLSAssetSecrets(ctx, p, assetStore)
	if err != nil {
		return errors.Wrap(err, "creating tls asset secret failed")
	}

	if err := c.createOrUpdateWebConfigSecret(ctx, p); err != nil {
		return errors.Wrap(err, "synchronizing web config secret failed")
	}

	// Create governing service if it doesn't exist.
	svcClient := c.kclient.CoreV1().Services(p.Namespace)
	if err := k8sutil.CreateOrUpdateService(ctx, svcClient, makeStatefulSetService(p, c.config)); err != nil {
		return errors.Wrap(err, "synchronizing governing service failed")
	}

	ssetClient := c.kclient.AppsV1().StatefulSets(p.Namespace)

	// Ensure we have a StatefulSet running Prometheus Agent deployed and that StatefulSet names are created correctly.
	expected := prompkg.ExpectedStatefulSetShardNames(p)
	for shard, ssetName := range expected {
		logger := log.With(logger, "statefulset", ssetName, "shard", fmt.Sprintf("%d", shard))
		level.Debug(logger).Log("msg", "reconciling statefulset")

		obj, err := c.ssetInfs.Get(prompkg.KeyToStatefulSetKey(p, key, shard))
		exists := !apierrors.IsNotFound(err)
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "retrieving statefulset failed")
		}

		existingStatefulSet := &appsv1.StatefulSet{}
		if obj != nil {
			existingStatefulSet = obj.(*appsv1.StatefulSet)
			if c.rr.DeletionInProgress(existingStatefulSet) {
				// We want to avoid entering a hot-loop of update/delete cycles
				// here since the sts was marked for deletion in foreground,
				// which means it may take some time before the finalizers
				// complete and the resource disappears from the API. The
				// deletion timestamp will have been set when the initial
				// delete request was issued. In that case, we avoid further
				// processing.
				continue
			}
		}

		newSSetInputHash, err := createSSetInputHash(*p, c.config, tlsAssets, existingStatefulSet.Spec)
		if err != nil {
			return err
		}

		sset, err := makeStatefulSet(
			logger,
			ssetName,
			p,
			&c.config,
			cg,
			newSSetInputHash,
			int32(shard),
			tlsAssets.ShardNames())
		if err != nil {
			return errors.Wrap(err, "making statefulset failed")
		}
		operator.SanitizeSTS(sset)

		if !exists {
			level.Debug(logger).Log("msg", "no current statefulset found")
			level.Debug(logger).Log("msg", "creating statefulset")
			if _, err := ssetClient.Create(ctx, sset, metav1.CreateOptions{}); err != nil {
				return errors.Wrap(err, "creating statefulset failed")
			}
			continue
		}

		if newSSetInputHash == existingStatefulSet.ObjectMeta.Annotations[prompkg.SSetInputHashName] {
			level.Debug(logger).Log("msg", "new statefulset generation inputs match current, skipping any actions")
			continue
		}

		level.Debug(logger).Log(
			"msg", "updating current statefulset because of hash divergence",
			"new_hash", newSSetInputHash,
			"existing_hash", existingStatefulSet.ObjectMeta.Annotations[prompkg.SSetInputHashName],
		)

		err = k8sutil.UpdateStatefulSet(ctx, ssetClient, sset)
		sErr, ok := err.(*apierrors.StatusError)

		if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
			c.metrics.StsDeleteCreateCounter().Inc()

			// Gather only reason for failed update
			failMsg := make([]string, len(sErr.ErrStatus.Details.Causes))
			for i, cause := range sErr.ErrStatus.Details.Causes {
				failMsg[i] = cause.Message
			}

			level.Info(logger).Log("msg", "recreating StatefulSet because the update operation wasn't possible", "reason", strings.Join(failMsg, ", "))

			propagationPolicy := metav1.DeletePropagationForeground
			if err := ssetClient.Delete(ctx, sset.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
				return errors.Wrap(err, "failed to delete StatefulSet to avoid forbidden action")
			}
			continue
		}

		if err != nil {
			return errors.Wrap(err, "updating StatefulSet failed")
		}
	}

	ssets := map[string]struct{}{}
	for _, ssetName := range expected {
		ssets[ssetName] = struct{}{}
	}

	err = c.ssetInfs.ListAllByNamespace(p.Namespace, labels.SelectorFromSet(labels.Set{prompkg.PrometheusNameLabelName: p.Name, prompkg.PrometheusModeLabeLName: prometheusMode}), func(obj interface{}) {
		s := obj.(*appsv1.StatefulSet)

		if _, ok := ssets[s.Name]; ok {
			// Do not delete statefulsets that we still expect to exist. This
			// is to cleanup StatefulSets when shards are reduced.
			return
		}

		if c.rr.DeletionInProgress(s) {
			return
		}

		propagationPolicy := metav1.DeletePropagationForeground
		if err := ssetClient.Delete(ctx, s.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}); err != nil {
			level.Error(c.logger).Log("err", err, "name", s.GetName(), "namespace", s.GetNamespace())
		}
	})
	if err != nil {
		return errors.Wrap(err, "listing StatefulSet resources failed")
	}

	return nil
}

func (c *Operator) createOrUpdateConfigurationSecret(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, cg *prompkg.ConfigGenerator, store *assets.Store) error {
	resourceSelector := prompkg.NewResourceSelector(c.logger, p, store, c.pmonInfs, c.smonInfs, c.probeInfs, c.nsMonInf, c.metrics)
	smons, err := resourceSelector.SelectServiceMonitors(ctx)
	if err != nil {
		return errors.Wrap(err, "selecting ServiceMonitors failed")
	}

	pmons, err := resourceSelector.SelectPodMonitors(ctx)
	if err != nil {
		return errors.Wrap(err, "selecting PodMonitors failed")
	}

	bmons, err := resourceSelector.SelectProbes(ctx)
	if err != nil {
		return errors.Wrap(err, "selecting Probes failed")
	}
	sClient := c.kclient.CoreV1().Secrets(p.Namespace)
	SecretsInPromNS, err := sClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i, remote := range p.Spec.RemoteWrite {
		if err := prompkg.ValidateRemoteWriteSpec(remote); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		key := fmt.Sprintf("remoteWrite/%d", i)
		if err := store.AddBasicAuth(ctx, p.GetNamespace(), remote.BasicAuth, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddOAuth2(ctx, p.GetNamespace(), remote.OAuth2, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddTLSConfig(ctx, p.GetNamespace(), remote.TLSConfig); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddAuthorizationCredentials(ctx, p.GetNamespace(), remote.Authorization, fmt.Sprintf("remoteWrite/auth/%d", i)); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
		if err := store.AddSigV4(ctx, p.GetNamespace(), remote.Sigv4, key); err != nil {
			return errors.Wrapf(err, "remote write %d", i)
		}
	}

	if p.Spec.APIServerConfig != nil {
		if err := store.AddBasicAuth(ctx, p.GetNamespace(), p.Spec.APIServerConfig.BasicAuth, "apiserver"); err != nil {
			return errors.Wrap(err, "apiserver config")
		}
		if err := store.AddAuthorizationCredentials(ctx, p.GetNamespace(), p.Spec.APIServerConfig.Authorization, "apiserver/auth"); err != nil {
			return errors.Wrapf(err, "apiserver config")
		}
	}

	additionalScrapeConfigs, err := c.loadConfigFromSecret(p.Spec.AdditionalScrapeConfigs, SecretsInPromNS)
	if err != nil {
		return errors.Wrap(err, "loading additional scrape configs from Secret failed")
	}

	// Update secret based on the most recent configuration.
	conf, err := cg.GenerateAgentConfiguration(
		smons,
		pmons,
		bmons,
		store,
		additionalScrapeConfigs,
	)
	if err != nil {
		return errors.Wrap(err, "generating config failed")
	}

	s := prompkg.MakeConfigSecret(p, c.config)
	s.ObjectMeta.Annotations = map[string]string{
		"generated": "true",
	}

	// Compress config to avoid 1mb secret limit for a while
	var buf bytes.Buffer
	if err = operator.GzipConfig(&buf, conf); err != nil {
		return errors.Wrap(err, "couldn't gzip config")
	}
	s.Data[prompkg.ConfigFilename] = buf.Bytes()

	level.Debug(c.logger).Log("msg", "updating Prometheus configuration secret")

	return k8sutil.CreateOrUpdateSecret(ctx, sClient, s)
}

func createSSetInputHash(p monitoringv1alpha1.PrometheusAgent, c operator.Config, tlsAssets *operator.ShardedSecret, ssSpec appsv1.StatefulSetSpec) (string, error) {
	var http2 *bool
	if p.Spec.Web != nil && p.Spec.Web.WebConfigFileFields.HTTPConfig != nil {
		http2 = p.Spec.Web.WebConfigFileFields.HTTPConfig.HTTP2
	}

	hash, err := hashstructure.Hash(struct {
		PrometheusLabels      map[string]string
		PrometheusAnnotations map[string]string
		PrometheusGeneration  int64
		PrometheusWebHTTP2    *bool
		Config                operator.Config
		StatefulSetSpec       appsv1.StatefulSetSpec
		Assets                []string `hash:"set"`
	}{
		PrometheusLabels:      p.Labels,
		PrometheusAnnotations: p.Annotations,
		PrometheusGeneration:  p.Generation,
		PrometheusWebHTTP2:    http2,
		Config:                c,
		StatefulSetSpec:       ssSpec,
		Assets:                tlsAssets.ShardNames(),
	},
		nil,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to calculate combined hash")
	}

	return fmt.Sprintf("%d", hash), nil
}

// UpdateStatus updates the status subresource of the object identified by the given
// key.
// UpdateStatus implements the operator.Syncer interface.
func (c *Operator) UpdateStatus(ctx context.Context, key string) error {
	pobj, err := c.promInfs.Get(key)

	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	p := pobj.(*monitoringv1alpha1.PrometheusAgent)
	p = p.DeepCopy()

	pStatus := monitoringv1.PrometheusStatus{
		Paused: p.Spec.Paused,
	}

	logger := log.With(c.logger, "key", key)
	level.Info(logger).Log("msg", "update prometheus status")

	var (
		availableCondition = monitoringv1.Condition{
			Type:   monitoringv1.Available,
			Status: monitoringv1.ConditionTrue,
			LastTransitionTime: metav1.Time{
				Time: time.Now().UTC(),
			},
			ObservedGeneration: p.Generation,
		}
		messages []string
		replicas = 1
	)

	if p.Spec.Replicas != nil {
		replicas = int(*p.Spec.Replicas)
	}

	for shard := range prompkg.ExpectedStatefulSetShardNames(p) {
		ssetName := prompkg.KeyToStatefulSetKey(p, key, shard)
		logger := log.With(logger, "statefulset", ssetName, "shard", shard)

		obj, err := c.ssetInfs.Get(ssetName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Object not yet in the store or already deleted.
				level.Info(logger).Log("msg", "not found")
				continue
			}
			return errors.Wrap(err, "failed to retrieve statefulset")
		}

		sset := obj.(*appsv1.StatefulSet)
		if c.rr.DeletionInProgress(sset) {
			continue
		}

		stsReporter, err := operator.NewStatefulSetReporter(ctx, c.kclient, sset)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve statefulset state")
		}

		pStatus.Replicas += int32(len(stsReporter.Pods))
		pStatus.UpdatedReplicas += int32(len(stsReporter.UpdatedPods()))
		pStatus.AvailableReplicas += int32(len(stsReporter.ReadyPods()))
		pStatus.UnavailableReplicas += int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods()))

		pStatus.ShardStatuses = append(
			pStatus.ShardStatuses,
			monitoringv1.ShardStatus{
				ShardID:             strconv.Itoa(shard),
				Replicas:            int32(len(stsReporter.Pods)),
				UpdatedReplicas:     int32(len(stsReporter.UpdatedPods())),
				AvailableReplicas:   int32(len(stsReporter.ReadyPods())),
				UnavailableReplicas: int32(len(stsReporter.Pods) - len(stsReporter.ReadyPods())),
			},
		)

		if len(stsReporter.ReadyPods()) >= replicas {
			// All pods are ready (or the desired number of replicas is zero).
			continue
		}

		if len(stsReporter.ReadyPods()) == 0 {
			availableCondition.Reason = "NoPodReady"
			availableCondition.Status = monitoringv1.ConditionFalse
		} else if availableCondition.Status != monitoringv1.ConditionFalse {
			availableCondition.Reason = "SomePodsNotReady"
			availableCondition.Status = monitoringv1.ConditionDegraded
		}

		for _, p := range stsReporter.Pods {
			if m := p.Message(); m != "" {
				messages = append(messages, fmt.Sprintf("shard %d: pod %s: %s", shard, p.Name, m))
			}
		}
	}

	availableCondition.Message = strings.Join(messages, "\n")

	// Compute the Reconciled ConditionType.
	reconciledCondition := monitoringv1.Condition{
		Type:   monitoringv1.Reconciled,
		Status: monitoringv1.ConditionTrue,
		LastTransitionTime: metav1.Time{
			Time: time.Now().UTC(),
		},
		ObservedGeneration: p.Generation,
	}
	reconciliationStatus, found := c.reconciliations.GetStatus(key)
	if !found {
		reconciledCondition.Status = monitoringv1.ConditionUnknown
		reconciledCondition.Reason = "NotFound"
		reconciledCondition.Message = fmt.Sprintf("object %q not found", key)
	} else {
		if !reconciliationStatus.Ok() {
			reconciledCondition.Status = monitoringv1.ConditionFalse
		}
		reconciledCondition.Reason = reconciliationStatus.Reason()
		reconciledCondition.Message = reconciliationStatus.Message()
	}

	// Update the last transition times only if the status of the available condition has changed.
	for _, condition := range p.Status.Conditions {
		if condition.Type == availableCondition.Type && condition.Status == availableCondition.Status {
			availableCondition.LastTransitionTime = condition.LastTransitionTime
			continue
		}

		if condition.Type == reconciledCondition.Type && condition.Status == reconciledCondition.Status {
			reconciledCondition.LastTransitionTime = condition.LastTransitionTime
		}
	}

	pStatus.Conditions = append(pStatus.Conditions, availableCondition, reconciledCondition)

	p.Status = pStatus
	if _, err = c.mclient.MonitoringV1alpha1().PrometheusAgents(p.Namespace).UpdateStatus(ctx, p, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(err, "failed to update status subresource")
	}

	return nil
}

func (c *Operator) createOrUpdateTLSAssetSecrets(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent, store *assets.Store) (*operator.ShardedSecret, error) {
	labels := c.config.Labels.Merge(prompkg.ManagedByOperatorLabels)
	template := prompkg.NewTLSAssetSecret(p, labels)

	sSecret := operator.NewShardedSecret(template, prompkg.TLSAssetsSecretName(p))

	for k, v := range store.TLSAssets {
		sSecret.AppendData(k.String(), []byte(v))
	}

	sClient := c.kclient.CoreV1().Secrets(p.Namespace)

	if err := sSecret.StoreSecrets(ctx, sClient); err != nil {
		return nil, errors.Wrapf(err, "failed to create TLS assets secret for Prometheus")
	}

	level.Debug(c.logger).Log("msg", "tls-asset secret: stored")

	return sSecret, nil
}

func (c *Operator) createOrUpdateWebConfigSecret(ctx context.Context, p *monitoringv1alpha1.PrometheusAgent) error {
	boolTrue := true

	var fields monitoringv1.WebConfigFileFields
	if p.Spec.Web != nil {
		fields = p.Spec.Web.WebConfigFileFields
	}

	webConfig, err := webconfig.New(
		prompkg.WebConfigDir,
		prompkg.WebConfigSecretName(p),
		fields,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize web config")
	}

	secretClient := c.kclient.CoreV1().Secrets(p.Namespace)
	ownerReference := metav1.OwnerReference{
		APIVersion:         p.APIVersion,
		BlockOwnerDeletion: &boolTrue,
		Controller:         &boolTrue,
		Kind:               p.Kind,
		Name:               p.Name,
		UID:                p.UID,
	}
	secretLabels := c.config.Labels.Merge(prompkg.ManagedByOperatorLabels)

	if err := webConfig.CreateOrUpdateWebConfigSecret(ctx, secretClient, secretLabels, ownerReference); err != nil {
		return errors.Wrap(err, "failed to reconcile web config secret")
	}

	return nil
}

func (c *Operator) loadConfigFromSecret(sks *v1.SecretKeySelector, s *v1.SecretList) ([]byte, error) {
	if sks == nil {
		return nil, nil
	}

	for _, secret := range s.Items {
		if secret.Name == sks.Name {
			if c, ok := secret.Data[sks.Key]; ok {
				return c, nil
			}

			return nil, fmt.Errorf("key %v could not be found in secret %v", sks.Key, sks.Name)
		}
	}

	if sks.Optional == nil || !*sks.Optional {
		return nil, fmt.Errorf("secret %v could not be found", sks.Name)
	}

	level.Debug(c.logger).Log("msg", fmt.Sprintf("secret %v could not be found", sks.Name))
	return nil, nil
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.AddEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.ServiceMonitor).ResourceVersion == cur.(*monitoringv1.ServiceMonitor).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor updated")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.UpdateEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleSmonDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ServiceMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.ServiceMonitorsKind, operator.DeleteEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor added")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.AddEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.PodMonitor).ResourceVersion == cur.(*monitoringv1.PodMonitor).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor updated")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.UpdateEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handlePmonDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "PodMonitor delete")
		c.metrics.TriggerByCounter(monitoringv1.PodMonitorsKind, operator.DeleteEvent).Inc()

		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonAdd(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe added")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.AddEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonUpdate(old, cur interface{}) {
	if old.(*monitoringv1.Probe).ResourceVersion == cur.(*monitoringv1.Probe).ResourceVersion {
		return
	}

	if o, ok := c.accessor.ObjectMetadata(cur); ok {
		level.Debug(c.logger).Log("msg", "Probe updated")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.UpdateEvent)
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Don't enqueue just for the namespace
func (c *Operator) handleBmonDelete(obj interface{}) {
	if o, ok := c.accessor.ObjectMetadata(obj); ok {
		level.Debug(c.logger).Log("msg", "Probe delete")
		c.metrics.TriggerByCounter(monitoringv1.ProbesKind, operator.DeleteEvent).Inc()
		c.enqueueForMonitorNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enqueue configmaps just for the namespace or in general?
func (c *Operator) handleConfigMapAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap added")
		c.metrics.TriggerByCounter("ConfigMap", operator.AddEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap deleted")
		c.metrics.TriggerByCounter("ConfigMap", operator.DeleteEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleConfigMapUpdate(old, cur interface{}) {
	if old.(*v1.ConfigMap).ResourceVersion == cur.(*v1.ConfigMap).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "ConfigMap updated")
		c.metrics.TriggerByCounter("ConfigMap", operator.UpdateEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

// TODO: Do we need to enqueue secrets just for the namespace or in general?
func (c *Operator) handleSecretDelete(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret deleted")
		c.metrics.TriggerByCounter("Secret", operator.DeleteEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretUpdate(old, cur interface{}) {
	if old.(*v1.Secret).ResourceVersion == cur.(*v1.Secret).ResourceVersion {
		return
	}

	o, ok := c.accessor.ObjectMetadata(cur)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret updated")
		c.metrics.TriggerByCounter("Secret", operator.UpdateEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) handleSecretAdd(obj interface{}) {
	o, ok := c.accessor.ObjectMetadata(obj)
	if ok {
		level.Debug(c.logger).Log("msg", "Secret added")
		c.metrics.TriggerByCounter("Secret", operator.AddEvent).Inc()

		c.enqueueForPrometheusNamespace(o.GetNamespace())
	}
}

func (c *Operator) enqueueForPrometheusNamespace(nsName string) {
	c.enqueueForNamespace(c.nsPromInf.GetStore(), nsName)
}

func (c *Operator) enqueueForMonitorNamespace(nsName string) {
	c.enqueueForNamespace(c.nsMonInf.GetStore(), nsName)
}

// enqueueForNamespace enqueues all Prometheus object keys that belong to the
// given namespace or select objects in the given namespace.
func (c *Operator) enqueueForNamespace(store cache.Store, nsName string) {
	nsObject, exists, err := store.GetByKey(nsName)
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "get namespace to enqueue Prometheus instances failed",
			"err", err,
		)
		return
	}
	if !exists {
		level.Error(c.logger).Log(
			"msg", fmt.Sprintf("get namespace to enqueue Prometheus instances failed: namespace %q does not exist", nsName),
		)
		return
	}
	ns := nsObject.(*v1.Namespace)

	err = c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		// Check for Prometheus Agent instances in the namespace.
		p := obj.(*monitoringv1alpha1.PrometheusAgent)
		if p.Namespace == nsName {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting ServiceMonitors in
		// the namespace.
		smNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ServiceMonitorNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert ServiceMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if smNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting PodMonitors in the NS.
		pmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.PodMonitorNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert PodMonitorNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if pmNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}

		// Check for Prometheus instances selecting Probes in the NS.
		bmNSSelector, err := metav1.LabelSelectorAsSelector(p.Spec.ProbeNamespaceSelector)
		if err != nil {
			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("failed to convert ProbeNamespaceSelector of %q to selector", p.Name),
				"err", err,
			)
			return
		}

		if bmNSSelector.Matches(labels.Set(ns.Labels)) {
			c.rr.EnqueueForReconciliation(p)
			return
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Prometheus instances from cache failed",
			"err", err,
		)
	}

}

func (c *Operator) handleMonitorNamespaceUpdate(oldo, curo interface{}) {
	old := oldo.(*v1.Namespace)
	cur := curo.(*v1.Namespace)

	level.Debug(c.logger).Log("msg", "update handler", "namespace", cur.GetName(), "old", old.ResourceVersion, "cur", cur.ResourceVersion)

	// Periodic resync may resend the Namespace without changes
	// in-between.
	if old.ResourceVersion == cur.ResourceVersion {
		return
	}

	level.Debug(c.logger).Log("msg", "Monitor namespace updated", "namespace", cur.GetName())
	c.metrics.TriggerByCounter("Namespace", operator.UpdateEvent).Inc()

	// Check for Prometheus Agent instances selecting ServiceMonitors, PodMonitors,
	// and Probes in the namespace.
	err := c.promInfs.ListAll(labels.Everything(), func(obj interface{}) {
		p := obj.(*monitoringv1alpha1.PrometheusAgent)

		for name, selector := range map[string]*metav1.LabelSelector{
			"PodMonitors":     p.Spec.PodMonitorNamespaceSelector,
			"Probes":          p.Spec.ProbeNamespaceSelector,
			"ServiceMonitors": p.Spec.ServiceMonitorNamespaceSelector,
		} {

			sync, err := k8sutil.LabelSelectionHasChanged(old.Labels, cur.Labels, selector)
			if err != nil {
				level.Error(c.logger).Log(
					"err", err,
					"name", p.Name,
					"namespace", p.Namespace,
					"subresource", name,
				)
				return
			}

			if sync {
				c.rr.EnqueueForReconciliation(p)
				return
			}
		}
	})
	if err != nil {
		level.Error(c.logger).Log(
			"msg", "listing all Prometheus Agent instances from cache failed",
			"err", err,
		)
	}
}

// getMissingPermissions returns the RBAC permissions that the controller would need to be
// granted to fulfill its mission. An empty map means that everything is ok.
func (c *Operator) getMissingPermissions(ctx context.Context) (map[string][]string, error) {
	verbs := map[string][]string{
		monitoringv1alpha1.PrometheusAgentName:                           {"get", "list", "watch"},
		fmt.Sprintf("%s/status", monitoringv1alpha1.PrometheusAgentName): {"update"},
	}
	var ssar *authv1.SelfSubjectAccessReview
	var ssarResponse *authv1.SelfSubjectAccessReview
	var err error

	missingPermissions := map[string][]string{}

	for ns := range c.config.Namespaces.PrometheusAllowList {
		for resource, verbs := range verbs {
			for _, verb := range verbs {
				ssar = &authv1.SelfSubjectAccessReview{
					Spec: authv1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &authv1.ResourceAttributes{
							Verb:     verb,
							Group:    monitoringv1alpha1.SchemeGroupVersion.Group,
							Resource: resource,
							// If ns is empty string, it will check cluster-wide
							Namespace: ns,
						},
					},
				}
				ssarResponse, err = c.kclient.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
				if !ssarResponse.Status.Allowed {
					missingPermissions[resource] = append(missingPermissions[resource], verb)
				}
			}
		}
	}

	return missingPermissions, nil
}
