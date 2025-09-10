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
	"time"
)

// #region TelemetrySpec
// Resource Specification Specific to Telemetry

type TelemetrySpec struct {
	DisableHostname                  bool          `json:"disableHostname,omitempty"`
	EnableHostnameLabel              bool          `json:"enable_hostname_label,omitempty"`
	MetricsPrefix                    string        `json:"metrics_prefix,omitempty"`
	LeaseMetricsEpsilon              time.Duration `json:"lease_metrics_epsilon,omitempty"`
	NumLeaseMetricsTimeBuckets       int32         `json:"num_lease_metrics_buckets,omitempty"`
	LeaseMetricsNameSpaceLabels      bool          `json:"add_lease_metrics_namespace_label,omitempty"`
	RollbackMetricsIncludeMountPoint bool          `json:"add_mount_point_rollback_metrics,omitempty"`
	FilterDefault                    bool          `json:"filter_default,omitempty"`
	PrefixFilter                     []string      `json:"prefix_filter,omitempty"`
	MaximumGaugeCardinality          int32         `json:"maximum_gauge_cardinality,omitempty"`
	UsageGaugePeriod                 time.Duration `json:"usage_gauge_period,omitempty"`

	Statsite    *StatSiteSpec    `json:"statsite,omitempty"`
	Statsd      *StatsdSpec      `json:"statsd,omitempty"`
	Circonus    *CirconusSpec    `json:"circonus,omitempty"`
	DogStatsD   *DogStatsDSpec   `json:"dogStatsD,omitempty"`
	StackDriver *StackDriverSpec `json:"stackDriver,omitempty"`
	Prometheus  *PrometheusSpec  `json:"prometheus,omitempty"`
}

type StatSiteSpec struct {
	Addr string `json:"address,omitempty"`
}

type StatsdSpec struct {
	Addr string `json:"address,omitempty"`
}

type CirconusSpec struct {
	Credentials                SecretSelector `json:"credentials"`
	ApiURL                     string         `json:"apiUrl"`
	SubmissionInterval         string         `json:"submissionInterval"`
	CheckSubmissionURL         string         `json:"submissionUrl"`
	CheckID                    string         `json:"checkId"`
	CheckForceMetricActivation string         `json:"checkForceMetricActivation"`
	CheckInstanceID            string         `json:"checkInstanceId"`
	CheckSearchTag             string         `json:"checkSearchTag"`
	CheckTags                  string         `json:"checkTags"`
	CheckDisplayName           string         `json:"checkDisplayName"`
	BrokerID                   string         `json:"brokerId"`
	BrokerSelectTag            string         `json:"brokerSelectTag"`
}

type DogStatsDSpec struct {
	Addr string   `json:"addr,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type PrometheusSpec struct {
	RetentionTime time.Duration `json:"retentionTime,omitempty"`
}

type StackDriverSpec struct {
	ProjectID string `json:"projectId,omitempty"`
	Location  string `json:"location,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	DebugLogs bool   `json:"debugLogs,omitempty"`
}
