// Copyright 2017 The prometheus-operator Authors
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
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
)

func TestListOptions(t *testing.T) {
	for i := 0; i < 1000; i++ {
		o := ListOptions("test")
		if o.LabelSelector != "app=prometheus,prometheus=test" && o.LabelSelector != "prometheus=test,app=prometheus" {
			t.Fatalf("LabelSelector not computed correctly\n\nExpected: \"app=prometheus,prometheus=test\"\n\nGot:      %#+v", o.LabelSelector)
		}
	}
}

func TestCreateStatefulSetChecksum(t *testing.T) {
	p1 := monitoringv1.Prometheus{}
	p1.Spec.Version = "v1.7.0"
	p2 := monitoringv1.Prometheus{}
	p2.Spec.Version = "v1.7.2"
	c := Config{}

	p1Checksum, err := createSSetInputChecksum(p1, c)
	if err != nil {
		t.Fatal(err)
	}
	p2Checksum, err := createSSetInputChecksum(p2, c)
	if err != nil {
		t.Fatal(err)
	}

	if p1Checksum == p2Checksum {
		t.Fatal("expected two different Prometheus CRDs to result in two different checksums but got equal checksums")
	}
}
