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
	"fmt"
	"maps"
	"net/url"
	"strconv"
	"strings"

	"github.com/jynolen/vault-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kythe.io/kythe/go/util/datasize"
)

// +kubebuilder:validation:Pattern=([12345][\dx][\dx])|(default)
type SpecificStatusCodeSpec string

type ListenerTelemetrySpec struct {
	UnauthenticatedMetricsAccess *bool `json:"unauthenticatedMetricsAccess,omitempty"`
}

type ListenerProfilingSpec struct {
	UnauthenticatedPprofAccess *bool `json:"unauthenticatedPprofAccess,omitempty"`
}

type ListenerInflightRequestLoggingSpec struct {
	UnauthenticatedInFlightRequestsAccess *bool `json:"unauthenticatedInFlightRequestsAccess,omitempty"`
}

type ListenerTCPSpec struct {
	ChrootNamespace                   *string `json:"chrootNamespace,omitempty"`
	RequireRequestHeader              *bool   `json:"requireRequestHeader,omitempty"`
	DisableReplicationStatusEndpoints *bool   `json:"disableReplicationStatusEndpoints,omitempty"`
	DisableRequestLimiter             *bool   `json:"disableRequestLimiter,omitempty"`

	MaxRequestLimit *ListernerMaxRequestLimitSpec `json:"maxRequestLimit,omitempty"`
	TLS             *ListenerTlsSpec              `json:"tls"`
	HTTPTimeout     *ListenerHTTPTimeoutSpec      `json:"httpTimeout,omitempty"`
	ProxyProtocol   *ListenerProxyProtocolSpec    `json:"proxyProcotol,omitempty"`
	XForwarded      *ListenerXForwardedForSpec    `json:"xForwardedFor,omitempty"`
	Cors            *ListenerCorsSpec             `json:"cors,omitempty"`
	CustomMaxJson   *CustomMaxJsonSpec            `json:"customMaxJson,omitempty"`
	Redact          *RedactSpec                   `json:"redact,omitempty"`

	Profiling              *ListenerProfilingSpec                         `json:"profiling,omitempty"`
	Telemetry              *ListenerTelemetrySpec                         `json:"telemetry,omitempty"`
	InFlightRequestsAccess *ListenerInflightRequestLoggingSpec            `json:"inFlightRequestsAccess,omitempty"`
	CustomResponseHeaders  map[SpecificStatusCodeSpec]map[string][]string `json:"customResponseHeaders,omitempty"`
}

func (l *ListenerTCPSpec) Address() *url.URL {
	return &url.URL{Host: "127.0.0.1:8200"}
}

func (l *ListenerTCPSpec) ClusterAddress() *url.URL {
	return &url.URL{Host: "127.0.0.1:8201"}
}

func (l *ListenerTCPSpec) MapValue() map[string]string {
	_m := map[string]string{
		"address":      strconv.Quote(fmt.Sprintf("%s:%s", l.Address().Hostname(), l.Address().Port())),
		"cluster_addr": strconv.Quote(fmt.Sprintf("%s:%s", l.ClusterAddress().Hostname(), l.ClusterAddress().Port())),
	}
	if l.ChrootNamespace != nil {
		_m["chroot_namespace"] = strconv.Quote(*l.ChrootNamespace)
	}
	if l.DisableReplicationStatusEndpoints != nil {
		_m["disable_replication_status_endpoints"] = strconv.FormatBool(*l.DisableReplicationStatusEndpoints)
	}
	if l.DisableRequestLimiter != nil {
		_m["disable_request_limiter"] = strconv.FormatBool(*l.DisableRequestLimiter)
	}
	if l.RequireRequestHeader != nil {
		_m["require_request_header"] = strconv.FormatBool(*l.RequireRequestHeader)
	}

	if l.MaxRequestLimit != nil {
		maps.Copy(_m, l.MaxRequestLimit.MapValue())
	}
	if l.TLS != nil {
		maps.Copy(_m, l.TLS.MapValue())
	}
	if l.HTTPTimeout != nil {
		maps.Copy(_m, l.HTTPTimeout.MapValue())
	}
	if l.ProxyProtocol != nil {
		maps.Copy(_m, l.ProxyProtocol.MapValue())
	}
	if l.XForwarded != nil {
		maps.Copy(_m, l.XForwarded.MapValue())
	}
	if l.Cors != nil {
		maps.Copy(_m, l.Cors.MapValue())
	}
	if l.CustomMaxJson != nil {
		maps.Copy(_m, l.CustomMaxJson.MapValue())
	}
	if l.Redact != nil {
		maps.Copy(_m, l.Redact.MapValue())
	}
	return _m
}

