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
	"strconv"
	"strings"

	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvFromSource represents the source of a set of ConfigMaps or Secrets

// VaultServer is the Schema for the vaultservers API.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type VaultServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   VaultServerSpec   `json:"spec"`
	Status VaultServerStatus `json:"status,omitempty"`
}

func (v *VaultServer) Volumes() ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, volumeMounts := []corev1.Volume{}, []corev1.VolumeMount{}
	return volumes, volumeMounts, nil
}

type VaultServerSpec struct {
	Size   int32             `json:"size"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels,omitempty"`

	PersistentVolumeClassName string                         `json:"persistentVolumeClassName,omitempty"`
	Config                    *VaultServerConfigSpec         `json:"config"`
	SecretOverride            *VaultServerSecretOverrideSpec `json:"secretMapOverride,omitempty"`
}

type VaultServerConfigSpec struct {
	// Operator Managed Config
	ClusterName                    *string            `json:"clusterName,omitempty"`
	Ui                             *bool              `json:"ui,omitempty"`
	DisableMLock                   *bool              `json:"disableMLock,omitempty"`
	CacheSize                      *int32             `json:"cacheSize,omitempty"`
	DisableCache                   *bool              `json:"disableCache,omitempty"`
	DefaultLeaseTTL                *metav1.Duration   `json:"defaultLeaseTTL,omitempty"`
	MaxLeaseTTL                    *metav1.Duration   `json:"maxLeaseTTL,omitempty"`
	DefaultMaxRequestDuration      *metav1.Duration   `json:"defaultMaxRequestDuration,omitempty"`
	RawStorageEndpoint             *bool              `json:"rawStorageEndpoint,omitempty"`
	IntrospectionEndpoint          *bool              `json:"introspectionEndpoint,omitempty"`
	EnableResponseHeaderHostname   *bool              `json:"enableResponseHeaderHostname,omitempty"`
	EnableResponseHeaderRaftNodeId *bool              `json:"enableResponseHeaderRaftNodeId,omitempty"`
	AllowAuditLogPrefixing         *bool              `json:"allowAuditLogPrefixing,omitempty"`
	Experiments                    []string           `json:"experiments,omitempty"`
	ImpreciseLeaseRoleTracking     *bool              `json:"impreciseLeaseRoleTracking,omitempty"`
	EnablePostUnsealTrace          *bool              `json:"enablePostUnsealTrace,omitempty"`
	PostUnsealTraceDirectory       *string            `json:"postUnsealTraceDirectory,omitempty"`
	DisableClustering              *bool              `json:"disableClustering,omitempty"`
	DisableSealwrap                *bool              `json:"disableSealwrap,omitempty"`
	DisablePerformanceStandby      *bool              `json:"disablePerformanceStandby,omitempty"`
	License                        *SecretKeySelector `json:"license,omitempty"`
	AdministrativeNamespacePath    *string            `json:"administrativeNamespacePath,omitempty"`
	RemoveIrrevocableLeaseAfter    *metav1.Duration   `json:"removeIrrevocableLeaseAfter,omitempty"`

	// +kubebuilder:validation:Enum=statelock;quotas;expiration
	DetectDeadlocks *string `json:"detectDeadlocks,omitempty"`
	// +kubebuilder:validation:Enum=trace;debug;info;warn;error
	LogLevel *string `json:"logLevel,omitempty"`
	// +kubebuilder:validation:Enum=trace;debug;info;warn;error;off
	LogRequestsLevel *string `json:"logRequestsLevel,omitempty"`
	// +kubebuilder:validation:Enum=standard;json
	LogFormat *string `json:"logFormat,omitempty"`

	// OSS features stanza
	ListenerTcp *ListenerTCPSpec  `json:"listenerTcp"`
	Telemetry   *TelemetrySpec    `json:"telemetry,omitempty"`
	UserLockout []UserLockoutSpec `json:"userLockout,omitempty"`
	// +kubebuilder:validation:MaxItems=2
	Seal                []SealSpec                `json:"seal,omitempty"`
	ServiceRegistration []ServiceRegistrationSpec `json:"serviceRegistration,omitempty"`
	Storage             *StorageSpec              `json:"storage"`

	// Enterprise features stanza
	KMSLibrary                 *KmsLibrarySpec                 `json:"kmsLibrary,omitempty"`
	Replication                *ReplicationSpec                `json:"replication,omitempty"`
	Reporting                  *ReportingSpec                  `json:"reporting,omitempty"`
	Sentinel                   *SentinelSpec                   `json:"sentinel,omitempty"`
	AdaptiveOverloadProtection *AdaptiveOverloadProtectionSpec `json:"adaptiveOverloadProtection,omitempty"`
}

