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
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/internal/utils"
	"github.com/samber/lo"
	"github.com/zeebo/xxh3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const vaultServerHclTemplate = `
{{ range $key,$val := .MapValue }}
{{ $key }} = {{ $val }}
{{ end}}
`

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

func (v *VaultServer) ServerConfigFilePath() string {
	return fmt.Sprintf("/vault/%s", v.ServerConfigFileName())
}

func (v *VaultServer) Labels() map[string]string {
	labels := v.ObjectMeta.Labels
	labels["vault-operator.io/instance"] = v.Name
	return labels
}

func (v *VaultServer) ResolveSecret(c *client.Client, ctx context.Context) error {
	if err := v.Spec.Config.ResolveSecret(c, ctx, v); err != nil {
		return err
	}
	return nil
}

func (v *VaultServer) ServerConfigFileName() string {
	return "vault.hcl"
}

func (v *VaultServer) Volumes(c *client.Client, ctx context.Context) ([]corev1.Volume, []corev1.VolumeMount, error) {
	mode, quantity, volumes, mounts := int32(420), lo.Must(resource.ParseQuantity("128Mi")), []corev1.Volume{}, []corev1.VolumeMount{}
	volumes = append(volumes, corev1.Volume{
		Name: fmt.Sprintf("%s-server-config", v.GetSecretNameForVaultConfig()),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  v.GetSecretNameForVaultConfig(),
				DefaultMode: &mode,
			},
		},
	})
	volumes = append(volumes, corev1.Volume{
		Name: fmt.Sprintf("%s-base-folder", v.GetSecretNameForVaultConfig()),
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: &quantity,
				Medium:    corev1.StorageMediumDefault,
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      fmt.Sprintf("%s-server-config", v.GetSecretNameForVaultConfig()),
		MountPath: v.ServerConfigFilePath(),
		SubPath:   v.ServerConfigFileName(),
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      fmt.Sprintf("%s-base-folder", v.GetSecretNameForVaultConfig()),
		MountPath: "/vault",
	})
	subVols, subVolMounts, err := v.Spec.Config.Volumes(c, ctx, v)
	if err != nil {
		return nil, nil, err
	}
	return append(volumes, subVols...), append(mounts, subVolMounts...), nil
}

func (v *VaultServer) PersistentVolumeClaimTemplates() []corev1.PersistentVolumeClaim {
	pvcTemplates := v.Spec.PersistentVolumeClaim
	return append(pvcTemplates, v.Spec.Config.PersistentVolumeClaim()...)
}

func (v *VaultServer) ConfigHash() string {
	hcl, _ := v.HclRender()
	return fmt.Sprintf("0x%x", xxh3.HashString(hcl))
}

type VaultServerSpec struct {
	//+default=15
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds"`
	//+default=3
	RevisionHistoryLimit int32             `json:"revisionHistoryLimit"`
	Replicas             int32             `json:"replicas"`
	Image                string            `json:"image"`
	Labels               map[string]string `json:"labels,omitempty"`

	PersistentVolumeClaim []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	Config                *VaultServerConfigSpec         `json:"config"`
	SecretOverride        *VaultServerSecretOverrideSpec `json:"secretMapOverride,omitempty"`
	Service               VaultServerServiceSpec         `json:"service,omitempty"`
}

type VaultServerServiceSpec struct {
	//+default="ClusterIP"
	Type corev1.ServiceType `json:"type,omitempty"`
}

