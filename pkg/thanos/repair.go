// Copyright 2025 The prometheus-operator Authors
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

package thanos

import (
	"context"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func (o *Operator) resolveStuckStatefulSet(ctx context.Context, logger *slog.Logger, _ *monitoringv1.ThanosRuler, sset *appsv1.StatefulSet) error {
	return operator.ResolveStuckStatefulSet(ctx, logger, o.kclient, sset, o.config.RepairPolicy)
}
