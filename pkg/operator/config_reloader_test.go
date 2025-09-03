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
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var reloaderConfig = ContainerConfig{
	CPURequests:    Quantity{q: resource.MustParse("100m")},
	CPULimits:      Quantity{q: resource.MustParse("100m")},
	MemoryRequests: Quantity{q: resource.MustParse("50Mi")},
	MemoryLimits:   Quantity{q: resource.MustParse("50Mi")},
	Image:          "quay.io/prometheus-operator/prometheus-config-reloader:latest",
}

func TestCreateConfigReloaderEnableProbes(t *testing.T) {
	reloaderConfigCopy := reloaderConfig
	reloaderConfigCopy.EnableProbes = true
	containerName := "config-reloader"
	var container = CreateConfigReloader(
		containerName,
		ReloaderConfig(reloaderConfigCopy),
		ReloaderURL(url.URL{
			Scheme: "http",
			Host:   "localhost:9093",
			Path:   "/-/reload",
		}),
		ListenLocal(true),
		LocalHost("localhost"),
	)

	if container.Name != "config-reloader" {
		t.Errorf("Expected container name %s, but found %s", containerName, container.Name)
	}

	if container.LivenessProbe == nil {
		t.Errorf("expected LivenessProbe but got none")
	}

	if container.ReadinessProbe == nil {
		t.Errorf("expected ReadinessProbe but got none")
	}

	if container.StartupProbe == nil {
		t.Errorf("expected StartupProbe but got none")
	}
}

func TestCreateInitConfigReloaderEnableProbes(t *testing.T) {
	reloaderConfigCopy := reloaderConfig
	reloaderConfigCopy.EnableProbes = true
	initContainerName := "init-config-reloader"
	var container = CreateConfigReloader(
		initContainerName,
		ReloaderConfig(reloaderConfigCopy),
		ReloaderURL(url.URL{
			Scheme: "http",
			Host:   "localhost:9093",
			Path:   "/-/reload",
		}),
		InitContainer(),
	)

	if container.Name != "init-config-reloader" {
		t.Errorf("Expected container name %s, but found %s", initContainerName, container.Name)
	}

	if container.LivenessProbe != nil {
		t.Errorf("expected no LivenessProbe but got %v", container.LivenessProbe)
	}

	if container.ReadinessProbe != nil {
		t.Errorf("expected no ReadinessProbe but got %v", container.ReadinessProbe)
	}

	if container.StartupProbe != nil {
		t.Errorf("expected no StartupProbe but got %v", container.StartupProbe)
	}
}

func TestCreateInitConfigReloader(t *testing.T) {
	initContainerName := "init-config-reloader"
	expectedImagePullPolicy := v1.PullAlways
	var container = CreateConfigReloader(
		initContainerName,
		ReloaderConfig(reloaderConfig),
		InitContainer(),
		ImagePullPolicy(v1.PullAlways),
	)

	assert.NotContains(t, container.Env, v1.EnvVar{
		Name: NodeNameEnvVar,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
		},
	})

	if container.Name != "init-config-reloader" {
		t.Errorf("Expected container name %s, but found %s", initContainerName, container.Name)
	}

	if !slices.Contains(container.Args, "--watch-interval=0") {
		t.Errorf("Expected '--watch-interval=0' does not exist in container arguments")
	}

	if container.Ports[0].ContainerPort != initConfigReloaderPort {
		t.Errorf("Expected port number to be %d, got %d", initConfigReloaderPort, container.Ports[0].ContainerPort)
	}

	if !slices.Contains(container.Args, "--listen-address=:8081") {
		t.Errorf("Expected '--listen-address=:8081' not found in %s", container.Args)
	}

	if container.ImagePullPolicy != expectedImagePullPolicy {
		t.Errorf("Expected imagePullPolicy %s, but found %s", expectedImagePullPolicy, container.ImagePullPolicy)
	}

	if container.LivenessProbe != nil {
		t.Errorf("expected no LivenessProbe but got %v", container.LivenessProbe)
	}

	if container.ReadinessProbe != nil {
		t.Errorf("expected no ReadinessProbe but got %v", container.ReadinessProbe)
	}

	if container.StartupProbe != nil {
		t.Errorf("expected no StartupProbe but got %v", container.StartupProbe)
	}
}

func TestCreateConfigReloader(t *testing.T) {
	containerName := "config-reloader"
	logFormat := "logFormat"
	logLevel := "logLevel"
	configFile := "configFile"
	webConfigFile := "webConfigFile"
	configEnvsubstFile := "configEnvsubstFile"
	watchedDirectories := []string{"directory1", "directory2"}
	shard := int32(1)
	expectedImagePullPolicy := v1.PullAlways
	var container = CreateConfigReloader(
		containerName,
		ReloaderConfig(reloaderConfig),
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
		WebConfigFile(webConfigFile),
		Shard(shard),
		ImagePullPolicy(expectedImagePullPolicy),
	)

	if container.Name != "config-reloader" {
		t.Errorf("Expected container name %s, but found %s", containerName, container.Name)
	}
	if !slices.Contains(container.Args, "--listen-address=localhost:8080") {
		t.Errorf("Expected '--listen-address=localhost:8080' not found in %s", container.Args)
	}
	if !slices.Contains(container.Args, "--reload-url=http://localhost:9093/-/reload") {
		t.Errorf("Expected '--reload-url=http://localhost:9093/-/reload' not found in %s", container.Args)
	}
	if !slices.Contains(container.Args, "--log-level=logLevel") {
		t.Errorf("Expected '--log-level=%s' not found in %s", logLevel, container.Args)
	}
	if !slices.Contains(container.Args, "--log-format=logFormat") {
		t.Errorf("Expected '--log-format=%s' not found in %s", logFormat, container.Args)
	}
	if !slices.Contains(container.Args, "--config-file=configFile") {
		t.Errorf("Expected '--config-file=%s' not found in %s", configFile, container.Args)
	}
	if !slices.Contains(container.Args, "--config-envsubst-file=configEnvsubstFile") {
		t.Errorf("Expected '--config-envsubst-file=%s' not found in %s", configEnvsubstFile, container.Args)
	}
	if !slices.Contains(container.Args, "--web-config-file=webConfigFile") {
		t.Errorf("Expected '--web-config-file=%s' not found in %s", webConfigFile, container.Args)
	}
	for _, dir := range watchedDirectories {
		if !slices.Contains(container.Args, fmt.Sprintf("--watched-dir=%s", dir)) {
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

	if container.ImagePullPolicy != expectedImagePullPolicy {
		t.Errorf("Expected imagePullPolicy %s, but found %s", expectedImagePullPolicy, container.ImagePullPolicy)
	}

	if container.LivenessProbe != nil {
		t.Errorf("expected no LivenessProbe but got %v", container.LivenessProbe)
	}

	if container.ReadinessProbe != nil {
		t.Errorf("expected no ReadinessProbe but got %v", container.ReadinessProbe)
	}

	if container.StartupProbe != nil {
		t.Errorf("expected no StartupProbe but got %v", container.StartupProbe)
	}
}

func TestCreateConfigReloaderForDaemonSet(t *testing.T) {
	var container = CreateConfigReloader(
		"config-reloader",
		WithDaemonSetMode(),
	)

	assert.Contains(t, container.Env, v1.EnvVar{
		Name: NodeNameEnvVar,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
		},
	})

	assert.Contains(t, container.Env, v1.EnvVar{
		Name:  ShardEnvVar,
		Value: strconv.Itoa(0),
	})
}