type VaultServerConfigSpec struct {
	*HclHelper `json:""`
	// Operator Managed Config
	ClusterName *string `json:"clusterName,omitempty" hcl:"cluster_name"`
	Ui          *bool   `json:"ui,omitempty" hcl:"ui"`
	// +default=true
	DisableMLock                   bool               `json:"disableMLock,omitempty" hcl:"disable_mlock"`
	CacheSize                      *int32             `json:"cacheSize,omitempty" hcl:"cache_size"`
	DisableCache                   *bool              `json:"disableCache,omitempty" hcl:"disable_cache"`
	DefaultLeaseTTL                *metav1.Duration   `json:"defaultLeaseTTL,omitempty" hcl:"default_lease_ttl"`
	MaxLeaseTTL                    *metav1.Duration   `json:"maxLeaseTTL,omitempty" hcl:"max_lease_ttl"`
	DefaultMaxRequestDuration      *metav1.Duration   `json:"defaultMaxRequestDuration,omitempty" hcl:"default_max_request_duration"`
	RawStorageEndpoint             *bool              `json:"rawStorageEndpoint,omitempty" hcl:"raw_storage_endpoint"`
	IntrospectionEndpoint          *bool              `json:"introspectionEndpoint,omitempty" hcl:"introspection_endpoint"`
	EnableResponseHeaderHostname   *bool              `json:"enableResponseHeaderHostname,omitempty" hcl:"enable_response_header_hostname"`
	EnableResponseHeaderRaftNodeId *bool              `json:"enableResponseHeaderRaftNodeId,omitempty" hcl:"enable_response_header_raft_node_id"`
	AllowAuditLogPrefixing         *bool              `json:"allowAuditLogPrefixing,omitempty" hcl:"allow_audit_log_prefixing"`
	Experiments                    []string           `json:"experiments,omitempty" hcl:"experiments"`
	ImpreciseLeaseRoleTracking     *bool              `json:"impreciseLeaseRoleTracking,omitempty" hcl:"imprecise_lease_role_tracking"`
	EnablePostUnsealTrace          *bool              `json:"enablePostUnsealTrace,omitempty" hcl:"enable_post_unseal_trace"`
	PostUnsealTraceDirectory       *string            `json:"postUnsealTraceDirectory,omitempty" hcl:"post_unseal_trace_directory"`
	DisableClustering              *bool              `json:"disableClustering,omitempty" hcl:"disable_clustering"`
	DisableSealwrap                *bool              `json:"disableSealwrap,omitempty" hcl:"disable_sealwrap"`
	DisablePerformanceStandby      *bool              `json:"disablePerformanceStandby,omitempty" hcl:"disable_performance_standby"`
	License                        *SecretKeySelector `json:"license,omitempty"`
	AdministrativeNamespacePath    *string            `json:"administrativeNamespacePath,omitempty" hcl:"administrative_namespace_path"`
	RemoveIrrevocableLeaseAfter    *metav1.Duration   `json:"removeIrrevocableLeaseAfter,omitempty" hcl:"remove_irrevocable_lease_after"`

	// +kubebuilder:validation:Enum=statelock;quotas;expiration
	DetectDeadlocks *string `json:"detectDeadlocks,omitempty" hcl:"detect_deadlocks"`
	// +kubebuilder:validation:Enum=trace;debug;info;warn;error
	LogLevel *string `json:"logLevel,omitempty" hcl:"log_level"`
	// +kubebuilder:validation:Enum=trace;debug;info;warn;error;off
	LogRequestsLevel *string `json:"logRequestsLevel,omitempty" hcl:"log_requests_level"`
	// +kubebuilder:validation:Enum=standard;json
	LogFormat *string `json:"logFormat,omitempty" hcl:"log_format"`

	// OSS features stanza
	ListenerTcp *ListenerTCPSpec `json:"listenerTcp"`
	Telemetry   *TelemetrySpec   `json:"telemetry,omitempty"`
	UserLockout UserLockoutList  `json:"userLockout,omitempty"`

	Seal                SealListSpec            `json:"seal,omitempty"`
	ServiceRegistration ServiceRegistrationList `json:"serviceRegistration,omitempty"`
	Storage             *StorageSpec            `json:"storage"`

	// Enterprise features stanza
	KMSLibrary                 *KmsLibrarySpec                 `json:"kmsLibrary,omitempty"`
	Replication                *ReplicationSpec                `json:"replication,omitempty"`
	Reporting                  *ReportingSpec                  `json:"reporting,omitempty"`
	Sentinel                   *SentinelSpec                   `json:"sentinel,omitempty"`
	AdaptiveOverloadProtection *AdaptiveOverloadProtectionSpec `json:"adaptiveOverloadProtection,omitempty"`
}

func (v *VaultServerConfigSpec) PersistentVolumeClaim() []corev1.PersistentVolumeClaim {
	return v.Storage.PersistentVolumeClaims()
}

