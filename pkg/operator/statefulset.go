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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const configReloaderPort = 8080

type ConfigReloader struct {
	name           string
	config         ReloaderConfig
	reloadURL      url.URL
	listenLocal    bool
	localHost      string
	logFormat      string
	logLevel       string
	additionalArgs []string
	volumeMounts   []v1.VolumeMount
	shard          int32
}

type option = func(*ConfigReloader)

func ReloaderRunOnce() option {
	return func(c *ConfigReloader) {
		c.additionalArgs = append(c.additionalArgs, fmt.Sprintf("--watch-interval=%d", 0))
	}
}

func ListenAddress() option {
	return func(c *ConfigReloader) {
		var listenAddress string
		if c.listenLocal {
			listenAddress = c.localHost
		}
		c.additionalArgs = append(c.additionalArgs, fmt.Sprintf("--listen-address=%s:%d", listenAddress, configReloaderPort))
	}
}

func ReloadURL() option {
	return func(c *ConfigReloader) {
		c.additionalArgs = append(c.additionalArgs, fmt.Sprintf("--reload-url=%s", c.reloadURL.String()))
	}
}

func ReloaderResources(rc ReloaderConfig) option {
	return func(c *ConfigReloader) {
		c.config = rc
	}
}

func ReloaderURL(reloadURL url.URL) option {
	return func(c *ConfigReloader) {
		c.reloadURL = reloadURL
	}
}

func ListenLocal(listenLocal bool) option {
	return func(c *ConfigReloader) {
		c.listenLocal = listenLocal
	}
}

func LocalHost(localHost string) option {
	return func(c *ConfigReloader) {
		c.localHost = localHost
	}
}

func LogFormat(logFormat string) option {
	return func(c *ConfigReloader) {
		c.logFormat = logFormat
	}
}

func LogLevel(logLevel string) option {
	return func(c *ConfigReloader) {
		c.logLevel = logLevel
	}
}

func AdditionalArgs(args []string) option {
	return func(c *ConfigReloader) {
		c.additionalArgs = append(c.additionalArgs, args...)
	}
}

func VolumeMount(mounts []v1.VolumeMount) option {
	return func(c *ConfigReloader) {
		c.volumeMounts = mounts
	}
}

func Shard(shard int32) option {
	return func(c *ConfigReloader) {
		c.shard = shard
	}
}

// CreateConfigReloader returns the definition of the config-reloader
// container. No shard environment variable will be passed to the reloader
// container if `-1` is passed to the shards parameter.
func CreateConfigReloader(name string, options ...option) v1.Container {
	configReloader := ConfigReloader{name: name}

	for _, option := range options {
		option(&configReloader)
	}

	var (
		ports []v1.ContainerPort
		args  = make([]string, 0)
	)

	ports = append(
		ports,
		v1.ContainerPort{
			Name:          "reloader-web",
			ContainerPort: configReloaderPort,
			Protocol:      v1.ProtocolTCP,
		},
	)

	if configReloader.logLevel != "" && configReloader.logLevel != "info" {
		args = append(args, fmt.Sprintf("--log-level=%s", configReloader.logLevel))
	}

	if configReloader.logFormat != "" && configReloader.logFormat != "logfmt" {
		args = append(args, fmt.Sprintf("--log-format=%s", configReloader.logFormat))
	}

	for i := range configReloader.additionalArgs {
		args = append(args, configReloader.additionalArgs[i])
	}

	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if configReloader.config.CPURequest != "0" {
		resources.Requests[v1.ResourceCPU] = resource.MustParse(configReloader.config.CPURequest)
	}
	if configReloader.config.CPULimit != "0" {
		resources.Limits[v1.ResourceCPU] = resource.MustParse(configReloader.config.CPULimit)
	}
	if configReloader.config.MemoryRequest != "0" {
		resources.Requests[v1.ResourceMemory] = resource.MustParse(configReloader.config.MemoryRequest)
	}
	if configReloader.config.MemoryLimit != "0" {
		resources.Limits[v1.ResourceMemory] = resource.MustParse(configReloader.config.MemoryLimit)
	}

	return v1.Container{
		Name:                     name,
		Image:                    configReloader.config.Image,
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		Env: []v1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			},
			{
				Name:  "SHARD",
				Value: fmt.Sprint(configReloader.shard),
			},
		},
		Command:      []string{"/bin/prometheus-config-reloader"},
		Args:         args,
		Ports:        ports,
		VolumeMounts: configReloader.volumeMounts,
		Resources:    resources,
	}
}
