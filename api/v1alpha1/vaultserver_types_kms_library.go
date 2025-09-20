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
	"strconv"
)

// Resource Specification Specific to KmsLibrary

type KmsLibrarySpec struct {
	// +kubebuilder:validation:Enum=pkcs11
	Type    string  `json:"type,omitempty"`
	Name    *string `json:"name,omitempty"`
	Library *string `json:"library,omitempty"`
}

func (k *KmsLibrarySpec) MapValue() map[string]any {
	_m := map[string]any{}
	if k.Name != nil {
		_m["name"] = strconv.Quote(*k.Name)
	}
	if k.Library != nil {
		_m["library"] = strconv.Quote(*k.Library)
	}
	return _m
}
