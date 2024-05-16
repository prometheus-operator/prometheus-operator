// Copyright 2024 The prometheus-operator Authors
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

package goruntime

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.uber.org/automaxprocs/maxprocs"
)

func SetMaxProcs(logger log.Logger) {
	l := func(format string, a ...interface{}) {
		level.Info(logger).Log("msg", fmt.Sprintf(strings.TrimPrefix(format, "maxprocs: "), a...))
	}

	if _, err := maxprocs.Set(maxprocs.Logger(l)); err != nil {
		level.Warn(logger).Log("msg", "Failed to set GOMAXPROCS automatically", "err", err)
	}
}
