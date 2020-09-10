module github.com/prometheus-operator/prometheus-operator

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/kube-rbac-proxy v0.5.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.10.0
	github.com/go-openapi/swag v0.19.9
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.0
	github.com/hashicorp/go-version v1.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452
	github.com/oklog/run v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-community/prom-label-proxy v0.1.1-0.20200616110844-0fbfa11fa8f3
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/prometheus v1.8.2-0.20200907175821-8219b442c864
	github.com/stretchr/testify v1.5.1
	github.com/thanos-io/thanos v0.11.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/component-base v0.18.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66
)

replace github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => ./pkg/apis/monitoring
