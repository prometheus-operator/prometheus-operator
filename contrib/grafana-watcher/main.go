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
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/coreos/prometheus-operator/contrib/grafana-watcher/grafana"
	"github.com/coreos/prometheus-operator/contrib/grafana-watcher/updater"
	"github.com/spf13/pflag"
)

type watchDirSet map[string]struct{}

func (w *watchDirSet) String() string {
	s := *w
	return strings.Join(s.asSlice(), ",")
}

func (w *watchDirSet) Set(value string) error {
	s := *w
	cols := strings.Split(value, ",")
	for _, col := range cols {
		s[col] = struct{}{}
	}
	return nil
}

func (w watchDirSet) asSlice() []string {
	cols := []string{}
	for col, _ := range w {
		cols = append(cols, col)
	}
	return cols
}

func (w watchDirSet) isEmpty() bool {
	return len(w.asSlice()) == 0
}

func (w *watchDirSet) Type() string {
	return "map[string]struct{}"
}

type options struct {
	grafanaUrl string
	watchDirs  watchDirSet
	help       bool
}

type volumeWatcher struct {
	watchDirs watchDirSet
	handlers  []updater.Updater
}

func newVolumeWatcher(watchDirs watchDirSet) *volumeWatcher {
	return &volumeWatcher{
		watchDirs: watchDirs,
	}
}

func (w *volumeWatcher) AddEventHandler(handler updater.Updater) {
	w.handlers = append(w.handlers, handler)
}

func (w *volumeWatcher) Run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		var handlersToRetry []updater.Updater
		retryTicker := time.NewTicker(10 * time.Second)
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						log.Println("ConfigMap modified")
						handlersToRetry = []updater.Updater{}
						for _, h := range w.handlers {
							err := h.OnModify()
							if err != nil {
								log.Println("error:", err)
								handlersToRetry = append(handlersToRetry, h)
							}
						}
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			case <-retryTicker.C:
				// Retry any operations that failed during the last update attempt
				// New events seen by the watcher will clear the retry list, though
				// it may take several cycles through the select statement before this happens.
				// See: https://golang.org/ref/spec#Select_statements
				if len(handlersToRetry) < 1 {
					break
				}
				log.Printf("Retrying %v previously failed operations...", len(handlersToRetry))
				var remainingHandlers []updater.Updater
				for _, h := range handlersToRetry {
					// Only retry if the watcher still cares about this handler; could have been removed
					found := false
					for _, h2 := range w.handlers {
						if h == h2 {
							found = true
							break
						}
					}
					if !found {
						continue
					}

					if err := h.OnModify(); err != nil {
						log.Println("error during retry attempt:", err)
						remainingHandlers = append(remainingHandlers, h)
					}
				}
				handlersToRetry = remainingHandlers
			}
		}
	}()

	log.Println("Starting...")
	for watchDir := range w.watchDirs {
		err = watcher.Add(watchDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

func main() {
	flag.CommandLine.Parse([]string{})

	options := &options{watchDirs: make(watchDirSet)}
	flags := pflag.NewFlagSet("", pflag.ExitOnError)

	flags.Var(&options.watchDirs, "watch-dir", "Directories to watch for updates.")
	flags.StringVar(&options.grafanaUrl, "grafana-url", "", "The url to issue requests to update dashboards to.")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
	}

	err := flags.Parse(os.Args)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	if options.help {
		flags.Usage()
		os.Exit(0)
	}

	watchDirs := options.watchDirs
	if len(watchDirs) == 0 {
		log.Fatal("No watch dir specified")
	}

	grafanaUrl := options.grafanaUrl
	if grafanaUrl == "" {
		log.Fatal("Missing grafana-url")
	}

	gUrl, err := url.Parse(grafanaUrl)
	if err != nil {
		log.Fatalf("Grafana URL could not be parsed: %s", grafanaUrl)
	}

	if os.Getenv("GRAFANA_USER") != "" && os.Getenv("GRAFANA_PASSWORD") == "" {
		gUrl.User = url.User(os.Getenv("GRAFANA_USER"))
	}

	if os.Getenv("GRAFANA_USER") != "" && os.Getenv("GRAFANA_PASSWORD") != "" {
		gUrl.User = url.UserPassword(os.Getenv("GRAFANA_USER"), os.Getenv("GRAFANA_PASSWORD"))
	}

	g := grafana.New(gUrl)

	for {
		_, err := g.Datasources().All()
		if err == nil {
			break
		}
		fmt.Fprintln(os.Stderr, err)
		time.Sleep(time.Second)
	}

	dirs := []string{}
	for dir, _ := range watchDirs {
		dirs = append(dirs, dir)
	}

	su := updater.NewGrafanaDatasourceUpdater(g.Datasources(), dirs)
	log.Println("Initializing datasources.")
	err = su.Init()
	if err != nil {
		log.Fatal(err)
	}

	du := updater.NewGrafanaDashboardUpdater(g.Dashboards(), dirs)
	log.Println("Initializing dashboards.")
	err = du.Init()
	if err != nil {
		log.Fatal(err)
	}

	w := newVolumeWatcher(watchDirs)

	w.AddEventHandler(du)
	w.AddEventHandler(su)

	w.Run()
}
