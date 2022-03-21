// Copyright 2016 The prometheus-operator Authors
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
	"fmt"
	"github.com/prometheus-operator/prometheus-operator/cmd/po-docgen/model"
	"log"
	"sort"
	"strings"
)

const (
	firstParagraph = `---
title: "API"
description: "Generated API docs for the Prometheus Operator"
lead: ""
date: 2021-03-08T08:49:31+00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 1000
toc: true
---

This Document documents the types introduced by the Prometheus Operator to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.`
)

func toSectionLink(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "-", -1)
	return name
}

func printAPIDocs(paths []string) {
	fmt.Println(firstParagraph)

	typeSetUnion := make(model.TypeSet)
	typeSets := make([]model.TypeSet, 0, len(paths))
	for _, path := range paths {
		typeSet, err := model.Load(path)
		if err != nil {
			log.Fatal(err)
		}

		typeSets = append(typeSets, typeSet)
		for k, v := range typeSet {
			typeSetUnion[k] = v
		}
	}

	fmt.Printf("\n## Table of Contents\n")
	for _, typeSet := range typeSets {
		for _, key := range typeSet.SortedKeys() {
			t := typeSet[key]
			if len(t.Fields) == 0 {
				continue
			}

			fmt.Printf("* [%s](#%s)\n", t.Name, toSectionLink(t.Name))
		}
	}

	for _, typeSet := range typeSets {
		for _, key := range typeSet.SortedKeys() {
			t := typeSet[key]
			if len(t.Fields) == 0 {
				continue
			}

			fmt.Printf("\n## %s\n\n%s\n\n", t.Name, t.Description())
			backlinks := getBacklinks(t, typeSetUnion)
			if len(backlinks) > 0 {
				fmt.Printf("\n<em>appears in: %s</em>\n\n", strings.Join(backlinks, ", "))
			}

			fmt.Println("| Field | Description | Scheme | Required |")
			fmt.Println("| ----- | ----------- | ------ | -------- |")
			for _, f := range t.Fields {
				fmt.Println("|", f.Name(), "|", f.Description(), "|", f.TypeLink(typeSetUnion), "|", f.IsRequired(), "|")
			}
			fmt.Println("")
			fmt.Println("[Back to TOC](#table-of-contents)")
		}
	}
}

func getBacklinks(t *model.StructType, typeSet model.TypeSet) []string {
	appearsIn := make(map[string]struct{})
	for _, v := range typeSet {
		if v.IsOnlyEmbedded() {
			continue
		}

		for _, f := range v.Fields {
			if f.TypeName() == t.Name {
				appearsIn[v.Name] = struct{}{}
			}
		}
	}

	var backlinks []string
	for item := range appearsIn {
		link := fmt.Sprintf("[%s](#%s)", item, toSectionLink(item))
		backlinks = append(backlinks, link)
	}
	sort.Strings(backlinks)

	return backlinks
}
