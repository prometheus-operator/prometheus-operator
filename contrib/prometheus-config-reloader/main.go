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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/cenkalti/backoff"
	"github.com/ericchiang/k8s"
	"github.com/go-kit/kit/log"
)

type config struct {
	configVolumeDir string
	ruleVolumeDir   string
	reloadUrl       string
}

type volumeWatcher struct {
	client *k8s.Client
	cfg    config
	logger log.Logger
}

func newVolumeWatcher(client *k8s.Client, cfg config, logger log.Logger) *volumeWatcher {
	return &volumeWatcher{
		client: client,
		cfg:    cfg,
		logger: logger,
	}
}

type ConfigMapReference struct {
	Key string `json:"key"`
}

type ConfigMapReferenceList struct {
	Items []*ConfigMapReference `json:"items"`
}

func (w *volumeWatcher) UpdateRuleFiles() error {
	file, err := os.Open(filepath.Join(w.cfg.configVolumeDir, "configmaps.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	configMaps := ConfigMapReferenceList{}
	err = json.NewDecoder(file).Decode(&configMaps)
	if err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir(w.cfg.ruleVolumeDir, "prometheus-rule-files")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	for i, cm := range configMaps.Items {
		err := w.writeRuleConfigMap(tmpdir, i, cm.Key)
		if err != nil {
			return err
		}
	}

	err = w.placeNewRuleFiles(tmpdir, w.cfg.ruleVolumeDir)
	if err != nil {
		return err
	}

	return nil
}

func (w *volumeWatcher) placeNewRuleFiles(tmpdir, ruleFileDir string) error {
	err := os.MkdirAll(ruleFileDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = w.removeOldRuleFiles(ruleFileDir, tmpdir)
	if err != nil {
		return err
	}
	d, err := os.Open(tmpdir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.Rename(filepath.Join(tmpdir, name), filepath.Join(ruleFileDir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *volumeWatcher) removeOldRuleFiles(dir string, tmpdir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		s := filepath.Join(dir, name)
		if s != tmpdir {
			err = os.RemoveAll(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *volumeWatcher) writeRuleConfigMap(rulesDir string, index int, configMap string) error {
	configMapParts := strings.Split(configMap, "/")
	if len(configMapParts) != 2 {
		return fmt.Errorf("Malformatted configmap key: %s. Format must be namespace/name.", configMap)
	}
	configMapNamespace := configMapParts[0]
	configMapName := configMapParts[1]

	cm, err := w.client.CoreV1().GetConfigMap(context.TODO(), configMapName, configMapNamespace)
	if err != nil {
		return err
	}

	dir := filepath.Join(rulesDir, fmt.Sprintf("rules-%d", index))
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	for filename, content := range cm.Data {
		err = w.writeConfigMapFile(filepath.Join(dir, filename), content)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *volumeWatcher) writeConfigMapFile(filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func (w *volumeWatcher) ReloadPrometheus() error {
	req, err := http.NewRequest("POST", w.cfg.reloadUrl, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Received response code %s, expected 200", resp.StatusCode)
	}
	return nil
}

func (w *volumeWatcher) Refresh() {
	w.logger.Log("msg", "Updating rule files...")
	err := w.UpdateRuleFiles()
	if err != nil {
		w.logger.Log("msg", "Updating rule files failed.", "err", err)
	} else {
		w.logger.Log("msg", "Rule files updated.")
	}

	w.logger.Log("msg", "Reloading Prometheus...")
	err = backoff.RetryNotify(w.ReloadPrometheus, backoff.NewExponentialBackOff(), func(err error, next time.Duration) {
		w.logger.Log("msg", "Reloading Prometheus temporarily failed.", "err", err, "next-retry", next)
	})
	if err != nil {
		w.logger.Log("msg", "Reloading Prometheus failed.", "err", err)
	} else {
		w.logger.Log("msg", "Prometheus successfully reloaded.")
	}
}

func (w *volumeWatcher) Run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Log("msg", "Creating a new watcher failed.", "err", err)
		os.Exit(1)
	}
	defer watcher.Close()

	w.logger.Log("msg", "Starting...")
	w.Refresh()
	err = watcher.Add(w.cfg.configVolumeDir)
	if err != nil {
		w.logger.Log("msg", "Adding config volume to be watched failed.", "err", err)
		os.Exit(1)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				if filepath.Base(event.Name) == "..data" {
					w.logger.Log("msg", "ConfigMap modified.")
					w.Refresh()
				}
			}
		case err := <-watcher.Errors:
			w.logger.Log("err", err)
		}
	}
}

func main() {
	logger := log.NewContext(log.NewLogfmtLogger(os.Stdout)).
		With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	cfg := config{}
	flags := flag.NewFlagSet("prometheus-config-reloader", flag.ExitOnError)
	flags.StringVar(&cfg.configVolumeDir, "config-volume-dir", "", "The directory to watch for changes to reload Prometheus.")
	flags.StringVar(&cfg.ruleVolumeDir, "rule-volume-dir", "", "The directory to write rule files to.")
	flags.StringVar(&cfg.reloadUrl, "reload-url", "", "The URL to call when intending to reload Prometheus.")
	flags.Parse(os.Args[1:])

	if cfg.ruleVolumeDir == "" {
		logger.Log("Missing directory to write rule files into\n")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.configVolumeDir == "" {
		logger.Log("Missing directory to watch for configuration changes\n")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.reloadUrl == "" {
		logger.Log("Missing URL to call when intending to reload Prometheus\n")
		flag.Usage()
		os.Exit(1)
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	newVolumeWatcher(client, cfg, logger.With("component", "volume-watcher")).Run()
}