func (l *ListenerTCPSpec) Blocks() map[string]map[string]string {
	_m := map[string]map[string]string{}

	if l.Telemetry != nil && l.Telemetry.UnauthenticatedMetricsAccess != nil {
		_m["telemetry"] = map[string]string{
			"unauthenticated_metrics_access": strconv.FormatBool(*l.Telemetry.UnauthenticatedMetricsAccess),
		}
	}
	if l.Profiling != nil && l.Profiling.UnauthenticatedPprofAccess != nil {
		_m["profiling"] = map[string]string{
			"unauthenticated_pprof_access": strconv.FormatBool(*l.Profiling.UnauthenticatedPprofAccess),
		}
	}
	if l.InFlightRequestsAccess != nil && l.InFlightRequestsAccess.UnauthenticatedInFlightRequestsAccess != nil {
		_m["inflight_requests_logging"] = map[string]string{
			"unauthenticated_in_flight_requests_access": strconv.FormatBool(*l.InFlightRequestsAccess.UnauthenticatedInFlightRequestsAccess),
		}
	}
	return _m
}

type ListernerMaxRequestLimitSpec struct {
	Size     *datasize.Size   `json:"size,omitempty"`
	Duration *metav1.Duration `json:"duration,omitempty"`
}

func (l *ListernerMaxRequestLimitSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if l.Size != nil {
		_m["max_request_size"] = fmt.Sprintf("%dkb", int(l.Size.Kilobytes()))
	}
	if l.Duration != nil {
		_m["max_request_duration"] = fmt.Sprintf("%s", l.Duration.Duration)
	}
	return _m
}

type ListenerTlsSpec struct {
	// +default=false
	Disable    bool            `json:"disable"`
	Cert       *SecretSelector `json:"certificate,omitempty"`
	MinVersion *TLSVersion     `json:"minVersion,omitempty"`
	MaxVersion *TLSVersion     `json:"maxVersion,omitempty"`

	// +kubebuilder:validation:Enum=TLS_RSA_WITH_RC4_128_SHA;TLS_RSA_WITH_3DES_EDE_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA;TLS_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA256;TLS_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_RC4_128_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_RSA_WITH_RC4_128_SHA;TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256;TLS_AES_128_GCM_SHA256;TLS_AES_256_GCM_SHA384;TLS_CHACHA20_POLY1305_SHA256
	CipherSuites               *string               `json:"cipherSuites,omitempty"`
	ClientCa                   *ConfigMapKeySelector `json:"clientCa,omitempty"`
	PreferServerCipherSuites   *bool                 `json:"preferServerCipherSuites,omitempty"`
	RequireAndVerifyClientCert *bool                 `json:"requireAndVerifyClientCert,omitempty"`
	DisableClientCerts         *bool                 `json:"disableClientCerts,omitempty"`
}

func (l *ListenerTlsSpec) MapValue() map[string]string {
	_m := map[string]string{
		"tls_disable": strconv.FormatBool(l.Disable),
	}
	if !l.Disable {
		_m["tls_cert_file"] = strconv.Quote("/tls/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/tls/tls.key")
		_m["tls_client_ca_file"] = strconv.Quote("/tls/ca.crt")
	}

	if l.MinVersion != nil {
		_m["tls_min_version"] = strconv.Quote(string(*l.MinVersion))
	}
	if l.MaxVersion != nil {
		_m["tls_max_version"] = strconv.Quote(string(*l.MaxVersion))
	}
	if l.CipherSuites != nil {
		_m["tls_prefer_server_cipher_suites"] = strconv.Quote(*l.CipherSuites)
	}
	if l.RequireAndVerifyClientCert != nil {
		_m["tls_require_and_verify_client_cert"] = strconv.FormatBool(*l.RequireAndVerifyClientCert)
	}
	if l.DisableClientCerts != nil {
		_m["tls_disable_client_certs"] = strconv.FormatBool(*l.DisableClientCerts)
	}
	return _m
}

type ListenerHTTPTimeoutSpec struct {
	Read       *metav1.Duration `json:"read,omitempty"`
	ReadHeader *metav1.Duration `json:"readHeader,omitempty"`
	Write      *metav1.Duration `json:"write,omitempty"`
	Idle       *metav1.Duration `json:"idle,omitempty"`
}

func (l *ListenerHTTPTimeoutSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if l.Idle != nil {
		_m["http_idle_timeout"] = strconv.Quote(fmt.Sprintf("%s", l.Idle.Duration))
	}
	if l.ReadHeader != nil {
		_m["http_read_header_timeout"] = strconv.Quote(fmt.Sprintf("%s", l.ReadHeader.Duration))
	}
	if l.Write != nil {
		_m["http_write_timeout"] = strconv.Quote(fmt.Sprintf("%s", l.Write.Duration))
	}
	if l.Read != nil {
		_m["http_read_timeout"] = strconv.Quote(fmt.Sprintf("%s", l.Read.Duration))
	}
	return _m
}

