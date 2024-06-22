// Copyright 2023 The prometheus-operator Authors
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

package prometheusagent

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prompkg "github.com/prometheus-operator/prometheus-operator/pkg/prometheus"
)

func buildAgentArgs(
	cpf monitoringv1.CommonPrometheusFields,
	cg *prompkg.ConfigGenerator,
) []monitoringv1.Argument {
	promArgs := prompkg.BuildCommonPrometheusArgs(cpf, cg)
	return appendAgentArgs(promArgs, cg, cpf.WALCompression)
}
