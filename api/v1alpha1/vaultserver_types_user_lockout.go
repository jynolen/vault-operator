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

type InternalUserLockoutSpec struct {
	Threshold      *int32           `json:"threshold,omitempty"`
	Duration       *metav1.Duration `json:"duration,omitempty"`
	CounterReset   *metav1.Duration `json:"counterReset,omitempty"`
	DisableLockout *bool            `json:"disableLockout,omitempty"`
}

func (s *InternalUserLockoutSpec) MapValue() map[string]any {
	m := map[string]any{}
	if s.Threshold != nil {
		m["lockout_threshold"] = strconv.Quote(strconv.FormatInt(int64(*s.Threshold), 10))
	}
	if s.Duration != nil {
		m["lockout_duration"] = fmt.Sprintf("%s", s.Duration.Duration)
	}
	if s.Duration != nil {
		m["lockout_counter_reset"] = fmt.Sprintf("%s", s.CounterReset.Duration)
	}
	if s.Duration != nil {
		m["disable_lockout"] = strconv.FormatBool(*s.DisableLockout)
	}
	return m
}

type UserLockoutSpec struct {
	All      *InternalUserLockoutSpec `json:"all,omitempty"`
	UserPass *InternalUserLockoutSpec `json:"userpass,omitempty"`
	LDAP     *InternalUserLockoutSpec `json:"ldap,omitempty"`
	AppRole  *InternalUserLockoutSpec `json:"appRole,omitempty"`
}

func (s *UserLockoutSpec) ListMapValue() map[string]map[string]any {
	_m := map[string]map[string]any{}
	if s.All != nil {
		_m["all"] = s.All.MapValue()
	}
	if s.UserPass != nil {
		_m["userpass"] = s.UserPass.MapValue()
	}
	if s.LDAP != nil {
		_m["ldap"] = s.LDAP.MapValue()
	}
	if s.AppRole != nil {
		_m["approle"] = s.AppRole.MapValue()
	}
	return _m
}
