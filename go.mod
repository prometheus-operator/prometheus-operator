module github.com/coreos/prometheus-operator

go 1.12

require (
	github.com/ant31/crd-validation v0.0.0-20180702145049-30f8a35d0ac2
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20190425155809-e8bd32d46b3d
	github.com/campoy/embedmd v1.0.0
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/go-openapi/spec v0.19.2
	github.com/go-openapi/swag v0.19.4
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/go-version v1.1.0
	github.com/improbable-eng/thanos v0.3.2
	github.com/jsonnet-bundler/jsonnet-bundler v0.1.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.0.0
	github.com/openshift/prom-label-proxy v0.1.1-0.20191016113035-b8153a7f39f1
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/prometheus v2.3.2+incompatible
	github.com/prometheus/tsdb v0.8.0 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/klog v0.4.0
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)

replace (
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.0.0-20190818123050-43acd0e2e93f
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/klog => k8s.io/klog v0.3.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
)
