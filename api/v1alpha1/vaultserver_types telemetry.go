/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package v1alpha1

import (
	"bytes"
	"context"
	"maps"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const telemetryHclTemplate = `
telemetry {
    {{ range $key,$val := .MapValue }}
    {{ $key }} = {{ $val }}
    {{ end}}
}
`

type TelemetrySpec struct {
	DisableHostname               *bool            `json:"disableHostname,omitempty" hcl:"disable_hostname"`
	EnableHostnameLabel           *bool            `json:"enableHostnameLabel,omitempty" hcl:"enable_hostname_label"`
	MetricsPrefix                 *string          `json:"metricsPrefix,omitempty" hcl:"metrics_prefix"`
	LeaseMetricsEpsilon           *metav1.Duration `json:"leaseMetricsEpsilon,omitempty" hcl:"lease_metrics_epsilon"`
	NumLeaseMetricsTimeBuckets    *int32           `json:"numLeaseMetricsBuckets,omitempty" hcl:"num_lease_metrics_buckets"`
	AddLeaseMetricsNamespaceLabel *bool            `json:"addLeaseMetricsNamespaceLabel,omitempty" hcl:"add_lease_metrics_namespace_labels"`
	AddMountPointRollbackMetrics  *bool            `json:"addMountPointRollbackMetrics,omitempty" hcl:"add_mount_point_rollback_metrics"`
	FilterDefault                 *bool            `json:"filterDefault,omitempty" hcl:"filter_default"`
	PrefixFilter                  []string         `json:"prefixFilter,omitempty" hcl:"prefix_filter"`
	MaximumGaugeCardinality       *int32           `json:"maximumGaugeCardinality,omitempty" hcl:"maximum_gauge_cardinality"`
	UsageGaugePeriod              *metav1.Duration `json:"usageGaugePeriod,omitempty" hcl:"usage_gauge_period"`

	Statsite    *StatSiteSpec    `json:"statsite,omitempty"`
	Statsd      *StatsdSpec      `json:"statsd,omitempty"`
	Circonus    *CirconusSpec    `json:"circonus,omitempty"`
	DogStatsD   *DogStatsDSpec   `json:"dogStatsD,omitempty"`
	StackDriver *StackDriverSpec `json:"stackDriver,omitempty"`
	Prometheus  *PrometheusSpec  `json:"prometheus,omitempty"`
}

func (s *TelemetrySpec) Volumes(c *client.Client, ctx context.Context, vs *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	for _, cb := range s.internalTelemetry() {
		if vol, mnt, err := cb.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}
	return volumes, mounts, nil
}

func (s *TelemetrySpec) internalTelemetry() []ConfigBuilderHelper {
	array := []ConfigBuilderHelper{}
	if s.Statsite != nil {
		array = append(array, s.Statsite)
	}
	if s.Statsd != nil {
		array = append(array, s.Statsd)
	}
	if s.Circonus != nil {
		array = append(array, s.Circonus)
	}
	if s.DogStatsD != nil {
		array = append(array, s.DogStatsD)
	}
	if s.StackDriver != nil {
		array = append(array, s.StackDriver)
	}
	if s.Prometheus != nil {
		array = append(array, s.Prometheus)
	}
	return array
}

func (s *TelemetrySpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	for _, telemetryOutput := range s.internalTelemetry() {
		if _s, err := telemetryOutput.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}

	return _m, nil
}

func (s *TelemetrySpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	for _, cb := range s.internalTelemetry() {
		if err := cb.Secrets(c, ctx, vaultServer); err != nil {
			return err
		}
	}
	return nil
}

type StatSiteSpec struct {
	Addr *string `json:"address,omitempty" hcl:"statsite_address"`
}

// Secrets implements ConfigBuilderHelper.
func (s StatSiteSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s StatSiteSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StatSiteSpec) Type() string {
	return "statsite"
}

func (s *StatSiteSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StatsdSpec struct {
	Addr *string `json:"address,omitempty" hcl:"statsd_address"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StatsdSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StatsdSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StatsdSpec) Type() string {
	return "statsd"
}

func (s *StatsdSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type CirconusSpec struct {
	ApiToken                   *string          `json:"-" hcl:"circonus_api_token"`
	ApiApp                     *string          `json:"-" hcl:"circonus_api_app"`
	Credentials                SecretSelector   `json:"credentials"`
	ApiURL                     *string          `json:"apiUrl" hcl:"circonus_api_url"`
	SubmissionInterval         *metav1.Duration `json:"submissionInterval" hcl:"circonus_submission_interval"`
	SubmissionURL              *string          `json:"submissionUrl" hcl:"circonus_submission_url"`
	CheckID                    *string          `json:"checkId" hcl:"circonus_check_id"`
	CheckForceMetricActivation *bool            `json:"checkForceMetricActivation" hcl:"circonus_check_force_metric_activation"`
	CheckInstanceID            *string          `json:"checkInstanceId" hcl:"circonus_check_instance_id"`
	CheckSearchTag             *string          `json:"checkSearchTag" hcl:"circonus_check_search_tag"`
	CheckTags                  []string         `json:"checkTags"`
	CheckDisplayName           *string          `json:"checkDisplayName" hcl:"circonus_check_display_name"`
	BrokerID                   *string          `json:"brokerId" hcl:"circonus_broker_id"`
	BrokerSelectTag            *string          `json:"brokerSelectTag" hcl:"circonus_broker_select_tag"`
}

func (s *CirconusSpec) Type() string {
	return "circonus"
}

func (s *CirconusSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"api_token": {Dest: s.ApiToken, Mandatory: true},
		"api_app":   {Dest: s.ApiApp},
	}
	return secretMappings.Apply(&secret, "SealTransitSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *CirconusSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *CirconusSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if s.CheckTags != nil {
		_m["circonus_check_tags"] = strconv.Quote(strings.Join(s.CheckTags, ","))
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type DogStatsDSpec struct {
	Addr *string  `json:"addr,omitempty" hcl:"dogstatsd_addr"`
	Tags []string `json:"tags,omitempty" hcl:"dogstatsd_tags"`
}

// Secrets implements ConfigBuilderHelper.
func (s *DogStatsDSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *DogStatsDSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *DogStatsDSpec) Type() string {
	return "dogstatsd"
}

func (s *DogStatsDSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type PrometheusSpec struct {
	RetentionTime   *metav1.Duration `json:"retentionTime,omitempty" hcl:"prometheus_retention_time"`
	DisableHostname *bool            `json:"disableHostname,omitempty" hcl:"disable_hostname"`
}

// Secrets implements ConfigBuilderHelper.
func (s *PrometheusSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *PrometheusSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *PrometheusSpec) Type() string {
	return "prometheus"
}

func (s *PrometheusSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StackDriverSpec struct {
	ProjectID *string `json:"projectId,omitempty" hcl:"stackdriver_project_id"`
	Location  *string `json:"location,omitempty" hcl:"stackdriver_location"`
	Namespace *string `json:"namespace,omitempty" hcl:"stackdriver_namespace"`
	DebugLogs *bool   `json:"debugLogs,omitempty" hcl:"stackdriver_debug_logs"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StackDriverSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StackDriverSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StackDriverSpec) Type() string {
	return "stackdriver"
}

func (s *StackDriverSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

func (t *TelemetrySpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(telemetryHclTemplate))
	if err := template.Execute(&buf, t); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}
