# glog

This repo contains a package that exposes an API subset of the [glog](https://github.com/golang/glog) package.
All logging state delivered to this package is shunted to the global [klog logger](https://github.com/kubernetes/klog).

This package makes it so we can intercept the calls to glog and redirect them to klog and thus produce
a consistent log for our processes.

This code was inspired by https://github.com/istio/glog/
