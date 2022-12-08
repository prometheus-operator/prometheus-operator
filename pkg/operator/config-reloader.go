// Copyright 2020 The prometheus-operator Authors
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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const configReloaderPort = 8080

// ConfigReloader contains the options to configure
// a config-reloader container
type ConfigReloader struct {
	name               string
	config             ReloaderConfig
	configFile         string
	configEnvsubstFile string
	imagePullPolicy    v1.PullPolicy
	listenLocal        bool
	localHost          string
	logFormat          string
	logLevel           string
	reloadURL          url.URL
	runOnce            bool
	shard              *int32
	volumeMounts       []v1.VolumeMount
	watchedDirectories []string
}

type ReloaderOption = func(*ConfigReloader)

// ReloaderRunOnce sets the runOnce option for the config-reloader container
func ReloaderRunOnce() ReloaderOption {
	return func(c *ConfigReloader) {
		c.runOnce = true
	}
}

// WatchedDirectories sets the watchedDirectories option for the config-reloader container
func WatchedDirectories(watchedDirectories []string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.watchedDirectories = watchedDirectories
	}
}

// ConfigFile sets the configFile option for the config-reloader container
func ConfigFile(configFile string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.configFile = configFile
	}
}

// ConfigEnvsubstFile sets the configEnvsubstFile option for the config-reloader container
func ConfigEnvsubstFile(configEnvsubstFile string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.configEnvsubstFile = configEnvsubstFile
	}
}

// ReloaderResources sets the config option for the config-reloader container
func ReloaderResources(rc ReloaderConfig) ReloaderOption {
	return func(c *ConfigReloader) {
		c.config = rc
	}
}

// ReloaderURL sets the reloaderURL option for the config-reloader container
func ReloaderURL(reloadURL url.URL) ReloaderOption {
	return func(c *ConfigReloader) {
		c.reloadURL = reloadURL
	}
}

// ListenLocal sets the listenLocal option for the config-reloader container
func ListenLocal(listenLocal bool) ReloaderOption {
	return func(c *ConfigReloader) {
		c.listenLocal = listenLocal
	}
}

// LocalHost sets the localHost option for the config-reloader container
func LocalHost(localHost string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.localHost = localHost
	}
}

// LogFormat sets the logFormat option for the config-reloader container
func LogFormat(logFormat string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.logFormat = logFormat
	}
}

// LogLevel sets the logLevel option for the config-reloader container\
func LogLevel(logLevel string) ReloaderOption {
	return func(c *ConfigReloader) {
		c.logLevel = logLevel
	}
}

// VolumeMounts sets the volumeMounts option for the config-reloader container
func VolumeMounts(mounts []v1.VolumeMount) ReloaderOption {
	return func(c *ConfigReloader) {
		c.volumeMounts = mounts
	}
}

// Shard sets the shard option for the config-reloader container
func Shard(shard int32) ReloaderOption {
	return func(c *ConfigReloader) {
		c.shard = &shard
	}
}

// ImagePullPolicy sets the imagePullPolicy option for the config-reloader container
func ImagePullPolicy(imagePullPolicy v1.PullPolicy) ReloaderOption {
	return func(c *ConfigReloader) {
		c.imagePullPolicy = imagePullPolicy
	}
}

// CreateConfigReloader returns the definition of the config-reloader
// container.
func CreateConfigReloader(name string, options ...ReloaderOption) v1.Container {
	configReloader := ConfigReloader{name: name}

	for _, option := range options {
		option(&configReloader)
	}

	var (
		args    = make([]string, 0)
		envVars = []v1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			},
		}
		ports []v1.ContainerPort
	)

	if configReloader.runOnce {
		args = append(args, fmt.Sprintf("--watch-interval=%d", 0))
	}

	if configReloader.listenLocal {
		args = append(args, fmt.Sprintf("--listen-address=%s:%d", configReloader.localHost, configReloaderPort))
	} else {
		args = append(args, fmt.Sprintf("--listen-address=:%d", configReloaderPort))
		ports = append(
			ports,
			v1.ContainerPort{
				Name:          "reloader-web",
				ContainerPort: configReloaderPort,
				Protocol:      v1.ProtocolTCP,
			},
		)
	}

	if len(configReloader.reloadURL.String()) > 0 {
		args = append(args, fmt.Sprintf("--reload-url=%s", configReloader.reloadURL.String()))
	}

	if len(configReloader.configFile) > 0 {
		args = append(args, fmt.Sprintf("--config-file=%s", configReloader.configFile))
	}

	if len(configReloader.configEnvsubstFile) > 0 {
		args = append(args, fmt.Sprintf("--config-envsubst-file=%s", configReloader.configEnvsubstFile))
	}

	if len(configReloader.watchedDirectories) > 0 {
		for _, directory := range configReloader.watchedDirectories {
			args = append(args, fmt.Sprintf("--watched-dir=%s", directory))
		}
	}

	if configReloader.logLevel != "" && configReloader.logLevel != "info" {
		args = append(args, fmt.Sprintf("--log-level=%s", configReloader.logLevel))
	}

	if configReloader.logFormat != "" && configReloader.logFormat != "logfmt" {
		args = append(args, fmt.Sprintf("--log-format=%s", configReloader.logFormat))
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

	if configReloader.shard != nil {
		envVars = append(envVars, v1.EnvVar{
			Name:  "SHARD",
			Value: strconv.Itoa(int(*configReloader.shard)),
		})
	}

	boolFalse := false
	boolTrue := true
	return v1.Container{
		Name:                     name,
		Image:                    configReloader.config.Image,
		ImagePullPolicy:          configReloader.imagePullPolicy,
		TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		Env:                      envVars,
		Command:                  []string{"/bin/prometheus-config-reloader"},
		Args:                     args,
		Ports:                    ports,
		VolumeMounts:             configReloader.volumeMounts,
		Resources:                resources,
		SecurityContext: &v1.SecurityContext{
			AllowPrivilegeEscalation: &boolFalse,
			ReadOnlyRootFilesystem:   &boolTrue,
			Capabilities: &v1.Capabilities{
				Drop: []v1.Capability{"ALL"},
			},
		},
	}
}
