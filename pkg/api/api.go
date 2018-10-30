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

package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-kit/kit/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/coreos/prometheus-operator/pkg/k8sutil"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
)

type API struct {
	kclient *kubernetes.Clientset
	mclient *v1.MonitoringV1Client
	logger  log.Logger
}

func New(conf prometheus.Config, l log.Logger) (*API, error) {
	cfg, err := k8sutil.NewClusterConfig(conf.Host, conf.TLSInsecure, &conf.TLSConfig)
	if err != nil {
		return nil, err
	}

	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	mclient, err := v1.NewForConfig(&conf.CrdKinds, conf.CrdGroup, cfg)
	if err != nil {
		return nil, err
	}

	return &API{
		kclient: kclient,
		mclient: mclient,
		logger:  l,
	}, nil
}

var (
	prometheusRoute   = regexp.MustCompile("/apis/monitoring.coreos.com/" + v1.Version + "/namespaces/(.*)/prometheuses/(.*)/status")
	alertmanagerRoute = regexp.MustCompile("/apis/monitoring.coreos.com/" + v1.Version + "/namespaces/(.*)/alertmanagers/(.*)/status")
)

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch {
		case prometheusRoute.MatchString(req.URL.Path):
			api.prometheusStatus(w, req)
		case alertmanagerRoute.MatchString(req.URL.Path):
			api.alertmanagerStatus(w, req)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

type objectReference struct {
	name      string
	namespace string
}

func parseURL(route *regexp.Regexp, path string) objectReference {
	matches := route.FindAllStringSubmatch(path, -1)
	ns := ""
	name := ""
	if len(matches) == 1 {
		if len(matches[0]) == 3 {
			ns = matches[0][1]
			name = matches[0][2]
		}
	}

	return objectReference{
		name:      name,
		namespace: ns,
	}
}

func (api *API) prometheusStatus(w http.ResponseWriter, req *http.Request) {
	or := parseURL(prometheusRoute, req.URL.Path)

	p, err := api.mclient.Prometheuses(or.namespace).Get(or.name, metav1.GetOptions{})
	if err != nil {
		api.logger.Log("error", err)
		if k8sutil.IsResourceNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p.Status, _, err = prometheus.PrometheusStatus(api.kclient, p)
	if err != nil {
		api.logger.Log("error", err)
	}

	b, err := json.Marshal(p)
	if err != nil {
		api.logger.Log("error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusFound)
	w.Write(b)
}

func (api *API) alertmanagerStatus(w http.ResponseWriter, req *http.Request) {
	or := parseURL(alertmanagerRoute, req.URL.Path)

	am, err := api.mclient.Alertmanagers(or.namespace).Get(or.name, metav1.GetOptions{})
	if err != nil {
		api.logger.Log("error", err)
		if k8sutil.IsResourceNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	am.Status, _, err = alertmanager.AlertmanagerStatus(api.kclient, am)
	if err != nil {
		api.logger.Log("error", err)
	}

	b, err := json.Marshal(am)
	if err != nil {
		api.logger.Log("error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusFound)
	w.Write(b)
}
