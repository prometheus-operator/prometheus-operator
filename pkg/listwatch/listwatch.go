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
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apimachinery/pkg/watch"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
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
	l log.Logger,
	k8sVersion version.Info,
	corev1Client corev1.CoreV1Interface,
	ssarClient authv1.SelfSubjectAccessReviewInterface,
	allowedNamespaces, deniedNamespaces map[string]struct{},
) (cache.ListerWatcher, bool, error) {
	if l == nil {
		l = log.NewNopLogger()
	}

	listWatchAllowed, reasons, err := k8sutil.IsAllowed(
		ctx,
		ssarClient,
		nil, // namespaces is a cluster-scoped resource.
		k8sutil.ResourceAttribute{
			Resource: "namespaces",
			Verbs:    []string{"list", "watch"},
		},
	)
	if err != nil {
		return nil, false, err
	}

	// The "kubernetes.io/metadata.name" label is GA since Kubernetes 1.22.
	var metadataNameLabelSupported bool
	v, err := semver.ParseTolerant(k8sVersion.String())
	if err != nil {
		level.Warn(l).Log("msg", "failed to parse Kubernetes version", "version", k8sVersion.String(), "err", err)
	} else {
		metadataNameLabelSupported = v.GTE(semver.MustParse("1.22.0"))
	}

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

		level.Debug(l).Log("msg", "using privileged namespace lister/watcher")
		return cache.NewFilteredListWatchFromClient(
			corev1Client.RESTClient(),
			"namespaces",
			metav1.NamespaceAll,
			tweak,
		), true, nil
	}

	if listWatchAllowed && metadataNameLabelSupported {
		level.Debug(l).Log("msg", "using privileged namespace lister/watcher")
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
	attrs := make([]k8sutil.ResourceAttribute, 0, len(allowedNamespaces))
	for n := range allowedNamespaces {
		attrs = append(attrs, k8sutil.ResourceAttribute{
			Verbs:    []string{"get"},
			Resource: "namespaces",
			Name:     n,
		})
	}

	getAllowed, reasons, err := k8sutil.IsAllowed(
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

		level.Warn(l).Log("msg", "the operator lacks required permissions which may result in degraded functionalities", "err", err)
	}

	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		list := &v1.NamespaceList{}
		for name := range allowedNamespaces {
			result, err := corev1Client.Namespaces().Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					level.Info(l).Log("msg", "namespace not found", "namespace", name)
					continue
				}

				return nil, fmt.Errorf("unexpected error while listing namespaces: %w", err)
			}

			list.Items = append(list.Items, *result)
		}
		return list, nil
	}

	// Since the client does not have Watch privileges, do not
	// actually watch anything. Use a watch.FakeWatcher here to
	// implement watch.Interface but not send any events.
	watchFunc := func(_ metav1.ListOptions) (watch.Interface, error) {
		// TODO(simonpasquier): implement a poll-based watcher that gets the
		// list of namespaces perdiocially and send watch.Event() whenever it
		// detects a change.
		return watch.NewFake(), nil
	}

	level.Debug(l).Log("msg", "using unprivileged namespace lister/watcher")
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}, false, nil
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

	var labels []string
	for value := range valueSet {
		labels = append(labels, value)
	}
	sort.Strings(labels)

	var op string
	switch filter {
	case IncludeFilterType:
		op = "in"
	case ExcludeFilterType:
		op = "notin"
	default:
		panic(fmt.Sprintf("unsupported filter: %q", filter))
	}
	selectors := []string{fmt.Sprintf("%s %s (%s)", label, op, strings.Join(labels, ","))}

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
	sort.Strings(selectors)

	if options.FieldSelector != "" {
		selectors = append(selectors, options.FieldSelector)
	}

	options.FieldSelector = strings.Join(selectors, ",")
}
