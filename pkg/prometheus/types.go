package prometheus

import (
	"fmt"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	//"github.com/pkg/errors"

	//v1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/operator"
)

const (
	resourcePrefix = "prometheus"
	labelName      = "prometheus.io/prometheus-name"
)

var (
	boolFalse = false
	boolTrue  = true

	int32Zero             int32
	minReplicas           int32 = 1
	defaultMaxConcurrency int32 = 20
	probeTimeoutSeconds   int32 = 3
)

// PrometheusServer objects are used to run Prometheus instances in server mode.
// Implements operator.PrometheusType interface
type PrometheusServer struct {
	*monitoringv1.Prometheus
}

// GetNameNomenclator implements PrometheusType interface
func (p PrometheusServer) GetNomenclator() *operator.Nomenclator {
	return operator.NewNomenclator(p.Kind, resourcePrefix, p.Name, labelName, p.Spec.Shards)
}

// GetCommonFields implements PrometheusType interface
func (p PrometheusServer) GetCommonFields() *monitoringv1.CommonPrometheusFields {
	return p.Spec.GetCommonPrometheusFields()
}

// DisableCompaction implements PrometheusType interface
func (p PrometheusServer) DisableCompaction() bool {
	if p.Spec.Thanos != nil {
		if p.Spec.Thanos.ObjectStorageConfig != nil || p.Spec.Thanos.ObjectStorageConfigFile != nil {
			// NOTE(bwplotka): As described in https://thanos.io/components/sidecar.md/ we have to turn off compaction of Prometheus
			// to avoid races during upload, if the uploads are configured.
			return true
		}
	}
	return p.Spec.DisableCompaction
}

// MakeCommandArgs returns map of command line arguments for Prometheus server mode
func (p PrometheusServer) MakeCommandArgs() (map[string]string, []string, error) {
	warns := []string{}
	args := map[string]string{}

	version, err := operator.ParseVersion(p.Spec.Version)
	if err != nil {
		return args, warns, err
	}

	args["config.file"] = path.Join(operator.PrometheusConfOutDir, operator.PrometheusConfEnvSubstFilename)

	args["storage.tsdb.path"] = operator.PrometheusStorageDir
	retentionTimeFlag := "storage.tsdb.retention"
	if version.GTE(semver.MustParse("2.7.0")) {
		retentionTimeFlag = "storage.tsdb.retention.time"
		if p.Spec.Retention == "" && p.Spec.RetentionSize == "" {
			args[retentionTimeFlag] = operator.DefaultRetention
		} else {
			if p.Spec.Retention != "" {
				args[retentionTimeFlag] = string(p.Spec.Retention)
			}

			if p.Spec.RetentionSize != "" {
				args["storage.tsdb.retention.size"] = string(p.Spec.RetentionSize)
			}
		}
	} else {
		if p.Spec.Retention == "" {
			args[retentionTimeFlag] = operator.DefaultRetention
		} else {
			args[retentionTimeFlag] = string(p.Spec.Retention)
		}
	}

	if p.Spec.Query != nil {
		if p.Spec.Query.LookbackDelta != nil {
			args["query.lookback-delta"] = *p.Spec.Query.LookbackDelta
		}

		if p.Spec.Query.MaxConcurrency != nil {
			if *p.Spec.Query.MaxConcurrency < 1 {
				p.Spec.Query.MaxConcurrency = &defaultMaxConcurrency
			}
			args["query.max-concurrency"] = fmt.Sprintf("%d", *p.Spec.Query.MaxConcurrency)
		}
		if p.Spec.Query.Timeout != nil {
			args["query.timeout"] = string(*p.Spec.Query.Timeout)
		}
		if version.Minor >= 5 {
			if p.Spec.Query.MaxSamples != nil {
				args["query.max-samples"] = fmt.Sprintf("%d", *p.Spec.Query.MaxSamples)
			}
		}
	}

	if version.Minor >= 4 {
		if p.Spec.Rules.Alert.ForOutageTolerance != "" {
			args["rules.alert.for-outage-tolerance"] = p.Spec.Rules.Alert.ForOutageTolerance
		}
		if p.Spec.Rules.Alert.ForGracePeriod != "" {
			args["rules.alert.for-grace-period"] = p.Spec.Rules.Alert.ForGracePeriod
		}
		if p.Spec.Rules.Alert.ResendDelay != "" {
			args["rules.alert.resend-delay"] = p.Spec.Rules.Alert.ResendDelay
		}
	}

	args["web.config.file"] = path.Join(operator.WebConfigDir, operator.WebConfigFilename)
	args["web.console.templates"] = operator.WebConsoleTemplatesDir
	args["web.console.libraries"] = operator.WebConsoleLibraryDir
	args["web.enable-lifecycle"] = ""
	if p.Spec.Web != nil {
		// TODO(simonpasquier): check that the Prometheus version supports the flag.
		if p.Spec.Web.PageTitle != nil {
			args["web.page-title"] = *p.Spec.Web.PageTitle
		}
	}

	if p.Spec.EnableAdminAPI {
		args["web.enable-admin-api"] = ""
	}

	if p.Spec.EnableRemoteWriteReceiver {
		if version.GTE(semver.MustParse("2.33.0")) {
			args["web.enable-remote-write-receiver"] = ""
		} else {
			msg := "ignoring 'enableRemoteWriteReceiver' supported by Prometheus v v2.33.0+"
			warns = append(warns, msg)
		}
	}

	if len(p.Spec.EnableFeatures) > 0 {
		args["enable-feature"] = strings.Join(p.Spec.EnableFeatures[:], ",")
	}

	if p.Spec.ExternalURL != "" {
		args["web.external-url"] = p.Spec.ExternalURL
	}

	webRoutePrefix := "/"
	if p.Spec.RoutePrefix != "" {
		webRoutePrefix = p.Spec.RoutePrefix
	}
	args["web.route-prefix"] = webRoutePrefix

	if p.Spec.LogLevel != "" && p.Spec.LogLevel != "info" {
		args["log.level"] = p.Spec.LogLevel
	}
	if version.GTE(semver.MustParse("2.6.0")) {
		if p.Spec.LogFormat != "" && p.Spec.LogFormat != "logfmt" {
			args["log.format"] = p.Spec.LogFormat
		}
	}

	if version.GTE(semver.MustParse("2.11.0")) && p.Spec.WALCompression != nil {
		if *p.Spec.WALCompression {
			args["storage.tsdb.wal-compression"] = ""
		} else {
			args["no-storage.tsdb.wal-compression"] = ""
		}
	}

	if version.GTE(semver.MustParse("2.8.0")) && p.Spec.AllowOverlappingBlocks {
		args["storage.tsdb.allow-overlapping-blocks"] = ""
	}

	if p.Spec.ListenLocal {
		args["web.listen-address"] = "127.0.0.1:9090"
	}

	if p.DisableCompaction() {
		args["storage.tsdb.max-block-duration"] = "2h"
		args["storage.tsdb.min-block-duration"] = "2h"
	}

	return args, warns, nil
}

