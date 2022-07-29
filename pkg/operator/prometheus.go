// Copyright 2022 The prometheus-operator Authors
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
	"path"
	"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	prometheusServerPrefix = "prometheus"
	//prometheusAgentPrefix = "prometheusagent"
)

// PrometheusType defines interface for both modes Prometheus is operating
type PrometheusType interface {
	// GetNomenclator returns object used for naming other objects unified way
	GetNomenclator() *Nomenclator
	// GetVersion returns object's spec version parsed by semver. If object's version is nil parses given default value/
	// If default value is "" returns nil
	GetVersion(string) (*semver.Version, error)
	// GetImage returns object's spec image
	GetImage() *string
	// GetDeprecatedImageInfo returns deprecated image-related spec fields such as base image, tag and sha
	GetDeprecatedImageInfo() (string, string, string)
	// GetQuery returns object's QuerySpec
	GetQuery() *promv1.QuerySpec
	// GetStorageSpec returns object's StorageSpec
	GetStorageSpec() *promv1.StorageSpec
	// GetVolumeMounts returns slice of object's VolumeMounts
	GetVolumeMounts() []v1.VolumeMount
	// GetVolumes returns slice of objects Volumes
	GetVolumes() []v1.Volume
	// GetWebSpec returns object's WebSpec
	GetWebSpec() *promv1.PrometheusWebSpec
	// GetSecrets returns slice of object's secrets
	GetSecrets() []string
	// GetConfigMaps returns slice of object's config maps
	GetConfigMaps() []string
	// GetPodMetadata returns object's EmbeddedMetadata
	GetPodMetadata() *promv1.EmbeddedObjectMetadata
	// GetThanosSpec returns object's ThanosSpec
	GetThanosSpec() *promv1.ThanosSpec
	// GetLoggerInfo returns setting for a logger, such as logging level and record format
	GetLoggerInfo() (string, string)
	// GetInitContainers returns slice of object's init Containers
	GetInitContainers() []v1.Container
	// GetContainers returns slice of object's Containers
	GetContainers() []v1.Container
	// GetResources returns object's ResourceRequirements
	GetResources() v1.ResourceRequirements
	// GetReplicas returns object's replica count
	GetReplicas() *int32
	// GetMinReadySeconds return's object's ...
	GetMinReadySeconds() *uint32
	// GetAdditionalArgs return's arguments which should be added to object's container command
	GetAdditionalArgs() []promv1.Argument
	// GetObjectMeta returns object's ObjectMeta
	GetObjectMeta() metav1.ObjectMeta
	// GetOwnerReference returns object's OwnerReference
	GetOwnerReference() metav1.OwnerReference
	// GetImagePullSecrets returns object's LocalObjectReference
	GetImagePullSecrets() []v1.LocalObjectReference
	// DisableCompaction returns true if DisableCompaction is true in spec or if Thanos object storage is defined
	DisableCompaction() bool
	// ListensOn returns object's PortName or "localhost" if ListenLocal is true
	ListensOn() string
	// UsesDefaultQueryLogVolume returns true if QueryLogFile is set to current directory (eg. ".")
	UsesDefaultQueryLogVolume() bool
	// MakeCommandArgs returns map of command line arguments for object's container, slice of warnings raised during generation process and error
	MakeCommandArgs() (map[string]string, []string, error)
	// MakePodSpec builds PodSpec from object's spec fields
	MakePodSpec(containers, initContainers []v1.Container, volumes []v1.Volume) v1.PodSpec
	// Duplicate is analogy to DeepCopy
	Duplicate() PrometheusType
	// SetDefaultPortname sets given value to Spec.Portname
	SetDefaultPortname(defaultPortName string)
	// SetDefaultReplicas sets given value to Spec.Replicas
	SetDefaultReplicas(minReplicas *int32)
	// SetResourceRequests sets given ResourceList as Spec.Resources.Requests
	SetResourceRequests(requests v1.ResourceList)
}

// PrometheusServer objects are used to run Prometheus instances in server mode
type PrometheusServer struct {
	*promv1.Prometheus
}

