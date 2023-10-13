// Copyright 2019 The prometheus-operator Authors
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

package admission

import (
	"encoding/json"
	"fmt"
)

func generatePatchesForNonStringLabelsAnnotations(content []byte) ([]string, error) {
	groups := &RuleGroups{}
	if err := json.Unmarshal(content, groups); err != nil {
		return nil, fmt.Errorf("cannot unmarshal RuleGroups: %w", err)
	}

	patches := new([]string)
	for gi := range groups.Groups {
		for ri, rule := range groups.Groups[gi].Rules {
			for key, val := range rule.Annotations {
				patchIfNotString(patches, gi, ri, "annotations", key, val)
			}
			for key, val := range rule.Labels {
				patchIfNotString(patches, gi, ri, "labels", key, val)
			}
		}
	}

	return *patches, nil
}

func patchIfNotString(patches *[]string, gi, ri int, typ, key string, val interface{}) {
	if _, ok := val.(string); ok || val == nil {
		// Kubernetes does not let nil values get this far.
		// Keeping it here for the sake of clarity of behavior.
		return
	}
	*patches = append(*patches,
		fmt.Sprintf(`{"op": "replace","path": "/spec/groups/%d/rules/%d/%s/%s","value": "%v"}`,
			gi, ri, typ, key, val))

}
