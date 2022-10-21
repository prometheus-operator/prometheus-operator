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

// ------------------------------------------------------------------
/*
// MakePrometheusStatefulSet creates StatefulSet from given Prometheus type object and Config
func MakePrometheusStatefulSet(logger log.Logger, name string, pt PrometheusType,
	config *Config, ruleConfigMapNames []string, inputHash string, shard int32,
	tlsAssetSecrets []string,
) (*appsv1.StatefulSet, error) {
	// p is passed in by value, not by reference. But p contains references like
	// to annotation map, that do not get copied on function invocation. Ensure to
	// prevent side effects before editing p by creating a deep copy. For more
	// details see https://github.com/prometheus-operator/prometheus-operator/issues/1659.
	pt = pt.Duplicate()
	nc := pt.GetNomenclator()

	version, err := pt.GetVersion(DefaultPrometheusVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse prometheus version")
	}

	pt.SetDefaultPortname(DefaultPrometheusPortName)

	replicas := pt.GetReplicas()
	if replicas == nil {
		pt.SetDefaultReplicas(&minReplicas)
	} else if *replicas < 0 {
		pt.SetDefaultReplicas(&int32Zero)
	}

	resources := pt.GetResources()
	requests := v1.ResourceList{}
	if resources.Requests != nil {
		requests = resources.Requests
	}
	_, memoryRequestFound := requests[v1.ResourceMemory]
	memoryLimit, memoryLimitFound := resources.Limits[v1.ResourceMemory]
	if !memoryRequestFound && version.Major == 1 {
		defaultMemoryRequest := resource.MustParse(DefaultMemoryRequestValue)
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		// If limit is given and smaller or equal to 2Gi, then set memory
		// request to the given limit. This is necessary as if limit < request,
		// then a Pod is not schedulable.
		if memoryLimitFound && compareResult <= 0 {
			requests[v1.ResourceMemory] = memoryLimit
		} else {
			requests[v1.ResourceMemory] = defaultMemoryRequest
		}
	}
	pt.SetResourceRequests(requests)

	spec, err := MakePrometheusStatefulsetSpec(pt, logger, config, shard, ruleConfigMapNames, tlsAssetSecrets)
	if err != nil {
		return nil, errors.Wrap(err, "make StatefulSet spec")
	}

	// do not transfer kubectl annotations to the statefulset so it is not
	// pruned by kubectl
	objectMeta := pt.GetObjectMeta()
	annotations := make(map[string]string)
	for key, value := range objectMeta.Annotations {
		if !strings.HasPrefix(key, "kubectl.kubernetes.io/") {
			annotations[key] = value
		}
	}
	labels := make(map[string]string)
	for key, value := range objectMeta.Labels {
		labels[key] = value
	}
	labels[shardLabelName] = fmt.Sprintf("%d", shard)
	labels[nc.NameLabelName()] = nc.BaseName()

	ownr := pt.GetOwnerReference()
	ownr.BlockOwnerDeletion = &boolTrue
	ownr.Controller = &boolTrue
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          config.Labels.Merge(labels),
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{ownr},
		},
		Spec: *spec,
	}

	if statefulset.ObjectMeta.Annotations == nil {
		statefulset.ObjectMeta.Annotations = map[string]string{
			StatefulSetInputHashName: inputHash,
		}
	} else {
		statefulset.ObjectMeta.Annotations[StatefulSetInputHashName] = inputHash
	}

	imagePullSecrets := pt.GetImagePullSecrets()
	if len(imagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = imagePullSecrets
	}
	storageSpec := pt.GetStorageSpec()
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else if storageSpec.Ephemeral != nil {
		ephemeral := storageSpec.Ephemeral
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: nc.VolumeName(),
			VolumeSource: v1.VolumeSource{
				Ephemeral: ephemeral,
			},
		})
	} else {
		pvcTemplate := MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = nc.VolumeName()
		}
		if storageSpec.VolumeClaimTemplate.Spec.AccessModes == nil {
			pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		} else {
			pvcTemplate.Spec.AccessModes = storageSpec.VolumeClaimTemplate.Spec.AccessModes
		}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, *pvcTemplate)
	}

	statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, pt.GetVolumes()...)

	return statefulset, nil
}*/
