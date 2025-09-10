package v1alpha1

import (
	"time"
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
	CaCert     ConfigMapKeySelector `json:"caCert,omitempty"`
	ClientCert SecretSelector       `json:"clientCert"`
	MinVersion TLSVersion           `json:"minVersion,omitempty"`
	SkipVerify bool                 `json:"skipVerify,omitempty"`
}

type ConsulSpec struct {
	Address             string            `json:"address,omitempty"`
	CheckTimeout        time.Duration     `json:"checkTimeout,omitempty"`
	DisableRegistration bool              `json:"disableRegistration,omitempty"`
	Service             string            `json:"service,omitempty"`
	ServiceTags         []string          `json:"serviceTags,omitempty"`
	ServiceMeta         map[string]string `json:"serviceMeta,omitempty"`
	ServiceAddress      string            `json:"serviceAddress,omitempty"`
	Token               SecretKeySelector `json:"clientCert"`
	Tls                 ConsulTlsSpec     `json:"tls,omitempty"`

	// +kubebuilder:validation:Enum=http;https
	Scheme string `json:"scheme,omitempty"`
}
