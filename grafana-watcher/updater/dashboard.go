package updater

import (
	"log"
	"os"
	"path/filepath"

	"github.com/coreos/kube-prometheus/grafana-watcher/grafana"
)

type Updater interface {
	Init() error
	OnModify() error
}

type GrafanaDashboardUpdater struct {
	client grafana.DashboardsInterface
	glob   string
}

func NewGrafanaDashboardUpdater(c grafana.DashboardsInterface, g string) Updater {
	return &GrafanaDashboardUpdater{
		client: c,
		glob:   g,
	}
}

func (u *GrafanaDashboardUpdater) Init() error {
	return u.updateDashboards()
}

func (u *GrafanaDashboardUpdater) OnModify() error {
	return u.updateDashboards()
}

func (u *GrafanaDashboardUpdater) updateDashboards() error {
	err := u.deleteAllDashboards()
	if err != nil {
		return err
	}
	err = u.createDashboardsFromFiles()
	if err != nil {
		return err
	}

	return nil
}

func (u *GrafanaDashboardUpdater) deleteAllDashboards() error {
	log.Println("Retrieving existing dashboards")
	searchResults, err := u.client.Search()
	if err != nil {
		return err
	}

	log.Printf("Deleting %d dashboards\n", len(searchResults))
	for _, d := range searchResults {
		log.Println("Deleting dashboard:", d.Slug())

		err := u.client.Delete(d.Slug())
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaDashboardUpdater) createDashboardsFromFiles() error {
	filePaths, err := filepath.Glob(u.glob)
	if err != nil {
		return err
	}

	for _, fp := range filePaths {
		u.createDashboardFromFile(fp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaDashboardUpdater) createDashboardFromFile(filePath string) error {
	log.Println("Creating dashboard from:", filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return u.client.Create(f)
}
