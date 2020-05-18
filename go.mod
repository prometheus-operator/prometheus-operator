module github.com/coreos/prometheus-operator

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20191212081931-bf2969bbd742
	github.com/brancz/kube-rbac-proxy v0.5.0
	github.com/campoy/embedmd v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.10.0
	github.com/go-openapi/swag v0.19.9
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.0
	github.com/hashicorp/go-version v1.2.0
	github.com/jsonnet-bundler/jsonnet-bundler v0.3.1
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.1.0
	github.com/openshift/prom-label-proxy v0.1.1-0.20191016113035-b8153a7f39f1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/prometheus v2.3.2+incompatible
	github.com/stretchr/testify v1.5.1
	github.com/thanos-io/thanos v0.11.0
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.2
	k8s.io/component-base v0.18.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-tools v0.2.4
)

replace (
	// Temporary until https://github.com/openshift/prom-label-proxy/pull/28 gets merged
	github.com/openshift/prom-label-proxy => github.com/vsliouniaev/prom-label-proxy v0.0.0-20200518104441-4fd7fe13454f
	// Prometheus 2.18.1
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.0.0-20200507164740-ecee9c8abfd1
	k8s.io/client-go => k8s.io/client-go v0.18.2
)
