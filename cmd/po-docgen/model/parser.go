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
	"go/doc"
	"go/parser"
	"go/token"
)

// Load parses a go file and returns a TypeSet.
func Load(path string) (TypeSet, error) {
	structs := make(map[string]*StructType)
	types, err := astFrom(path)
	if err != nil {
		return nil, err
	}

	for _, t := range types.Types {
		structType, ok := t.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
		if !ok {
			continue
		}

		structs[t.Name] = &StructType{
			Name:   t.Name,
			Fields: make([]Field, 0),

			doc:       t.Doc,
			rawFields: structType.Fields.List,
			appearsIn: make(map[string]struct{}),
		}
	}

	for _, currentStruct := range structs {
		for _, field := range currentStruct.rawFields {
			field := Field(*field)
			if field.IsInlined() {
				if field.HasInternalType() {
					mergeFields(currentStruct, structs[field.Name()], structs)
				}
				continue
			}

			fieldName := field.Name()
			if fieldName == "-" {
				continue
			}

			currentStruct.Fields = append(currentStruct.Fields, field)
			fieldStruct, ok := structs[field.TypeName()]
			if !ok {
				continue
			}
			fieldStruct.referenceCount++
		}
	}

	for k, s := range structs {
		// Remove types that are used only as embedded
		if s.IsOnlyEmbedded() {
			delete(structs, k)
		}
	}

	return structs, nil
}

func mergeFields(to *StructType, from *StructType, structs TypeSet) {
	from.embeddedCount++
	from.referenceCount++
	for _, field := range from.rawFields {
		field := Field(*field)
		if field.Name() == "-" {
			continue
		}

		to.Fields = append(to.Fields, field)
		if fieldStruct, ok := structs[field.TypeName()]; ok {
			fieldStruct.appearsIn[to.Name] = struct{}{}
		}
	}
}

func astFrom(filePath string) (*doc.Package, error) {
	fset := token.NewFileSet()
	m := make(map[string]*ast.File)

	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	m[filePath] = f
	apkg, _ := ast.NewPackage(fset, m, nil, nil)

	return doc.New(apkg, "", 0), nil
}
