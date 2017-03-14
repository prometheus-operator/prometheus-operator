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
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/ericchiang/k8s"
)

type config struct {
	configMapFile string
	ruleFileDir   string
	namespace     string
}

type volumeWatcher struct {
	client *k8s.Client
	cfg    config
}

func newVolumeWatcher(client *k8s.Client, cfg config) *volumeWatcher {
	return &volumeWatcher{
		client: client,
		cfg:    cfg,
	}
}

func (w *volumeWatcher) UpdateRuleFiles() error {
	file, err := os.Open(w.cfg.configMapFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpdir, err := ioutil.TempDir(w.cfg.ruleFileDir, "prometheus-rule-files")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		err := w.writeRuleConfigMap(tmpdir, i, scanner.Text())
		if err != nil {
			log.Println("err", err)
		}
		i++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err = w.placeNewRuleFiles(tmpdir, w.cfg.ruleFileDir)
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

func (w *volumeWatcher) writeRuleConfigMap(rulesDir string, index int, configMapName string) error {
	cm, err := w.client.CoreV1().GetConfigMap(context.TODO(), configMapName, w.cfg.namespace)
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

	_, err = f.Write([]byte(content))
	return err
}

func (w *volumeWatcher) Run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						log.Println("ConfigMap modified. Updating rule files...")
						err := w.UpdateRuleFiles()
						if err != nil {
							log.Println("err", err)
						}
						log.Println("Rule files updated.")
					}
				}
			case err := <-watcher.Errors:
				log.Println("err", err)
			}
		}
	}()

	log.Println("Starting...")
	err = watcher.Add(filepath.Dir(w.cfg.configMapFile))
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func main() {
	cfg := config{}
	flags := flag.NewFlagSet("prometheus-watcher", flag.ExitOnError)
	flags.StringVar(&cfg.configMapFile, "configmap-file", "", "A file containing a list of ConfigMap names.")
	flags.StringVar(&cfg.ruleFileDir, "rule-file-dir", "", "A directory where rule files will be written to.")
	flags.StringVar(&cfg.namespace, "namespace", "", "The namespace to get ConfigMaps from.")
	flags.Parse(os.Args[1:])

	if cfg.configMapFile == "" {
		log.Println("Missing file to watch for changes\n")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.ruleFileDir == "" {
		log.Println("Missing directory to write rule files into\n")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.namespace == "" {
		log.Println("Missing namespace to get ConfigMaps from\n")
		flag.Usage()
		os.Exit(1)
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}

	newVolumeWatcher(client, cfg).Run()
}
