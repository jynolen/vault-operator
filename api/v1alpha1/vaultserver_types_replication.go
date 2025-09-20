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
	"kythe.io/kythe/go/util/datasize"
)

// #region ReplicationSpec
// Resource Specification Specific to Replication

type ReplicationSpec struct {
	ResolverDiscoverServers               *bool            `json:"resolverDiscoverServers,omitempty"`
	LogshipperBufferLength                *int32           `json:"logshipperBufferLength,omitempty"`
	LogshipperBufferSize                  *datasize.Size   `json:"logshipperBufferSize,omitempty"`
	AllowForwardingViaHeader              *bool            `json:"allowForwardingViaHeader,omitempty"`
	BestEffortWalWaitDuration             *metav1.Duration `json:"bestEffortWalWaitDuratione,omitempty"`
	AllowForwardingViaToken               *string          `json:"allowForwardingViaToken,omitempty"`
	ReplicationCanaryWriteIntervalSeconds *int32           `json:"replicationCanaryWriteIntervalSeconds,omitempty"`
}

func (r *ReplicationSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if r.ResolverDiscoverServers != nil {
		_m["resolver_discover_servers"] = strconv.FormatBool(*r.ResolverDiscoverServers)
	}
	if r.LogshipperBufferLength != nil {
		_m["logshipper_buffer_length"] = strconv.FormatInt(int64(*r.LogshipperBufferLength), 10)
	}
	if r.LogshipperBufferSize != nil {
		_m["logshipper_buffer_size"] = strconv.Quote(fmt.Sprintf("%dkb", int(r.LogshipperBufferSize.Kilobytes())))
	}
	if r.AllowForwardingViaHeader != nil {
		_m["allow_forwarding_via_header"] = strconv.FormatBool(*r.AllowForwardingViaHeader)
	}
	if r.BestEffortWalWaitDuration != nil {
		_m["best_effort_wal_wait_duration"] = strconv.Quote(fmt.Sprintf("%s", r.BestEffortWalWaitDuration.Duration))
	}
	if r.AllowForwardingViaToken != nil {
		_m["allow_forwarding_via_token"] = strconv.Quote(*r.AllowForwardingViaToken)
	}
	if r.ReplicationCanaryWriteIntervalSeconds != nil {
		_m["replication_canary_write_interval_seconds"] = strconv.FormatInt(int64(*r.ReplicationCanaryWriteIntervalSeconds), 10)
	}
	return _m
}
