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
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIdenticalNamespaces(t *testing.T) {
	for _, tc := range []struct {
		a, b map[string]struct{}
		ret  bool
	}{
		{
			a: map[string]struct{}{
				"foo": {},
			},
			b: map[string]struct{}{
				"foo": {},
			},
			ret: true,
		},
		{
			a: map[string]struct{}{
				"foo": {},
			},
			b: map[string]struct{}{
				"bar": {},
			},
			ret: false,
		},
	} {
		tc := tc
		t.Run("", func(t *testing.T) {
			ret := IdenticalNamespaces(tc.a, tc.b)
			if ret != tc.ret {
				t.Fatalf("expecting IdenticalNamespaces() to return %v, got %v", tc.ret, ret)
			}
		})
	}
}

func TestDenyTweak(t *testing.T) {
	for _, tc := range []struct {
		options  metav1.ListOptions
		field    string
		valueSet map[string]struct{}

		exp string
	}{
		{
			field:    "metadata.name",
			valueSet: map[string]struct{}{},
			exp:      "",
		},
		{
			options: metav1.ListOptions{
				FieldSelector: "metadata.namespace=foo",
			},
			field:    "metadata.name",
			valueSet: map[string]struct{}{},
			exp:      "metadata.namespace=foo",
		},
		{
			field: "metadata.name",
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "metadata.name!=bar,metadata.name!=foo",
		},
		{
			options: metav1.ListOptions{
				FieldSelector: "metadata.namespace=foo",
			},
			field: "metadata.name",
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "metadata.name!=bar,metadata.name!=foo,metadata.namespace=foo",
		},
	} {
		t.Run("", func(t *testing.T) {
			options := tc.options
			DenyTweak(&options, tc.field, tc.valueSet)
			require.Equal(t, tc.exp, options.FieldSelector)
		})
	}
}

func TestOnlyTweak(t *testing.T) {
	for _, tc := range []struct {
		options  metav1.ListOptions
		label    string
		filter   FilterType
		valueSet map[string]struct{}

		exp string
	}{
		{
			label:    "kubernetes.io/metadata.name",
			filter:   IncludeFilterType,
			valueSet: map[string]struct{}{},
			exp:      "",
		},
		{
			options: metav1.ListOptions{
				LabelSelector: "foo=bar",
			},
			label:    "kubernetes.io/metadata.name",
			filter:   IncludeFilterType,
			valueSet: map[string]struct{}{},
			exp:      "foo=bar",
		},
		{
			label:  "kubernetes.io/metadata.name",
			filter: IncludeFilterType,
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "kubernetes.io/metadata.name in (bar,foo)",
		},
		{
			options: metav1.ListOptions{
				LabelSelector: "foo=bar",
			},
			label:  "kubernetes.io/metadata.name",
			filter: IncludeFilterType,
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "kubernetes.io/metadata.name in (bar,foo),foo=bar",
		},
		{
			label:    "kubernetes.io/metadata.name",
			filter:   ExcludeFilterType,
			valueSet: map[string]struct{}{},
			exp:      "",
		},
		{
			options: metav1.ListOptions{
				LabelSelector: "foo=bar",
			},
			label:    "kubernetes.io/metadata.name",
			filter:   ExcludeFilterType,
			valueSet: map[string]struct{}{},
			exp:      "foo=bar",
		},
		{
			label:  "kubernetes.io/metadata.name",
			filter: ExcludeFilterType,
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "kubernetes.io/metadata.name notin (bar,foo)",
		},
		{
			options: metav1.ListOptions{
				LabelSelector: "foo=bar",
			},
			label:  "kubernetes.io/metadata.name",
			filter: ExcludeFilterType,
			valueSet: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
			exp: "kubernetes.io/metadata.name notin (bar,foo),foo=bar",
		},
	} {
		t.Run("", func(t *testing.T) {
			options := tc.options
			TweakByLabel(&options, tc.label, tc.filter, tc.valueSet)
			require.Equal(t, tc.exp, options.LabelSelector)
		})
	}
}
