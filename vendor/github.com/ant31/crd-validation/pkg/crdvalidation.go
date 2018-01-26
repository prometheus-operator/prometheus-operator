package crdvalidation

import (
	spec "github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	common "k8s.io/kube-openapi/pkg/common"
)

// OpenAPIRefCallBack returns a jsonref using the input string without modification
func OpenAPIRefCallBack(name string) spec.Ref {
	return spec.MustCreateRef(name)
}

// GetCustomResourceValidations returns a CRD validation spec map. It took the openapi generated definition from kube-openapi as argument
func GetCustomResourceValidations(fn func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition) map[string]*extensionsobj.CustomResourceValidation {
	openapiSpec := fn(OpenAPIRefCallBack)
	var definitions map[string]*extensionsobj.CustomResourceValidation
	definitions = make(map[string]*extensionsobj.CustomResourceValidation)
	for key, definition := range openapiSpec {
		schema := definition.Schema
		definitions[key] = &extensionsobj.CustomResourceValidation{
			OpenAPIV3Schema: SchemaPropsToJSONProps(&schema, openapiSpec, true),
		}
	}
	return definitions
}

// GetCustomResourceValidation returns the validation definition for a CRD name
func GetCustomResourceValidation(name string, fn func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition) *extensionsobj.CustomResourceValidation {
	openapiSpec := fn(OpenAPIRefCallBack)
	schema := openapiSpec[name].Schema
	return &extensionsobj.CustomResourceValidation{
		OpenAPIV3Schema: SchemaPropsToJSONProps(&schema, openapiSpec, true),
	}

}
