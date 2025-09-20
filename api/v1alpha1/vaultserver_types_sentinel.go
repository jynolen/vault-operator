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
	"strings"

	"github.com/jynolen/vault-operator/internal/utils"
)

// #region SentinelSpec
// Resource Specification Specific to Sentinel

type SentinelSpec struct {
	AdditionalEnabledModules []string `json:"additionalEnabledModules,omitempty"`
}

func (s *SentinelSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if len(s.AdditionalEnabledModules) > 0 {
		_m["additional_enabled_modules"] = fmt.Sprintf("[%s]", strings.Join(utils.Map(strconv.Quote, s.AdditionalEnabledModules), ","))
	}
	return _m
}

// #endregion
