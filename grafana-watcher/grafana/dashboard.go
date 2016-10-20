package grafana

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type DashboardsInterface interface {
	Search() ([]GrafanaDashboard, error)
	Create(dashboardJson io.Reader) error
	Delete(slug string) error
}

type DashboardsClient struct {
	BaseUrl    string
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

func NewDashboardsClient(baseUrl string, c *http.Client) DashboardsInterface {
	return &DashboardsClient{
		BaseUrl:    baseUrl,
		HTTPClient: c,
	}
}

func (c *DashboardsClient) Search() ([]GrafanaDashboard, error) {
	searchUrl := c.BaseUrl + "/api/search"
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
	deleteUrl := c.BaseUrl + "/api/dashboards/db/" + slug
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	_, err = c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *DashboardsClient) Create(dashboardJson io.Reader) error {
	importDashboardUrl := c.BaseUrl + "/api/dashboards/import"
	_, err := c.HTTPClient.Post(importDashboardUrl, "application/json", dashboardJson)
	if err != nil {
		return err
	}

	return nil
}
