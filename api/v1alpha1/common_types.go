package v1alpha1

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"

	// "github.com/jynolen/vault-operator/internal/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DataReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type DatakeyReference struct {
	DataReference `json:",inline"`
	Key           string `json:"key"`
}

type SecretKeySelector struct {
	SecretRef DatakeyReference `json:"secretKeyRef"`
}

type SecretSelector struct {
	SecretRef DataReference `json:"secretRef"`
}

type ConfigMapSelector struct {
	ConfigMapRef DataReference `json:"configMapRef"`
}

type ConfigMapKeySelector struct {
	ConfigMapRef DatakeyReference `json:"configMapRef"`
}

// +kubebuilder:validation:Enum=tls10;tls11;tls12;tls13
type TLSVersion string

type ConsulTlsSpec struct {
	CaCert     *ConfigMapKeySelector `json:"caCert,omitempty"`
	ClientCert *SecretSelector       `json:"clientCert"`
	MinVersion *TLSVersion           `json:"minVersion,omitempty"`
	SkipVerify *bool                 `json:"skipVerify,omitempty"`
}

func (s *ConsulTlsSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.CaCert != nil {
		_m["tls_ca_file"] = strconv.Quote("/consul/ca.crt")
	}
	if s.ClientCert != nil {
		_m["tls_cert_file"] = strconv.Quote("/consul/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/consul/tls.key")
	}
	if s.MinVersion != nil {
		_m["tls_min_version"] = strconv.Quote(string(*s.MinVersion))
	}
	if s.SkipVerify != nil {
		_m["tls_skip_verify"] = strconv.FormatBool(*s.SkipVerify)
	}
	return _m
}

type ConsulSpec struct {
	Token               *string            `json:"-"`
	Address             *string            `json:"address,omitempty"`
	CheckTimeout        *metav1.Duration   `json:"checkTimeout,omitempty"`
	DisableRegistration *bool              `json:"disableRegistration,omitempty"`
	Service             *string            `json:"service,omitempty"`
	ServiceTags         []string           `json:"serviceTags,omitempty"`
	ServiceMeta         map[string]string  `json:"serviceMeta,omitempty"`
	ServiceAddress      *string            `json:"serviceAddress,omitempty"`
	TokenSecret         *SecretKeySelector `json:"clientCert"`
	Tls                 *ConsulTlsSpec     `json:"tls,omitempty"`

	// +kubebuilder:validation:Enum=http;https
	Scheme *string `json:"scheme,omitempty"`
}

func (s *ConsulSpec) Type() string {
	return "consul"
}

func (s *ConsulSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Address != nil {
		_m["address"] = strconv.Quote(*s.Address)
	}
	if s.CheckTimeout != nil {
		_m["check_timeout"] = strconv.Quote(fmt.Sprintf("%s", s.CheckTimeout.Duration))
	}
	if s.DisableRegistration != nil {
		_m["disable_registration"] = strconv.FormatBool(*s.DisableRegistration)
	}
	if s.Scheme != nil {
		_m["scheme"] = strconv.Quote(*s.Scheme)
	}
	if s.Service != nil {
		_m["service"] = strconv.Quote(*s.Service)
	}
	if len(s.ServiceTags) > 0 {
		_m["service_tags"] = strconv.Quote(strings.Join(s.ServiceTags, ","))
	}
	if len(s.ServiceMeta) > 0 {
		_m["service_meta"] = strconv.Quote("TODO")
	}
	if s.ServiceAddress != nil {
		_m["service_address"] = strconv.Quote(*s.ServiceAddress)
	}
	if s.TokenSecret != nil {
		_m["token"] = strconv.Quote("TODO")
	}
	if s.Tls != nil {
		maps.Copy(_m, s.Tls.MapValue())
	}
	return _m
}

// +kubebuilder:object:root=false
// +kubebuilder:object:generate:false
// +k8s:deepcopy-gen:interfaces=nil
// +k8s:deepcopy-gen=nil

type ConfigBuilderHelper interface {
	Type() string
	MapValue() map[string]any
	Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error)
	Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error
}
