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
	"io"
	"net/http"
	"net/url"
	"strings"
)

type DashboardsInterface interface {
	Search() ([]GrafanaDashboard, error)
	Create(dashboardJson io.Reader) error
	Delete(slug string) error
}

type DashboardsClient struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
}

type GrafanaDashboard struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Uri   string `json:"uri"`
}

func (d *GrafanaDashboard) Slug() string {
	// The uri in the search result contains the slug.
	// http://docs.grafana.org/v3.1/http_api/dashboard/#search-dashboards
	return strings.TrimPrefix(d.Uri, "db/")
}

func NewDashboardsClient(baseUrl *url.URL, c *http.Client) DashboardsInterface {
	return &DashboardsClient{
		BaseUrl:    baseUrl,
		HTTPClient: c,
	}
}

func (c *DashboardsClient) Search() ([]GrafanaDashboard, error) {
	searchUrl := makeUrl(c.BaseUrl, "/api/search")
	resp, err := c.HTTPClient.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	searchResult := make([]GrafanaDashboard, 0)
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return nil, err
	}

	return searchResult, nil
}

func (c *DashboardsClient) Delete(slug string) error {
	deleteUrl := makeUrl(c.BaseUrl, "/api/dashboards/db/"+slug)
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return doRequest(c.HTTPClient, req)
}

func (c *DashboardsClient) Create(dashboardJson io.Reader) error {
	importDashboardUrl := makeUrl(c.BaseUrl, "/api/dashboards/import")
	req, err := http.NewRequest("POST", importDashboardUrl, dashboardJson)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	return doRequest(c.HTTPClient, req)
}
