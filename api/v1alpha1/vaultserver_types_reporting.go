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
)

const reportingHclTemplate = `
reporting {
    {{ range $key,$val := .MapValue }}
    {{ $key }} = {{ $val }}
    {{ end}}
    {{ if .License }}
    license {
        {{ range $key,$val := .License.MapValue }}
        {{ $key }} = {{ $val }}
        {{ end}}
    }
    {{ end}}
}
`

type ReportingLicenseSpec struct {
	Enabled               *bool  `json:"enabled,omitempty" hcl:"enabled"`
	BillingStartTimestamp *int32 `json:"billingStartTimestamp,omitempty" hcl:"billing_start_timestamp"`
	DevelopmentCluster    *bool  `json:"developmentCluster,omitempty" hcl:"development_cluster"`
}

func (r *ReportingLicenseSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*r); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type ReportingSpec struct {
	SnapshotRetentionTime        *metav1.Duration      `json:"snapshotRetentionTime,omitempty" hcl:"snapshot_retention_time"`
	DisableProductUsageReporting *bool                 `json:"disableProductUsageReporting,omitempty" hcl:"disable_product_usage_reporting"`
	License                      *ReportingLicenseSpec `json:"license,omitempty"`
}

func (r *ReportingSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*r); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

func (r *ReportingSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(reportingHclTemplate))
	if err := template.Execute(&buf, r); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

// #endregion
