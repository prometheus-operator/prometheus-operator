// Copyright 2016 The prometheus-operator Authors
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

package listwatch

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"

	sortutil "github.com/prometheus-operator/prometheus-operator/internal/sortutil"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8s"
)

const (
	pollInterval = 15 * time.Second
)

// NewNamespaceListWatchFromClient mimics
// cache.NewListWatchFromClient. It allows for the creation of a
// cache.ListWatch for namespaces from a client that does not have `List`
// privileges. If the slice of namespaces contains only v1.NamespaceAll, then
// this func assumes that the client has List and Watch privileges and returns
// a regular cache.ListWatch, since there is no other way to get all
// namespaces.
//
// The allowed namespaces and denied namespaces are mutually exclusive.
//
// If allowed namespaces contain multiple items, the given denied namespaces have no effect.
//
// If the allowed namespaces includes exactly one entry with the value v1.NamespaceAll (empty string),
// the given denied namespaces are applied.
func NewNamespaceListWatchFromClient(
	ctx context.Context,
	l *slog.Logger,
	k8sVersion semver.Version,
	corev1Client corev1.CoreV1Interface,
	ssarClient authv1.SelfSubjectAccessReviewInterface,
	allowedNamespaces, deniedNamespaces map[string]struct{},
) (cache.ListerWatcher, bool, error) {
	if l == nil {
		l = slog.New(slog.DiscardHandler)
	}

	listWatchAllowed, reasons, err := k8s.IsAllowed(
		ctx,
		ssarClient,
		nil, // namespaces is a cluster-scoped resource.
		k8s.ResourceAttribute{
			Resource: "namespaces",
			Verbs:    []string{"list", "watch"},
		},
	)
	if err != nil {
		return nil, false, err
	}

	// The "kubernetes.io/metadata.name" label is GA since Kubernetes 1.22.
	metadataNameLabelSupported := k8sVersion.GTE(semver.MustParse("1.22.0"))

	if IsAllNamespaces(allowedNamespaces) {
		if !listWatchAllowed {
			err := fmt.Errorf("missing list/watch permissions on the 'namespaces' resource")
			for _, r := range reasons {
				err = fmt.Errorf("%w: %w", err, r)
			}

			return nil, false, err
		}

		// deniedNamespaces is only supported with allowedNamespaces = "all".
		tweak := func(options *metav1.ListOptions) {
			DenyTweak(options, "metadata.name", deniedNamespaces)
		}
		if metadataNameLabelSupported {
			// Using a label selector is more efficient but requires Kubernetes 1.22 at least.
			tweak = func(options *metav1.ListOptions) {
				TweakByLabel(options, "kubernetes.io/metadata.name", ExcludeFilterType, deniedNamespaces)
			}
		}

		return cache.NewFilteredListWatchFromClient(
			corev1Client.RESTClient(),
			"namespaces",
			metav1.NamespaceAll,
			tweak,
		), true, nil
	}

	if listWatchAllowed && metadataNameLabelSupported {
		l.Debug("using privileged namespace lister/watcher")
		return cache.NewFilteredListWatchFromClient(
			corev1Client.RESTClient(),
			"namespaces",
			metav1.NamespaceAll,
			func(options *metav1.ListOptions) {
				TweakByLabel(options, "kubernetes.io/metadata.name", IncludeFilterType, allowedNamespaces)
			},
		), true, nil
	}

	// At this point, the operator has no list/watch permissions on the
	// namespaces resource. Check if it has at least the get permission to
	// emulate the list/watch operations.
	attrs := make([]k8s.ResourceAttribute, 0, len(allowedNamespaces))
	for n := range allowedNamespaces {
		attrs = append(attrs, k8s.ResourceAttribute{
			Verbs:    []string{"get"},
			Resource: "namespaces",
			Name:     n,
		})
	}

	getAllowed, reasons, err := k8s.IsAllowed(
		ctx,
		ssarClient,
		nil, // namespaces is a cluster-scoped resource.
		attrs...,
	)
	if err != nil {
		return nil, false, err
	}

	// Only log a warning to preserve backward compatibility.
	if !getAllowed {
		err := fmt.Errorf("missing permissions")
		for _, r := range reasons {
			err = fmt.Errorf("%w: %w", err, r)
		}

		l.Warn("the operator lacks required permissions which may result in degraded functionalities", "err", err)
	}

	var namespaces []string
	for ns := range allowedNamespaces {
		namespaces = append(namespaces, ns)
	}

	return newPollBasedListerWatcher(ctx, l, corev1Client, namespaces), false, nil
}

// IsAllNamespaces checks if the given map of namespaces
// contains only v1.NamespaceAll.
func IsAllNamespaces(namespaces map[string]struct{}) bool {
	_, ok := namespaces[v1.NamespaceAll]
	return ok && len(namespaces) == 1
}

// IdenticalNamespaces returns true if a and b are identical.
func IdenticalNamespaces(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}

	return true
}

type FilterType string

const (
	IncludeFilterType = "Include"
	ExcludeFilterType = "Exclude"
)

