package crdvalidation

import (
	"fmt"
	spec "github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	common "k8s.io/kube-openapi/pkg/common"
)

// SchemaPropsToJSONPropsArray converts []Schema to []JSONSchemaProps
func SchemaPropsToJSONPropsArray(schemas []spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) []extensionsobj.JSONSchemaProps {
	var s []extensionsobj.JSONSchemaProps
	for _, schema := range schemas {
		s = append(s, *SchemaPropsToJSONProps(&schema, openapiSpec, nested))
	}
	return s
}

// StringOrArrayToString converts StringOrArray to string
func StringOrArrayToString(strOrArray spec.StringOrArray) string {
	if len(strOrArray) > 0 {
		return strOrArray[0]
	}
	return ""
}

// EnumJSON converts []interface{} to []JSON
func EnumJSON(enum []interface{}) []extensionsobj.JSON {
	var s []extensionsobj.JSON
	for _, elt := range enum {
		s = append(s, extensionsobj.JSON{
			Raw: []byte(fmt.Sprintf("%v", elt)),
		})
	}
	return s
}

// SchemaOrArrayToJSONItems converts *SchemaOrArray to *JSONSchemaPropsOrArray
func SchemaOrArrayToJSONItems(schemaOrArray *spec.SchemaOrArray, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaPropsOrArray {
	var array *extensionsobj.JSONSchemaPropsOrArray
	if schemaOrArray == nil {
		return array
	}
	return &extensionsobj.JSONSchemaPropsOrArray{
		Schema:      SchemaPropsToJSONProps(schemaOrArray.Schema, openapiSpec, nested),
		JSONSchemas: SchemaPropsToJSONPropsArray(schemaOrArray.Schemas, openapiSpec, nested),
	}
}

// SchemaOrBoolToJSONProps converts *SchemaOrBool to *JSONSchemaPropsOrBool
func SchemaOrBoolToJSONProps(schemaOrBool *spec.SchemaOrBool, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaPropsOrBool {
	var s *extensionsobj.JSONSchemaPropsOrBool
	if schemaOrBool == nil {
		return s
	}
	return &extensionsobj.JSONSchemaPropsOrBool{
		Schema: SchemaPropsToJSONProps(schemaOrBool.Schema, openapiSpec, nested),
		Allows: schemaOrBool.Allows,
	}
}

// SchemPropsMapToJSONMap converts map[string]Schema to map[string]JSONSchemaProps
func SchemPropsMapToJSONMap(schemaMap map[string]spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) map[string]extensionsobj.JSONSchemaProps {
	var m map[string]extensionsobj.JSONSchemaProps
	m = make(map[string]extensionsobj.JSONSchemaProps)
	for key, schema := range schemaMap {
		m[key] = *SchemaPropsToJSONProps(&schema, openapiSpec, nested)
	}
	return m
}

// SchemaPropsToJSONProps converts a SchemaProps to a JSONProps
func SchemaPropsToJSONProps(schema *spec.Schema, openapiSpec map[string]common.OpenAPIDefinition, nested bool) *extensionsobj.JSONSchemaProps {
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
			return SchemaPropsToJSONProps(&propref, openapiSpec, nested)
		}
		ref = new(string)
		*ref = schemaProps.Ref.String()
	}

	props = &extensionsobj.JSONSchemaProps{
		Ref:              ref,
		ID:               schemaProps.ID,
		Schema:           extensionsobj.JSONSchemaURL(string(schema.Schema)),
		Description:      schemaProps.Description,
		Type:             StringOrArrayToString(schemaProps.Type),
		Format:           schemaProps.Format,
		Title:            schemaProps.Title,
		Maximum:          schemaProps.Maximum,
		ExclusiveMaximum: schemaProps.ExclusiveMaximum,
		Minimum:          schemaProps.Minimum,
		ExclusiveMinimum: schemaProps.ExclusiveMinimum,
		MaxLength:        schemaProps.MaxLength,
		MinLength:        schemaProps.MinLength,
		Pattern:          schemaProps.Pattern,
		MaxItems:         schemaProps.MaxItems,
		MinItems:         schemaProps.MinItems,
		UniqueItems:      schemaProps.UniqueItems,
		MultipleOf:       schemaProps.MultipleOf,
		Enum:             EnumJSON(schemaProps.Enum),
		MaxProperties:    schemaProps.MaxProperties,
		MinProperties:    schemaProps.MinProperties,
		Required:         schemaProps.Required,
		Items:            SchemaOrArrayToJSONItems(schemaProps.Items, openapiSpec, nested),
		AllOf:            SchemaPropsToJSONPropsArray(schemaProps.AllOf, openapiSpec, nested),
		OneOf:            SchemaPropsToJSONPropsArray(schemaProps.OneOf, openapiSpec, nested),
		AnyOf:            SchemaPropsToJSONPropsArray(schemaProps.AnyOf, openapiSpec, nested),
		Not:              SchemaPropsToJSONProps(schemaProps.Not, openapiSpec, nested),
		Properties:       SchemPropsMapToJSONMap(schemaProps.Properties, openapiSpec, nested),
		// @TODO(01-25-2018) Field not accepted by the current CRD Validation Spec
		// AdditionalProperties: SchemaOrBoolToJSONProps(schemaProps.AdditionalProperties, openapiSpec, nested),
		PatternProperties: SchemPropsMapToJSONMap(schemaProps.PatternProperties, openapiSpec, nested),
		AdditionalItems:   SchemaOrBoolToJSONProps(schemaProps.AdditionalItems, openapiSpec, nested),
	}
	return props
}