// PrometheusServer objects are used to run Prometheus instances in agent mode
/*type PrometheusAgent struct {
	*promv1a1.PrometheusAgent
}*/

//----------------------------- server mode getters ----------------------------

// GetNameNomenclator implements PrometheusType interface
func (p PrometheusServer) GetNomenclator() *Nomenclator {
	return NewNomenclator(p.Kind, prometheusServerPrefix, p.Name, p.Spec.Shards)
}

// GetImage implements PrometheusType interface
func (p PrometheusServer) GetImage() *string {
	return p.Spec.Image
}

// GetDeprecatedImageInfo implements PrometheusType interface
func (p PrometheusServer) GetDeprecatedImageInfo() (string, string, string) {
	return p.Spec.BaseImage, p.Spec.Tag, p.Spec.SHA
}

// GetVersion implements PrometheusType interface
func (p PrometheusServer) GetVersion(defValue string) (*semver.Version, error) {
	vstr := p.Spec.Version
	if strings.TrimSpace(vstr) == "" {
		if defValue == "" {
			return nil, nil
		}
		vstr = defValue
	}

	version, err := semver.ParseTolerant(vstr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Prometheus version")
	}
	return &version, nil
}

// GetQuery implements PrometheusType interface
func (p PrometheusServer) GetQuery() *promv1.QuerySpec {
	return p.Spec.Query
}

// GetStorageSpec implements PrometheusType interface
func (p PrometheusServer) GetStorageSpec() *promv1.StorageSpec {
	return p.Spec.Storage
}

// GetVolumeMounts implements PrometheusType interface
func (p PrometheusServer) GetVolumeMounts() []v1.VolumeMount {
	return p.Spec.VolumeMounts
}

// GetVolumes implements PrometheusType interface
func (p PrometheusServer) GetVolumes() []v1.Volume {
	return p.Spec.Volumes
}

// GetWebSpec implements PrometheusType interface
func (p PrometheusServer) GetWebSpec() *promv1.PrometheusWebSpec {
	return p.Spec.Web
}

// GetSecrets implements PrometheusType interface
func (p PrometheusServer) GetSecrets() []string {
	return p.Spec.Secrets
}

// GetConfigMaps implements PrometheusType interface
func (p PrometheusServer) GetConfigMaps() []string {
	return p.Spec.ConfigMaps
}

// GetPodMetadata implements PrometheusType interface
func (p PrometheusServer) GetPodMetadata() *promv1.EmbeddedObjectMetadata {
	return p.Spec.PodMetadata
}

// GetThanosSpec implements PrometheusType interface
func (p PrometheusServer) GetThanosSpec() *promv1.ThanosSpec {
	return p.Spec.Thanos
}

// GetLoggerInfo implements PrometheusType interface
func (p PrometheusServer) GetLoggerInfo() (string, string) {
	return p.Spec.LogLevel, p.Spec.LogFormat
}

// GetInitContainers implements PrometheusType interface
func (p PrometheusServer) GetInitContainers() []v1.Container {
	return p.Spec.InitContainers
}

// GetContainers implements PrometheusType interface
func (p PrometheusServer) GetContainers() []v1.Container {
	return p.Spec.Containers
}

// GetResources implements PrometheusType interface
func (p PrometheusServer) GetResources() v1.ResourceRequirements {
	return p.Spec.Resources
}

// GetReplicas implements PrometheusType interface
func (p PrometheusServer) GetReplicas() *int32 {
	return p.Spec.Replicas
}

// GetMinReadySeconds implements PrometheusType interface
func (p PrometheusServer) GetMinReadySeconds() *uint32 {
	return p.Spec.MinReadySeconds
}

// GetAdditionalArgs implements PrometheusType interface
func (p PrometheusServer) GetAdditionalArgs() []promv1.Argument {
	return p.Spec.AdditionalArgs
}

// GetObjectMeta implements PrometheusType interface
func (p PrometheusServer) GetObjectMeta() metav1.ObjectMeta {
	return p.ObjectMeta
}

// GetOwnerReference implements PrometheusType interface
func (p PrometheusServer) GetOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: p.APIVersion,
		Kind:       p.Kind,
		Name:       p.Name,
		UID:        p.UID,
	}
}

