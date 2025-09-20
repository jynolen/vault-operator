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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// #region ReportingSpec
// Resource Specification Specific to Reporting

type ReportingLicenseSpec struct {
	Enabled               *bool  `json:"enabled,omitempty"`
	BillingStartTimestamp *int32 `json:"billingStartTimestamp,omitempty"`
	DevelopmentCluster    *bool  `json:"developmentCluster,omitempty"`
}

func (r *ReportingLicenseSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if r.DevelopmentCluster != nil {
		_m["development_cluster"] = strconv.FormatBool(*r.DevelopmentCluster)
	}
	if r.Enabled != nil {
		_m["enabled"] = strconv.FormatBool(*r.Enabled)
	}
	if r.BillingStartTimestamp != nil {
		_m["billing_start_timestamp"] = strconv.FormatInt(int64(*r.BillingStartTimestamp), 10)
	}
	return _m
}

type ReportingSpec struct {
	SnapshotRetentionTime        *metav1.Duration      `json:"snapshotRetentionTime,omitempty"`
	DisableProductUsageReporting *bool                 `json:"disableProductUsageReporting,omitempty"`
	License                      *ReportingLicenseSpec `json:"license,omitempty"`
}

func (r *ReportingSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if r.SnapshotRetentionTime != nil {
		_m["snapshot_retention_time"] = strconv.Quote(fmt.Sprintf("%s", r.SnapshotRetentionTime.Duration))
	}
	if r.DisableProductUsageReporting != nil {
		_m["disable_product_usage_reporting"] = strconv.FormatBool(*r.DisableProductUsageReporting)
	}
	return _m
}

// #endregion