// MakeThanosCommandArgs returns slice of Thanos command arguments for Thanos sidecar
func MakeThanosCommandArgs(p monitoringv1.Prometheus, c *operator.Config) (out, warns []string, err error) {
	thanos := p.Spec.Thanos
	uriScheme := operator.GetURIScheme(p.Spec.Web)

	prefix := "/"
	if p.Spec.RoutePrefix != "" {
		prefix = p.Spec.RoutePrefix
	}

	bindAddress := "" // Listen to all available IP addresses by default
	if thanos.ListenLocal {
		bindAddress = "127.0.0.1"
	}

	args := map[string]string{
		"prometheus.url": fmt.Sprintf("%s://%s:9090%s", uriScheme, c.LocalHost, path.Clean(prefix)),
		"grpc-address":   fmt.Sprintf("%s:%d", bindAddress, DefaultThanosGRPCPort),
		"http-address":   fmt.Sprintf("%s:%d", bindAddress, DefaultThanosHTTPPort),
	}

	if thanos.GRPCServerTLSConfig != nil {
		tls := thanos.GRPCServerTLSConfig
		if tls.CertFile != "" {
			args["grpc-server-tls-cert"] = tls.CertFile
		}
		if tls.KeyFile != "" {
			args["grpc-server-tls-key"] = tls.KeyFile
		}
		if tls.CAFile != "" {
			args["grpc-server-tls-client-ca"] = tls.CAFile
		}
	}

	if thanos.ObjectStorageConfig != nil || thanos.ObjectStorageConfigFile != nil {
		if thanos.ObjectStorageConfigFile != nil {
			args["objstore.config-file"] = *thanos.ObjectStorageConfigFile
		} else {
			args["objstore.config"] = fmt.Sprintf("$(%s)", ThanosObjStoreEnvVar)
		}
		args["tsdb.path"] = operator.PrometheusStorageDir
	}

	if thanos.TracingConfig != nil || len(thanos.TracingConfigFile) > 0 {
		traceConfig := fmt.Sprintf("$(%s)", ThanosTraceConfigEnvVar)
		if len(thanos.TracingConfigFile) > 0 {
			traceConfig = thanos.TracingConfigFile
		}
		args["tracing.config"] = traceConfig
	}

	logLevel := p.Spec.LogLevel
	if thanos.LogLevel != "" {
		logLevel = thanos.LogLevel
	}
	if logLevel != "" {
		args["log.level"] = logLevel
	}

	logFormat := p.Spec.LogFormat
	if thanos.LogFormat != "" {
		logFormat = thanos.LogFormat
	}
	if logFormat != "" {
		args["log.format"] = logFormat
	}

	if thanos.MinTime != "" {
		args["min-time"] = thanos.MinTime
	}

	if thanos.ReadyTimeout != "" {
		args["prometheus.ready_timeout"] = string(thanos.ReadyTimeout)
	}

	out, err = operator.ProcessCommandArgs(args, thanos.AdditionalArgs)
	if err != nil {
		return out, warns, err
	}

	out = append([]string{"sidecar"}, out...)
	return out, warns, err
}
