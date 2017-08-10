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

package updater

import (
	"log"
	"os"
	"path/filepath"

	"github.com/coreos/prometheus-operator/contrib/grafana-watcher/grafana"
)

type GrafanaDatasourceUpdater struct {
	client grafana.DatasourcesInterface
	globs  []string
}

func NewGrafanaDatasourceUpdater(c grafana.DatasourcesInterface, g []string) Updater {
	return &GrafanaDatasourceUpdater{
		client: c,
		globs:  g,
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
		log.Println("Deleting datasource:", d.Name)

		err := u.client.Delete(d.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaDatasourceUpdater) createDatasourcesFromFiles() error {
	for _, glob := range u.globs {
		filePaths, err := filepath.Glob(filepath.Join(glob, "*-datasource.json"))
		if err != nil {
			return err
		}

		for _, fp := range filePaths {
			err = u.createDatasourceFromFile(fp)
			if err != nil {
				return err
			}
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
