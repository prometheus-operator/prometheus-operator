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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericchiang/k8s"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/improbable-eng/thanos/pkg/reloader"
	"github.com/oklog/run"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	app := kingpin.New("prometheus-config-reloader", "")

	cfgFile := app.Flag("config-file", "config file watched by the reloader").
		String()

	ruleListFile := app.Flag("rule-list-file", "file listing configmaps of rules to load dynamically").
		String()

	cfgSubstFile := app.Flag("config-envsubst-file", "output file for environment variable substituted config file").
		String()

	ruleDir := app.Flag("rule-dir", "rule directory for the reloader to refresh").String()

	reloadURL := app.Flag("reload-url", "reload URL to trigger Prometheus reload on").
		Default("http://127.0.0.1:9090/-/reload").URL()

	if _, err := app.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := os.MkdirAll(*ruleDir, 0777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var g run.Group
	{
		ctx, cancel := context.WithCancel(context.Background())
		rel := reloader.New(logger, *reloadURL, *cfgFile, *cfgSubstFile, *ruleDir)

		g.Add(func() error {
			return rel.Watch(ctx)
		}, func(error) {
			cancel()
		})
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		tick := time.NewTicker(1 * time.Minute)

		rfet := newRuleFetcher(client, *ruleListFile, *ruleDir)

		g.Add(func() error {
			defer tick.Stop()

			for {
				select {
				case <-tick.C:
					if err := rfet.Refresh(ctx); err != nil {
						level.Error(logger).Log("msg", "updating rules failed", "err", err)
					}
				case <-ctx.Done():
					return nil
				}
			}
		}, func(error) {
			cancel()
		})
	}

	if err := g.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type configMapRef struct {
	Key string `json:"key"`
}

type ruleFetcher struct {
	client   *k8s.Client
	listFile string
	outDir   string

	lastHash [sha256.Size]byte
}

func newRuleFetcher(client *k8s.Client, listFile, outDir string) *ruleFetcher {
	return &ruleFetcher{
		client:   client,
		listFile: listFile,
		outDir:   outDir,
	}
}

func (rf *ruleFetcher) Refresh(ctx context.Context) error {
	b, err := ioutil.ReadFile(rf.listFile)
	if err != nil {
		return err
	}

	h := sha256.Sum256(b)
	if rf.lastHash == h {
		return nil
	}

	var cms struct {
		Items []*configMapRef `json:"items"`
	}
	err = json.Unmarshal(b, &cms)
	if err != nil {
		return err
	}
	if err := rf.refresh(ctx, cms.Items); err != nil {
		return err
	}

	rf.lastHash = h
	return nil
}

func (rf *ruleFetcher) refresh(ctx context.Context, cms []*configMapRef) error {
	tmpdir := rf.outDir + ".tmp"

	if err := os.MkdirAll(tmpdir, 0777); err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	for i, cm := range cms {
		parts := strings.Split(cm.Key, "/")
		if len(parts) != 2 {
			return fmt.Errorf("malformatted configmap key: %s, must be namespace/name", cm.Key)
		}
		namespace, name := parts[0], parts[1]

		cm, err := rf.client.CoreV1().GetConfigMap(ctx, name, namespace)
		if err != nil {
			return err
		}
		dir := filepath.Join(tmpdir, fmt.Sprintf("rules-%d", i))

		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		for fn, content := range cm.Data {
			if err := ioutil.WriteFile(filepath.Join(dir, fn), []byte(content), 0666); err != nil {
				return err
			}
		}
	}

	if err := os.RemoveAll(rf.outDir); err != nil {
		return err
	}
	return os.Rename(tmpdir, rf.outDir)
}
