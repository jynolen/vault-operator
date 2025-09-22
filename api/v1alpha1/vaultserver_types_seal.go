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
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const sealHclTemplate = `
{{ if gt (len .) 1 }}
enable_multiseal = true
{{ end }}
{{ range $key, $object := . }}
seal "{{ $object.Type }}"{
		{{ if gt (len $object) 1 }}
    priority = $key
		{{ end }}
		{{ range $k,$v := $object.MapValue }}
    {{ $k }} = {{ $v }}
		{{ end}}
}
{{ end }}
`

// #region SealSpec
// Resource Specification Specific to Seal

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type SealSpec struct {
	AliCloudKms   *SealAliCloudKmsSpec   `json:"aliCloudKms,omitempty"`
	AWSKms        *SealAwsKmsSpec        `json:"awsKms,omitempty"`
	AzureKeyVault *SealAzureKeyVaultSpec `json:"azureKeyVault,omitempty"`
	GCPKms        *SealGcpKmsSpec        `json:"gcpKms,omitempty"`
	OCIKms        *SealOciKmsSpec        `json:"ociKms,omitempty"`
	PKCS11        *SealPKCS11Spec        `json:"pkcs11,omitempty"`
	Transit       *SealTransitSpec       `json:"transit,omitempty"`
}

// +kubebuilder:validation:MaxItems=2
type SealListSpec []SealSpec

