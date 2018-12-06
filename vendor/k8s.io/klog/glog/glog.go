// Copyright 2017 Istio Authors
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

// Package glog exposes an API subset of the [glog](https://github.com/golang/glog) package.
// All logging state delivered to this package is shunted to the global [klog logger](ttps://github.com/kubernetes/klog).
//
// We depend on some downstream components that use glog for logging. This package makes it so we can intercept
// the calls to glog and redirect them to klog and thus produce a consistent log for our processes.
package glog

import (
	"k8s.io/klog"
)

// Level is a shim
type Level int32

// Verbose is a shim
type Verbose bool

// Flush is a shim
func Flush() {
}

// V is a shim
func V(level Level) Verbose {
	return Verbose(bool(klog.V(klog.Level(int32(level)))))
}

// Info is a shim
func (v Verbose) Info(args ...interface{}) {
	if v {
		klog.Info(args...)
	}
}

// Infoln is a shim
func (v Verbose) Infoln(args ...interface{}) {
	if v {
		klog.Infoln(args...)
	}
}

// Infof is a shim
func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		klog.Infof(format, args...)
	}
}

// Info is a shim
func Info(args ...interface{}) {
	klog.Info(args...)
}

// InfoDepth is a shim
func InfoDepth(depth int, args ...interface{}) {
	klog.InfoDepth(depth, args...)
}

// Infoln is a shim
func Infoln(args ...interface{}) {
	klog.Infoln(args...)
}

// Infof is a shim
func Infof(format string, args ...interface{}) {
	klog.Infof(format, args...)
}

// Warning is a shim
func Warning(args ...interface{}) {
	klog.Warning(args...)
}

// WarningDepth is a shim
func WarningDepth(depth int, args ...interface{}) {
	klog.WarningDepth(depth, args...)
}

// Warningln is a shim
func Warningln(args ...interface{}) {
	klog.Warningln(args...)
}

// Warningf is a shim
func Warningf(format string, args ...interface{}) {
	klog.Warningf(format, args...)
}

// Error is a shim
func Error(args ...interface{}) {
	klog.Error(args...)
}

// ErrorDepth is a shim
func ErrorDepth(depth int, args ...interface{}) {
	klog.ErrorDepth(depth, args...)
}

// Errorln is a shim
func Errorln(args ...interface{}) {
	klog.Errorln(args...)
}

// Errorf is a shim
func Errorf(format string, args ...interface{}) {
	klog.Errorf(format, args...)
}

// Fatal is a shim
func Fatal(args ...interface{}) {
	klog.Fatal(args...)
}

// FatalDepth is a shim
func FatalDepth(depth int, args ...interface{}) {
	klog.FatalDepth(depth, args...)
}

// Fatalln is a shim
func Fatalln(args ...interface{}) {
	klog.Fatalln(args...)
}

// Fatalf is a shim
func Fatalf(format string, args ...interface{}) {
	klog.Fatalf(format, args...)
}

// Exit is a shim
func Exit(args ...interface{}) {
	klog.Exit(args...)
}

// ExitDepth is a shim
func ExitDepth(depth int, args ...interface{}) {
	klog.ExitDepth(depth, args...)
}

// Exitln is a shim
func Exitln(args ...interface{}) {
	klog.Exitln(args...)
}

// Exitf is a shim
func Exitf(format string, args ...interface{}) {
	klog.Exitf(format, args...)
}
