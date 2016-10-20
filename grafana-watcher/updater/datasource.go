package updater

import (
	"log"
	"os"
	"path/filepath"

	"github.com/coreos/kube-prometheus/grafana-watcher/grafana"
)

type GrafanaDatasourceUpdater struct {
	client grafana.DatasourcesInterface
	glob   string
}

func NewGrafanaDatasourceUpdater(c grafana.DatasourcesInterface, g string) Updater {
	return &GrafanaDatasourceUpdater{
		client: c,
		glob:   g,
	}
}

func (u *GrafanaDatasourceUpdater) Init() error {
	return u.updateDatasources()
}

func (u *GrafanaDatasourceUpdater) OnModify() error {
	return u.updateDatasources()
}

func (u *GrafanaDatasourceUpdater) updateDatasources() error {
	err := u.deleteAllDatasources()
	if err != nil {
		return err
	}
	err = u.createDatasourcesFromFiles()
	if err != nil {
		return err
	}

	return nil
}

func (u *GrafanaDatasourceUpdater) deleteAllDatasources() error {
	log.Println("Retrieving existing datasources")
	datasources, err := u.client.All()
	if err != nil {
		return err
	}

	log.Printf("Deleting %d datasources\n", len(datasources))
	for _, d := range datasources {
		log.Println("Deleting datasource:", d.Id)

		err := u.client.Delete(d.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaDatasourceUpdater) createDatasourcesFromFiles() error {
	filePaths, err := filepath.Glob(u.glob)
	if err != nil {
		return err
	}

	for _, fp := range filePaths {
		u.createDatasourceFromFile(fp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaDatasourceUpdater) createDatasourceFromFile(filePath string) error {
	log.Println("Creating datasource from:", filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return u.client.Create(f)
}
