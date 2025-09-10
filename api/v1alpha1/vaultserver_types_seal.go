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

// #region SealSpec
// Resource Specification Specific to Seal

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type SealSpec struct {
	AliCloudKMS   *AliCloudKMSSealSpec   `json:"aliCloudKms,omitempty"`
	AWSKms        *AWSKmsSealSpec        `json:"awsKms,omitempty"`
	AzureKeyVault *AzureKeyVaultSealSpec `json:"azureKeyVault,omitempty"`
	GCPKms        *GcpKmsSealSpec        `json:"gcpKms,omitempty"`
	OCIKms        *OCIKmsSealSpec        `json:"ociKms,omitempty"`
	PKCS11        *PKCS11SealSpec        `json:"pkcs11,omitempty"`
	Transit       *TransitSealSpec       `json:"transit,omitempty"`
}

type InternalSealSpec struct {
	Disabled    bool           `json:"disabled,omitempty"`
	Credentials SecretSelector `json:"credentials"`
}

type AliCloudKMSSealSpec struct {
	InternalSealSpec `json:",inline"`
	Region           string `json:"region,omitempty"`
	Domain           string `json:"domain,omitempty"`
	KmsKeyId         string `json:"kmsKeyId"`
}

type AWSKmsSealSpec struct {
	InternalSealSpec `json:",inline"`
	Region           string `json:"region,omitempty"`
	Endpoint         string `json:"endpoint,omitempty"`
	KmsKeyId         string `json:"kmsKeyId"`
}

type AzureKeyVaultSealSpec struct {
	InternalSealSpec `json:",inline"`
	Environment      string `json:"environment,omitempty"`
	VaultName        string `json:"vaultName"`
	KeyName          string `json:"keyName"`
	Resource         string `json:"resource,omitempty"`
}

type GcpKmsSealSpec struct {
	InternalSealSpec `json:",inline"`
	Project          string `json:"project"`
	Region           string `json:"region"`
	KeyRing          string `json:"keyRing"`
	CryptoKey        string `json:"cryptoKey"`
}

type OCIKmsSealSpec struct {
	InternalSealSpec `json:",inline"`
	AuthTypeApiKey   bool `json:"authTypeApiKey,omitempty"`
}

type PKCS11SealSpec struct {
	Lib                 ConfigMapKeySelector `json:"lib"`
	Slot                string               `json:"slot"`
	TokenLabel          string               `json:"tokenLabel"`
	Pin                 SecretKeySelector    `json:"pin"`
	KeyLabel            string               `json:"keyLabel"`
	KeyId               string               `json:"keyId"`
	HMACKeyLabel        string               `json:"hmacKeyLabel"`
	DefaultKeyLabel     string               `json:"defaultKeyLabel,omitempty"`
	DefaultHMACKeyLabel string               `json:"defaultHmacKeyLabel,omitempty"`
	HmacKeyId           string               `json:"hmacKeyId,omitempty"`
	GenerateKey         bool                 `json:"generateKey,omitempty"`
	ForceRwSession      bool                 `json:"forceRwSession,omitempty"`
	MaxParallel         int32                `json:"maxParallel,omitempty"`
	Disabled            bool                 `json:"disabled,omitempty"`
	RsaEncryptLocal     bool                 `json:"rsaEncryptLocal,omitempty"`

	// +kubebuilder:validation:Enum=sha1;sha256;sha224;sha384;sha512
	RsaOaepHash string `json:"rsaOaepHash,omitempty"`
	// +kubebuilder:validation:Enum="0x1085";"0x1082";"0x1087";"0x0009";"0x0001"
	Mechanism string `json:"mechanism,omitempty"`
	// +kubebuilder:validation:Enum="0x0251"
	HMACMechanism string `json:"hmacMechanism,omitempty"`
}

type TlsTransitSpec struct {
	ServerName string               `json:"serverName"`
	CaCert     ConfigMapKeySelector `json:"caCert,omitempty"`
	ClientCert SecretSelector       `json:"clientCert"`
	SkipVerify bool                 `json:"skipVerify,omitempty"`
}

type TransitSealSpec struct {
	InternalSealSpec `json:",inline"`
	KeyName          string         `json:"KeyName"`
	Address          string         `json:"Address"`
	KeyIdPrefix      string         `json:"KeyIdPrefix,omitempty"`
	MountPath        string         `json:"MountPath"`
	Namespace        string         `json:"Namespace,omitempty"`
	DisableRenewal   string         `json:"DisableRenewal,omitempty"`
	Tls              TlsTransitSpec `json:"tls,omitempty"`
}

// #endregion
