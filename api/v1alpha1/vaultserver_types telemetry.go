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
	"fmt"
	"maps"
	"reflect"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// #region TelemetrySpec
// Resource Specification Specific to Telemetry

type TelemetrySpec struct {
	DisableHostname               *bool            `json:"disableHostname,omitempty"`
	EnableHostnameLabel           *bool            `json:"enableHostnameLabel,omitempty"`
	MetricsPrefix                 *string          `json:"metricsPrefix,omitempty"`
	LeaseMetricsEpsilon           *metav1.Duration `json:"leaseMetricsEpsilon,omitempty"`
	NumLeaseMetricsTimeBuckets    *int32           `json:"numLeaseMetricsBuckets,omitempty"`
	AddLeaseMetricsNamespaceLabel *bool            `json:"addLeaseMetricsNamespaceLabel,omitempty"`
	AddMountPointRollbackMetrics  *bool            `json:"addMountPointRollbackMetrics,omitempty"`
	FilterDefault                 *bool            `json:"filterDefault,omitempty"`
	PrefixFilter                  []string         `json:"prefixFilter,omitempty"`
	MaximumGaugeCardinality       *int32           `json:"maximumGaugeCardinality,omitempty"`
	UsageGaugePeriod              *metav1.Duration `json:"usageGaugePeriod,omitempty"`

	Statsite    *StatSiteSpec    `json:"statsite,omitempty"`
	Statsd      *StatsdSpec      `json:"statsd,omitempty"`
	Circonus    *CirconusSpec    `json:"circonus,omitempty"`
	DogStatsD   *DogStatsDSpec   `json:"dogStatsD,omitempty"`
	StackDriver *StackDriverSpec `json:"stackDriver,omitempty"`
	Prometheus  *PrometheusSpec  `json:"prometheus,omitempty"`
}

func (s *TelemetrySpec) InternalTelemetry() []HclHelper {
	v, t := reflect.ValueOf(*s), reflect.TypeOf(*s)
	reSlice := make([]HclHelper, 0)

	for i := range t.NumField() {
		intf := v.Field(i)
		if !intf.IsNil() {
			reSlice = append(reSlice, intf.Interface().(HclHelper))
		}
	}
	return reSlice
}

func (s *TelemetrySpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.UsageGaugePeriod != nil {
		_m["usage_gauge_period"] = strconv.Quote(fmt.Sprintf("%s", s.UsageGaugePeriod.Duration))
	}
	if s.MaximumGaugeCardinality != nil {
		_m["maximum_gauge_cardinality"] = strconv.FormatInt(int64(*s.MaximumGaugeCardinality), 10)
	}
	if s.DisableHostname != nil {
		_m["disable_hostname"] = strconv.FormatBool(*s.DisableHostname)
	}
	if s.EnableHostnameLabel != nil {
		_m["enable_hostname_label"] = strconv.FormatBool(*s.EnableHostnameLabel)
	}
	if s.MetricsPrefix != nil {
		_m["metrics_prefix"] = strconv.Quote(*s.MetricsPrefix)
	}
	if s.LeaseMetricsEpsilon != nil {
		_m["lease_metrics_epsilon"] = strconv.Quote(fmt.Sprintf("%s", s.LeaseMetricsEpsilon.Duration))
	}
	if s.NumLeaseMetricsTimeBuckets != nil {
		_m["num_lease_metrics_buckets"] = strconv.FormatInt(int64(*s.NumLeaseMetricsTimeBuckets), 10)
	}
	if s.AddLeaseMetricsNamespaceLabel != nil {
		_m["add_lease_metrics_namespace_labels"] = strconv.FormatBool(*s.AddLeaseMetricsNamespaceLabel)
	}
	if s.AddMountPointRollbackMetrics != nil {
		_m["add_mount_point_rollback_metrics"] = strconv.FormatBool(*s.AddMountPointRollbackMetrics)
	}
	if s.FilterDefault != nil {
		_m["filter_default"] = strconv.FormatBool(*s.FilterDefault)
	}
	if len(s.PrefixFilter) > 0 {
		_m["prefix_filter"] = fmt.Sprintf("[%s]", strings.Join(s.PrefixFilter, ","))
	}

	for _, telemetryOutput := range s.InternalTelemetry() {
		maps.Copy(_m, telemetryOutput.MapValue())
	}

	return _m
}

type StatSiteSpec struct {
	Addr *string `json:"address,omitempty"`
}

func (s *StatSiteSpec) Type() string {
	return "statsite"
}

func (s *StatSiteSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Addr != nil {
		_m["statsite_address"] = strconv.Quote(*s.Addr)
	}
	return _m
}

type StatsdSpec struct {
	Addr *string `json:"address,omitempty"`
}

func (s *StatsdSpec) Type() string {
	return "statsd"
}

func (s *StatsdSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Addr != nil {
		_m["statsd_address"] = strconv.Quote(*s.Addr)
	}
	return _m
}