func (s *VaultServerConfigSpec) MapValue() map[string]any {
	scheme := "https"
	if s.ListenerTcp.TLS.Disable {
		scheme = "http"
	}

	_m := map[string]any{
		"cluster_addr":     strconv.Quote(fmt.Sprintf("%s://127.0.0.1:8201", scheme)),
		"api_addr":         strconv.Quote(fmt.Sprintf("%s://127.0.0.1:8200", scheme)),
		"pid_file":         strconv.Quote("/run/vault.pid"),
		"log_file":         strconv.Quote("/dev/stdout"),
		"plugin_tmpdir":    strconv.Quote("/plugins/tmp"),
		"plugin_directory": strconv.Quote("/plugins"),
	}

	if s.Ui != nil {
		_m["ui"] = strconv.FormatBool(*s.Ui)
	}
	if s.ClusterName != nil {
		_m["cluster_name"] = strconv.Quote(*s.ClusterName)
	}
	if s.DisableMLock != nil {
		_m["disable_mlock"] = strconv.FormatBool(*s.DisableMLock)
	}
	if s.CacheSize != nil {
		_m["cache_sizes"] = strconv.FormatInt(int64(*s.CacheSize), 10)
	}
	if s.DisableCache != nil {
		_m["disable_cache"] = strconv.FormatBool(*s.DisableCache)
	}
	if s.DefaultLeaseTTL != nil {
		_m["default_lease_ttl"] = strconv.Quote(fmt.Sprintf("%s", s.DefaultLeaseTTL.Duration))
	}
	if s.MaxLeaseTTL != nil {
		_m["max_lease_ttl"] = strconv.Quote(fmt.Sprintf("%s", s.MaxLeaseTTL.Duration))
	}
	if s.DefaultMaxRequestDuration != nil {
		_m["default_max_request_duration"] = strconv.Quote(fmt.Sprintf("%s", s.DefaultMaxRequestDuration.Duration))
	}
	if s.RawStorageEndpoint != nil {
		_m["raw_storage_endpoint"] = strconv.FormatBool(*s.RawStorageEndpoint)
	}
	if s.IntrospectionEndpoint != nil {
		_m["introspection_endpoint"] = strconv.FormatBool(*s.IntrospectionEndpoint)
	}
	if s.EnableResponseHeaderHostname != nil {
		_m["enable_response_header_hostname"] = strconv.FormatBool(*s.EnableResponseHeaderHostname)
	}
	if s.EnableResponseHeaderRaftNodeId != nil {
		_m["enable_response_header_raft_node_id"] = strconv.FormatBool(*s.EnableResponseHeaderRaftNodeId)
	}
	if s.AllowAuditLogPrefixing != nil {
		_m["allow_audit_log_prefixing"] = strconv.FormatBool(*s.AllowAuditLogPrefixing)
	}
	if len(s.Experiments) > 0 {
		_m["disable_cache"] = fmt.Sprintf("[%s]", strings.Join(utils.Map(strconv.Quote, s.Experiments), ","))
	}
	if s.ImpreciseLeaseRoleTracking != nil {
		_m["imprecise_lease_role_tracking"] = strconv.FormatBool(*s.ImpreciseLeaseRoleTracking)
	}
	if s.EnablePostUnsealTrace != nil {
		_m["enable_post_unseal_trace"] = strconv.FormatBool(*s.EnablePostUnsealTrace)
	}
	if s.PostUnsealTraceDirectory != nil {
		_m["post_unseal_trace_directory"] = strconv.Quote(*s.PostUnsealTraceDirectory)
	}
	if s.DisableClustering != nil {
		_m["disable_clustering"] = strconv.FormatBool(*s.DisableClustering)
	}
	if s.DisableSealwrap != nil {
		_m["disable_sealwrap"] = strconv.FormatBool(*s.DisableSealwrap)
	}
	if s.DisablePerformanceStandby != nil {
		_m["disable_performance_standby"] = strconv.FormatBool(*s.DisablePerformanceStandby)
	}
	if s.License != nil {
		_m["license_path"] = strconv.Quote("/vault-license")
	}
	if s.AdministrativeNamespacePath != nil {
		_m["administrative_namespace_path"] = strconv.Quote(*s.AdministrativeNamespacePath)
	}
	if s.RemoveIrrevocableLeaseAfter != nil {
		_m["remove_irrevocable_lease_after"] = strconv.Quote(fmt.Sprintf("%s", s.RemoveIrrevocableLeaseAfter.Duration))
	}
	if s.DetectDeadlocks != nil {
		_m["detect_deadlocks"] = strconv.Quote(*s.DetectDeadlocks)
	}
	if s.LogLevel != nil {
		_m["log_level"] = strconv.Quote(*s.LogLevel)
	}
	if s.LogRequestsLevel != nil {
		_m["log_requests_level"] = strconv.Quote(*s.LogRequestsLevel)
	}
	if s.LogFormat != nil {
		_m["log_format"] = strconv.Quote(*s.LogFormat)
	}
	return _m
}

type VaultServerSecretOverrideSpec struct {
	Name string            `json:"configMapOverride,omitempty"`
	Data map[string][]byte `json:"data,omitempty"`
}

// VaultServerStatus defines the observed state of VaultServer.
type VaultServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	metav1.ListMeta `json:"metadata,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// VaultServerList contains a list of VaultServer.
// +kubebuilder:object:root=true
type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items"`
}

func (v *VaultServer) GetConfigMapNameForVaultConfig() string {
	generateName := fmt.Sprintf("%s-config", v.GetObjectMeta().GetName())
	if v.Spec.SecretOverride == nil || v.Spec.SecretOverride.Name == "" {
		return generateName
	}
	return v.Spec.SecretOverride.Name
}

func init() {
	SchemeBuilder.Register(&VaultServer{}, &VaultServerList{})
}
