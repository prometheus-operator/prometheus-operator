package grafana

import (
	"net/http"
)

type Interface interface {
	Dashboards() DashboardsInterface
	Datasources() DatasourcesInterface
}

type Clientset struct {
	BaseUrl    string
	HTTPClient *http.Client
}

func New(baseUrl string) Interface {
	return &Clientset{
		BaseUrl:    baseUrl,
		HTTPClient: http.DefaultClient,
	}
}

func (c *Clientset) Dashboards() DashboardsInterface {
	return NewDashboardsClient(c.BaseUrl, c.HTTPClient)
}

func (c *Clientset) Datasources() DatasourcesInterface {
	return NewDatasourcesClient(c.BaseUrl, c.HTTPClient)
}
