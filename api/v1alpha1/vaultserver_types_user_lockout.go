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

type InternalUserLockoutSpec struct {
	Threshold      int32         `json:"threshold,omitempty"`
	Duration       time.Duration `json:"duration,omitempty"`
	CounterReset   time.Duration `json:"counterReset,omitempty"`
	DisableLockout bool          `json:"disableLockout,omitempty"`
}

type UserLockoutSpec struct {
	All      *InternalUserLockoutSpec `json:"all,omitempty"`
	UserPass *InternalUserLockoutSpec `json:"userpass,omitempty"`
	LDAP     *InternalUserLockoutSpec `json:"ldap,omitempty"`
	AppRole  *InternalUserLockoutSpec `json:"appRole,omitempty"`
}
