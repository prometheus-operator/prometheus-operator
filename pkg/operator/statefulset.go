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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const configReloaderPort = 8080

// CreateConfigReloader returns the definition of the config-reloader
// container. No shard environment variable will be passed to the reloader
// container if `-1` is passed to the shards parameter.
func CreateConfigReloader(
	config ReloaderConfig,
	reloadURL url.URL,
	listenLocal bool,
	localHost string,
	logFormat, logLevel string,
	additionalArgs []string,
	volumeMounts []v1.VolumeMount,
	shard int32,
) v1.Container {
	var (
		ports []v1.ContainerPort
		args  = make([]string, 0, len(additionalArgs))
	)

	var listenAddress string
	if listenLocal {
		listenAddress = localHost
	} else {
		ports = append(
			ports,
			v1.ContainerPort{
				Name:          "reloader-web",
				ContainerPort: configReloaderPort,
				Protocol:      v1.ProtocolTCP,
			},
		)
	}
	args = append(args, fmt.Sprintf("--listen-address=%s:%d", listenAddress, configReloaderPort))

	args = append(args, fmt.Sprintf("--reload-url=%s", reloadURL.String()))

	if logLevel != "" && logLevel != "info" {
		args = append(args, fmt.Sprintf("--log-level=%s", logLevel))
	}

	if logFormat != "" && logFormat != "logfmt" {
		args = append(args, fmt.Sprintf("--log-format=%s", logFormat))
	}

	for i := range additionalArgs {
		args = append(args, additionalArgs[i])
	}

	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if config.CPURequest != "0" {
		resources.Requests[v1.ResourceCPU] = resource.MustParse(config.CPURequest)
	}
	if config.CPULimit != "0" {
		resources.Limits[v1.ResourceCPU] = resource.MustParse(config.CPULimit)
	}
	if config.MemoryRequest != "0" {
		resources.Requests[v1.ResourceMemory] = resource.MustParse(config.MemoryRequest)
	}
	if config.MemoryLimit != "0" {
		resources.Limits[v1.ResourceMemory] = resource.MustParse(config.MemoryLimit)
	}

	return v1.Container{
		Name:                     "config-reloader",
		Image:                    config.Image,
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
				Value: fmt.Sprint(shard),
			},
		},
		Command:      []string{"/bin/prometheus-config-reloader"},
		Args:         args,
		Ports:        ports,
		VolumeMounts: volumeMounts,
		Resources:    resources,
	}
}