// GetImagePullSecrets implements PrometheusType interface
func (p PrometheusServer) GetImagePullSecrets() []v1.LocalObjectReference {
	return p.Spec.ImagePullSecrets
}

// DisableCompaction implements PrometheusType interface
func (p PrometheusServer) DisableCompaction() bool {
	if p.Spec.Thanos != nil {
		if p.Spec.Thanos.ObjectStorageConfig != nil || p.Spec.Thanos.ObjectStorageConfigFile != nil {
			// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/ we have to turn off compaction of Prometheus
			// to avoid races during upload, if the uploads are configured.
			return true
		}
	}
	return p.Spec.DisableCompaction
}

// ListensOn implements PrometheusType interface
func (p PrometheusServer) ListensOn() string {
	if p.Spec.ListenLocal {
		return "localhost"
	}
	return p.Spec.PortName
}

// UsesDefaultQueryLogVolume implements PrometheusType interface
func (p PrometheusServer) UsesDefaultQueryLogVolume() bool {
	return p.Spec.QueryLogFile != "" && filepath.Dir(p.Spec.QueryLogFile) == "."
}

// MakeCommandArgs returns map of command line arguments for Prometheus server mode
func (p PrometheusServer) MakeCommandArgs() (map[string]string, []string, error) {
	warns := []string{}
	args := map[string]string{}

	version, err := p.GetVersion(DefaultPrometheusVersion)
	if err != nil {
		return args, warns, err
	}

	args["config.file"] = path.Join(PrometheusConfOutDir, PrometheusConfEnvSubstFilename)

	args["storage.tsdb.path"] = PrometheusStorageDir
	retentionTimeFlag := "storage.tsdb.retention"
	if version.GTE(semver.MustParse("2.7.0")) {
		retentionTimeFlag = "storage.tsdb.retention.time"
		if p.Spec.Retention == "" && p.Spec.RetentionSize == "" {
			args[retentionTimeFlag] = DefaultRetention
		} else {
			if p.Spec.Retention != "" {
				args[retentionTimeFlag] = string(p.Spec.Retention)
			}

			if p.Spec.RetentionSize != "" {
				args["storage.tsdb.retention.size"] = string(p.Spec.RetentionSize)
			}
		}
	} else {
		if p.Spec.Retention == "" {
			args[retentionTimeFlag] = DefaultRetention
		} else {
			args[retentionTimeFlag] = string(p.Spec.Retention)
		}
	}

	if p.Spec.Query != nil {
		if p.Spec.Query.LookbackDelta != nil {
			args["query.lookback-delta"] = *p.Spec.Query.LookbackDelta
		}

		if p.Spec.Query.MaxConcurrency != nil {
			if *p.Spec.Query.MaxConcurrency < 1 {
				p.Spec.Query.MaxConcurrency = &defaultMaxConcurrency
			}
			args["query.max-concurrency"] = fmt.Sprintf("%d", *p.Spec.Query.MaxConcurrency)
		}
		if p.Spec.Query.Timeout != nil {
			args["query.timeout"] = string(*p.Spec.Query.Timeout)
		}
		if version.Minor >= 5 {
			if p.Spec.Query.MaxSamples != nil {
				args["query.max-samples"] = fmt.Sprintf("%d", *p.Spec.Query.MaxSamples)
			}
		}
	}

	if version.Minor >= 4 {
		if p.Spec.Rules.Alert.ForOutageTolerance != "" {
			args["rules.alert.for-outage-tolerance"] = p.Spec.Rules.Alert.ForOutageTolerance
		}
		if p.Spec.Rules.Alert.ForGracePeriod != "" {
			args["rules.alert.for-grace-period"] = p.Spec.Rules.Alert.ForGracePeriod
		}
		if p.Spec.Rules.Alert.ResendDelay != "" {
			args["rules.alert.resend-delay"] = p.Spec.Rules.Alert.ResendDelay
		}
	}

	args["web.config.file"] = path.Join(WebConfigDir, WebConfigFilename)
	args["web.console.templates"] = WebConsoleTemplatesDir
	args["web.console.libraries"] = WebConsoleLibraryDir
	args["web.enable-lifecycle"] = ""
	if p.Spec.Web != nil {
		// TODO(simonpasquier): check that the Prometheus version supports the flag.
		if p.Spec.Web.PageTitle != nil {
			args["web.page-title"] = *p.Spec.Web.PageTitle
		}
	}

	if p.Spec.EnableAdminAPI {
		args["web.enable-admin-api"] = ""
	}

	if p.Spec.EnableRemoteWriteReceiver {
		if version.GTE(semver.MustParse("2.33.0")) {
			args["web.enable-remote-write-receiver"] = ""
		} else {
			msg := "ignoring 'enableRemoteWriteReceiver' supported by Prometheus v v2.33.0+"
			warns = append(warns, msg)
		}
	}

	if len(p.Spec.EnableFeatures) > 0 {
		args["enable-feature"] = strings.Join(p.Spec.EnableFeatures[:], ",")
	}

	if p.Spec.ExternalURL != "" {
		args["web.external-url"] = p.Spec.ExternalURL
	}

	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}
	args["web.route-prefix"] = webRoutePrefix

	if p.Spec.LogLevel != "" && p.Spec.LogLevel != "info" {
		args["log.level"] = p.Spec.LogLevel
	}
	if version.GTE(semver.MustParse("2.6.0")) {
		if p.Spec.LogFormat != "" && p.Spec.LogFormat != "logfmt" {
			args["log.format"] = p.Spec.LogFormat
		}
	}

	if version.GTE(semver.MustParse("2.11.0")) && p.Spec.WALCompression != nil {
		if *p.Spec.WALCompression {
			args["storage.tsdb.wal-compression"] = ""
		} else {
			args["no-storage.tsdb.wal-compression"] = ""
		}
	}

	if version.GTE(semver.MustParse("2.8.0")) && p.Spec.AllowOverlappingBlocks {
		args["storage.tsdb.allow-overlapping-blocks"] = ""
	}

	if p.Spec.ListenLocal {
		args["web.listen-address"] = "127.0.0.1:9090"
	}

	if p.DisableCompaction() {
		args["storage.tsdb.max-block-duration"] = "2h"
		args["storage.tsdb.min-block-duration"] = "2h"
	}

	return args, warns, nil
}

