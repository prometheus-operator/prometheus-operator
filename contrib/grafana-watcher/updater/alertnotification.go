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

type GrafanaAlertNotificationUpdater struct {
	client grafana.AlertNotificationsInterface
	globs  []string
}

func NewGrafanaAlertNotificationUpdater(c grafana.AlertNotificationsInterface, g []string) Updater {
	return &GrafanaAlertNotificationUpdater{
		client: c,
		globs:  g,
	}
}

func (u *GrafanaAlertNotificationUpdater) Init() error {
	return u.updateAlertNotifications()
}

func (u *GrafanaAlertNotificationUpdater) OnModify() error {
	return u.updateAlertNotifications()
}

func (u *GrafanaAlertNotificationUpdater) updateAlertNotifications() error {
	err := u.deleteAllAlertNotifications()
	if err != nil {
		return err
	}
	err = u.createAlertNotificationsFromFiles()
	if err != nil {
		return err
	}

	return nil
}

func (u *GrafanaAlertNotificationUpdater) deleteAllAlertNotifications() error {
	log.Println("Retrieving existing alertnotifications")
	alertnotifications, err := u.client.All()
	if err != nil {
		return err
	}

	log.Printf("Deleting %d alertnotifications\n", len(alertnotifications))
	for _, d := range alertnotifications {
		log.Println("Deleting alertnotification:", d.Name)

		err := u.client.Delete(d.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *GrafanaAlertNotificationUpdater) createAlertNotificationsFromFiles() error {
	for _, glob := range u.globs {
		filePaths, err := filepath.Glob(filepath.Join(glob, "*-alertnotification.json"))
		if err != nil {
			return err
		}

		for _, fp := range filePaths {
			err = u.createAlertNotificationFromFile(fp)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *GrafanaAlertNotificationUpdater) createAlertNotificationFromFile(filePath string) error {
	log.Println("Creating alertnotification from:", filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return u.client.Create(f)
}
