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

func AlertmanagerCreated() {
	send(ga.NewEvent(category, "alertmanager_created"))
}

func AlertmanagerDeleted() {
	send(ga.NewEvent(category, "alertmanager_deleted"))
}
