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
	"bytes"
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

const kubeAPIVersion = "v1.23"

var externalLinks = map[string]string{
	"metav1.ObjectMeta":        fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#objectmeta-v1-meta", kubeAPIVersion),
	"metav1.ListMeta":          fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#listmeta-v1-meta", kubeAPIVersion),
	"metav1.LabelSelector":     fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#labelselector-v1-meta", kubeAPIVersion),
	"v1.ResourceRequirements":  fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#resourcerequirements-v1-core", kubeAPIVersion),
	"v1.LocalObjectReference":  fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#localobjectreference-v1-core", kubeAPIVersion),
	"v1.SecretKeySelector":     fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#secretkeyselector-v1-core", kubeAPIVersion),
	"v1.PersistentVolumeClaim": fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#persistentvolumeclaim-v1-core", kubeAPIVersion),
	"v1.EmptyDirVolumeSource":  fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#emptydirvolumesource-v1-core", kubeAPIVersion),
	"apiextensionsv1.JSON":     fmt.Sprintf("https://kubernetes.io/docs/reference/generated/kubernetes-api/%s/#json-v1-apiextensions-k8s-io", kubeAPIVersion),
}

// Field is a representation of a struct field.
type Field ast.Field

// Name returns the name of the field as it should appear in JSON format
// "-" indicates that this field is not part of the JSON representation.
func (f *Field) Name() string {
	jsonTag := reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1]).Get("json") // Delete first and last quotation
	jsonTag = strings.Split(jsonTag, ",")[0]                                      // This can return "-"
	if jsonTag != "" {
		return jsonTag
	}

	if f.Names != nil {
		return f.Names[0].Name
	}

	return f.Type.(*ast.Ident).Name
}

// Description returns the description of the field inferred from the comment string preceding it.
func (f *Field) Description() interface{} {
	return fmtRawDoc(f.Doc.Text())
}

func fmtRawDoc(rawDoc string) string {
	var buffer bytes.Buffer
	delPrevChar := func() {
		if buffer.Len() > 0 {
			buffer.Truncate(buffer.Len() - 1) // Delete the last " " or "\n"
		}
	}

	// Ignore all lines after ---
	rawDoc = strings.Split(rawDoc, "---")[0]

	for _, line := range strings.Split(rawDoc, "\n") {
		line = strings.TrimRight(line, " ")
		leading := strings.TrimLeft(line, " ")
		switch {
		case len(line) == 0: // Keep paragraphs
			delPrevChar()
			buffer.WriteString("\n\n")
		case strings.HasPrefix(leading, "TODO"): // Ignore one line TODOs
		case strings.HasPrefix(leading, "+"): // Ignore instructions to go2idl
		default:
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				delPrevChar()
				line = "\n" + line + "\n" // Replace it with newline. This is useful when we have a line with: "Example:\n\tJSON-someting..."
			} else {
				line += " "
			}
			buffer.WriteString(line)
		}
	}

	postDoc := strings.TrimRight(buffer.String(), "\n")
	postDoc = strings.Replace(postDoc, "\\\"", "\"", -1) // replace user's \" to "
	postDoc = strings.Replace(postDoc, "\"", "\\\"", -1) // Escape "
	postDoc = strings.Replace(postDoc, "\n", "\\n", -1)
	postDoc = strings.Replace(postDoc, "\t", "\\t", -1)
	postDoc = strings.Replace(postDoc, "|", "\\|", -1)

	return postDoc
}

// IsInlined returns true when the field has the `inline` annotation.
func (f *Field) IsInlined() bool {
	jsonTag := reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1]).Get("json") // Delete first and last quotation
	return strings.Contains(jsonTag, "inline")
}

// TypeName returns the name of the field's data type.
func (f *Field) TypeName() string {
	return typeName(f.Type)
}

func typeName(expr ast.Expr) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.SelectorExpr:
		pkg := typ.X.(*ast.Ident)
		t := typ.Sel
		return pkg.Name + "." + t.Name
	case *ast.StarExpr:
		return typeName(typ.X)
	case *ast.ArrayType:
		return typeName(typ.Elt)
	}

	return ""
}

// HasInternalType returns true when the data type of the field is
// defined in the prometheus-operator project.
func (f *Field) HasInternalType() bool {
	return isInternalType(f.Type)
}

func isInternalType(typ ast.Expr) bool {
	switch typ := typ.(type) {
	case *ast.SelectorExpr:
		pkg := typ.X.(*ast.Ident)
		return strings.HasPrefix(pkg.Name, "monitoring")
	case *ast.StarExpr:
		return isInternalType(typ.X)
	case *ast.ArrayType:
		return isInternalType(typ.Elt)
	case *ast.MapType:
		return isInternalType(typ.Key) && isInternalType(typ.Value)
	default:
		return true
	}
}

// IsRequired returns true if the field is mandatory.
func (f *Field) IsRequired() interface{} {
	jsonTag := ""
	if f.Tag == nil {
		return false
	}

	jsonTag = reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1]).Get("json") // Delete first and last quotation
	return !strings.Contains(jsonTag, "omitempty")
}

// TypeLink returns a link to the data type of the field in the API documentation.
func (f *Field) TypeLink(typeSet TypeSet) string {
	return typeLink(f.Type, typeSet)
}

func typeLink(expr ast.Expr, typeSet TypeSet) string {
	switch typ := expr.(type) {
	case *ast.Ident:
		return toLink(typ.Name, typeSet)
	case *ast.StarExpr:
		return "*" + typeLink(typ.X, typeSet)
	case *ast.SelectorExpr:
		pkg := typ.X.(*ast.Ident)
		t := typ.Sel
		return toLink(pkg.Name+"."+t.Name, typeSet)
	case *ast.ArrayType:
		return "[]" + typeLink(typ.Elt, typeSet)
	case *ast.MapType:
		return "map[" + typeLink(typ.Key, typeSet) + "]" + typeLink(typ.Value, typeSet)
	default:
		return ""
	}
}

func toLink(typeName string, structs map[string]*StructType) string {
	link, hasLink := externalLinks[typeName]
	if hasLink {
		return wrapInLink(typeName, link)
	}

	s, ok := structs[typeName]
	if ok {
		return wrapInLink(typeName, "#"+strings.ToLower(s.Name))
	}

	return typeName
}

func wrapInLink(text, link string) string {
	return fmt.Sprintf("[%s](%s)", text, link)
}
