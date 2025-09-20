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
	"reflect"
	"strconv"
)

// #region ServiceRegistrationSpec
// Resource Specification Specific to ServiceRegistration

type ServiceRegistrationKubernetesSpec struct {
	Namespace *string `json:"namespace,omitempty"`
	PodName   *string `json:"podname,omitempty"`
}

func (s *ServiceRegistrationKubernetesSpec) Type() string {
	return "kubernetes"
}

func (s *ServiceRegistrationKubernetesSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Namespace != nil {
		_m["namespace"] = strconv.Quote(*s.Namespace)
	}
	if s.PodName != nil {
		_m["pod_name"] = strconv.Quote(*s.PodName)
	}
	return _m
}

// +kubebuilder:validation:MinProperties=1
type ServiceRegistrationSpec struct {
	Consul     *ConsulSpec                        `json:"consul"`
	Kubernetes *ServiceRegistrationKubernetesSpec `json:"kubernetes"`
}

func (s *ServiceRegistrationSpec) InternalServiceRegistration() HclHelper {
	v := reflect.ValueOf(*s)
	t := reflect.TypeOf(*s)

	for i := range t.NumField() {
		intf := v.Field(i)
		if !intf.IsNil() {
			return intf.Interface().(HclHelper)
		}
	}
	return nil
}

// #endregion
