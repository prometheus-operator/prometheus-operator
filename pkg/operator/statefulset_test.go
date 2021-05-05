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
	"fmt"
	"net/url"
	"strconv"
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
	logFormat := "logFormat"
	logLevel := "logLevel"
	configFile := "configFile"
	configEnvsubstFile := "configEnvsubstFile"
	watchedDirectories := []string{"directory1", "directory2"}
	shard := int32(1)
	var container = CreateConfigReloader(
		containerName,
		ReloaderResources(reloaderConfig),
		ReloaderURL(url.URL{
			Scheme: "http",
			Host:   "localhost:9093",
			Path:   "/-/reload",
		}),
		ListenLocal(true),
		LocalHost("localhost"),
		LogFormat(logFormat),
		LogLevel(logLevel),
		ConfigFile(configFile),
		ConfigEnvsubstFile(configEnvsubstFile),
		WatchedDirectories(watchedDirectories),
		Shard(shard),
	)
	if container.Name != "config-reloader" {
		t.Errorf("Expected container name %s, but found %s", containerName, container.Name)
	}
	if !contains(container.Args, "--listen-address=localhost:8080") {
		t.Errorf("Expected '--listen-address=localhost:8080' not found in %s", container.Args)
	}
	if !contains(container.Args, "--reload-url=http://localhost:9093/-/reload") {
		t.Errorf("Expected '--reload-url=http://localhost:9093/-/reload' not found in %s", container.Args)
	}
	if !contains(container.Args, "--log-level=logLevel") {
		t.Errorf("Expected '--log-level=%s' not found in %s", logLevel, container.Args)
	}
	if !contains(container.Args, "--log-format=logFormat") {
		t.Errorf("Expected '--log-format=%s' not found in %s", logFormat, container.Args)
	}
	if !contains(container.Args, "--config-file=configFile") {
		t.Errorf("Expected '--config-file=%s' not found in %s", configFile, container.Args)
	}
	if !contains(container.Args, "--config-envsubst-file=configEnvsubstFile") {
		t.Errorf("Expected '--config-envsubst-file=%s' not found in %s", configEnvsubstFile, container.Args)
	}
	for _, dir := range watchedDirectories {
		if !contains(container.Args, fmt.Sprintf("--watched-dir=%s", dir)) {
			t.Errorf("Expected '--watched-dir=%s' not found in %s", dir, container.Args)
		}
	}

	flag := false
	for _, val := range container.Env {
		if val.Value == strconv.Itoa(int(shard)) {
			flag = true
		}
	}
	if !flag {
		t.Errorf("Expected shard value '%d' not found in %s", shard, container.Env)
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
