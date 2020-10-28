module github.com/prometheus-operator/prometheus-operator

go 1.15

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/brancz/kube-rbac-proxy v0.7.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/evanphx/json-patch/v5 v5.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.10.0
	github.com/go-openapi/swag v0.19.10
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.2
	github.com/hashicorp/go-version v1.2.1
	github.com/kylelemons/godebug v1.1.0
	github.com/mitchellh/hashstructure v1.0.1-0.20200508175121-8fdbea448aa6
	github.com/oklog/run v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-community/prom-label-proxy v0.2.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.43.0
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/prometheus/prometheus v1.8.2-0.20201015110737-0a7fdd3b7696
	github.com/stretchr/testify v1.6.1
	github.com/thanos-io/thanos v0.16.0-rc.1
	golang.org/x/sync v0.0.0-20201008141435-b3e1573b7520
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.19.2
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.3.0
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
)

replace (
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => ./pkg/apis/monitoring
	// A replace directive is needed for k8s.io/client-go because Cortex (which
	// is an indirect dependency through Thanos) has a requirement on v12.0.0.
	k8s.io/client-go => k8s.io/client-go v0.19.2
	k8s.io/klog/v2 => github.com/simonpasquier/klog-gokit/v2 v2.0.1
)
