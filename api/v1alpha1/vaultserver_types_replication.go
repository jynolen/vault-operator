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
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kythe.io/kythe/go/util/datasize"
)

const replicationHclTemplate = `
replication {
    {{ range $key,$val := .MapValue }}
    {{ $key }} = {{ $val }}
    {{ end}}
}
`

type ReplicationSpec struct {
	ResolverDiscoverServers               *bool            `json:"resolverDiscoverServers,omitempty" hcl:"resolver_discover_servers"`
	LogshipperBufferLength                *int32           `json:"logshipperBufferLength,omitempty" hcl:"logshipper_buffer_length"`
	LogshipperBufferSize                  *datasize.Size   `json:"logshipperBufferSize,omitempty" hcl:"logshipper_buffer_size"`
	AllowForwardingViaHeader              *bool            `json:"allowForwardingViaHeader,omitempty" hcl:"allow_forwarding_via_header"`
	BestEffortWalWaitDuration             *metav1.Duration `json:"bestEffortWalWaitDuration,omitempty" hcl:"best_effort_wal_wait_duration"`
	AllowForwardingViaToken               *string          `json:"allowForwardingViaToken,omitempty" hcl:"allow_forwarding_via_token"`
	ReplicationCanaryWriteIntervalSeconds *int32           `json:"replicationCanaryWriteIntervalSeconds,omitempty" hcl:"replication_canary_write_interval_seconds"`
}

func (r *ReplicationSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*r); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

func (r *ReplicationSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(replicationHclTemplate))
	if err := template.Execute(&buf, r); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}
