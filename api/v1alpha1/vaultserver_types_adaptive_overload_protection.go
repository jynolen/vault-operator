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
)

const adaptiveOverloadProtectionHclTemplate = `
adaptive_overload_protection {
    {{ range $key,$val := .AdaptiveOverloadProtection }}
    {{ $key }} = {{ $val }}
    {{ end}}
}`

type AdaptiveOverloadProtectionSpec struct {
	DisableWriteController *bool `json:"disableWriteController,omitempty" hcl:"disable_write_controller"`
}

func (a *AdaptiveOverloadProtectionSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*a); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

func (a *AdaptiveOverloadProtectionSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(adaptiveOverloadProtectionHclTemplate))
	if err := template.Execute(&buf, a); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}