// TweakByLabel modifies the given list options by adding a label selector
// matching/excluding the given values.
func TweakByLabel(options *metav1.ListOptions, label string, filter FilterType, valueSet map[string]struct{}) {
	if len(valueSet) == 0 {
		return
	}

	var op string
	switch filter {
	case IncludeFilterType:
		op = "in"
	case ExcludeFilterType:
		op = "notin"
	default:
		panic(fmt.Sprintf("unsupported filter: %q", filter))
	}
	selectors := []string{fmt.Sprintf("%s %s (%s)", label, op, strings.Join(sortutil.SortedKeys(valueSet), ","))}

	if options.LabelSelector != "" {
		selectors = append(selectors, options.LabelSelector)
	}

	options.LabelSelector = strings.Join(selectors, ",")
}

// DenyTweak modifies the given list options by adding a field selector *not*
// matching the given values.
func DenyTweak(options *metav1.ListOptions, field string, valueSet map[string]struct{}) {
	if len(valueSet) == 0 {
		return
	}

	var selectors []string
	for value := range valueSet {
		selectors = append(selectors, field+"!="+value)
	}
	slices.Sort(selectors)

	if options.FieldSelector != "" {
		selectors = append(selectors, options.FieldSelector)
	}

	options.FieldSelector = strings.Join(selectors, ",")
}

type pollBasedListerWatcher struct {
	corev1Client corev1.CoreV1Interface
	ch           chan watch.Event

	ctx context.Context
	l   *slog.Logger

	cache map[string]cacheEntry
}

type cacheEntry struct {
	present bool
	ns      *v1.Namespace
}

var _ = watch.Interface(&pollBasedListerWatcher{})
var _ = cache.ListerWatcher(&pollBasedListerWatcher{})

func newPollBasedListerWatcher(ctx context.Context, l *slog.Logger, corev1Client corev1.CoreV1Interface, namespaces []string) *pollBasedListerWatcher {
	if l == nil {
		l = slog.New(slog.DiscardHandler)
	}

	pblw := &pollBasedListerWatcher{
		corev1Client: corev1Client,
		ch:           make(chan watch.Event, 1),
		ctx:          ctx,
		l:            l,
		cache:        make(map[string]cacheEntry, len(namespaces)),
	}

	for _, ns := range namespaces {
		pblw.cache[ns] = cacheEntry{}
	}

	return pblw
}

func (pblw *pollBasedListerWatcher) List(_ metav1.ListOptions) (runtime.Object, error) {
	list := &v1.NamespaceList{}

	for ns := range pblw.cache {
		result, err := pblw.corev1Client.Namespaces().Get(pblw.ctx, ns, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				pblw.l.Info("namespace not found", "namespace", ns)
				continue
			}

			return nil, fmt.Errorf("unexpected error while listing namespaces: %w", err)
		}

		pblw.cache[ns] = cacheEntry{
			present: true,
			ns:      result,
		}
		list.Items = append(list.Items, *result)
	}

	return list, nil
}

func (pblw *pollBasedListerWatcher) Watch(_ metav1.ListOptions) (watch.Interface, error) {
	return pblw, nil
}

func (pblw *pollBasedListerWatcher) Stop() {}

func (pblw *pollBasedListerWatcher) ResultChan() <-chan watch.Event {
	go func() {
		jitter, err := rand.Int(rand.Reader, big.NewInt(int64(pollInterval)))
		if err == nil {
			time.Sleep(time.Duration(jitter.Int64()))
		} else {
			pblw.l.Info("failed to generate random jitter", "err", err)
		}

		_ = wait.PollUntilContextCancel(pblw.ctx, pollInterval, false, pblw.poll)
	}()

	return pblw.ch
}

func (pblw *pollBasedListerWatcher) poll(ctx context.Context) (bool, error) {
	var (
		updated []*v1.Namespace
		deleted []string
	)

	for ns, entry := range pblw.cache {
		result, err := pblw.corev1Client.Namespaces().Get(ctx, ns, metav1.GetOptions{ResourceVersion: entry.ns.ResourceVersion})
		if err != nil {
			switch {
			case apierrors.IsNotFound(err):
				if entry.present {
					deleted = append(deleted, ns)
				}
			default:
				pblw.l.Warn("watch error", "err", err, "namespace", ns)
			}
			continue
		}

		if entry.ns.ResourceVersion != result.ResourceVersion {
			updated = append(updated, result)
		}
	}

	for _, ns := range deleted {
		entry := pblw.cache[ns]

		pblw.ch <- watch.Event{
			Type:   watch.Deleted,
			Object: entry.ns,
		}

		pblw.cache[ns] = cacheEntry{
			present: false,
		}
	}

	for _, ns := range updated {
		var (
			eventType = watch.Modified
			entry     = pblw.cache[ns.Name]
		)

		switch {
		case !entry.present:
			eventType = watch.Added
		case ns.ResourceVersion == entry.ns.ResourceVersion:
			continue
		}

		pblw.ch <- watch.Event{
			Type:   eventType,
			Object: ns,
		}

		pblw.cache[ns.Name] = cacheEntry{
			ns:      ns,
			present: true,
		}
	}

	return false, nil
}
