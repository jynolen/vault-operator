package v1alpha1

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"strconv"

	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func (s *SecretKeySelector) IsKind(c *client.Client, ctx context.Context, namespace string, kind corev1.SecretType) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.SecretRef.Name, Namespace: namespace}, &secret); err != nil {
		return nil, err
	}
	if secret.Type != kind {
		return nil, fmt.Errorf("ListenerTCP.TLS SecretType is not a %s", kind)
	}
	return &secret, nil
}

func (s *SecretKeySelector) ContainsKey(c *client.Client, ctx context.Context, namespace string) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.SecretRef.Name, Namespace: namespace}, &secret); err != nil {
		return nil, err
	}

	if _, ok := secret.Data[s.SecretRef.Key]; !ok {
		return nil, fmt.Errorf("Secret `%s` does not contains key `%s`", secret.Name, s.SecretRef.Key)
	}
	return &secret, nil
}

type SecretSelector struct {
	SecretRef DataReference `json:"secretRef"`
}

func (s *SecretSelector) ContainsKey(c *client.Client, ctx context.Context, namespace string, key string) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.SecretRef.Name, Namespace: namespace}, &secret); err != nil {
		return nil, err
	}

	if _, ok := secret.Data[key]; !ok {
		return nil, fmt.Errorf("Secret `%s` does not contains key `%s`", secret.Name, key)
	}
	return &secret, nil
}

func (s *SecretSelector) IsKind(c *client.Client, ctx context.Context, namespace string, kind corev1.SecretType) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.SecretRef.Name, Namespace: namespace}, &secret); err != nil {
		return nil, err
	}
	if secret.Type != kind {
		return nil, fmt.Errorf("Secret %s is not a %s", s.SecretRef.Name, kind)
	}
	return &secret, nil
}

type ConfigMapSelector struct {
	ConfigMapRef DataReference `json:"configMapRef"`
}

func (s *ConfigMapSelector) ContainsKey(c *client.Client, ctx context.Context, namespace string, key string) (*corev1.ConfigMap, error) {
	var cm corev1.ConfigMap
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.ConfigMapRef.Name, Namespace: namespace}, &cm); err != nil {
		return nil, err
	}

	if _, ok := cm.Data[key]; !ok {
		return nil, fmt.Errorf("ConfigMap `%s` does not contains key `%s`", cm.Name, key)
	}
	return &cm, nil
}

type ConfigMapKeySelector struct {
	ConfigMapRef DatakeyReference `json:"configMapRef"`
}

func (s *ConfigMapKeySelector) ContainsKey(c *client.Client, ctx context.Context, namespace string) (*corev1.ConfigMap, error) {
	var cm corev1.ConfigMap
	if err := (*c).Get(ctx, types.NamespacedName{Name: s.ConfigMapRef.Name, Namespace: namespace}, &cm); err != nil {
		return nil, err
	}

	if _, ok := cm.Data[s.ConfigMapRef.Key]; !ok {
		return nil, fmt.Errorf("ConfigMap `%s` does not contains key `%s`", cm.Name, s.ConfigMapRef.Key)
	}
	return &cm, nil
}

// +kubebuilder:validation:Enum=tls10;tls11;tls12;tls13
type TLSVersion string

type ConsulTlsSpec struct {
	CaCert     *ConfigMapKeySelector `json:"caCert,omitempty"`
	ClientCert *SecretSelector       `json:"clientCert"`
	MinVersion *TLSVersion           `json:"minVersion,omitempty" hcl:""`
	SkipVerify *bool                 `json:"skipVerify,omitempty" hcl:""`
}

func (s *ConsulTlsSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if s.CaCert != nil {
		_m["tls_ca_file"] = strconv.Quote("/consul/ca.crt")
	}
	if s.ClientCert != nil {
		_m["tls_cert_file"] = strconv.Quote("/consul/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/consul/tls.key")
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type ConsulSpec struct {
	Token               *string            `json:"-" hcl:"token"`
	Address             *string            `json:"address,omitempty" hcl:"address"`
	CheckTimeout        *metav1.Duration   `json:"checkTimeout,omitempty" hcl:"check_timeout"`
	DisableRegistration *bool              `json:"disableRegistration,omitempty" hcl:"disable_registration"`
	Service             *string            `json:"service,omitempty" hcl:"service"`
	ServiceTags         []string           `json:"serviceTags,omitempty" hcl:"service_tags"`
	ServiceMeta         map[string]string  `json:"serviceMeta,omitempty" hcl:"service_meta"`
	ServiceAddress      *string            `json:"serviceAddress,omitempty" hcl:"service_address"`
	TokenSecret         *SecretKeySelector `json:"clientCert"`
	Tls                 *ConsulTlsSpec     `json:"tls,omitempty"`

	// +kubebuilder:validation:Enum=http;https
	Scheme *string `json:"scheme,omitempty" hcl:"scheme"`
}

// Volumes implements ConfigBuilderHelper.
func (s *ConsulSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	panic("unimplemented")
}

func (s ConsulSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.TokenSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"connection_url": {Dest: s.Token, Mandatory: true, Src: s.TokenSecret.SecretRef.Key},
	}
	return secretMappings.Apply(&secret, "StorageCockroachDBSpec.Credentials")
}

func (s *ConsulSpec) Type() string {
	return "consul"
}

func (s *ConsulSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.Tls != nil {
		if _s, err := s.Tls.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

// +kubebuilder:object:root=false
// +kubebuilder:object:generate:false
// +k8s:deepcopy-gen:interfaces=nil
// +k8s:deepcopy-gen=nil

type ConfigBuilderHelper interface {
	Type() string
	MapValue() (map[string]any, error)
	Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error)
	Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error
}

type HclHelper struct{}

func (h *HclHelper) HclExport() (map[string]any, error) {
	_m := map[string]any{}
	val := reflect.Indirect(reflect.ValueOf(h))
	for i := range val.NumField() {
		field := reflect.TypeOf(h).Field(i)
		if hcl, ok := field.Tag.Lookup("hcl"); ok {
			if v, err := utils.HclEscape(reflect.ValueOf(h).Field(i)); err != nil {
				return nil, err
			} else if v != nil {
				_m[hcl] = v
			}
		}
	}
	return _m, nil
}