type CirconusSpec struct {
	ApiToken                   *string
	ApiApp                     *string
	Credentials                SecretSelector   `json:"credentials"`
	ApiURL                     *string          `json:"apiUrl"`
	SubmissionInterval         *metav1.Duration `json:"submissionInterval"`
	SubmissionURL              *string          `json:"submissionUrl"`
	CheckID                    *string          `json:"checkId"`
	CheckForceMetricActivation *bool            `json:"checkForceMetricActivation"`
	CheckInstanceID            *string          `json:"checkInstanceId"`
	CheckSearchTag             *string          `json:"checkSearchTag"`
	CheckTags                  []string         `json:"checkTags"`
	CheckDisplayName           *string          `json:"checkDisplayName"`
	BrokerID                   *string          `json:"brokerId"`
	BrokerSelectTag            *string          `json:"brokerSelectTag"`
}

func (s *CirconusSpec) Type() string {
	return "statsd"
}

func (s *CirconusSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.ApiURL != nil {
		_m["circonus_api_token"] = strconv.Quote(*s.ApiToken)
	}
	if s.ApiApp != nil {
		_m["circonus_api_app"] = strconv.Quote(*s.ApiApp)
	}
	if s.ApiURL != nil {
		_m["circonus_api_url"] = strconv.Quote(*s.ApiURL)
	}
	if s.ApiURL != nil {
		_m["circonus_api_url"] = strconv.Quote(*s.ApiURL)
	}
	if s.SubmissionInterval != nil {
		_m["circonus_submission_interval"] = strconv.Quote(fmt.Sprintf("%d", s.SubmissionInterval))
	}
	if s.SubmissionURL != nil {
		_m["circonus_submission_url"] = strconv.Quote(*s.SubmissionURL)
	}
	if s.CheckID != nil {
		_m["circonus_check_id"] = strconv.Quote(*s.CheckID)
	}
	if s.CheckForceMetricActivation != nil {
		_m["circonus_check_force_metric_activation"] = strconv.FormatBool(*s.CheckForceMetricActivation)
	}
	if s.CheckInstanceID != nil {
		_m["circonus_check_instance_id"] = strconv.Quote(*s.CheckInstanceID)
	}
	if s.CheckSearchTag != nil {
		_m["circonus_check_search_tag"] = strconv.Quote(*s.CheckSearchTag)
	}
	if s.CheckDisplayName != nil {
		_m["circonus_check_display_name"] = strconv.Quote(*s.CheckDisplayName)
	}
	if s.CheckTags != nil {
		_m["circonus_check_tags"] = strconv.Quote(strings.Join(s.CheckTags, ","))
	}
	if s.BrokerID != nil {
		_m["circonus_broker_id"] = strconv.Quote(*s.BrokerID)
	}
	if s.BrokerSelectTag != nil {
		_m["circonus_broker_select_tag"] = strconv.Quote(*s.BrokerSelectTag)
	}
	return _m
}

type DogStatsDSpec struct {
	Addr *string  `json:"addr,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

func (s *DogStatsDSpec) Type() string {
	return "dogstatsd"
}

func (s *DogStatsDSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Addr != nil {
		_m["dogstatsd_addr"] = strconv.Quote(*s.Addr)
	}
	if len(s.Tags) > 0 {
		_m["dogstatsd_tags"] = fmt.Sprintf("[%s]", strings.Join(s.Tags, ","))
	}
	return _m
}

type PrometheusSpec struct {
	RetentionTime   *metav1.Duration `json:"retentionTime,omitempty"`
	DisableHostname *bool            `json:"disableHostname,omitempty"`
}

func (s *PrometheusSpec) Type() string {
	return "prometheus"
}

func (s *PrometheusSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.RetentionTime != nil {
		_m["prometheus_retention_time"] = strconv.Quote(fmt.Sprintf("%s", s.RetentionTime.Duration))
	}
	if s.DisableHostname != nil {
		_m["disable_hostname"] = strconv.FormatBool(*s.DisableHostname)
	}
	return _m
}

type StackDriverSpec struct {
	ProjectID *string `json:"projectId,omitempty"`
	Location  *string `json:"location,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
	DebugLogs *bool   `json:"debugLogs,omitempty"`
}

func (s *StackDriverSpec) Type() string {
	return "stackdriver"
}

func (s *StackDriverSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.ProjectID != nil {
		_m["stackdriver_project_id"] = strconv.Quote(*s.ProjectID)
	}
	if s.Location != nil {
		_m["stackdriver_location"] = strconv.Quote(*s.Location)
	}
	if s.Namespace != nil {
		_m["stackdriver_namespace"] = strconv.Quote(*s.Namespace)
	}
	if s.DebugLogs != nil {
		_m["stackdriver_debug_logs"] = strconv.FormatBool(*s.DebugLogs)
	}
	return _m
}
