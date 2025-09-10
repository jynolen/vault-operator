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
	"time"

	url "github.com/jynolen/vault-operator/internal/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvFromSource represents the source of a set of ConfigMaps or Secrets

// VaultServer is the Schema for the vaultservers API.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type VaultServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   VaultServerSpec   `json:"spec"`
	Status VaultServerStatus `json:"status"`
}

type VaultServerSpec struct {
	Size   int32             `json:"size"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels,omitempty"`

	Config            VaultServerConfigSpec             `json:"config"`
	ConfigMapOverride *VaultServerConfigMapOverrideSpec `json:"configMapOverride,omitempty"`
}

type VaultServerConfigSpec struct {
	// Operator Managed Config
	ClusterAddr *url.URL `json:"-"`
	ApiAddr     *url.URL `json:"-"`

	// +default=/plugins/lib
	PluginDirectory string `json:"-"`
	// +default=/plugins/tmp
	PluginTmpdir string `json:"-"`
	// +default=0
	PluginFileUid int32 `json:"-"`

	// +default=/run/vault.pid
	PidFile string `json:"-"`
	// +default=/dev/stdout
	LogFile string `json:"-"`

	// User Managed Config
	ClusterName string `json:"clusterName,omitempty"`

	// +default=false
	Ui bool `json:"ui,omitempty"`
	// +default=false
	DisableMLock bool `json:"disableMLock,omitempty"`

	// +default=131072
	CacheSize int32 `json:"cacheSize,omitempty"`
	// +default=false
	DisableSize bool `json:"disableCache,omitempty"`
	// +default=768h
	DefaultLeaseTTL time.Duration `json:"defaultLeaseTTL,omitempty"`
	// +default=768h
	MaxLeaseTTL time.Duration `json:"maxLeaseTTL,omitempty"`
	// +default=90s
	DefaultMaxRequestDuration time.Duration `json:"defaultMaxRequestDuration,omitempty"`

	// +kubebuilder:validation:Enum=statelock;quotas;expiration
	DetectDeadlocks string `json:"detectDeadlocks,omitempty"`
	// +default=false
	RawStorageEndpoint bool `json:"rawStorageEndpoint,omitempty"`
	// +default=false
	IntrospectionEndpoint bool `json:"introspectionEndpoint,omitempty"`
	// +default=false
	EnableResponseHeaderHostname bool `json:"enableResponseHeaderHostname,omitempty"`
	// +default=false
	EnableResponseHeaderRaftNodeId bool `json:"enableResponseHeaderRaftNodeId,omitempty"`

	// +kubebuilder:validation:Enum=trace;debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// +kubebuilder:validation:Enum=trace;debug;info;warn;error;off
	LogRequestsLevel string `json:"logRequestsLevel,omitempty"`
	// +default=standard
	AllowAuditLogPrefixing bool `json:"allowAuditLogPrefixing,omitempty"`

	// +kubebuilder:validation:Enum=standard;json
	// +default=standard
	LogFormat   string   `json:"logFormat,omitempty"`
	Experiments []string `json:"experiments,omitempty"`
	// +default=false
	ImpreciseLeaseRoleTracking bool `json:"impreciseLeaseRoleTracking,omitempty"`
	// +default=false
	EnablePostUnsealTrace    bool   `json:"enablePostUnsealTrace,omitempty"`
	PostUnsealTraceDirectory string `json:"postUnsealTraceDirectory,omitempty"`

	// +default=false
	DisableClustering bool `json:"disableClustering,omitempty"`
	// +default=false
	DisableSealwrap bool `json:"disableSealwrap,omitempty"`
	// +default=false
	DisablePerformanceStandby   bool               `json:"disablePerformanceStandby,omitempty"`
	License                     *SecretKeySelector `json:"license,omitempty"`
	AdministrativeNamespacePath string             `json:"administrativeNamespacePath,omitempty"`
	// +default=2d
	RemoveIrrevocableLeaseAfter time.Duration `json:"removeIrrevocableLeaseAfter,omitempty"`

	// OSS features stanza
	Listener            *ListenerSpec            `json:"listener"`
	Telemetry           *TelemetrySpec           `json:"telemetry,omitempty"`
	UserLockout         []UserLockoutSpec        `json:"userLockout,omitempty"`
	Seal                *SealSpec                `json:"seal,omitempty"`
	ServiceRegistration *ServiceRegistrationSpec `json:"serviceRegistration,omitempty"`
	Storage             *StorageSpec             `json:"storage,omitempty"`

	// Enterprise features stanza
	Entropy                    *EntropySpec                    `json:"entropy,omitempty"`
	KMSLibrary                 *KmsLibrarySpec                 `json:"kmsLibrary,omitempty"`
	Replication                *ReplicationSpec                `json:"replication,omitempty"`
	Reporting                  *ReportingSpec                  `json:"reporting,omitempty"`
	SentinelSpec               *SentinelSpec                   `json:"sentinel,omitempty"`
	AdaptiveOverloadProtection *AdaptiveOverloadProtectionSpec `json:"adaptiveOverloadProtection,omitempty"`
}

type VaultServerConfigMapOverrideSpec struct {
	Name string            `json:"configMapOverride,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

// VaultServerStatus defines the observed state of VaultServer.
type VaultServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	metav1.ListMeta `json:"metadata,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true

// VaultServerList contains a list of VaultServer.
type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items"`
}

func (v *VaultServer) GetConfigMapNameForVaultConfig() string {
	generateName := fmt.Sprintf("%s-config", v.GetObjectMeta().GetName())
	if v.Spec.ConfigMapOverride == nil || v.Spec.ConfigMapOverride.Name == "" {
		return generateName
	}
	return v.Spec.ConfigMapOverride.Name
}

func init() {
	SchemeBuilder.Register(&VaultServer{}, &VaultServerList{})
}
