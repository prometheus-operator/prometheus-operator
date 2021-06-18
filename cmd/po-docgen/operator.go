// Copyright 2021 The prometheus-operator Authors
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

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type FlagDoc struct {
	name         string
	defaultValue string
	description  string
}

func fileToASTPackage(filePaths []string) *ast.Package {
	fset := token.NewFileSet()
	m := make(map[string]*ast.File)

	for _, filePath := range filePaths {

		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		m[filePath] = f
	}
	apkg, _ := ast.NewPackage(fset, m, nil, nil)

	return apkg
}

func unquoteLiteral(value string) string {
	if len(value) > 0 && value[0] == '"' {
		value = value[1:]
	}

	if len(value) > 0 && value[len(value)-1] == '"' {
		value = value[:len(value)-1]
	}
	return value
}

func resolveDurationExpr(expr ast.Expr) (string, error) {
	var exprIdent *ast.Ident

	switch exprCast := expr.(type) {
	case *ast.SelectorExpr:
		exprIdent = exprCast.Sel
	case *ast.Ident:
		exprIdent = exprCast
	default:
		return "", errors.New("No Identifier to Resolve")
	}

	if exprIdent.Obj == nil {
		return exprIdent.String(), nil
	}

	retValue := ""

	if exprIdent.Obj.Kind == ast.Con {
		valSpec := exprIdent.Obj.Decl.(*ast.ValueSpec)
		for _, value := range valSpec.Values {
			retValue += unquoteLiteral(value.(*ast.BasicLit).Value)
		}
	}
	return retValue, nil

}

func resolveBoolExpr(expr ast.Expr) (string, error) {
	exprIdent, ok := expr.(*ast.Ident)
	if ok {
		return exprIdent.Name, nil
	}

	return "", errors.New("No Identifier to Resolve")
}

func resolveConstStringExpr(expr ast.Expr) (string, error) {

	switch exprCast := expr.(type) {
	case *ast.BasicLit:
		return unquoteLiteral(exprCast.Value), nil
	case *ast.Ident:
		retValue := ""
		if exprCast.Obj.Kind == ast.Con {
			valSpec := exprCast.Obj.Decl.(*ast.ValueSpec)
			for _, value := range valSpec.Values {
				retValue += unquoteLiteral(value.(*ast.BasicLit).Value)
			}
		}
		return retValue, nil
	case *ast.BinaryExpr:
		operandX, err := resolveConstStringExpr(exprCast.X)
		if err != nil {
			return "", err
		}
		operandY, err := resolveConstStringExpr(exprCast.Y)
		if err != nil {
			return "", err
		}
		if exprCast.Op == token.ADD {
			return operandX + operandY, nil
		}
		return "", errors.New(fmt.Sprint("unhandled OP", exprCast.Op))

	}

	return "", nil
}

func operatorCodeToDoc(paths []string) ([]FlagDoc, error) {
	astPkg := fileToASTPackage(paths)
	flagDocs := make([]FlagDoc, 0)

	for _, file := range astPkg.Files {
		fmt.Println(file.Name.Name)
		for _, decl := range file.Decls {
			funcdecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcdecl.Name.Name == "init" {
				for _, statement := range funcdecl.Body.List {
					exprStatement, ok := statement.(*ast.ExprStmt)
					if !ok {
						continue
					}
					exprCall, ok := exprStatement.X.(*ast.CallExpr)
					if !ok {
						continue
					}

					if exprCall.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "flagset" {
						selectorExpr, ok := exprCall.Fun.(*ast.SelectorExpr)
						if !ok {
							continue
						}

						var argName, argDefaultValue, argDescription string
						var err error
						switch selectorExpr.Sel.Name {
						case "StringVar":
							argName = unquoteLiteral(exprCall.Args[1].(*ast.BasicLit).Value)
							argDefaultValue, err = resolveConstStringExpr(exprCall.Args[2])
							if err != nil {
								return flagDocs, err
							}
							if len(argDefaultValue) == 0 {
								argDefaultValue = "\"\""
							}
							argDescription, err = resolveConstStringExpr(exprCall.Args[3])
							if err != nil {
								return flagDocs, err
							}
						case "Var":
							argName = unquoteLiteral(exprCall.Args[1].(*ast.BasicLit).Value)
							argDefaultValue = "N/A"
							argDescription, err = resolveConstStringExpr(exprCall.Args[2])
							if err != nil {
								return flagDocs, err
							}
						case "DurationVar":
							argName = unquoteLiteral(exprCall.Args[1].(*ast.BasicLit).Value)
							argDefaultValue, err = resolveDurationExpr(exprCall.Args[2])
							if err != nil {
								return flagDocs, err
							}
							argDescription, err = resolveConstStringExpr(exprCall.Args[3])
							if err != nil {
								return flagDocs, err
							}
						case "BoolVar":
							argName = unquoteLiteral(exprCall.Args[1].(*ast.BasicLit).Value)
							argDefaultValue, err = resolveBoolExpr(exprCall.Args[2])
							if err != nil {
								return flagDocs, err
							}
							argDescription, err = resolveConstStringExpr(exprCall.Args[3])
							if err != nil {
								return flagDocs, err
							}
						default:
							return flagDocs, errors.New(fmt.Sprint("Unhandled argument type ", selectorExpr.Sel.Name))
						}

						flagdoc := FlagDoc{
							name:         argName,
							defaultValue: argDefaultValue,
							description:  argDescription,
						}

						flagDocs = append(flagDocs, flagdoc)

					}

				}
			}

		}
	}
	return flagDocs, nil

}

func printOperatorDocs(paths []string) {

	flagdocs, err := operatorCodeToDoc(paths)
	if err != nil {
		fmt.Print("Error when generating documents! ", err)
		return
	}

	fmt.Print(`---
title: "Operator CLI Flags"
description: "Lists of possible arguments passed to operator executable."
date: 2021-06-18T14:12:33-00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 1000
toc: false
---

`)
	fmt.Println("Operator CLI Flags")
	fmt.Println("=================")
	fmt.Println("This article lists arguments of operator executable.")
	fmt.Println("> Note this document is automatically generated from the `cmd/operator/main.go` file and shouldn't be edited directly.")
	fmt.Println("")
	fmt.Println("| Argument | Description | Default Value |")
	fmt.Println("| -------- | ----------- | ------------- |")

	for _, flagdoc := range flagdocs {
		fmt.Printf("| %s | %s | %s |\n",
			flagdoc.name, flagdoc.description, flagdoc.defaultValue)
	}
}
