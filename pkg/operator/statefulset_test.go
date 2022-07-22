// Copyright 2022 The prometheus-operator Authors
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

package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestPrometheusCommandArgs(t *testing.T) {
	fooString := "foo"
	boolTrue := true
	int32val := int32(2)
	dur2s := promv1.Duration("2s")

	testMatrix := []struct {
		Object   PrometheusType
		Expected []string
	}{
		{
			Object: PrometheusServer{&promv1.Prometheus{
				Spec: promv1.PrometheusSpec{
					CommonPrometheusFields: promv1.CommonPrometheusFields{
						Version:        "2.33.0",
						LogLevel:       "warning",
						LogFormat:      "wutheva",
						EnableFeatures: []string{},
						ExternalURL:    "http://external.url",
						RoutePrefix:    "prefix",
						Web: &promv1.PrometheusWebSpec{
							PageTitle: &fooString,
						},
						EnableRemoteWriteReceiver: true,
						ListenLocal:               true,
						WALCompression:            &boolTrue,
					},
					EnableAdminAPI: true,
					Retention:      promv1.Duration("15d"),
					RetentionSize:  promv1.ByteSize("2GB"),
					//DisableCompaction: false,
					Rules: promv1.Rules{
						Alert: promv1.RulesAlert{
							ForOutageTolerance: "1",
							ForGracePeriod:     "1",
							ResendDelay:        "1",
						},
					},
					Query: &promv1.QuerySpec{
						LookbackDelta:  &fooString,
						MaxConcurrency: &int32val,
						MaxSamples:     &int32val,
						Timeout:        &dur2s,
					},
					AllowOverlappingBlocks: true,
				},
			}},
			Expected: []string{
				"--config.file=/etc/prometheus/config_out/prometheus.env.yaml",
				"--web.config.file=/etc/prometheus/web_config/web-config.yaml",
				"--web.console.templates=/etc/prometheus/consoles",
				"--web.console.libraries=/etc/prometheus/console_libraries",
				"--web.enable-lifecycle",
				"--web.page-title=foo",
				"--web.enable-admin-api",
				"--web.external-url=http://external.url",
				"--web.enable-remote-write-receiver",
				"--web.listen-address=127.0.0.1:9090",
				"--web.route-prefix=prefix",
				"--storage.tsdb.path=/prometheus",
				"--storage.tsdb.wal-compression",
				"--storage.tsdb.retention.time=15d",
				"--storage.tsdb.retention.size=2GB",
				"--storage.tsdb.allow-overlapping-blocks",
				"--log.level=warning",
				"--log.format=wutheva",
				"--query.lookback-delta=foo",
				"--query.max-concurrency=2",
				"--query.timeout=2s",
				"--query.max-samples=2",
				"--rules.alert.for-outage-tolerance=1",
				"--rules.alert.for-grace-period=1",
				"--rules.alert.resend-delay=1",
			},
		},
		{
			Object: PrometheusServer{&promv1.Prometheus{
				Spec: promv1.PrometheusSpec{
					CommonPrometheusFields: promv1.CommonPrometheusFields{
						Version:        "2.33.0",
						EnableFeatures: []string{},
					},
				},
			}},
			Expected: []string{
				"--config.file=/etc/prometheus/config_out/prometheus.env.yaml",
				"--web.route-prefix=/",
				"--storage.tsdb.path=/prometheus",
				"--storage.tsdb.retention.time=24h",
			},
		},
	}

	t.Run("Test correct Prometheus command agrument generation", func(t *testing.T) {
		for _, testCase := range testMatrix {
			actual, warns, err := MakePrometheusCommandArgs(testCase.Object)
			require.NoError(t, err)
			assert.Empty(t, warns)
			assert.ElementsMatch(t, testCase.Expected, actual)
		}
	})
}

func TestThanosCommandArgs(t *testing.T) {
	dur2s := promv1.Duration("2s")
	testMatrix := []struct {
		Object   PrometheusType
		Config   Config
		Expected []string
	}{
		{
			Object: PrometheusServer{&promv1.Prometheus{
				Spec: promv1.PrometheusSpec{
					CommonPrometheusFields: promv1.CommonPrometheusFields{
						Version:        "2.33.0",
						EnableFeatures: []string{},
						RoutePrefix:    "/prefix",
						Web: &promv1.PrometheusWebSpec{
							WebConfigFileFields: promv1.WebConfigFileFields{
								TLSConfig: &promv1.WebTLSConfig{},
							},
						},
					},
					Thanos: &promv1.ThanosSpec{
						ObjectStorageConfig: &v1.SecretKeySelector{},
						ListenLocal:         true,
						TracingConfig:       &v1.SecretKeySelector{},
						GRPCServerTLSConfig: &promv1.TLSConfig{
							CAFile:   "tls_ca_file",
							CertFile: "tls_cert_file",
							KeyFile:  "tls_key_file",
						},
						LogLevel:     "warning",
						LogFormat:    "wutheva",
						MinTime:      "2h",
						ReadyTimeout: dur2s,
					},
				},
			}},
			Config: Config{
				LocalHost: "local",
			},
			Expected: []string{
				"sidecar",
				"--prometheus.url=https://local:9090/prefix",
				"--grpc-address=127.0.0.1:10901",
				"--http-address=127.0.0.1:10902",
				"--grpc-server-tls-cert=tls_cert_file",
				"--grpc-server-tls-key=tls_key_file",
				"--grpc-server-tls-client-ca=tls_ca_file",
				"--objstore.config=$(OBJSTORE_CONFIG)",
				"--tsdb.path=/prometheus",
				"--tracing.config=$(TRACING_CONFIG)",
				"--log.level=warning",
				"--log.format=wutheva",
				"--min-time=2h",
				"--prometheus.ready_timeout=2s",
			},
		},
		{
			Object: PrometheusServer{&promv1.Prometheus{
				Spec: promv1.PrometheusSpec{
					CommonPrometheusFields: promv1.CommonPrometheusFields{
						Version:   "2.33.0",
						LogLevel:  "fatal",
						LogFormat: "foobar",
					},
					Thanos: &promv1.ThanosSpec{},
				},
			}},
			Config: Config{
				LocalHost: "local",
			},
			Expected: []string{
				"sidecar",
				"--prometheus.url=http://local:9090/",
				"--log.level=fatal",
				"--log.format=foobar",
				"--grpc-address=:10901",
				"--http-address=:10902",
			},
		},
	}

	t.Run("Test correct Thanos command agrument generation", func(t *testing.T) {
		for _, testCase := range testMatrix {
			actual, warns, err := MakeThanosCommandArgs(testCase.Object, &testCase.Config)
			require.NoError(t, err)
			assert.Empty(t, warns)
			assert.ElementsMatch(t, testCase.Expected, actual)
		}
	})
}
