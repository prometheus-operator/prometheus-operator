module github.com/coreos/prometheus-operator

go 1.12

require (
	github.com/ant31/crd-validation v0.0.0-20180702145049-30f8a35d0ac2
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20190425155809-e8bd32d46b3d
	github.com/campoy/embedmd v1.0.0
	github.com/client9/misspell v0.3.4
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815 // indirect
	github.com/emicklei/go-restful v2.6.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/go-openapi/spec v0.19.2
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/go-version v1.1.0
	github.com/jsonnet-bundler/jsonnet-bundler v0.1.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/prometheus v1.8.2-0.20190819201610-48b2c9c8eae2
	github.com/raviqqe/liche v0.0.0-20181124191719-2a2e6e56f6c6
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/thanos-io/thanos v0.7.0
	github.com/valyala/fasthttp v1.5.0 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/russross/blackfriday.v2 v2.0.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190813020757-36bff7324fb7
	k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery v0.0.0-20190809020650-423f5d784010
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/klog v0.4.0
	k8s.io/kube-openapi v0.0.0-20190722073852-5e22f3d471e6
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/klog => k8s.io/klog v0.3.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