func (v *VaultServerConfigSpec) Volumes(c *client.Client, ctx context.Context, vs *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if vol, mnt, err := v.ListenerTcp.Volumes(c, ctx, vs); err != nil {
		return nil, nil, err
	} else {
		volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
	}
	if vol, mnt, err := v.Storage.Volumes(c, ctx, vs); err != nil {
		return nil, nil, err
	} else {
		volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
	}
	if v.Telemetry != nil {
		if vol, mnt, err := v.Telemetry.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}
	if v.Telemetry != nil {
		if vol, mnt, err := v.Telemetry.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}
	if v.Seal != nil {
		if vol, mnt, err := v.Seal.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}
	if v.ServiceRegistration != nil {
		if vol, mnt, err := v.ServiceRegistration.Volumes(c, ctx, vs); err != nil {
			return nil, nil, err
		} else {
			volumes, mounts = append(volumes, vol...), append(mounts, mnt...)
		}
	}

	if v.License != nil {
		_, err := v.License.ContainsKey(c, ctx, vs.Namespace)
		if err != nil {
			return nil, nil, err
		}
		volumes = append(volumes, corev1.Volume{
			Name: "vault-license",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: v.License.SecretRef.Name,
					Items: []corev1.KeyToPath{{
						Key:  v.License.SecretRef.Key,
						Path: "vault.license",
					}},
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "seal-transit-client-tls",
			ReadOnly:  true,
			MountPath: "/vault.license",
			SubPath:   "vault.license",
		})
	}
	return volumes, mounts, nil
}

func (v *VaultServerConfigSpec) ResolveSecret(c *client.Client, ctx context.Context, vs *VaultServer) error {
	if err := v.ListenerTcp.Secrets(c, ctx, vs); err != nil {
		return err
	}
	if err := v.Storage.Secrets(c, ctx, vs); err != nil {
		return err
	}
	if v.Telemetry != nil {
		if err := v.Telemetry.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	if v.Telemetry != nil {
		if err := v.Telemetry.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	if v.Seal != nil {
		if err := v.Seal.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	if v.ServiceRegistration != nil {
		if err := v.ServiceRegistration.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	if v.ServiceRegistration != nil {
		if err := v.ServiceRegistration.Secrets(c, ctx, vs); err != nil {
			return err
		}
	}
	return nil
}

func (s *VaultServerConfigSpec) MapValue() (map[string]any, error) {
	scheme := "https"
	if s.ListenerTcp.TLS.Disable {
		scheme = "http"
	}

	_m := map[string]any{
		"cluster_addr":     strconv.Quote(fmt.Sprintf("%s://0.0.0.0:8201", scheme)),
		"api_addr":         strconv.Quote(fmt.Sprintf("%s://0.0.0.0:8200", scheme)),
		"pid_file":         strconv.Quote("/vault/vault.pid"),
		"log_file":         strconv.Quote("/dev/stdout"),
		"plugin_tmpdir":    strconv.Quote("/plugins/tmp"),
		"plugin_directory": strconv.Quote("/plugins"),
	}

	if s.License != nil {
		_m["license_path"] = strconv.Quote("/vault.license")
	}

	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
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

func (v *VaultServer) GetSecretNameForVaultConfig() string {
	generateName := fmt.Sprintf("%s-config", v.GetObjectMeta().GetName())
	if v.Spec.SecretOverride == nil || v.Spec.SecretOverride.Name == "" {
		return generateName
	}
	return v.Spec.SecretOverride.Name
}

func (v *VaultServerConfigSpec) internalHclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(vaultServerHclTemplate))
	if err := template.Execute(&buf, v); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

func (v *VaultServerConfigSpec) HclRender() (string, error) {
	var sb strings.Builder
	if s, err := v.internalHclRender(); err != nil {
		return "", err
	} else {
		sb.WriteString(s)
	}

	if s, err := v.ListenerTcp.HclRender(); err != nil {
		return "", err
	} else {
		sb.WriteString(s)
	}

	if s, err := v.Storage.HclRender(); err != nil {
		return "", err
	} else {
		sb.WriteString(s)
	}

	if len(v.UserLockout) > 0 {
		if s, err := v.UserLockout.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if len(v.Seal) > 0 {
		if s, err := v.Seal.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if len(v.ServiceRegistration) > 0 {
		if s, err := v.ServiceRegistration.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.Telemetry != nil {
		if s, err := v.Telemetry.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.KMSLibrary != nil {
		if s, err := v.KMSLibrary.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.Replication != nil {
		if s, err := v.Replication.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.Reporting != nil {
		if s, err := v.Reporting.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.Sentinel != nil {
		if s, err := v.Sentinel.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	if v.AdaptiveOverloadProtection != nil {
		if s, err := v.AdaptiveOverloadProtection.HclRender(); err != nil {
			return "", err
		} else {
			sb.WriteString(s)
		}
	}

	return sb.String(), nil
}

func (v *VaultServer) HclRender() (string, error) {
	return v.Spec.Config.HclRender()
}

func init() {
	SchemeBuilder.Register(&VaultServer{}, &VaultServerList{})
}
