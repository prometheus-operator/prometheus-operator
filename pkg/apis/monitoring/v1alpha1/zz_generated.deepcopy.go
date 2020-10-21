// +build !ignore_autogenerated

// Copyright The prometheus-operator Authors
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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/api/core/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertmanagerConfig) DeepCopyInto(out *AlertmanagerConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertmanagerConfig.
func (in *AlertmanagerConfig) DeepCopy() *AlertmanagerConfig {
	if in == nil {
		return nil
	}
	out := new(AlertmanagerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertmanagerConfigList) DeepCopyInto(out *AlertmanagerConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*AlertmanagerConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(AlertmanagerConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertmanagerConfigList.
func (in *AlertmanagerConfigList) DeepCopy() *AlertmanagerConfigList {
	if in == nil {
		return nil
	}
	out := new(AlertmanagerConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertmanagerConfigSpec) DeepCopyInto(out *AlertmanagerConfigSpec) {
	*out = *in
	if in.Route != nil {
		in, out := &in.Route, &out.Route
		*out = new(Route)
		(*in).DeepCopyInto(*out)
	}
	if in.Receivers != nil {
		in, out := &in.Receivers, &out.Receivers
		*out = make([]Receiver, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InhibitRules != nil {
		in, out := &in.InhibitRules, &out.InhibitRules
		*out = make([]InhibitRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertmanagerConfigSpec.
func (in *AlertmanagerConfigSpec) DeepCopy() *AlertmanagerConfigSpec {
	if in == nil {
		return nil
	}
	out := new(AlertmanagerConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPConfig) DeepCopyInto(out *HTTPConfig) {
	*out = *in
	if in.BasicAuth != nil {
		in, out := &in.BasicAuth, &out.BasicAuth
		*out = new(monitoringv1.BasicAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.BearerTokenSecret != nil {
		in, out := &in.BearerTokenSecret, &out.BearerTokenSecret
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.TLSConfig != nil {
		in, out := &in.TLSConfig, &out.TLSConfig
		*out = new(monitoringv1.SafeTLSConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ProxyURL != nil {
		in, out := &in.ProxyURL, &out.ProxyURL
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPConfig.
func (in *HTTPConfig) DeepCopy() *HTTPConfig {
	if in == nil {
		return nil
	}
	out := new(HTTPConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InhibitRule) DeepCopyInto(out *InhibitRule) {
	*out = *in
	if in.TargetMatch != nil {
		in, out := &in.TargetMatch, &out.TargetMatch
		*out = make([]Matcher, len(*in))
		copy(*out, *in)
	}
	if in.SourceMatch != nil {
		in, out := &in.SourceMatch, &out.SourceMatch
		*out = make([]Matcher, len(*in))
		copy(*out, *in)
	}
	if in.Equal != nil {
		in, out := &in.Equal, &out.Equal
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InhibitRule.
func (in *InhibitRule) DeepCopy() *InhibitRule {
	if in == nil {
		return nil
	}
	out := new(InhibitRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Matcher) DeepCopyInto(out *Matcher) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Matcher.
func (in *Matcher) DeepCopy() *Matcher {
	if in == nil {
		return nil
	}
	out := new(Matcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerDutyConfig) DeepCopyInto(out *PagerDutyConfig) {
	*out = *in
	if in.SendResolved != nil {
		in, out := &in.SendResolved, &out.SendResolved
		*out = new(bool)
		**out = **in
	}
	if in.RoutingKey != nil {
		in, out := &in.RoutingKey, &out.RoutingKey
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.ServiceKey != nil {
		in, out := &in.ServiceKey, &out.ServiceKey
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.URL != nil {
		in, out := &in.URL, &out.URL
		*out = new(string)
		**out = **in
	}
	if in.Client != nil {
		in, out := &in.Client, &out.Client
		*out = new(string)
		**out = **in
	}
	if in.ClientURL != nil {
		in, out := &in.ClientURL, &out.ClientURL
		*out = new(string)
		**out = **in
	}
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.Severity != nil {
		in, out := &in.Severity, &out.Severity
		*out = new(string)
		**out = **in
	}
	if in.Class != nil {
		in, out := &in.Class, &out.Class
		*out = new(string)
		**out = **in
	}
	if in.Group != nil {
		in, out := &in.Group, &out.Group
		*out = new(string)
		**out = **in
	}
	if in.Component != nil {
		in, out := &in.Component, &out.Component
		*out = new(string)
		**out = **in
	}
	if in.Details != nil {
		in, out := &in.Details, &out.Details
		*out = make([]PagerDutyConfigDetail, len(*in))
		copy(*out, *in)
	}
	if in.HTTPConfig != nil {
		in, out := &in.HTTPConfig, &out.HTTPConfig
		*out = new(HTTPConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerDutyConfig.
func (in *PagerDutyConfig) DeepCopy() *PagerDutyConfig {
	if in == nil {
		return nil
	}
	out := new(PagerDutyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerDutyConfigDetail) DeepCopyInto(out *PagerDutyConfigDetail) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerDutyConfigDetail.
func (in *PagerDutyConfigDetail) DeepCopy() *PagerDutyConfigDetail {
	if in == nil {
		return nil
	}
	out := new(PagerDutyConfigDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Receiver) DeepCopyInto(out *Receiver) {
	*out = *in
	if in.PagerDutyConfigs != nil {
		in, out := &in.PagerDutyConfigs, &out.PagerDutyConfigs
		*out = make([]PagerDutyConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.WebhookConfigs != nil {
		in, out := &in.WebhookConfigs, &out.WebhookConfigs
		*out = make([]WebhookConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Receiver.
func (in *Receiver) DeepCopy() *Receiver {
	if in == nil {
		return nil
	}
	out := new(Receiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
	if in.GroupBy != nil {
		in, out := &in.GroupBy, &out.GroupBy
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Matchers != nil {
		in, out := &in.Matchers, &out.Matchers
		*out = make([]Matcher, len(*in))
		copy(*out, *in)
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebhookConfig) DeepCopyInto(out *WebhookConfig) {
	*out = *in
	if in.SendResolved != nil {
		in, out := &in.SendResolved, &out.SendResolved
		*out = new(bool)
		**out = **in
	}
	if in.URL != nil {
		in, out := &in.URL, &out.URL
		*out = new(string)
		**out = **in
	}
	if in.URLSecret != nil {
		in, out := &in.URLSecret, &out.URLSecret
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTPConfig != nil {
		in, out := &in.HTTPConfig, &out.HTTPConfig
		*out = new(HTTPConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.MaxAlerts != nil {
		in, out := &in.MaxAlerts, &out.MaxAlerts
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebhookConfig.
func (in *WebhookConfig) DeepCopy() *WebhookConfig {
	if in == nil {
		return nil
	}
	out := new(WebhookConfig)
	in.DeepCopyInto(out)
	return out
}
