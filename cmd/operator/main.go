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
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/coreos/prometheus-operator/pkg/alertmanager"
	"github.com/coreos/prometheus-operator/pkg/analytics"
	"github.com/coreos/prometheus-operator/pkg/api"
	"github.com/coreos/prometheus-operator/pkg/prometheus"
	"github.com/go-kit/kit/log"
)

var (
	cfg              prometheus.Config
	analyticsEnabled bool
)

func init() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flagset.StringVar(&cfg.Host, "apiserver", "", "API Server addr, e.g. ' - NOT RECOMMENDED FOR PRODUCTION - http://127.0.0.1:8080'. Omit parameter to run in on-cluster mode and utilize the service account token.")
	flagset.StringVar(&cfg.TLSConfig.CertFile, "cert-file", "", " - NOT RECOMMENDED FOR PRODUCTION - Path to public TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.KeyFile, "key-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to private TLS certificate file.")
	flagset.StringVar(&cfg.TLSConfig.CAFile, "ca-file", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to TLS CA file.")
	flagset.BoolVar(&cfg.TLSInsecure, "tls-insecure", false, "- NOT RECOMMENDED FOR PRODUCTION - Don't verify API server's CA certificate.")
	flagset.BoolVar(&analyticsEnabled, "analytics", true, "Send analytical event (Cluster Created/Deleted etc.) to Google Analytics")

	flagset.Parse(os.Args[1:])
}

func Main() int {
	logger := log.NewContext(log.NewLogfmtLogger(os.Stdout)).
		With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	if analyticsEnabled {
		analytics.Enable()
	}

	po, err := prometheus.New(cfg, logger.With("component", "prometheusoperator"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	ao, err := alertmanager.New(cfg, logger.With("component", "alertmanageroperator"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	web, err := api.New(cfg, logger.With("component", "api"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	mux := http.DefaultServeMux
	web.Register(mux)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	stopc := make(chan struct{})
	errc := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if err := po.Run(stopc); err != nil {
			errc <- err
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := ao.Run(stopc); err != nil {
			errc <- err
		}
		wg.Done()
	}()

	go func() {
		if err := http.Serve(l, nil); err != nil {
			errc <- err
		}
	}()

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		logger.Log("msg", "Received SIGTERM, exiting gracefully...")
		l.Close()
		close(stopc)
		wg.Wait()
	case err := <-errc:
		logger.Log("msg", "Unhandled error received. Exiting...", "err", err)
		l.Close()
		close(stopc)
		wg.Wait()
		return 1
	}

	return 0
}

func main() {
	os.Exit(Main())
}
