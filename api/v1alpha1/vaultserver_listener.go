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
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kythe.io/kythe/go/util/datasize"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const listenerTcpHclTemplate = `
listener "tcp" {
    {{ range $key,$val := .MapValue }}
    {{ $key }} = {{ $val }}
    {{ end}}
    {{ range $block,$blockVal := .Blocks }}
    {{ $block }} {
    {{ range $key,$val := $blockVal }}
        {{ $key }} = {{ $val }}
    {{ end}}
    }
    {{ end}}
    {{ if .CustomResponseHeaders }}
        {{ with .CustomResponseHeaders }}
    custom_response_headers {
            {{ range $code,$headers := . }}
        "{{ $code }}" {
                {{ range $key,$val := . }}
            {{ $key | quote }} = [ {{ $val | mapquote | join ","  }} ]
                {{ end}}
        }
            {{ end}}
        {{ end}}
    }
    {{ end }}
}
`

// +kubebuilder:validation:Pattern=([12345][\dx][\dx])|(default)
type SpecificStatusCodeSpec string

type ListenerTelemetrySpec struct {
	UnauthenticatedMetricsAccess *bool `json:"unauthenticatedMetricsAccess,omitempty" hcl:"unauthenticated_metrics_access"`
}

type ListenerProfilingSpec struct {
	UnauthenticatedPprofAccess *bool `json:"unauthenticatedPprofAccess,omitempty" hcl:"unauthenticated_pprof_access"`
}

type ListenerInflightRequestLoggingSpec struct {
	UnauthenticatedInFlightRequestsAccess *bool `json:"unauthenticatedInFlightRequestsAccess,omitempty" hcl:"unauthenticated_in_flight_requests_access"`
}

type ListenerTCPSpec struct {
	ChrootNamespace                   *string `json:"chrootNamespace,omitempty" hcl:"chroot_namespace"`
	RequireRequestHeader              *bool   `json:"requireRequestHeader,omitempty" hcl:"require_request_header"`
	DisableReplicationStatusEndpoints *bool   `json:"disableReplicationStatusEndpoints,omitempty" hcl:"disable_replication_status_endpoints"`
	DisableRequestLimiter             *bool   `json:"disableRequestLimiter,omitempty" hcl:"disable_request_limiter"`

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

func (l *ListenerTCPSpec) Scheme() string {
	if l.TLS.Disable {
		return "HTTP"
	}
	return "HTTPS"
}

func (l *ListenerTCPSpec) Address() *url.URL {
	return &url.URL{Host: "0.0.0.0:8200"}
}

func (l *ListenerTCPSpec) ClusterAddress() *url.URL {
	return &url.URL{Host: "0.0.0.0:8201"}
}

func (l *ListenerTCPSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"address":         strconv.Quote(fmt.Sprintf("%s:%s", l.Address().Hostname(), l.Address().Port())),
		"cluster_address": strconv.Quote(fmt.Sprintf("%s:%s", l.ClusterAddress().Hostname(), l.ClusterAddress().Port())),
	}
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if l.MaxRequestLimit != nil {
		if _s, err := l.MaxRequestLimit.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.TLS != nil {
		if _s, err := l.TLS.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.HTTPTimeout != nil {
		if _s, err := l.HTTPTimeout.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.ProxyProtocol != nil {
		if _s, err := l.ProxyProtocol.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.XForwarded != nil {
		if _s, err := l.XForwarded.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.Cors != nil {
		if _s, err := l.Cors.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.CustomMaxJson != nil {
		if _s, err := l.CustomMaxJson.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if l.Redact != nil {
		if _s, err := l.Redact.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
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

func (l *ListenerTCPSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

func (l *ListenerTCPSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if l.TLS.Cert == nil {
		return []corev1.Volume{}, []corev1.VolumeMount{}, nil
	}
	return l.TLS.Volumes(c, ctx, vaultServer.Namespace)
}

func (l *ListenerTCPSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Funcs(map[string]any{"mapquote": utils.MapQuote}).Parse(listenerTcpHclTemplate))
	if err := template.Execute(&buf, l); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

type ListernerMaxRequestLimitSpec struct {
	Size     *datasize.Size   `json:"size,omitempty" hcl:"max_request_size"`
	Duration *metav1.Duration `json:"duration,omitempty" hcl:"max_request_duration"`
}

func (l *ListernerMaxRequestLimitSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type ListenerTlsSpec struct {
	// +default=false
	Disable    bool            `json:"disable" hcl:"tls_disable"`
	Cert       *SecretSelector `json:"certificate,omitempty"`
	MinVersion *TLSVersion     `json:"minVersion,omitempty" hcl:"tls_min_version"`
	MaxVersion *TLSVersion     `json:"maxVersion,omitempty" hcl:"tls_max_version"`

	// +kubebuilder:validation:Enum=TLS_RSA_WITH_RC4_128_SHA;TLS_RSA_WITH_3DES_EDE_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA;TLS_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA256;TLS_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_RC4_128_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_RSA_WITH_RC4_128_SHA;TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256;TLS_AES_128_GCM_SHA256;TLS_AES_256_GCM_SHA384;TLS_CHACHA20_POLY1305_SHA256
	CipherSuites               *string               `json:"cipherSuites,omitempty" hcl:"tls_cipher_suites"`
	ClientCa                   *ConfigMapKeySelector `json:"clientCa,omitempty"`
	PreferServerCipherSuites   *bool                 `json:"preferServerCipherSuites,omitempty" hcl:"tls_prefer_server_cipher_suites"`
	RequireAndVerifyClientCert *bool                 `json:"requireAndVerifyClientCert,omitempty" hcl:"tls_require_and_verify_client_cert"`
	DisableClientCerts         *bool                 `json:"disableClientCerts,omitempty" hcl:"tls_disable_client_certs"`
}

func (l *ListenerTlsSpec) Volumes(c *client.Client, ctx context.Context, namespace string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	_, err := l.Cert.IsKind(c, ctx, namespace, corev1.SecretTypeTLS)
	if err != nil {
		return nil, nil, err
	}
	volume := corev1.Volume{
		Name: "listenertcp-tls",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: l.Cert.SecretRef.Name,
			},
		},
	}
	volumeMount := corev1.VolumeMount{
		Name:      "listenertcp-tls",
		ReadOnly:  true,
		MountPath: "/tls",
	}
	return []corev1.Volume{volume}, []corev1.VolumeMount{volumeMount}, nil
}

func (l *ListenerTlsSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if !l.Disable {
		_m["tls_cert_file"] = strconv.Quote("/tls/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/tls/tls.key")
		_m["tls_client_ca_file"] = strconv.Quote("/tls/ca.crt")
	}
	return _m, nil
}

type ListenerHTTPTimeoutSpec struct {
	Read       *metav1.Duration `json:"read,omitempty" hcl:"http_read_timeout"`
	ReadHeader *metav1.Duration `json:"readHeader,omitempty" hcl:"http_read_header_timeout"`
	Write      *metav1.Duration `json:"write,omitempty" hcl:"http_write_timeout"`
	Idle       *metav1.Duration `json:"idle,omitempty" hcl:"http_idle_timeout"`
}

func (l *ListenerHTTPTimeoutSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type ListenerProxyProtocolSpec struct {
	// +kubebuilder:validation:Enum=use_always;allow_authorized;deny_unauthorized
	Behavior *string `json:"behavior" hcl:"proxy_protocol_behavior"`
	// +kubebuilder:validation:MinItems=1
	AuthorizedAddrs []string `json:"authorizedAddrs" hcl:"proxy_protocol_authorized_addrs"`
}

func (l *ListenerProxyProtocolSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type ListenerXForwardedForSpec struct {
	// +kubebuilder:validation:Enum=BASE64;DER;URL
	ClientCertHeaderDecoders []string `json:"clientCertHeaderDecoders,omitempty"`
	AuthorizedAddrs          []string `json:"authorizedAddrs,omitempty"`
	HopSkips                 *string  `json:"hopSkips,omitempty" hcl:"x_forwarded_for_hop_skips"`
	RejectNotPresent         *bool    `json:"rejectNotPresent,omitempty" hcl:"x_forwarded_for_reject_not_present"`
	RejectNotAuthorized      *bool    `json:"rejectNotAuthorized,omitempty" hcl:"x_forwarded_for_reject_not_authorized"`
	ClientCertHeader         *string  `json:"clientCertHeader,omitempty" hcl:"x_forwarded_for_client_cert_header"`
}

func (l *ListenerXForwardedForSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if len(l.AuthorizedAddrs) > 0 {
		_m["x_forwarded_for_authorized_addrs"] = strings.Join(l.AuthorizedAddrs, ",")
	}
	if len(l.ClientCertHeaderDecoders) > 0 {
		_m["x_forwarded_for_client_cert_header_decoders"] = strings.Join(l.ClientCertHeaderDecoders, ",")
	}
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type ListenerCorsSpec struct {
	Enabled        *bool    `json:"enabled,omitempty" hcl:"cors_enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty" hcl:"cors_allowed_origins"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty" hcl:"cors_allowed_headers"`
}

func (l *ListenerCorsSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type RedactSpec struct {
	Addresses   *bool `json:"addresses,omitempty" hcl:"redact_addresses"`
	ClusterName *bool `json:"clusterName,omitempty" hcl:"redact_cluster_name"`
	Version     *bool `json:"version,omitempty" hcl:"redact_version"`
}

func (l *RedactSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type CustomMaxJsonSpec struct {
	Depth             *int32 `json:"depth,omitempty" hcl:"max_json_depth"`
	StringValueLength *int32 `json:"stringValueLength,omitempty" hcl:"max_json_string_value_length"`
	ObjectEntryCount  *int32 `json:"objectEntryCount,omitempty" hcl:"max_json_object_entry_count"`
	ArrayElementCount *int32 `json:"arrayElementCount,omitempty" hcl:"max_json_array_element_count"`
}

func (l *CustomMaxJsonSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*l); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}
