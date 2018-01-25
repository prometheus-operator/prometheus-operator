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

package k8sutil

import (
	"fmt"
	spec "github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	common "k8s.io/kube-openapi/pkg/common"
	"strings"
)

func OpenAPIRefCallBack(name string) spec.Ref {
	defName := name[strings.LastIndex(name, "/")+1:]
	return spec.MustCreateRef("#/definitions/" + common.EscapeJsonPointer(defName))
}

// SchemaPropsToJsonProps converts an array of Schema to an Array of JsonProps
func SchemaPropsToJsonPropsArray(schemas []spec.Schema) []extensionsobj.JSONSchemaProps {
	var s []extensionsobj.JSONSchemaProps
	for _, schema := range schemas {
		s = append(s, *SchemaPropsToJsonProps(&schema))
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

func SchemaOrArrayToJsonItems(schemaOrArray *spec.SchemaOrArray) *extensionsobj.JSONSchemaPropsOrArray {
	var array *extensionsobj.JSONSchemaPropsOrArray
	if schemaOrArray == nil {
		return array
	}
	return &extensionsobj.JSONSchemaPropsOrArray{
		Schema:      SchemaPropsToJsonProps(schemaOrArray.Schema),
		JSONSchemas: SchemaPropsToJsonPropsArray(schemaOrArray.Schemas),
	}
}

func SchemaOrBoolToJsonProps(schemaOrBool *spec.SchemaOrBool) *extensionsobj.JSONSchemaPropsOrBool {
	var s *extensionsobj.JSONSchemaPropsOrBool
	if schemaOrBool == nil {
		return s
	}
	return &extensionsobj.JSONSchemaPropsOrBool{
		Schema: SchemaPropsToJsonProps(schemaOrBool.Schema),
		Allows: schemaOrBool.Allows,
	}
}

func SchemPropsMapToJsonMap(schemaMap map[string]spec.Schema) map[string]extensionsobj.JSONSchemaProps {
	var m map[string]extensionsobj.JSONSchemaProps
	m = make(map[string]extensionsobj.JSONSchemaProps)
	for key, schema := range schemaMap {
		m[key] = *SchemaPropsToJsonProps(&schema)
	}
	return m
}

// SchemaPropsToJsonProps converts a SchemaProps to a JsonProps
func SchemaPropsToJsonProps(schema *spec.Schema) *extensionsobj.JSONSchemaProps {
	var props *extensionsobj.JSONSchemaProps
	if schema == nil {
		return props
	}
	schemaProps := &schema.SchemaProps
	var ref *string
	if schemaProps.Ref.String() != "" {
		ref = new(string)
		*ref = schemaProps.Ref.String()
	}
	props = &extensionsobj.JSONSchemaProps{
		ID:                   schema.ID,
		Ref:                  ref,
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
		Items:                SchemaOrArrayToJsonItems(schemaProps.Items),
		AllOf:                SchemaPropsToJsonPropsArray(schemaProps.AllOf),
		OneOf:                SchemaPropsToJsonPropsArray(schemaProps.OneOf),
		AnyOf:                SchemaPropsToJsonPropsArray(schemaProps.AnyOf),
		Not:                  SchemaPropsToJsonProps(schemaProps.Not),
		Properties:           SchemPropsMapToJsonMap(schemaProps.Properties),
		AdditionalProperties: SchemaOrBoolToJsonProps(schemaProps.AdditionalProperties),
		PatternProperties:    SchemPropsMapToJsonMap(schemaProps.PatternProperties),
		AdditionalItems:      SchemaOrBoolToJsonProps(schemaProps.AdditionalItems),
	}
	return props
}

// Type: ,
// Format:
// Title
// Default
// Maximum
// ExclusiveMaximum
// Minimum
// ExclusiveMinimum
// MaxLength
// MinLength
// Pattern
// MaxItems
// MinItems
// UniqueItems
// MultipleOf
// Enum
// MaxProperties
// MinProperties
// Required
// Items
// AllOf
// OneOf
// AnyOf
// Not
// Properties
// AdditionalProperties
// PatternProperties
// Dependencies
// AdditionalItems
// Definitions
// ExternalDocs
// Example
// 	}
// 	return &props

// type JSONSchemaProps struct {
// 	ID                   string                     `json:"id,omitempty" protobuf:"bytes,1,opt,name=id"`
// 	Schema               JSONSchemaURL              `json:"$schema,omitempty" protobuf:"bytes,2,opt,name=schema"`
// 	Ref                  *string                    `json:"$ref,omitempty" protobuf:"bytes,3,opt,name=ref"`
// 	Description          string                     `json:"description,omitempty" protobuf:"bytes,4,opt,name=description"`
// 	Type                 string                     `json:"type,omitempty" protobuf:"bytes,5,opt,name=type"`
// 	Format               string                     `json:"format,omitempty" protobuf:"bytes,6,opt,name=format"`
// 	Title                string                     `json:"title,omitempty" protobuf:"bytes,7,opt,name=title"`
// 	Default              *JSON                      `json:"default,omitempty" protobuf:"bytes,8,opt,name=default"`
// 	Maximum              *float64                   `json:"maximum,omitempty" protobuf:"bytes,9,opt,name=maximum"`
// 	ExclusiveMaximum     bool                       `json:"exclusiveMaximum,omitempty" protobuf:"bytes,10,opt,name=exclusiveMaximum"`
// 	Minimum              *float64                   `json:"minimum,omitempty" protobuf:"bytes,11,opt,name=minimum"`
// 	ExclusiveMinimum     bool                       `json:"exclusiveMinimum,omitempty" protobuf:"bytes,12,opt,name=exclusiveMinimum"`
// 	MaxLength            *int64                     `json:"maxLength,omitempty" protobuf:"bytes,13,opt,name=maxLength"`
// 	MinLength            *int64                     `json:"minLength,omitempty" protobuf:"bytes,14,opt,name=minLength"`
// 	Pattern              string                     `json:"pattern,omitempty" protobuf:"bytes,15,opt,name=pattern"`
// 	MaxItems             *int64                     `json:"maxItems,omitempty" protobuf:"bytes,16,opt,name=maxItems"`
// 	MinItems             *int64                     `json:"minItems,omitempty" protobuf:"bytes,17,opt,name=minItems"`
// 	UniqueItems          bool                       `json:"uniqueItems,omitempty" protobuf:"bytes,18,opt,name=uniqueItems"`
// 	MultipleOf           *float64                   `json:"multipleOf,omitempty" protobuf:"bytes,19,opt,name=multipleOf"`
// 	Enum                 []JSON                     `json:"enum,omitempty" protobuf:"bytes,20,rep,name=enum"`
// 	MaxProperties        *int64                     `json:"maxProperties,omitempty" protobuf:"bytes,21,opt,name=maxProperties"`
// 	MinProperties        *int64                     `json:"minProperties,omitempty" protobuf:"bytes,22,opt,name=minProperties"`
// 	Required             []string                   `json:"required,omitempty" protobuf:"bytes,23,rep,name=required"`
// 	Items                *JSONSchemaPropsOrArray    `json:"items,omitempty" protobuf:"bytes,24,opt,name=items"`
// 	AllOf                []JSONSchemaProps          `json:"allOf,omitempty" protobuf:"bytes,25,rep,name=allOf"`
// 	OneOf                []JSONSchemaProps          `json:"oneOf,omitempty" protobuf:"bytes,26,rep,name=oneOf"`
// 	AnyOf                []JSONSchemaProps          `json:"anyOf,omitempty" protobuf:"bytes,27,rep,name=anyOf"`
// 	Not                  *JSONSchemaProps           `json:"not,omitempty" protobuf:"bytes,28,opt,name=not"`
// 	Properties           map[string]JSONSchemaProps `json:"properties,omitempty" protobuf:"bytes,29,rep,name=properties"`
// 	AdditionalProperties *JSONSchemaPropsOrBool     `json:"additionalProperties,omitempty" protobuf:"bytes,30,opt,name=additionalProperties"`
// 	PatternProperties    map[string]JSONSchemaProps `json:"patternProperties,omitempty" protobuf:"bytes,31,rep,name=patternProperties"`
// 	Dependencies         JSONSchemaDependencies     `json:"dependencies,omitempty" protobuf:"bytes,32,opt,name=dependencies"`
// 	AdditionalItems      *JSONSchemaPropsOrBool     `json:"additionalItems,omitempty" protobuf:"bytes,33,opt,name=additionalItems"`
// 	Definitions          JSONSchemaDefinitions      `json:"definitions,omitempty" protobuf:"bytes,34,opt,name=definitions"`
// 	ExternalDocs         *ExternalDocumentation     `json:"externalDocs,omitempty" protobuf:"bytes,35,opt,name=externalDocs"`
// 	Example              *JSON                      `json:"example,omitempty" protobuf:"bytes,36,opt,name=example"`
// }

// type SchemaProps struct {
// 	ID                   string            `json:"id,omitempty"`
// 	Ref                  Ref               `json:"-"`
// 	Schema               SchemaURL         `json:"-"`
// 	Description          string            `json:"description,omitempty"`
// 	Type                 StringOrArray     `json:"type,omitempty"`
// 	Format               string            `json:"format,omitempty"`
// 	Title                string            `json:"title,omitempty"`
// 	Default              interface{}       `json:"default,omitempty"`
// 	Maximum              *float64          `json:"maximum,omitempty"`
// 	ExclusiveMaximum     bool              `json:"exclusiveMaximum,omitempty"`
// 	Minimum              *float64          `json:"minimum,omitempty"`
// 	ExclusiveMinimum     bool              `json:"exclusiveMinimum,omitempty"`
// 	MaxLength            *int64            `json:"maxLength,omitempty"`
// 	MinLength            *int64            `json:"minLength,omitempty"`
// 	Pattern              string            `json:"pattern,omitempty"`
// 	MaxItems             *int64            `json:"maxItems,omitempty"`
// 	MinItems             *int64            `json:"minItems,omitempty"`
// 	UniqueItems          bool              `json:"uniqueItems,omitempty"`
// 	MultipleOf           *float64          `json:"multipleOf,omitempty"`
// 	Enum                 []interface{}     `json:"enum,omitempty"`
// 	MaxProperties        *int64            `json:"maxProperties,omitempty"`
// 	MinProperties        *int64            `json:"minProperties,omitempty"`
// 	Required             []string          `json:"required,omitempty"`
// 	Items                *SchemaOrArray    `json:"items,omitempty"`
// 	AllOf                []Schema          `json:"allOf,omitempty"`
// 	OneOf                []Schema          `json:"oneOf,omitempty"`
// 	AnyOf                []Schema          `json:"anyOf,omitempty"`
// 	Not                  *Schema           `json:"not,omitempty"`
// 	Properties           map[string]Schema `json:"properties,omitempty"`
// 	AdditionalProperties *SchemaOrBool     `json:"additionalProperties,omitempty"`
// 	PatternProperties    map[string]Schema `json:"patternProperties,omitempty"`
// 	Dependencies         Dependencies      `json:"dependencies,omitempty"`
// 	AdditionalItems      *SchemaOrBool     `json:"additionalItems,omitempty"`
// 	Definitions          Definitions       `json:"definitions,omitempty"`
// }