// MakePodSpec implements PrometheusType interface
func (p PrometheusServer) MakePodSpec(containers, initContainers []v1.Container, volumes []v1.Volume) v1.PodSpec {
	boolTrue := true
	terminationGracePeriod := int64(600)
	return v1.PodSpec{
		Containers:                    containers,
		InitContainers:                initContainers,
		SecurityContext:               p.Spec.SecurityContext,
		ServiceAccountName:            p.Spec.ServiceAccountName,
		AutomountServiceAccountToken:  &boolTrue,
		NodeSelector:                  p.Spec.NodeSelector,
		PriorityClassName:             p.Spec.PriorityClassName,
		TerminationGracePeriodSeconds: &terminationGracePeriod,
		Volumes:                       volumes,
		Tolerations:                   p.Spec.Tolerations,
		Affinity:                      p.Spec.Affinity,
		TopologySpreadConstraints:     p.Spec.TopologySpreadConstraints,
		HostAliases:                   MakeHostAliases(p.Spec.HostAliases),
	}
}

// Duplicate implements PrometheusType interface
func (p PrometheusServer) Duplicate() PrometheusType {
	return PrometheusServer{p.DeepCopy()}
}

// SetDefaultPortname implements PrometheusType interface
func (p PrometheusServer) SetDefaultPortname(defaultPortName string) {
	if p.Spec.PortName == "" {
		p.Spec.PortName = defaultPortName
	}
}

// SetDefaultReplicas implements PrometheusType interface
func (p PrometheusServer) SetDefaultReplicas(minReplicas *int32) {
	if p.Spec.Replicas != nil {
		p.Spec.Replicas = minReplicas
	}
}

// SetResourceRequests implements PrometheusType interface
func (p PrometheusServer) SetResourceRequests(requests v1.ResourceList) {
	p.Spec.Resources.Requests = requests
}

//----------------------------- agent mode getters -----------------------------
// TBD
