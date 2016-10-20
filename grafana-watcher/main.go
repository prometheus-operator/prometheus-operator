package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"

	"github.com/coreos/kube-prometheus/grafana-watcher/grafana"
	"github.com/coreos/kube-prometheus/grafana-watcher/updater"
)

var (
	watchDir   = flag.String("watch-dir", "", "The directory the ConfigMap is mounted into to watch for updates.")
	grafanaUrl = flag.String("grafana-url", "", "The url to issue requests to update dashboards to.")
)

type volumeWatcher struct {
	watchDir string
	handlers []updater.Updater
}

func newVolumeWatcher(d string) *volumeWatcher {
	return &volumeWatcher{
		watchDir: d,
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
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						log.Println("ConfigMap modified")
						for _, h := range w.handlers {
							err := h.OnModify()
							if err != nil {
								log.Println("error:", err)
							}
						}
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Println("Starting...")
	err = watcher.Add(*watchDir)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func main() {
	flag.Parse()

	if *watchDir == "" {
		log.Println("Missing watch-dir\n")
		flag.Usage()
		os.Exit(1)
	}
	if *grafanaUrl == "" {
		log.Println("Missing grafana-url\n")
		flag.Usage()
		os.Exit(1)
	}

	g := grafana.New(*grafanaUrl)

	for {
		log.Println("Waiting for Grafana to be available.")
		_, err := g.Datasources().All()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	su := updater.NewGrafanaDatasourceUpdater(g.Datasources(), filepath.Join(*watchDir, "*-datasource.json"))
	log.Println("Initializing datasources.")
	err := su.Init()
	if err != nil {
		log.Fatal(err)
	}

	du := updater.NewGrafanaDashboardUpdater(g.Dashboards(), filepath.Join(*watchDir, "*-dashboard.json"))
	log.Println("Initializing dashboards.")
	err = du.Init()
	if err != nil {
		log.Fatal(err)
	}

	w := newVolumeWatcher(*watchDir)

	w.AddEventHandler(du)
	w.AddEventHandler(su)

	w.Run()
}
