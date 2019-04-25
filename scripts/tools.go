//+build tools

// Package tools tracks dependencies for tools that are required to generate the protobuf code.
// See https://github.com/golang/go/issues/25922
package tools

import (
	_ "github.com/campoy/embedmd"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "k8s.io/kube-openapi/cmd/openapi-gen"
)
