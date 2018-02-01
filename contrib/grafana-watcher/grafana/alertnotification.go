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

package grafana

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type AlertNotificationsInterface interface {
	All() ([]GrafanaAlertNotification, error)
	Create(alertNotificationJson io.Reader) error
	Delete(id int) error
}

// AlertNotificationsClient is an implementation of the AlertNotificationsInterface. The
// alertNotifications HTTP API of Grafana requires admin access.
type AlertNotificationsClient struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
}

type GrafanaAlertNotification struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func NewAlertNotificationsClient(baseUrl *url.URL, c *http.Client) AlertNotificationsInterface {
	return &AlertNotificationsClient{
		BaseUrl:    baseUrl,
		HTTPClient: c,
	}
}

func (c *AlertNotificationsClient) All() ([]GrafanaAlertNotification, error) {
	allUrl := makeUrl(c.BaseUrl, "/api/alert-notifications")
	resp, err := c.HTTPClient.Get(allUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		err := errors.New("Invalid credentials. Update and reapply the grafana-credentials manifest, then redeploy grafana.")
		return nil, err
	}
	alertNotifications := make([]GrafanaAlertNotification, 0)

	err = json.NewDecoder(resp.Body).Decode(&alertNotifications)
	if err != nil {
		return nil, err
	}

	return alertNotifications, nil
}

func (c *AlertNotificationsClient) Delete(id int) error {
	deleteUrl := makeUrl(c.BaseUrl, "/api/alert-notifications/"+strconv.Itoa(id))
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return doRequest(c.HTTPClient, req)
}

func (c *AlertNotificationsClient) Create(alertNotificationJson io.Reader) error {
	createUrl := makeUrl(c.BaseUrl, "/api/alert-notifications")
	req, err := http.NewRequest("POST", createUrl, alertNotificationJson)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	return doRequest(c.HTTPClient, req)
}
