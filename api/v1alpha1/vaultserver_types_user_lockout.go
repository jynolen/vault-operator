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
	"errors"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const userLockoutHclTemplate = `
{{ range $_, $val := . }}
user_lockout "{{ $val.Type }}"{
    {{ range $k,$v := $val }}
    {{ $k }} = {{ $v }}
    {{ end}}
}
{{ end }}
`

type InternalUserLockoutSpec struct {
	Threshold      *int32           `json:"threshold,omitempty" hcl:"lockout_threshold"`
	Duration       *metav1.Duration `json:"duration,omitempty" hcl:"lockout_duration"`
	CounterReset   *metav1.Duration `json:"counterReset,omitempty" hcl:"lockout_counter_reset"`
	DisableLockout *bool            `json:"disableLockout,omitempty" hcl:"disable_lockout"`
}

func (s *InternalUserLockoutSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type UserLockoutList []UserLockoutSpec

func (s *UserLockoutList) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(userLockoutHclTemplate))
	if err := template.Execute(&buf, s); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

type UserLockoutSpec struct {
	All      *InternalUserLockoutSpec `json:"all,omitempty"`
	UserPass *InternalUserLockoutSpec `json:"userpass,omitempty"`
	LDAP     *InternalUserLockoutSpec `json:"ldap,omitempty"`
	AppRole  *InternalUserLockoutSpec `json:"appRole,omitempty"`
}

func (s *UserLockoutSpec) Type() string {
	if s.All != nil {
		return "all"
	}
	if s.UserPass != nil {
		return "userpass"
	}
	if s.LDAP != nil {
		return "ldap"
	}
	if s.AppRole != nil {
		return "approle"
	}
	return ""
}

func (s *UserLockoutSpec) MapValue() (map[string]any, error) {
	if s.All != nil {
		return s.All.MapValue()
	}
	if s.UserPass != nil {
		return s.UserPass.MapValue()
	}
	if s.LDAP != nil {
		return s.LDAP.MapValue()
	}
	if s.AppRole != nil {
		return s.AppRole.MapValue()
	}
	return nil, errors.New("Undefined UserLockout")
}
