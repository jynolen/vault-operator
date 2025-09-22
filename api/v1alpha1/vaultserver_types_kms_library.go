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

// Resource Specification Specific to KmsLibrary

const kmsHclTemplate = `
kms_library "{{ .Type }}" {
    {{ range $key,$val := .MapValue }}
    {{ $key }} = {{ $val }}
    {{ end}}
}
`

type KmsLibrarySpec struct {
	// +kubebuilder:validation:Enum=pkcs11
	Type    string  `json:"type,omitempty"`
	Name    *string `json:"name,omitempty" hcl:"name"`
	Library *string `json:"library,omitempty" hcl:"library"`
}

func (k *KmsLibrarySpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*k); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

func (k *KmsLibrarySpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(kmsHclTemplate))
	if err := template.Execute(&buf, k); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}
