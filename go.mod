module github.com/coreos/prometheus-operator

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20190425155809-e8bd32d46b3d
	github.com/campoy/embedmd v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-kit/kit v0.9.0
	github.com/go-openapi/swag v0.19.5
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/go-version v1.1.0
	github.com/jsonnet-bundler/jsonnet-bundler v0.2.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.0.0
	github.com/openshift/prom-label-proxy v0.1.1-0.20191016113035-b8153a7f39f1
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/prometheus v2.3.2+incompatible
	github.com/stretchr/testify v1.4.0
	github.com/thanos-io/thanos v0.10.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.17.3
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-tools v0.2.4
)

replace (
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.0.0-20190818123050-43acd0e2e93f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
)
