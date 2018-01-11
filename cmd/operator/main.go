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

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"k8s.io/api/core/v1"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/api"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	prometheuscontroller "github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/go-kit/kit/log"
)

var (
	cfg prometheuscontroller.Config
)

func init() {
	cfg.CrdKinds = monitoringv1.DefaultCrdKinds
	flagset := flag.CommandLine
	flagset.StringVar(&cfg.Host, "apiserver", "", "API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.")
	flagset.StringVar(&cfg.TLSConfig.CertFile, "cert-file", "", " - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.KeyFile, "key-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.CAFile, "ca-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.")
	flagset.StringVar(&cfg.KubeletObject, "kubelet-service", "", "Service/Endpoints object to write kubelets into in format \"namespace/name\"")
	flagset.BoolVar(&cfg.TLSInsecure, "tls-insecure", false, "- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.")
	flagset.StringVar(&cfg.PrometheusConfigReloader, "prometheus-config-reloader", "quay.io/coreos/prometheus-config-reloader:v0.0.2", "Config and rule reload image")
	flagset.StringVar(&cfg.ConfigReloaderImage, "config-reloader-image", "quay.io/coreos/configmap-reload:v0.0.1", "Reload Image")
	flagset.StringVar(&cfg.AlertmanagerDefaultBaseImage, "alertmanager-default-base-image", "quay.io/prometheus/alertmanager", "Alertmanager default base image")
	flagset.StringVar(&cfg.PrometheusDefaultBaseImage, "prometheus-default-base-image", "quay.io/prometheus/prometheus", "Prometheus default base image")
	flagset.StringVar(&cfg.Namespace, "namespace", v1.NamespaceAll, "Namespace to scope the interaction of the Prometheus Operator and the apiserver.")
	flagset.Var(&cfg.Labels, "labels", "Labels to be add to all resources created by the operator")
	flagset.StringVar(&cfg.CrdGroup, "crd-apigroup", monitoringv1.Group, "prometheus CRD  API group name")
	flagset.Var(&cfg.CrdKinds, "crd-kinds", " - EXPERIMENTAL (could be removed in future releases) - customize CRD kind names")
	flagset.Parse(os.Args[1:])

}

func Main() int {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewGoCollector())

	po, err := prometheuscontroller.New(cfg, log.With(logger, "component", "prometheusoperator"))
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating prometheus controller failed: ", err)
		return 1
	}

	ao, err := alertmanager.New(cfg, log.With(logger, "component", "alertmanageroperator"))
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating alertmanager controller failed: ", err)
		return 1
	}

	mux := http.NewServeMux()
	web, err := api.New(cfg, log.With(logger, "component", "api"))
	if err != nil {
		fmt.Fprint(os.Stderr, "instantiating api failed: ", err)
		return 1
	}

	web.Register(mux)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Fprint(os.Stderr, "listening port 8080 failed", err)
		return 1
	}

	po.RegisterMetrics(r)
	ao.RegisterMetrics(r)
	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error { return po.Run(ctx.Done()) })
	wg.Go(func() error { return ao.Run(ctx.Done()) })

	srv := &http.Server{Handler: mux}
	go srv.Serve(l)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		logger.Log("msg", "Received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	cancel()
	if err := wg.Wait(); err != nil {
		logger.Log("msg", "Unhandled error received. Exiting...", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(Main())
}
