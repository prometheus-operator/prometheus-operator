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

package prometheus

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func makeExpectedProbeHandler(probePath string) v1.ProbeHandler {
	return v1.ProbeHandler{
		HTTPGet: &v1.HTTPGetAction{
			Path:   probePath,
			Port:   intstr.FromString("web"),
			Scheme: "HTTPS",
		},
	}
}

func MakeExpectedStartupProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
}

func MakeExpectedLivenessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
}

func MakeExpectedReadinessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
}

func NewLogger() log.Logger {
	return level.NewFilter(log.NewLogfmtLogger(os.Stdout), level.AllowWarn())
}
