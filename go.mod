module github.com/coreos/prometheus-operator

go 1.12

require (
	github.com/ant31/crd-validation v0.0.0-20180702145049-30f8a35d0ac2
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20190425155809-e8bd32d46b3d
	github.com/campoy/embedmd v1.0.0
	github.com/emicklei/go-restful v2.6.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.8.0
	github.com/go-openapi/spec v0.17.2
	github.com/golang/protobuf v1.3.1
	github.com/googleapis/gnostic v0.1.0 // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/improbable-eng/thanos v0.5.0
	github.com/jsonnet-bundler/jsonnet-bundler v0.1.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/prometheus v2.9.2+incompatible
	github.com/stretchr/testify v1.3.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190425145619-16072639606e // indirect
	golang.org/x/text v0.3.1 // indirect
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2 // indirect
	golang.org/x/tools v0.0.0-20190425150028-36563e24a262 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v2.0.0-alpha.0.0.20181121191925-a47917edff34+incompatible
	k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/klog v0.3.1
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
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
