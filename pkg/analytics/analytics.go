package analytics

import (
	"sync"

	ga "github.com/jpillora/go-ogle-analytics"
)

const (
	id       = "UA-42684979-8"
	category = "prometheus-operator"
)

var (
	mu     sync.Mutex
	client *ga.Client
)

func Enable() {
	mu.Lock()
	defer mu.Unlock()
	client = mustNewClient()
}

func Disable() {
	mu.Lock()
	defer mu.Unlock()
	client = nil
}

func send(e *ga.Event) {
	mu.Lock()
	c := client
	mu.Unlock()

	if c == nil {
		return
	}
	// error is ignored intentionally. we try to send event to GA in a best effort approach.
	c.Send(e)
}

func mustNewClient() *ga.Client {
	client, err := ga.NewClient(id)
	if err != nil {
		panic(err)
	}
	return client
}

func PrometheusCreated() {
	send(ga.NewEvent(category, "prometheus_created"))
}

func PrometheusDeleted() {
	send(ga.NewEvent(category, "prometheus_deleted"))
}
