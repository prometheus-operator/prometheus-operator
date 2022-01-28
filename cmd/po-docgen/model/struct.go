// Copyright 2022 The prometheus-operator Authors
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

package model

import (
	"go/ast"
	"sort"
)

// TypeSet is a map from type name to *StructType.
type TypeSet map[string]*StructType

// Keys returns the keys of the TypeSet.
func (ts TypeSet) Keys() []string {
	keys := make([]string, 0, len(ts))
	for k := range ts {
		keys = append(keys, k)
	}

	return keys
}

// SortedKeys returns a lexicographically sirted slice with the keys of the TypeSet.
func (ts TypeSet) SortedKeys() []string {
	keys := ts.Keys()
	sort.Strings(keys)
	return keys
}

// StructType is a struct data type.
type StructType struct {
	Name   string
	Fields []Field

	doc            string
	rawFields      []*ast.Field
	appearsIn      map[string]struct{}
	embeddedCount  int
	referenceCount int
}

// Description returns the description of the struct inferred from the comment preceding it.
func (s *StructType) Description() string {
	return fmtRawDoc(s.doc)
}

// IsOnlyEmbedded returns true if the struct is only used as an embedded field in other structs.
func (s StructType) IsOnlyEmbedded() bool {
	return s.referenceCount != 0 && s.referenceCount == s.embeddedCount
}
