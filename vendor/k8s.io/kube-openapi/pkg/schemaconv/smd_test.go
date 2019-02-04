/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schemaconv

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"k8s.io/kube-openapi/pkg/util/proto"
	prototesting "k8s.io/kube-openapi/pkg/util/proto/testing"
)

var (
	fakeSchema            = prototesting.Fake{Path: filepath.Join("testdata", "swagger.json")}
	expectedNewSchemaPath = filepath.Join("testdata", "new-schema.yaml")
)

func TestToSchema(t *testing.T) {
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatal(err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := ToSchema(models)
	if err != nil {
		t.Fatal(err)
	}
	got, err := yaml.Marshal(ns)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(got))

	expect, err := ioutil.ReadFile(expectedNewSchemaPath)
	if err != nil {
		t.Fatalf("Unable to read golden data file %q: %v", expectedNewSchemaPath, err)
	}

	if string(expect) != string(got) {
		t.Errorf("Computed schema did not match %q.", expectedNewSchemaPath)
		t.Logf("To recompute this file, run:\n\tgo run ./cmd/openapi2smd/openapi2smd.go < pkg/util/proto/testdata/swagger.json > pkg/schemaconv/testdata/new-schema.yaml")
	}
}
