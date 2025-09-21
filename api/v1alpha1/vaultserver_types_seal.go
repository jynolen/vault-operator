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
	"context"
	"fmt"
	"maps"
	"strconv"

	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

func (s *SealSpec) InternalSeal() ConfigBuilderHelper {
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
	return nil
}

type InternalSealSpec struct {
	Disabled    *bool           `json:"disabled,omitempty"`
	Credentials *SecretSelector `json:"credentials"`
}

type SealAliCloudKmsSpec struct {
	InternalSealSpec `json:",inline"`
	AccessKey        *string `json:"-"`
	SecretKey        *string `json:"-"`
	Region           *string `json:"region,omitempty"`
	Domain           *string `json:"domain,omitempty"`
	KmsKeyId         *string `json:"kmsKeyId"`
}

// Secrets implements ConfigBuilderHelper.
func (s *SealAliCloudKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"access_key": {Dest: s.AccessKey, Mandatory: true},
		"secret_key": {Dest: s.SecretKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "SealAliCloudKmsSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil

}

// Volumes implements ConfigBuilderHelper.
func (s *SealAliCloudKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealAliCloudKmsSpec) Type() string {
	return "alicloudkms"
}

func (s *SealAliCloudKmsSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.AccessKey != nil {
		_m["access_key"] = strconv.Quote(*s.AccessKey)
	}
	if s.SecretKey != nil {
		_m["secret_key"] = strconv.Quote(*s.SecretKey)
	}
	if s.Domain != nil {
		_m["domain"] = strconv.Quote(*s.Domain)
	}
	if s.KmsKeyId != nil {
		_m["kms_key_id"] = strconv.Quote(*s.KmsKeyId)
	}
	return _m
}

type SealAwsKmsSpec struct {
	InternalSealSpec `json:",inline"`
	AccessKey        *string `json:"-"`
	SecretKey        *string `json:"-"`
	SessionToken     *string `json:"-"`
	Region           *string `json:"region,omitempty"`
	Endpoint         *string `json:"endpoint,omitempty"`
	KmsKeyId         *string `json:"kmsKeyId"`
}

// Secrets implements ConfigBuilderHelper.
func (s *SealAwsKmsSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"access_key":    {Dest: s.AccessKey, Mandatory: true},
		"secret_key":    {Dest: s.SecretKey, Mandatory: true},
		"session_token": {Dest: s.SessionToken},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "SealAwsKmsSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *SealAwsKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealAwsKmsSpec) Type() string {
	return "awskms"
}

func (s *SealAwsKmsSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.AccessKey != nil {
		_m["access_key"] = strconv.Quote(*s.AccessKey)
	}
	if s.SecretKey != nil {
		_m["secret_key"] = strconv.Quote(*s.SecretKey)
	}
	if s.SessionToken != nil {
		_m["session_token"] = strconv.Quote(*s.SessionToken)
	}
	if s.Endpoint != nil {
		_m["endpoint"] = strconv.Quote(*s.Endpoint)
	}
	if s.KmsKeyId != nil {
		_m["kms_key_id"] = strconv.Quote(*s.KmsKeyId)
	}
	return _m
}

type SealAzureKeyVaultSpec struct {
	InternalSealSpec `json:",inline"`
	TenantID         *string `json:"-"`
	ClientID         *string `json:"-"`
	ClientSecret     *string `json:"-"`
	SessionToken     *string `json:"-"`
	Environment      *string `json:"environment,omitempty"`
	VaultName        *string `json:"vaultName"`
	KeyName          *string `json:"keyName"`
	Resource         *string `json:"resource,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *SealAzureKeyVaultSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"client_id":     {Dest: s.ClientID, Mandatory: true},
		"client_secret": {Dest: s.ClientSecret, Mandatory: true},
		"tenant_id":     {Dest: s.TenantID, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "SealAzureKeyVaultSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *SealAzureKeyVaultSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealAzureKeyVaultSpec) Type() string {
	return "azurekeyvault"
}

func (s *SealAzureKeyVaultSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Environment != nil {
		_m["environment"] = strconv.Quote(*s.Environment)
	}
	if s.VaultName != nil {
		_m["vault_name"] = strconv.Quote(*s.VaultName)
	}
	if s.KeyName != nil {
		_m["key_name"] = strconv.Quote(*s.KeyName)
	}
	if s.Resource != nil {
		_m["resource"] = strconv.Quote(*s.Resource)
	}
	return _m
}

type SealGcpKmsSpec struct {
	InternalSealSpec `json:",inline"`
	Project          *string `json:"project"`
	Region           *string `json:"region"`
	KeyRing          *string `json:"keyRing"`
	CryptoKey        *string `json:"cryptoKey"`
}

// Secrets implements ConfigBuilderHelper.
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

// Volumes implements ConfigBuilderHelper.
func (s *SealGcpKmsSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealGcpKmsSpec) Type() string {
	return "gcpckms"
}

func (s *SealGcpKmsSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Credentials != nil {
		_m["credentials"] = strconv.Quote("/gcp/client_secret.json")
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.KeyRing != nil {
		_m["key_ring"] = strconv.Quote(*s.KeyRing)
	}
	if s.CryptoKey != nil {
		_m["crypto_key"] = strconv.Quote(*s.CryptoKey)
	}
	if s.Project != nil {
		_m["project"] = strconv.Quote(*s.Project)
	}
	return _m
}

type SealOciKmsSpec struct {
	InternalSealSpec   `json:",inline"`
	KeyId              *string `json:"keyId"`
	CryptoEndpoint     *string `json:"cryptoEndpoint"`
	ManagementEndpoint *string `json:"managementEndpoint"`
	AuthTypeApiKey     *bool   `json:"authTypeApiKey,omitempty"`
}

func (s *SealOciKmsSpec) Type() string {
	return "ocikms"
}

func (s *SealOciKmsSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.KeyId != nil {
		_m["key_id"] = strconv.Quote(*s.KeyId)
	}
	if s.CryptoEndpoint != nil {
		_m["crypto_endpoint"] = strconv.Quote(*s.CryptoEndpoint)
	}
	if s.AuthTypeApiKey != nil {
		_m["auth_type_api_key"] = strconv.FormatBool(*s.AuthTypeApiKey)
	}
	if s.ManagementEndpoint != nil {
		_m["management_endpoint"] = strconv.Quote(*s.ManagementEndpoint)
	}
	return _m
}

type SealPKCS11Spec struct {
	Pin                 *string            `json:"-"`
	Lib                 *string            `json:"lib"`
	Slot                *string            `json:"slot"`
	TokenLabel          *string            `json:"tokenLabel"`
	PinSecret           *SecretKeySelector `json:"pinSecret"`
	KeyLabel            *string            `json:"keyLabel"`
	KeyId               *string            `json:"keyId"`
	HMACKeyLabel        *string            `json:"hmacKeyLabel"`
	DefaultKeyLabel     *string            `json:"defaultKeyLabel,omitempty"`
	DefaultHMACKeyLabel *string            `json:"defaultHmacKeyLabel,omitempty"`
	HmacKeyId           *string            `json:"hmacKeyId,omitempty"`
	MaxParallel         *int32             `json:"maxParallel,omitempty"`
	Disabled            *bool              `json:"disabled,omitempty"`
	GenerateKey         *bool              `json:"generateKey,omitempty"`
	ForceRwSession      *bool              `json:"forceRwSession,omitempty"`
	RsaEncryptLocal     *bool              `json:"rsaEncryptLocal,omitempty"`
	// +kubebuilder:validation:Enum=sha1;sha256;sha224;sha384;sha512
	RsaOaepHash *string `json:"rsaOaepHash,omitempty"`
	// +kubebuilder:validation:Enum="0x1085";"0x1082";"0x1087";"0x0009";"0x0001"
	Mechanism *string `json:"mechanism,omitempty"`
	// +kubebuilder:validation:Enum="0x0251"
	HMACMechanism *string `json:"hmacMechanism,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *SealPKCS11Spec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: pkcs11.PinSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"pin": {Dest: s.Pin, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "SealPKCS11Spec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil

}

// Volumes implements ConfigBuilderHelper.
func (s *SealPKCS11Spec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealPKCS11Spec) Type() string {
	return "pkcs11"
}

func (s *SealPKCS11Spec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Lib != nil {
		_m["lib"] = strconv.Quote(*s.Lib)
	}
	if s.Slot != nil {
		_m["slot"] = strconv.Quote(*s.Slot)
	}
	if s.TokenLabel != nil {
		_m["token"] = strconv.Quote(*s.TokenLabel)
	}
	if s.Pin != nil {
		_m["pin"] = strconv.Quote(*s.Pin)
	}
	if s.KeyLabel != nil {
		_m["key_label"] = strconv.Quote(*s.KeyLabel)
	}
	if s.DefaultKeyLabel != nil {
		_m["default_key_label"] = strconv.Quote(*s.DefaultKeyLabel)
	}
	if s.KeyId != nil {
		_m["key_id"] = strconv.Quote(*s.KeyId)
	}
	if s.HMACKeyLabel != nil {
		_m["hmac_key_label"] = strconv.Quote(*s.HMACKeyLabel)
	}
	if s.DefaultHMACKeyLabel != nil {
		_m["default_hmac_key_label"] = strconv.Quote(*s.DefaultHMACKeyLabel)
	}
	if s.HmacKeyId != nil {
		_m["hmac_key_id"] = strconv.Quote(*s.HmacKeyId)
	}
	if s.Mechanism != nil {
		_m["mechanism"] = strconv.Quote(*s.Mechanism)
	}
	if s.HMACMechanism != nil {
		_m["hmac_mechanism"] = strconv.Quote(*s.HMACMechanism)
	}
	if s.RsaOaepHash != nil {
		_m["rsa_oaep_hash"] = strconv.Quote(*s.RsaOaepHash)
	}
	if s.GenerateKey != nil {
		_m["generate_key"] = strconv.FormatBool(*s.GenerateKey)
	}
	if s.ForceRwSession != nil {
		_m["force_rw_session"] = strconv.FormatBool(*s.ForceRwSession)
	}
	if s.RsaEncryptLocal != nil {
		_m["rsa_encrypt_local"] = strconv.FormatBool(*s.RsaEncryptLocal)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	return _m
}

type TlsTransitSpec struct {
	ServerName *string               `json:"serverName,omitempty"`
	CaCert     *ConfigMapKeySelector `json:"caCert"`
	ClientCert *SecretSelector       `json:"clientCert"`
	SkipVerify *bool                 `json:"skipVerify,omitempty"`
}

func (s *TlsTransitSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.ServerName != nil {
		_m["tls_server_name"] = strconv.Quote(*s.ServerName)
	}
	if s.SkipVerify != nil {
		_m["tls_skip_verify"] = strconv.FormatBool(*s.SkipVerify)
	}
	if s.CaCert != nil {
		_m["tls_ca_cert"] = strconv.Quote("/seal/transit/ca.crt")
	}
	if s.ClientCert != nil {
		_m["tls_client_cert"] = strconv.Quote("/seal/transit/tls.crt")
		_m["tls_client_key"] = strconv.Quote("/seal/transit/tls.key")
	}
	return _m
}

type SealTransitSpec struct {
	InternalSealSpec `json:",inline"`
	Token            *string         `json:"-"`
	KeyName          *string         `json:"keyName"`
	Address          *string         `json:"address"`
	KeyIdPrefix      *string         `json:"keyIdPrefix,omitempty"`
	MountPath        *string         `json:"mountPath"`
	Namespace        *string         `json:"namespace,omitempty"`
	DisableRenewal   *bool           `json:"disableRenewal,omitempty"`
	Tls              *TlsTransitSpec `json:"tls,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *SealTransitSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	panic("unimplemented")
}

// Volumes implements ConfigBuilderHelper.
func (s *SealTransitSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]v1.Volume, []v1.VolumeMount, error) {
	panic("unimplemented")
}

func (s *SealTransitSpec) Type() string {
	return "transit"
}

func (s *SealTransitSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Disabled != nil {
		_m["disabled"] = strconv.FormatBool(*s.Disabled)
	}
	if s.Address != nil {
		_m["address"] = strconv.Quote(*s.Address)
	}
	if s.Address != nil {
		_m["key_name"] = strconv.Quote(*s.KeyName)
	}
	if s.Token != nil {
		_m["token"] = strconv.Quote(*s.Token)
	}

	if s.Address != nil {
		_m["key_id_prefix"] = strconv.Quote(*s.KeyIdPrefix)
	}
	if s.Address != nil {
		_m["mount_path"] = strconv.Quote(*s.MountPath)
	}
	if s.Address != nil {
		_m["namespace"] = strconv.Quote(*s.Namespace)
	}
	if s.Address != nil {
		_m["disable_renewal"] = strconv.FormatBool(*s.DisableRenewal)
	}
	if s.Tls != nil {
		maps.Copy(_m, s.Tls.MapValue())
	}
	return _m
}

// #endregion