func (s SealListSpec) Secrets(c *client.Client, ctx context.Context, vs *VaultServer) error {
	for _, cb := range s {
		if err := cb.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	return nil
}

func (s SealListSpec) Volumes(c *client.Client, ctx context.Context, vs *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
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

func (s *SealListSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(sealHclTemplate))
	if err := template.Execute(&buf, s); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

func (s *SealSpec) internalSeal() ConfigBuilderHelper {
	if s.AliCloudKms != nil {
		return s.AliCloudKms
	}
	if s.AWSKms != nil {
		return s.AWSKms
	}
	if s.AzureKeyVault != nil {
		return s.AzureKeyVault
	}
	if s.GCPKms != nil {
		return s.GCPKms
	}
	if s.PKCS11 != nil {
		return s.PKCS11
	}
	if s.Transit != nil {
		return s.Transit
	}
	if s.OCIKms != nil {
		return s.OCIKms
	}
	return nil
}

func (s *SealSpec) Type() string {
	return s.internalSeal().Type()
}

func (s *SealSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return s.internalSeal().Volumes(c, ctx, vaultServer)
}

func (s *SealSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return s.internalSeal().Secrets(c, ctx, vaultServer)
}

type InternalSealSpec struct {
	Disabled    *bool           `json:"disabled,omitempty" hcl:"disabled"`
	Credentials *SecretSelector `json:"credentials"`
}

type SealAliCloudKmsSpec struct {
	InternalSealSpec `json:",inline"`
	AccessKey        *string `json:"-" hcl:"access_key"`
	SecretKey        *string `json:"-" hcl:"secret_key"`
	Region           *string `json:"region,omitempty" hcl:"region"`
	Domain           *string `json:"domain,omitempty" hcl:"domain"`
	KmsKeyId         *string `json:"kmsKeyId" hcl:"kms_key_id"`
}

func (s *SealAliCloudKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"access_key": {Dest: s.AccessKey, Mandatory: true},
		"secret_key": {Dest: s.SecretKey, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "SealAliCloudKmsSpec.Credentials")
}

func (s *SealAliCloudKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealAliCloudKmsSpec) Type() string {
	return "alicloudkms"
}

func (s *SealAliCloudKmsSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type SealAwsKmsSpec struct {
	InternalSealSpec `json:",inline"`
	AccessKey        *string `json:"-" hcl:"access_key"`
	SecretKey        *string `json:"-" hcl:"secret_key"`
	SessionToken     *string `json:"-" hcl:"session_token"`
	Region           *string `json:"region,omitempty" hcl:"region"`
	Endpoint         *string `json:"endpoint,omitempty" hcl:"endpoint"`
	KmsKeyId         *string `json:"kmsKeyId" hcl:"kms_key_id"`
}

func (s *SealAwsKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"access_key":    {Dest: s.AccessKey, Mandatory: true},
		"secret_key":    {Dest: s.SecretKey, Mandatory: true},
		"session_token": {Dest: s.SessionToken},
	}
	return secretMappings.Apply(&secret, "SealAwsKmsSpec.Credentials")
}

func (s *SealAwsKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealAwsKmsSpec) Type() string {
	return "awskms"
}

func (s *SealAwsKmsSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type SealAzureKeyVaultSpec struct {
	InternalSealSpec `json:",inline"`
	TenantID         *string `json:"-" hcl:"tenand_id"`
	ClientID         *string `json:"-" hcl:"client_id"`
	ClientSecret     *string `json:"-" hcl:"client_secret"`
	SessionToken     *string `json:"-" hcl:"session_token"`
	Environment      *string `json:"environment,omitempty" hcl:"environment"`
	VaultName        *string `json:"vaultName" hcl:"vault_name"`
	KeyName          *string `json:"keyName" hcl:"key_name"`
	Resource         *string `json:"resource,omitempty" hcl:"resource"`
}

func (s *SealAzureKeyVaultSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"client_id":     {Dest: s.ClientID, Mandatory: true},
		"client_secret": {Dest: s.ClientSecret, Mandatory: true},
		"tenant_id":     {Dest: s.TenantID, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "SealAzureKeyVaultSpec.Credentials")
}

func (s *SealAzureKeyVaultSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealAzureKeyVaultSpec) Type() string {
	return "azurekeyvault"
}

func (s *SealAzureKeyVaultSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type SealGcpKmsSpec struct {
	InternalSealSpec `json:",inline"`
	Project          *string `json:"project" hcl:"project"`
	Region           *string `json:"region" hcl:"region"`
	KeyRing          *string `json:"keyRing" hcl:"key_ring"`
	CryptoKey        *string `json:"cryptoKey" hcl:"crypto_key"`
}

func (s *SealGcpKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", secret.Name)
	}
	_, ok := secret.Data["credentials.json"]
	if !ok {
		return fmt.Errorf("%s does not contains key `credentials.json`", secret.Name)
	}
	return nil
}

func (s *SealGcpKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealGcpKmsSpec) Type() string {
	return "gcpckms"
}

func (s *SealGcpKmsSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type SealOciKmsSpec struct {
	InternalSealSpec   `json:",inline"`
	KeyId              *string `json:"keyId" hcl:"key_id"`
	CryptoEndpoint     *string `json:"cryptoEndpoint" hcl:"crypto_endpoint"`
	ManagementEndpoint *string `json:"managementEndpoint" hcl:"management_endpoint"`
	AuthTypeApiKey     *bool   `json:"authTypeApiKey,omitempty" hcl:"auth_type_api_key"`
}

func (s *SealOciKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", secret.Name)
	}
	_, ok := secret.Data["config"]
	if !ok {
		return fmt.Errorf("%s does not contains key `config`", secret.Name)
	}
	return nil
}

func (s *SealOciKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealOciKmsSpec) Type() string {
	return "ocikms"
}

func (s *SealOciKmsSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type SealPKCS11Spec struct {
	Pin                 *string            `json:"-" hcl:"pin"`
	Lib                 *string            `json:"lib" hcl:"lib"`
	Slot                *string            `json:"slot" hcl:"slot"`
	TokenLabel          *string            `json:"tokenLabel" hcl:"token_label"`
	PinSecret           *SecretKeySelector `json:"pinSecret"`
	KeyLabel            *string            `json:"keyLabel" hcl:"key_label"`
	KeyId               *string            `json:"keyId" hcl:"key_id"`
	HMACKeyLabel        *string            `json:"hmacKeyLabel" hcl:"hmac_key_label"`
	DefaultKeyLabel     *string            `json:"defaultKeyLabel,omitempty" hcl:"default_key_label"`
	DefaultHMACKeyLabel *string            `json:"defaultHmacKeyLabel,omitempty" hcl:"default_hmac_key_label"`
	HmacKeyId           *string            `json:"hmacKeyId,omitempty" hcl:"hmac_key_id"`
	MaxParallel         *int32             `json:"maxParallel,omitempty" hcl:"max_parallel"`
	Disabled            *bool              `json:"disabled,omitempty" hcl:"disabled"`
	GenerateKey         *bool              `json:"generateKey,omitempty" hcl:"generate_key"`
	ForceRwSession      *bool              `json:"forceRwSession,omitempty" hcl:"force_rw_session"`
	RsaEncryptLocal     *bool              `json:"rsaEncryptLocal,omitempty" hcl:"rsa_encrypt_local"`
	// +kubebuilder:validation:Enum=sha1;sha256;sha224;sha384;sha512
	RsaOaepHash *string `json:"rsaOaepHash,omitempty" hcl:"rsa_oaep_hash"`
	// +kubebuilder:validation:Enum="0x1085";"0x1082";"0x1087";"0x0009";"0x0001"
	Mechanism *string `json:"mechanism,omitempty" hcl:"mechanism"`
	// +kubebuilder:validation:Enum="0x0251"
	HMACMechanism *string `json:"hmacMechanism,omitempty" hcl:"hmac_mechanism"`
}

func (s *SealPKCS11Spec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.PinSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"pin": {Dest: s.Pin, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "SealPKCS11Spec.Credentials")
}

func (s *SealPKCS11Spec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *SealPKCS11Spec) Type() string {
	return "pkcs11"
}

func (s *SealPKCS11Spec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type TlsTransitSpec struct {
	ServerName *string               `json:"serverName,omitempty" hcl:"tls_server_name"`
	CaCert     *ConfigMapKeySelector `json:"caCert"`
	ClientCert *SecretSelector       `json:"clientCert"`
	SkipVerify *bool                 `json:"skipVerify,omitempty" hcl:"tls_skip_verify"`
}

func (s *TlsTransitSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.ClientCert != nil {
		_, err := s.ClientCert.IsKind(c, ctx, vaultServer.Namespace, corev1.SecretTypeTLS)
		if err != nil {
			return nil, nil, err
		}
		volumes = append(volumes, corev1.Volume{
			Name: "seal-transit-client-tls",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: vaultServer.Spec.Config.ListenerTcp.TLS.Cert.SecretRef.Name,
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "seal-transit-client-tls",
			ReadOnly:  true,
			MountPath: "/seal/transit",
		})
	}

	if s.CaCert != nil {
		_, err := s.CaCert.ContainsKey(c, ctx, vaultServer.Namespace)
		if err != nil {
			return nil, nil, err
		}
		volumes = append(volumes, corev1.Volume{
			Name: "seal-transit-ca-tls",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: s.CaCert.ConfigMapRef.Name},
					Items: []corev1.KeyToPath{{
						Key:  s.CaCert.ConfigMapRef.Key,
						Path: "ca.crt",
					}},
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "seal-transit-client-tls",
			ReadOnly:  true,
			MountPath: "/seal/transit/ca.crt",
			SubPath:   "ca.crt",
		})
	}

	return volumes, mounts, nil
}

func (s *TlsTransitSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.CaCert != nil {
		_m["tls_ca_cert"] = strconv.Quote("/seal/transit/ca.crt")
	}
	if s.ClientCert != nil {
		_m["tls_client_cert"] = strconv.Quote("/seal/transit/tls.crt")
		_m["tls_client_key"] = strconv.Quote("/seal/transit/tls.key")
	}
	return _m, nil
}

type SealTransitSpec struct {
	InternalSealSpec `json:",inline"`
	Token            *string         `json:"-" hcl:"token"`
	KeyName          *string         `json:"keyName" hcl:"key_name"`
	Address          *string         `json:"address" hcl:"address"`
	KeyIdPrefix      *string         `json:"keyIdPrefix,omitempty" hcl:"key_id_prefix"`
	MountPath        *string         `json:"mountPath" hcl:"mount_path"`
	Namespace        *string         `json:"namespace,omitempty" hcl:"namespace"`
	DisableRenewal   *bool           `json:"disableRenewal,omitempty" hcl:"disable_renewal"`
	Tls              *TlsTransitSpec `json:"tls,omitempty"`
}

func (s *SealTransitSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"token": {Dest: s.Token, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "SealTransitSpec.Credentials")
}

func (s *SealTransitSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.Tls == nil {
		return volumes, mounts, nil
	}
	return s.Tls.Volumes(c, ctx, vaultServer)
}

func (s *SealTransitSpec) Type() string {
	return "transit"
}

func (s *SealTransitSpec) MapValue() (map[string]any, error) {
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

// #endregion
