// Copyright 2018 Istio Authors
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

package glog

import (
	"testing"
)

func TestAll(t *testing.T) {
	// just making sure this stuff doesn't crash...

	Errorf("%s %s", "One", "Two")
	Error("One", "Two")
	Errorln("One", "Two")
	ErrorDepth(2, "One", "Two")

	Warningf("%s %s", "One", "Two")
	Warning("One", "Two")
	Warningln("One", "Two")
	WarningDepth(2, "One", "Two")

	Infof("%s %s", "One", "Two")
	Info("One", "Two")
	Infoln("One", "Two")
	InfoDepth(2, "One", "Two")

	for i := 0; i < 10; i++ {
		V(Level(i)).Infof("%s %s", "One", "Two")
		V(Level(i)).Info("One", "Two")
		V(Level(i)).Infoln("One", "Two")
	}

	Flush()
}
