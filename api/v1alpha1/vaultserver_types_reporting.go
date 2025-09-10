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

// #region ReportingSpec
// Resource Specification Specific to Reporting

type ReportingLicenseSpec struct {
	Enabled               bool  `json:"enabled,omitempty"`
	BillingStartTimestamp int32 `json:"billingStartTimestamp,omitempty"`
	DevelopmentCluster    bool  `json:"developmentCluster,omitempty"`
}

type ReportingSpec struct {
	SnapshotRetentionTime        time.Duration         `json:"snapshotRetentionTime,omitempty"`
	DisableProductUsageReporting bool                  `json:"disableProductUsageReporting,omitempty"`
	License                      *ReportingLicenseSpec `json:"license,omitempty"`
}

// #endregion
