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

	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

func printCompatMatrixDocs() {
	fmt.Println(`---
title: "Compatibility"
description: "The Prometheus Operator supports a number of Kubernetes and Prometheus releases."
lead: ""
date: 2021-03-08T08:49:31+00:00
draft: false
images: []
menu:
  docs:
    parent: "operator"
weight: 200
toc: true
---

The Prometheus Operator supports a number of Kubernetes and Prometheus releases.

## Kubernetes

The Prometheus Operator uses client-go to communicate with Kubernetes clusters. The supported Kubernetes cluster version is determined by client-go. The compatibility matrix for client-go and Kubernetes clusters can be found [here](https://github.com/kubernetes/client-go#compatibility-matrix). All additional compatibility is only best effort, or happens to still/already be supported. The currently used client-go version is "v4.0.0-beta.0".

Due to the use of CustomResourceDefinitions Kubernetes >= v1.7.0 is required.

Due to the use of apiextensions.k8s.io/v1 CustomResourceDefinitions, prometheus-operator v0.39.0 onward requires Kubernetes >= v1.16.0.

## Prometheus

The versions of Prometheus compatible to be run with the Prometheus Operator are:`)
	fmt.Println("")

	for _, v := range operator.PrometheusCompatibilityMatrix {
		fmt.Printf("* %s\n", v)
	}

	fmt.Println()

	fmt.Println(`## Alertmanager

We only support Alertmanager v0.15 and above. Everything below v0.15 is on a
best effort basis.`)
}
