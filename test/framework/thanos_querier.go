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

package framework

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func MakeThanosQuerier(endpoints ...string) (*appsv1.Deployment, error) {
	d, err := MakeDeployment("../../example/thanos/query-deployment.yaml")
	if err != nil {
		return nil, err
	}

	d.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("quay.io/thanos/thanos:%s", operator.DefaultThanosVersion)

	var args []string
	for _, arg := range d.Spec.Template.Spec.Containers[0].Args {
		if strings.HasPrefix(arg, "--endpoint=") {
			continue
		}
		args = append(args, arg)
	}
	for _, endpoint := range endpoints {
		args = append(args, fmt.Sprintf("--endpoint=%s", endpoint))
	}
	d.Spec.Template.Spec.Containers[0].Args = args

	return d, nil
}
