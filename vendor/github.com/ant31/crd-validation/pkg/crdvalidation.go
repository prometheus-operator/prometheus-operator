package crdvalidation

import (
	spec "github.com/go-openapi/spec"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	common "k8s.io/kube-openapi/pkg/common"
)

// CustomResourceDefinitionTypeMeta set the default kind/apiversion of CRD
var CustomResourceDefinitionTypeMeta = metav1.TypeMeta{
	Kind:       "CustomResourceDefinition",
	APIVersion: "apiextensions.k8s.io/v1beta1",
}

// OpenAPIRefCallBack returns a jsonref using the input string without modification
func OpenAPIRefCallBack(name string) spec.Ref {
	return spec.MustCreateRef(name)
}

// GetAPIDefinition is a function returning a map with all Definition
type GetAPIDefinitions func(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition

// GetCustomResourceValidations returns a CRD validation spec map. It took the openapi generated definition from kube-openapi as argument
func GetCustomResourceValidations(fn GetAPIDefinitions) map[string]*extensionsobj.CustomResourceValidation {
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
	fixKnownTypes(openapiSpec)
	schema := openapiSpec[name].Schema
	crv := &extensionsobj.CustomResourceValidation{
		OpenAPIV3Schema: SchemaPropsToJSONProps(&schema, openapiSpec, true),
	}
	crv.OpenAPIV3Schema.Description = ""
	crv.OpenAPIV3Schema.Required = nil
	return crv
}

// ref: https://github.com/kubernetes/kubernetes/issues/62329
func fixKnownTypes(openapiSpec map[string]common.OpenAPIDefinition) {
	openapiSpec["k8s.io/apimachinery/pkg/util/intstr.IntOrString"] = common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				AnyOf: []spec.Schema{
					{
						SchemaProps: spec.SchemaProps{
							Type: []string{"string"},
						},
					},
					{
						SchemaProps: spec.SchemaProps{
							Type: []string{"integer"},
						},
					},
				},
			},
		},
	}
}

