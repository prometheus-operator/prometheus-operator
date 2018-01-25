// Copyright 2018 The prometheus-operator Authors
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

package crdvalidation

import (
	"fmt"
	spec "github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	common "k8s.io/kube-openapi/pkg/common"
	//"strings"
)

func OpenAPIRefCallBack(name string) spec.Ref {
	return spec.MustCreateRef(name)
}

// SchemaPropsToJsonProps converts an array of Schema to an Array of JsonProps
func SchemaPropsToJsonPropsArray(schemas []spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) []extensionsobj.JSONSchemaProps {
	var s []extensionsobj.JSONSchemaProps
	for _, schema := range schemas {
		s = append(s, *SchemaPropsToJsonProps(&schema, openapiSpec, nested))
	}
	return s
}

func StringOrArrayToString(strOrArray spec.StringOrArray) string {
	if len(strOrArray) > 0 {
		return strOrArray[0]
	}
	return ""
}

func EnumJSON(enum []interface{}) []extensionsobj.JSON {
	var s []extensionsobj.JSON
	for _, elt := range enum {
		s = append(s, extensionsobj.JSON{
			Raw: []byte(fmt.Sprintf("%v", elt)),
		})
	}
	return s
}

func SchemaOrArrayToJsonItems(schemaOrArray *spec.SchemaOrArray, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaPropsOrArray {
	var array *extensionsobj.JSONSchemaPropsOrArray
	if schemaOrArray == nil {
		return array
	}
	return &extensionsobj.JSONSchemaPropsOrArray{
		Schema:      SchemaPropsToJsonProps(schemaOrArray.Schema, openapiSpec, nested),
		JSONSchemas: SchemaPropsToJsonPropsArray(schemaOrArray.Schemas, openapiSpec, nested),
	}
}

func SchemaOrBoolToJsonProps(schemaOrBool *spec.SchemaOrBool, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaPropsOrBool {
	var s *extensionsobj.JSONSchemaPropsOrBool
	if schemaOrBool == nil {
		return s
	}
	return &extensionsobj.JSONSchemaPropsOrBool{
		Schema: SchemaPropsToJsonProps(schemaOrBool.Schema, openapiSpec, nested),
		Allows: schemaOrBool.Allows,
	}
}

func SchemPropsMapToJsonMap(schemaMap map[string]spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) map[string]extensionsobj.JSONSchemaProps {
	var m map[string]extensionsobj.JSONSchemaProps
	m = make(map[string]extensionsobj.JSONSchemaProps)
	for key, schema := range schemaMap {
		m[key] = *SchemaPropsToJsonProps(&schema, openapiSpec, nested)
	}
	return m
}

// SchemaPropsToJsonProps converts a SchemaProps to a JsonProps
func SchemaPropsToJsonProps(schema *spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaProps {
	var props *extensionsobj.JSONSchemaProps
	if schema == nil {
		return props
	}
	schemaProps := &schema.SchemaProps

	var ref *string
	if schemaProps.Ref.String() != "" {
		if nested {
			propref := openapiSpec[schemaProps.Ref.String()].Schema
			// If nested just return a pointer to the reference
			return SchemaPropsToJsonProps(&propref, openapiSpec, nested)
		}
		ref = new(string)
		*ref = schemaProps.Ref.String()
	}

	props = &extensionsobj.JSONSchemaProps{
		Ref:                  ref,
		ID:                   schemaProps.ID,
		Schema:               extensionsobj.JSONSchemaURL(string(schema.Schema)),
		Description:          schemaProps.Description,
		Type:                 StringOrArrayToString(schemaProps.Type),
		Format:               schemaProps.Format,
		Title:                schemaProps.Title,
		Maximum:              schemaProps.Maximum,
		ExclusiveMaximum:     schemaProps.ExclusiveMaximum,
		Minimum:              schemaProps.Minimum,
		ExclusiveMinimum:     schemaProps.ExclusiveMinimum,
		MaxLength:            schemaProps.MaxLength,
		MinLength:            schemaProps.MinLength,
		Pattern:              schemaProps.Pattern,
		MaxItems:             schemaProps.MaxItems,
		MinItems:             schemaProps.MinItems,
		UniqueItems:          schemaProps.UniqueItems,
		MultipleOf:           schemaProps.MultipleOf,
		Enum:                 EnumJSON(schemaProps.Enum),
		MaxProperties:        schemaProps.MaxProperties,
		MinProperties:        schemaProps.MinProperties,
		Required:             schemaProps.Required,
		Items:                SchemaOrArrayToJsonItems(schemaProps.Items, openapiSpec, nested),
		AllOf:                SchemaPropsToJsonPropsArray(schemaProps.AllOf, openapiSpec, nested),
		OneOf:                SchemaPropsToJsonPropsArray(schemaProps.OneOf, openapiSpec, nested),
		AnyOf:                SchemaPropsToJsonPropsArray(schemaProps.AnyOf, openapiSpec, nested),
		Not:                  SchemaPropsToJsonProps(schemaProps.Not, openapiSpec, nested),
		Properties:           SchemPropsMapToJsonMap(schemaProps.Properties, openapiSpec, nested),
		AdditionalProperties: SchemaOrBoolToJsonProps(schemaProps.AdditionalProperties, openapiSpec, nested),
		PatternProperties:    SchemPropsMapToJsonMap(schemaProps.PatternProperties, openapiSpec, nested),
		AdditionalItems:      SchemaOrBoolToJsonProps(schemaProps.AdditionalItems, openapiSpec, nested),
	}
	return props
}

// GetOpenAPICrdDefinitions returns a CRD validation spec map. It took the openapi generated definition from kube-openapi as argument
func GetOpenAPICrdDefinitions(fn func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition) map[string]*extensionsobj.JSONSchemaProps {
	openapiSpec := fn(OpenAPIRefCallBack)
	var definitions map[string]*extensionsobj.JSONSchemaProps
	definitions = make(map[string]*extensionsobj.JSONSchemaProps)
	for key, definition := range openapiSpec {
		schema := definition.Schema
		definitions[key] = SchemaPropsToJsonProps(&schema, openapiSpec, true)
	}
	return definitions
}
