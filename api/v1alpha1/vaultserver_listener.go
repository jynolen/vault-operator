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
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// +kubebuilder:validation:Pattern=[12345]\d\d
type SpecificStatusCodeSpec string

// +kubebuilder:validation:Pattern=[12345x][x\d][x\d]
type CollectiveStatusCodeSpec string

type ListenerSpec struct {
	Type                                  string                     `json:"-"`
	MaxRequestLimit                       *MaxRequestLimitSpec       `json:"maxRequestLimit,omitempty"`
	RequireRequestHeader                  bool                       `json:"requireRequestHeader,omitempty"`
	TLS                                   *ListenerTlsSpec           `json:"tls,omitempty"`
	HTTPimeout                            *HTTPTimeoutSpec           `json:"httpTimeout,omitempty"`
	ProxyProtocol                         *ProxyProtocolSpec         `json:"proxyProcotol,omitempty"`
	XForwarded                            *XForwardedForSpec         `json:"xForwardedFor,omitempty"`
	UnauthenticatedMetricsAccess          bool                       `json:"unauthenticatedMetricsAccess,omitempty"`
	UnauthenticatedPprofAccess            bool                       `json:"unauthenticatedPprofAccess,omitempty"`
	UnauthenticatedInFlightRequestsAccess bool                       `json:"unauthenticatedInFlightRequestsAccess,omitempty"`
	Cors                                  *CorsSpec                  `json:"cors,omitempty"`
	CustomResponseHeaders                 *CustomResponseHeadersSpec `json:"customResponseHeaders,omitempty"`
	ChrootNamespace                       string                     `json:"chrootNamespace,omitempty"`
	Redact                                *RedactSpec                `json:"redact,omitempty"`
	DisableReplicationStatusEndpoints     bool                       `json:"disableReplicationStatusEndpoints,omitempty"`
	DisableRequestLimiter                 bool                       `json:"disableRequestLimiter,omitempty"`
	CustomMaxJson                         *CustomMaxJsonSpec         `json:"customMaxJson,omitempty"`
}

type MaxRequestLimitSpec struct {
	Size     resource.Quantity `json:"size,omitempty"`
	Duration time.Duration     `json:"duration,omitempty"`
}

type ListenerTlsSpec struct {
	Disable    bool            `json:"disable,omitempty"`
	Cert       *SecretSelector `json:"certificate,omitempty"`
	MinVersion TLSVersion      `json:"minVersion,omitempty"`
	MaxVersion TLSVersion      `json:"maxVersion,omitempty"`

	// +kubebuilder:validation:Enum=TLS_RSA_WITH_RC4_128_SHA;TLS_RSA_WITH_3DES_EDE_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA;TLS_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA256;TLS_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_RC4_128_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_RSA_WITH_RC4_128_SHA;TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256;TLS_AES_128_GCM_SHA256;TLS_AES_256_GCM_SHA384;TLS_CHACHA20_POLY1305_SHA256
	CipherSuites               string                `json:"cipherSuites,omitempty"`
	ClientCa                   *ConfigMapKeySelector `json:"license,omitempty"`
	RequireAndVerifyClientCert bool                  `json:"requireAndVerifyClientCert,omitempty"`
	DisableClientCerts         bool                  `json:"disableClientCerts,omitempty"`
}

type HTTPTimeoutSpec struct {
	Read       time.Duration `json:"read,omitempty"`
	ReadHeader time.Duration `json:"readHeader,omitempty"`
	Write      time.Duration `json:"write,omitempty"`
	Idle       time.Duration `json:"idle,omitempty"`
}

type ProxyProtocolSpec struct {
	// +kubebuilder:validation:Enum=use_always;allow_authorized;deny_unauthorized
	Behavior string `json:"behavior"`
	// +kubebuilder:validation:MinItems=1
	AuthorizedAddrs []string `json:"authorizedAddrs"`
}

type XForwardedForSpec struct {
	// +kubebuilder:validation:Enum=BASE64;DER;URL
	ClientCertHeaderDecoders []string `json:"ClientCertHeaderDecoders,omitempty"`
	AuthorizedAddrs          []string `json:"AuthorizedAddrs,omitempty"`
	HopSkips                 string   `json:"HopSkips,omitempty"`
	RejectNotPresent         bool     `json:"RejectNotPresent,omitempty"`
	RejectNotAuthorized      bool     `json:"RejectNotAuthorized,omitempty"`
	ClientCertHeader         string   `json:"ClientCertHeader,omitempty"`
}

type CorsSpec struct {
	Enabled        bool     `json:"Enabled,omitempty"`
	AllowedOrigins []string `json:"AllowedOrigins,omitempty"`
	AllowedHeaders []string `json:"AllowedHeaders,omitempty"`
}

type CustomResponseHeadersSpec struct {
	Default              map[string]string                   `json:"default,omitempty"`
	SpecificStatusCode   map[SpecificStatusCodeSpec]string   `json:"specificStatusCode,omitempty"`
	CollectiveStatusCode map[CollectiveStatusCodeSpec]string `json:"collectiveStatusCode,omitempty"`
}

type RedactSpec struct {
	Addresses   bool `json:"Addresses,omitempty"`
	ClusterName bool `json:"ClusterName,omitempty"`
	Version     bool `json:"Version,omitempty"`
}

type CustomMaxJsonSpec struct {
	Depth             int32 `json:"Depth,omitempty"`
	StringValueLength int32 `json:"StringValueLength,omitempty"`
	ObjectEntryCount  int32 `json:"ObjectEntryCount,omitempty"`
	ArrayElementCount int32 `json:"ArrayElementCount,omitempty"`
}
