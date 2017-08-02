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
	"path"
	"strconv"
)

type DatasourcesInterface interface {
	All() ([]GrafanaDatasource, error)
	Create(datasourceJson io.Reader) error
	Delete(id int) error
}

// DatasourcesClient is an implementation of the DatasourcesInterface. The
// datasources HTTP API of Grafana requires admin access.
type DatasourcesClient struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
}

type GrafanaDatasource struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func NewDatasourcesClient(baseUrl *url.URL, c *http.Client) DatasourcesInterface {
	return &DatasourcesClient{
		BaseUrl:    baseUrl,
		HTTPClient: c,
	}
}

func (c *DatasourcesClient) All() ([]GrafanaDatasource, error) {
	allUrl := makeUrl(c.BaseUrl, "/api/datasources")
	resp, err := c.HTTPClient.Get(allUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		err := errors.New("Invalid credentials. Update and reapply the grafana-credentials manifest, then redeploy grafana.")
		return nil, err
	}
	datasources := make([]GrafanaDatasource, 0)

	err = json.NewDecoder(resp.Body).Decode(&datasources)
	if err != nil {
		return nil, err
	}

	return datasources, nil
}

func (c *DatasourcesClient) Delete(id int) error {
	deleteUrl := makeUrl(c.BaseUrl, "/api/datasources/"+strconv.Itoa(id))
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return doRequest(c.HTTPClient, req)
}

func (c *DatasourcesClient) Create(datasourceJson io.Reader) error {
	createUrl := makeUrl(c.BaseUrl, "/api/datasources")
	req, err := http.NewRequest("POST", createUrl, datasourceJson)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	return doRequest(c.HTTPClient, req)
}

func makeUrl(baseUrl *url.URL, endpoint string) string {
	result := &url.URL{}
	*result = *baseUrl

	result.Path = path.Join(result.Path, endpoint)

	return result.String()
}
