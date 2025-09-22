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
	"context"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// #region ServiceRegistrationSpec
// Resource Specification Specific to ServiceRegistration

const serviceRegistrationSpecHclTempate = `
{{ range $key, $object := . }}
service_registration "{{ $object.InternalServiceRegistration.Type }}"{
        {{ range $k,$v := $object.InternalServiceRegistration.MapValue }}
    {{ $k }} = {{ $v }}
        {{ end}}
}
{{ end }}
`

type ServiceRegistrationList []ServiceRegistrationSpec

func (s ServiceRegistrationList) Volumes(c *client.Client, ctx context.Context, vs *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	for _, cb := range s {
		if vol, mnt, err := cb.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}
	return volumes, mounts, nil
}

func (s ServiceRegistrationList) Secrets(c *client.Client, ctx context.Context, vs *VaultServer) error {
	for _, cb := range s {
		if err := cb.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceRegistrationList) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(serviceRegistrationSpecHclTempate))
	if err := template.Execute(&buf, s); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

type ServiceRegistrationKubernetesSpec struct {
	Namespace *string `json:"namespace,omitempty" hcl:"namespace"`
	PodName   *string `json:"podname,omitempty" hcl:"pod_name"`
}

// HclRender implements ConfigBuilderHelper.
func (s *ServiceRegistrationKubernetesSpec) HclRender() (string, error) {
	return "", nil
}

// Secrets implements ConfigBuilderHelper.
func (s *ServiceRegistrationKubernetesSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *ServiceRegistrationKubernetesSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	return []v1.Volume{}, []v1.VolumeMount{}, nil
}

func (s *ServiceRegistrationKubernetesSpec) Type() string {
	return "kubernetes"
}

func (s *ServiceRegistrationKubernetesSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

// +kubebuilder:validation:MinProperties=1
type ServiceRegistrationSpec struct {
	Consul     *ConsulSpec                        `json:"consul"`
	Kubernetes *ServiceRegistrationKubernetesSpec `json:"kubernetes"`
}

func (s *ServiceRegistrationSpec) internalServiceRegistration() ConfigBuilderHelper {
	if s.Kubernetes != nil {
		return s.Kubernetes
	}
	if s.Consul != nil {
		return s.Consul
	}
	return nil
}

func (s *ServiceRegistrationSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return s.internalServiceRegistration().Secrets(c, ctx, vaultServer)
}

func (s *ServiceRegistrationSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return s.internalServiceRegistration().Volumes(c, ctx, vaultServer)
}

// #endregion
