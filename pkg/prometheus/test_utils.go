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
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	logging "github.com/prometheus-operator/prometheus-operator/internal/log"
)

func makeExpectedProbeHandler(probePath string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   probePath,
			Port:   intstr.FromString("web"),
			Scheme: "HTTPS",
		},
	}
}

func MakeExpectedStartupProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    15,
		FailureThreshold: 60,
	}
}

func MakeExpectedLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/healthy"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 6,
	}
}

func MakeExpectedReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:     makeExpectedProbeHandler("/-/ready"),
		TimeoutSeconds:   3,
		PeriodSeconds:    5,
		FailureThreshold: 3,
	}
}

func NewLogger() *slog.Logger {
	l, err := logging.NewLoggerSlog(logging.Config{
		Level:  logging.LevelWarn,
		Format: logging.FormatLogFmt,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return l
}