type ListenerProxyProtocolSpec struct {
	// +kubebuilder:validation:Enum=use_always;allow_authorized;deny_unauthorized
	Behavior *string `json:"behavior"`
	// +kubebuilder:validation:MinItems=1
	AuthorizedAddrs []string `json:"authorizedAddrs"`
}

func (l *ListenerProxyProtocolSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if l.Behavior != nil {
		_m["proxy_protocol_behavior"] = strconv.Quote(*l.Behavior)
	}
	if len(l.AuthorizedAddrs) > 0 {
		_m["proxy_protocol_authorized_addrs"] = fmt.Sprintf("[%s]", strings.Join(utils.Map(strconv.Quote, l.AuthorizedAddrs), ","))
	}
	return _m
}

type ListenerXForwardedForSpec struct {
	// +kubebuilder:validation:Enum=BASE64;DER;URL
	ClientCertHeaderDecoders []string `json:"clientCertHeaderDecoders,omitempty"`
	AuthorizedAddrs          []string `json:"authorizedAddrs,omitempty"`
	HopSkips                 *string  `json:"hopSkips,omitempty"`
	RejectNotPresent         *bool    `json:"rejectNotPresent,omitempty"`
	RejectNotAuthorized      *bool    `json:"rejectNotAuthorized,omitempty"`
	ClientCertHeader         *string  `json:"clientCertHeader,omitempty"`
}

func (l *ListenerXForwardedForSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if len(l.AuthorizedAddrs) > 0 {
		_m["x_forwarded_for_authorized_addrs"] = strings.Join(l.AuthorizedAddrs, ",")
	}
	if len(l.ClientCertHeaderDecoders) > 0 {
		_m["x_forwarded_for_client_cert_header_decoders"] = strings.Join(l.ClientCertHeaderDecoders, ",")
	}
	if l.ClientCertHeader != nil {
		_m["x_forwarded_for_client_cert_header"] = strconv.Quote(*l.ClientCertHeader)
	}
	if l.HopSkips != nil {
		_m["x_forwarded_for_hop_skips"] = strconv.Quote(*l.HopSkips)
	}
	if l.RejectNotAuthorized != nil {
		_m["x_forwarded_for_reject_not_authorized"] = strconv.FormatBool(*l.RejectNotAuthorized)
	}
	if l.RejectNotAuthorized != nil {
		_m["x_forwarded_for_reject_not_present"] = strconv.FormatBool(*l.RejectNotPresent)
	}
	return _m
}

type ListenerCorsSpec struct {
	Enabled        *bool    `json:"enabled,omitempty"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
}

func (l *ListenerCorsSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if len(l.AllowedOrigins) > 0 {
		_m["cors_allowed_origins"] = fmt.Sprintf("[%s]", strings.Join(utils.Map(strconv.Quote, l.AllowedOrigins), ","))
	}
	if len(l.AllowedHeaders) > 0 {
		_m["cors_allowed_headers"] = fmt.Sprintf("[%s]", strings.Join(utils.Map(strconv.Quote, l.AllowedHeaders), ","))
	}
	if l.Enabled != nil {
		_m["cors_enabled"] = strconv.FormatBool(*l.Enabled)
	}
	return _m
}

type RedactSpec struct {
	Addresses   *bool `json:"Addresses,omitempty"`
	ClusterName *bool `json:"ClusterName,omitempty"`
	Version     *bool `json:"Version,omitempty"`
}

func (l *RedactSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if l.Addresses != nil {
		_m["redact_addresses"] = strconv.FormatBool(*l.Addresses)
	}
	if l.ClusterName != nil {
		_m["redact_cluster_name"] = strconv.FormatBool(*l.ClusterName)
	}
	if l.Version != nil {
		_m["redact_version"] = strconv.FormatBool(*l.Version)
	}
	return _m
}

type CustomMaxJsonSpec struct {
	Depth             *int32 `json:"Depth,omitempty"`
	StringValueLength *int32 `json:"StringValueLength,omitempty"`
	ObjectEntryCount  *int32 `json:"ObjectEntryCount,omitempty"`
	ArrayElementCount *int32 `json:"ArrayElementCount,omitempty"`
}

func (l *CustomMaxJsonSpec) MapValue() map[string]string {
	_m := map[string]string{}
	if l.Depth != nil {
		_m["max_json_depth"] = strconv.FormatInt(int64(*l.Depth), 10)
	}
	if l.StringValueLength != nil {
		_m["max_json_string_value_length"] = strconv.FormatInt(int64(*l.StringValueLength), 10)
	}
	if l.ObjectEntryCount != nil {
		_m["max_json_object_entry_count"] = strconv.FormatInt(int64(*l.ObjectEntryCount), 10)
	}
	if l.ArrayElementCount != nil {
		_m["max_json_array_element_count"] = strconv.FormatInt(int64(*l.ArrayElementCount), 10)
	}
	return _m
}
