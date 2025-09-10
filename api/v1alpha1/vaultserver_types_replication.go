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

	"k8s.io/apimachinery/pkg/api/resource"
)

// #region ReplicationSpec
// Resource Specification Specific to Replication

type ReplicationSpec struct {
	ResolverDiscoverServers   bool              `json:"resolverDiscoverServers,omitempty"`
	LogshipperBufferLength    int32             `json:"logshipperBufferLength,omitempty"`
	LogshipperBufferSize      resource.Quantity `json:"logshipperBufferSize,omitempty"`
	AllowForwardingViaHeader  bool              `json:"allowForwardingViaHeader,omitempty"`
	BestEffortWalWaitDuration time.Duration     `json:"bestEffortWalWaitDuratione,omitempty"`
}

// #endregion
