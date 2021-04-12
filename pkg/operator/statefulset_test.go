// Copyright 2021 The prometheus-operator Authors
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

package operator

import (
	"net/url"
	"testing"
)

var reloaderConfig = ReloaderConfig{
	CPURequest:    "100m",
	CPULimit:      "100m",
	MemoryRequest: "50Mi",
	MemoryLimit:   "50Mi",
	Image:         "quay.io/prometheus-operator/prometheus-config-reloader:latest",
}

func TestCreateInitConfigReloader(t *testing.T) {
	initContainerName := "init-config-reloader"
	var container = CreateConfigReloader(
		initContainerName,
		ReloaderResources(reloaderConfig),
		ReloaderRunOnce())
	if container.Name != "init-config-reloader" {
		t.Errorf("Expected container name %s, but found %s", initContainerName, container.Name)
	}
	if !contains(container.Args, "--watch-interval=0") {
		t.Errorf("Expected '--watch-interval=0' does not exist in container arguments")
	}
}

func TestCreateConfigReloader(t *testing.T) {
	containerName := "config-reloader"
	var container = CreateConfigReloader(
		containerName,
		ReloaderResources(reloaderConfig),
		ReloaderURL(url.URL{
			Scheme: "http",
			Host:   "localhost:9093",
			Path:   "/-/reload",
		}),
		ListenAddress(),
		ReloadURL(),
	)
	if container.Name != "config-reloader" {
		t.Errorf("Expected container name %s, but found %s", containerName, container.Name)
	}
	if !contains(container.Args, "--listen-address=:8080") {
		t.Errorf("Expected '--listen-address=:8080' not found in %s", container.Args)
	}
	if !contains(container.Args, "--reload-url=http://localhost:9093/-/reload") {
		t.Errorf("Expected '--reload-url=http://localhost:9093/-/reload' not found in %s", container.Args)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
