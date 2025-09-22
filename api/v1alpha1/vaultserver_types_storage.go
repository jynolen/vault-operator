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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kythe.io/kythe/go/util/datasize"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// #region StorageSpec

const storageHclTempate = `
storage "{{ .Type }}" {
    {{ range $key,$val := .MapValue }}
        {{ if $val | typeIs "map[string]string" }}
    {{ $key }} = { 
            {{ range $k,$v := $val }}
        {{ $k }} = {{ $v }}
            {{ end }}
    }
        {{ else }}
    {{ $key }} = {{ $val }}
        {{ end }}
    {{ end }}
    {{ if eq .Type "raft" }}
        {{ range .Raft.RetryJoin }}  
    retry_join {
            {{ range $key,$val := .MapValue }}
        {{ $key }} = {{ $val }}
            {{ end }}
    }
        {{ end }}
    {{ end }}
}
`

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type StorageSpec struct {
	Aerospike          *StorageAerospikeSpec          `json:"aerospike,omitempty"`
	AlicloudOss        *StorageAlicloudOssSpec        `json:"alicloudOss,omitempty"`
	InMem              *StorageInMemSpec              `json:"inMem,omitempty"`
	FileSystem         *StorageFileSystemSpec         `json:"fileSystem,omitempty"`
	Azure              *StorageAzureSpec              `json:"azure,omitempty"`
	Cassandra          *StorageCassandraSpec          `json:"cassandra,omitempty"`
	CockroachDB        *StorageCockroachDBSpec        `json:"cockroachDb,omitempty"`
	Consul             *StorageConsulSpec             `json:"consul,omitempty"`
	CouchDB            *StorageCouchDBSpec            `json:"couchDb,omitempty"`
	DynamoDB           *StorageDynamoDBSpec           `json:"dynamoDb,omitempty"`
	Etcd               *StorageEtcdSpec               `json:"etcd,omitempty"`
	FoundationDb       *StorageFoundationDbSpec       `json:"foundationDb,omitempty"`
	GoogleCloudSpanner *StorageGoogleCloudSpannerSpec `json:"googleCloudSpanner,omitempty"`
	GoogleCloudStorage *StorageGoogleCloudStorageSpec `json:"googleCloudStorage,omitempty"`
	Raft               *StorageRaftSpec               `json:"raft,omitempty"`
	Manta              *StorageMantaSpec              `json:"manta,omitempty"`
	MsSql              *StorageMsSqlSpec              `json:"msSql,omitempty"`
	MySql              *StorageMySqlSpec              `json:"mySql,omitempty"`
	OCIObjectStorage   *StorageOCIObjectStorageSpec   `json:"ociObjectStorage,omitempty"`
	PostgreSql         *StoragePostgreSqlSpec         `json:"postgreSql,omitempty"`
	S3                 *StorageS3Spec                 `json:"s3,omitempty"`
	Swift              *StorageSwiftSpec              `json:"swift,omitempty"`
	ZooKeeper          *StorageZooKeeperSpec          `json:"zooKeeper,omitempty"`
}

func (s *StorageSpec) Type() string {
	return s.internalStorage().Type()
}

func (s *StorageSpec) MapValue() (map[string]any, error) {
	return s.internalStorage().MapValue()
}

func (s *StorageSpec) Secrets(c *client.Client, ctx context.Context, v *VaultServer) error {
	return s.internalStorage().Secrets(c, ctx, v)
}

func (s *StorageSpec) Volumes(c *client.Client, ctx context.Context, v *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return s.internalStorage().Volumes(c, ctx, v)
}

func (s *StorageSpec) internalStorage() ConfigBuilderHelper {
	if s.Aerospike != nil {
		return s.Aerospike
	}
	if s.AlicloudOss != nil {
		return s.AlicloudOss
	}
	if s.InMem != nil {
		return s.InMem
	}
	if s.FileSystem != nil {
		return s.FileSystem
	}
	if s.Azure != nil {
		return s.Azure
	}
	if s.Cassandra != nil {
		return s.Cassandra
	}
	if s.CockroachDB != nil {
		return s.CockroachDB
	}
	if s.Consul != nil {
		return s.Consul
	}
	if s.CouchDB != nil {
		return s.CouchDB
	}
	if s.DynamoDB != nil {
		return s.DynamoDB
	}
	if s.Etcd != nil {
		return s.Etcd
	}
	if s.FoundationDb != nil {
		return s.FoundationDb
	}
	if s.GoogleCloudSpanner != nil {
		return s.GoogleCloudSpanner
	}
	if s.GoogleCloudStorage != nil {
		return s.GoogleCloudStorage
	}
	if s.Raft != nil {
		return s.Raft
	}
	if s.Manta != nil {
		return s.Manta
	}
	if s.MsSql != nil {
		return s.MsSql
	}
	if s.MySql != nil {
		return s.MySql
	}
	if s.OCIObjectStorage != nil {
		return s.OCIObjectStorage
	}
	if s.PostgreSql != nil {
		return s.PostgreSql
	}
	if s.S3 != nil {
		return s.S3
	}
	if s.Swift != nil {
		return s.Swift
	}
	if s.ZooKeeper != nil {
		return s.ZooKeeper
	}
	return nil
}

func (s *StorageSpec) HclRender() (string, error) {
	var buf bytes.Buffer
	template := template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(storageHclTempate))
	if err := template.Execute(&buf, s); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\n\s*\n`)
	return re.ReplaceAllString(buf.String(), "\n"), nil
}

type InternalStorageSpec struct {
	MaxParallel *int32 `json:"maxParallel,omitempty" hcl:"max_parallel"`
}

type StorageRaftRetrySpec struct {
	LeaderApiAddr       *string         `json:"leaderApiAddr,omitempty" hcl:"leader_api_addr"`
	AutoJoin            *string         `json:"autoJoin,omitempty" hcl:"auto_join"`
	AutoJoinPort        *int32          `json:"autoJoinPort,omitempty" hcl:"auto_join_port"`
	LeaderTlsServername *string         `json:"leaderTlsServername,omitempty" hcl:"leader_tls_servername"`
	LeaderTls           *SecretSelector `json:"leaderTls,omitempty"`
	// +kubebuilder:validation:Enum=http;https
	AutoJoinScheme *string `json:"autoJoinScheme,omitempty" hcl:"auto_join_scheme"`
}

func (s *StorageRaftRetrySpec) Volumes(c *client.Client, ctx context.Context, namespace string) (*corev1.Volume, *corev1.VolumeMount, error) {
	_, err := s.LeaderTls.IsKind(c, ctx, namespace, corev1.SecretTypeTLS)
	if err != nil {
		return nil, nil, err
	}
	hash := fmt.Sprintf("%x", xxh3.HashString(*s.LeaderApiAddr))
	v := corev1.Volume{
		Name: fmt.Sprintf("storage-raft-leader-tls-%s", hash),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.LeaderTls.SecretRef.Name,
			},
		},
	}
	m := corev1.VolumeMount{
		Name:      fmt.Sprintf("storage-raft-leader-tls-%s", hash),
		ReadOnly:  true,
		MountPath: fmt.Sprintf("/raft/retry/%x", hash),
	}

	return &v, &m, nil
}

func (s *StorageRaftRetrySpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.LeaderTls != nil {
		_tls := map[string]any{
			"leader_ca_cert_file":     strconv.Quote(fmt.Sprintf("/raft/retry/%x/ca.crt", xxh3.HashString(*s.LeaderApiAddr))),
			"leader_client_cert_file": strconv.Quote(fmt.Sprintf("/raft/retry/%x/tls.crt", xxh3.HashString(*s.LeaderApiAddr))),
			"leader_client_key_file":  strconv.Quote(fmt.Sprintf("/raft/retry/%x/tls.key", xxh3.HashString(*s.LeaderApiAddr))),
		}
		maps.Copy(_m, _tls)
	}
	return _m, nil
}

type StorageRaftSpec struct {
	NodeId                             *string                `json:"nodeId,omitempty" hcl:"node_id"`
	AutopilotUpgradeVersion            *string                `json:"autopilotUpgradeVersion,omitempty" hcl:"autopilot_upgrade_version"`
	AutopilotRedundancyZone            *string                `json:"autopilotRedundancyZone,omitempty" hcl:"autopilot_redundancy_zone"`
	Experimental                       *map[string]string     `json:"experimental,omitempty"`
	RetryJoin                          []StorageRaftRetrySpec `json:"retryJoin,omitempty"`
	PerformanceMultiplier              *int32                 `json:"performanceMultiplier,omitempty" hcl:"performance_multiplier"`
	TrailingLogs                       *int32                 `json:"trailingLogs,omitempty" hcl:"trailing_logs"`
	SnapshotThreshold                  *int32                 `json:"snapshotThreshold,omitempty" hcl:"snapshot_threshold"`
	SnapshotInterval                   *int32                 `json:"snapshotInterval,omitempty" hcl:"snapshot_interval"`
	RetryJoinAsNonVoter                *bool                  `json:"retryJoinAsNonVoter,omitempty" hcl:"retry_join_as_non_voter"`
	MaxEntrySize                       *int32                 `json:"maxEntrySize,omitempty" hcl:"max_entry_size"`
	MaxMountAndNamespaceTableEntrySize *int32                 `json:"maxMountAndNamespaceTableEntrySize,omitempty" hcl:"max_mount_and_namespace_table_entry_size"`
	AutopilotReconcileInterval         *metav1.Duration       `json:"autopilotReconcileInterval,omitempty" hcl:"autopilot_reconcile_interval"`
	AutopilotUpdateInterval            *metav1.Duration       `json:"autopilotUpdateInterval,omitempty" hcl:"autopilot_update_interval"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageRaftSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageRaftSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	for _, retry := range lo.Filter(s.RetryJoin, func(o StorageRaftRetrySpec, _ int) bool { return o.LeaderTls != nil }) {
		v, m, err := retry.Volumes(c, ctx, vaultServer.Namespace)
		if err != nil {
			return nil, nil, err
		}
		if v != nil && m != nil {
			volumes, mounts = append(volumes, *v), append(mounts, *m)
		}
	}
	mounts = append(mounts, corev1.VolumeMount{
		Name:      vaultServer.Spec.PersistentVolumeClaim.Spec.VolumeName,
		MountPath: "/raft",
	})
	return volumes, mounts, nil
}

func (s *StorageRaftSpec) Type() string {
	return "raft"
}

func (s *StorageRaftSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"path": strconv.Quote("/raft/"),
	}

	if s.Experimental != nil {
		for k, v := range *s.Experimental {
			_m[k] = v
		}
	}

	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StorageZooKeeperTLSpec struct {
	Enabled     *bool           `json:"enabled,omitempty" hcl:"tls_enabled"`
	Certificate *SecretSelector `json:"certificate,omitempty"`
	MinVersion  *TLSVersion     `json:"minVersion,omitempty" hcl:"tls_min_version"`
	SkipVerify  *bool           `json:"skipVerify,omitempty" hcl:"tls_skip_verify"`
	VerifyIP    *bool           `json:"verifyIp,omitempty" hcl:"tls_verify_ip"`
}

func (s *StorageZooKeeperTLSpec) Volumes(c *client.Client, ctx context.Context, namespace string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.Certificate == nil {
		return volumes, mounts, nil
	}
	_, err := s.Certificate.IsKind(c, ctx, namespace, corev1.SecretTypeTLS)
	if err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "seal-transit-client-tls",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Certificate.SecretRef.Name,
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "seal-transit-client-tls",
		ReadOnly:  true,
		MountPath: "/seal/transit",
	})
	return volumes, mounts, nil
}

func (s *StorageZooKeeperTLSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"tls_enabled":   strconv.FormatBool(true),
		"tls_ca_file":   strconv.Quote("/zookeeper/ca.crt"),
		"tls_cert_file": strconv.Quote("/zookeeper/tls.crt"),
		"tls_key_file":  strconv.Quote("/zookeeper/tls.key"),
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StorageZooKeeperSpec struct {
	ZnodeOwner       *string                 `json:"-"  hcl:"znode_owner"`
	AuthInfo         *string                 `json:"-"  hcl:"auth_info"`
	Address          *string                 `json:"address,omitempty"  hcl:"address"`
	Path             *string                 `json:"path,omitempty"  hcl:"path"`
	RedirectAddr     *string                 `json:"redirect_addr,omitempty"  hcl:"redirect_addr"`
	ZnodeOwnerSecret *SecretKeySelector      `json:"znodeOwner,omitempty"`
	AuthInfoSecret   *SecretKeySelector      `json:"authInfo,omitempty"`
	Tls              *StorageZooKeeperTLSpec `json:"tls,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageZooKeeperSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.ZnodeOwnerSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"znode_owner": {Dest: s.ZnodeOwner, Mandatory: true},
	}
	if err := secretMappings.Apply(&secret, "SealAliCloudKmsSpec.Credentials"); err != nil {
		return err
	}

	t = types.NamespacedName{Name: s.AuthInfoSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}

	secretMappings = utils.SecretKvMapping{
		"auth_info": {Dest: s.AuthInfo, Mandatory: true},
	}
	if err := secretMappings.Apply(&secret, "SealAliCloudKmsSpec.Credentials"); err != nil {
		return err
	}
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageZooKeeperSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if s.Tls == nil {
		return []corev1.Volume{}, []corev1.VolumeMount{}, nil
	}
	return s.Tls.Volumes(c, ctx, vaultServer.Namespace)
}

func (s *StorageZooKeeperSpec) Type() string {
	return "zookeeper"
}

func (s *StorageZooKeeperSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"tls_enabled": strconv.FormatBool(false),
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if s.Tls != nil {
		if _s, err := s.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

type StorageSwiftTenantSpec struct {
	Name *string `json:"name,omitempty" hcl:"tenant"`
	ID   *string `json:"id,omitempty" hcl:"tenant_id"`
}

func (s *StorageSwiftTenantSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageSwiftDomainNameSpec struct {
	Project *string `json:"project,omitempty" hcl:"project-domain"`
	User    *string `json:"user,omitempty" hcl:"domain"`
}

func (s *StorageSwiftDomainNameSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageSwiftSpec struct {
	InternalStorageSpec `json:",inline"`
	Username            *string                     `json:"-" hcl:"username"`
	Password            *string                     `json:"-" hcl:"password"`
	AuthToken           *string                     `json:"-" hcl:"auth_token"`
	AuthUrl             *string                     `json:"authUrl,omitempty" hcl:"auth_url"`
	Container           *string                     `json:"container,omitempty" hcl:"container"`
	StorageUrl          *string                     `json:"storageUrl,omitempty" hcl:"storage_url"`
	Region              *string                     `json:"region,omitempty" hcl:"region"`
	Tenant              *StorageSwiftTenantSpec     `json:"tenant,omitempty"`
	DomainName          *StorageSwiftDomainNameSpec `json:"domainName,omitempty"`
	TrustId             *string                     `json:"trustId,omitempty" hcl:"trust_id"`
	Credentials         *SecretSelector             `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageSwiftSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username":   {Dest: s.Username, Mandatory: true},
		"password":   {Dest: s.Password, Mandatory: true},
		"auth_token": {Dest: s.AuthToken, Mandatory: false},
	}
	return secretMappings.Apply(&secret, "StorageSwiftSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageSwiftSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageSwiftSpec) Type() string {
	return "swift"
}

func (s *StorageSwiftSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.Tenant != nil {
		if _s, err := s.Tenant.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if s.DomainName != nil {
		if _s, err := s.DomainName.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

type StorageS3Spec struct {
	InternalStorageSpec `json:",inline"`
	AccessKey           *string         `json:"-" hcl:"access_key"`
	SecretKey           *string         `json:"-" hcl:"secret_key"`
	SessionToken        *string         `json:"-" hcl:"session_token"`
	Bucket              *string         `json:"bucket" hcl:"bucket"`
	KmsKeyId            *string         `json:"kmsKeyId,omitempty" hcl:"kms_key_id"`
	Path                *string         `json:"path,omitempty" hcl:"path"`
	Endpoint            *string         `json:"endpoint,omitempty" hcl:"endpoint"`
	Region              *string         `json:"region,omitempty" hcl:"region"`
	S3ForcePathStyle    *bool           `json:"s3ForcePathStyle,omitempty" hcl:"s3_force_path_style"`
	DisableSsl          *bool           `json:"disableSsl,omitempty" hcl:"disable_ssl"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageS3Spec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
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
	return secretMappings.Apply(&secret, "StorageS3Spec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageS3Spec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageS3Spec) Type() string {
	return "s3"
}

func (s *StorageS3Spec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageOCIObjectStorageSpec struct {
	Region         *string            `json:"region,omitempty" hcl:"region"`
	NamespaceName  *string            `json:"namespaceName" hcl:"namespace_name"`
	BucketName     *string            `json:"bucketName" hcl:"bucket_name"`
	HaEnabled      *bool              `json:"haEnabled" hcl:"ha_enabled"`
	LockBucketName *string            `json:"lockBucketName" hcl:"lock_bucket_name"`
	Credentials    *SecretKeySelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageOCIObjectStorageSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageOCIObjectStorageSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	_, err := s.Credentials.ContainsKey(c, ctx, vaultServer.Namespace)
	if err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-oci-credentials",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Credentials.SecretRef.Name,
				Items: []corev1.KeyToPath{{
					Key:  s.Credentials.SecretRef.Key,
					Path: "config",
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-oci-credentials",
		ReadOnly:  true,
		MountPath: "/oci/config",
		SubPath:   "config",
	})
	return volumes, mounts, nil
}

func (s *StorageOCIObjectStorageSpec) Type() string {
	return "oci"
}

func (s *StorageOCIObjectStorageSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StoragePostgreSqlAwsSpec struct {
	DbRegion *string `json:"awsDbRegion,omitempty" hcl:"aws_db_region"`
}

func (s *StoragePostgreSqlAwsSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"auth_mode": strconv.Quote("aws_iam"),
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StoragePostgreSqlAzureSpec struct {
	ClientId *string `json:"clientId,omitempty" hcl:"client_id"`
}

func (s *StoragePostgreSqlAzureSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"auth_mode": strconv.Quote("azure_msi"),
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StoragePostgreSqlGcpSpec struct{}

func (s *StoragePostgreSqlGcpSpec) MapValue() (map[string]any, error) {
	return map[string]any{
		"auth_mode": strconv.Quote("gcp_iam"),
	}, nil
}

type StoragePostgreSqlStandardSpec struct{}

func (s *StoragePostgreSqlStandardSpec) MapValue() (map[string]any, error) {
	return map[string]any{
		"auth_mode": strconv.Quote("standard"),
	}, nil
}

// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type StoragePostgreSqlAutModeSpec struct {
	Aws      *StoragePostgreSqlAwsSpec      `json:"aws,omitempty"`
	Azure    *StoragePostgreSqlAzureSpec    `json:"azure,omitempty"`
	Gcp      *StoragePostgreSqlGcpSpec      `json:"gcp,omitempty"`
	Standard *StoragePostgreSqlStandardSpec `json:"standard,omitempty"`
}

func (s *StoragePostgreSqlAutModeSpec) MapValue() (map[string]any, error) {
	if s.Aws != nil {
		return s.Aws.MapValue()
	}
	if s.Azure != nil {
		return s.Azure.MapValue()
	}
	if s.Gcp != nil {
		return s.Gcp.MapValue()
	}
	if s.Standard != nil {
		return s.Standard.MapValue()
	}
	return map[string]any{}, nil
}

type StoragePostgreSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       *string                       `json:"-" hcl:"connection_url"`
	ConnectionUrlSecret *SecretKeySelector            `json:"connectionUrl,omitempty"`
	AuthMode            *StoragePostgreSqlAutModeSpec `json:"authMode,omitempty"`
	HaEnabled           *bool                         `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	HATable             *string                       `json:"haTable,omitempty" hcl:"ha_table"`
	Table               *string                       `json:"table,omitempty" hcl:"table"`
	MaxIdleConnections  *int32                        `json:"maxIdleConnections,omitempty" hcl:"max_idle_connections"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StoragePostgreSqlSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.ConnectionUrlSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"connection_url": {Dest: s.ConnectionUrl, Mandatory: true, Src: s.ConnectionUrlSecret.SecretRef.Key},
	}
	return secretMappings.Apply(&secret, "SealTransitSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StoragePostgreSqlSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StoragePostgreSqlSpec) Type() string {
	return "postgresql"
}

func (s *StoragePostgreSqlSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if s.AuthMode != nil {
		if _s, err := s.AuthMode.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

type StorageDynamoDBCapacitySpec struct {
	Read  *int32 `json:"read,omitempty" hcl:"read_capacity"`
	Write *int32 `json:"write,omitempty" hcl:"write_capacity"`
}

func (s *StorageDynamoDBCapacitySpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageDynamoDBSpec struct {
	InternalStorageSpec `json:",inline"`

	AccessKey            *string                      `json:"-" hcl:"access_key"`
	SecretKey            *string                      `json:"-" hcl:"secret_key"`
	SessionToken         *string                      `json:"-" hcl:"session_token"`
	DynamodbAllowUpdates *string                      `json:"dynamodbAllowUpdates,omitempty" hcl:"dynamodb_allow_updates"`
	Credentials          *SecretSelector              `json:"credentials,omitempty"`
	Capacity             *StorageDynamoDBCapacitySpec `json:"capacity,omitempty"`
	Endpoint             *string                      `json:"endpoint,omitempty" hcl:"endpoint"`
	HaEnabled            *bool                        `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	Region               *string                      `json:"region,omitempty" hcl:"region"`
	Table                *string                      `json:"table,omitempty" hcl:"table"`
	// +kubebuilder:validation:Enum=PROVISIONED;PAY_PER_REQUEST
	BillingMode *string `json:"billingMode,omitempty" hcl:"billing_mode"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageDynamoDBSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
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

// Volumes implements ConfigBuilderHelper.
func (s *StorageDynamoDBSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageDynamoDBSpec) Type() string {
	return "dynamodb"
}

func (s *StorageDynamoDBSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if s.Capacity != nil {
		if _s, err := s.Capacity.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

type StorageMySqlSpec struct {
	InternalStorageSpec        `json:",inline"`
	Username                   *string               `json:"-" hcl:"username"`
	Password                   *string               `json:"-" hcl:"password"`
	Address                    *string               `json:"address" hcl:"address"`
	Credentials                *SecretSelector       `json:"credentials,omitempty"`
	Database                   *string               `json:"database,omitempty" hcl:"database"`
	Table                      *string               `json:"table,omitempty" hcl:"table"`
	MaxIdleConnections         *int32                `json:"maxIdleConnections,omitempty" hcl:"max_idle_connections"`
	MaxConnectionLifetime      *int32                `json:"maxConnectionLifetime,omitempty" hcl:"max_connection_lifetime"`
	PlaintextConnectionAllowed *string               `json:"plaintextConnectionAllowed,omitempty" hcl:"plaintext_connection_allowed"`
	TlsCa                      *ConfigMapKeySelector `json:"tlsCa,omitempty"`
	HaEnabled                  *bool                 `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	LockTable                  *string               `json:"lockTable,omitempty" hcl:"lock_table"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageMySqlSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "StorageMySqlSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageMySqlSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.TlsCa == nil {
		return volumes, mounts, nil
	}
	_, err := s.TlsCa.ContainsKey(c, ctx, vaultServer.Namespace)
	if err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-mysql-tls",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: s.TlsCa.ConfigMapRef.Name},
				Items: []corev1.KeyToPath{{
					Key:  s.TlsCa.ConfigMapRef.Key,
					Path: "ca.crt",
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "seal-transit-client-tls",
		ReadOnly:  true,
		MountPath: "/mysql/ca.crt",
		SubPath:   "ca.crt",
	})
	return volumes, mounts, nil
}

func (s *StorageMySqlSpec) Type() string {
	return "mysql"
}

func (s *StorageMySqlSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.TlsCa != nil {
		_m["tls_ca_file"] = strconv.Quote("/mysql/ca.crt")
	}
	return _m, nil
}

type StorageMsSqlSpec struct {
	InternalStorageSpec `json:",inline"`

	Username          *string         `json:"-" hcl:"username"`
	Password          *string         `json:"-" hcl:"password"`
	Server            *string         `json:"server" hcl:"server"`
	Port              *int32          `json:"port,omitempty" hcl:"port"`
	Credentials       *SecretSelector `json:"credentials,omitempty"`
	Database          *string         `json:"database,omitempty" hcl:"database"`
	Table             *string         `json:"table,omitempty" hcl:"table"`
	Schema            *string         `json:"schema,omitempty" hcl:"schema"`
	ConnectionTimeout *int32          `json:"connectionTimeout,omitempty" hcl:"connectionTimeout"`
	AppName           *string         `json:"appName,omitempty" hcl:"appname"`

	// +kubebuilder:validation:Maximum=63
	// +kubebuilder:validation:Minimum=0
	LogLevel *int32 `json:"logLevel,omitempty" hcl:"logLevel"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageMsSqlSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "StorageMsSqlSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageMsSqlSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageMsSqlSpec) Type() string {
	return "mssql"
}

func (s *StorageMsSqlSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageEtcdTimeoutSpec struct {
	Request *metav1.Duration `json:"request,omitempty" hcl:"request_timeout"`
	Lock    *metav1.Duration `json:"lock,omitempty" hcl:"lock_timeout"`
}

func (s *StorageEtcdTimeoutSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageEtcdMaxSpec struct {
	ReceiveSize *datasize.Size `json:"receiveSize,omitempty" hcl:"max_receive_size"`
	SendSize    *datasize.Size `json:"sendSize,omitempty" hcl:"max_send_size"`
}

func (s *StorageEtcdMaxSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageEtcdSpec struct {
	Username         *string                 `json:"-" hcl:"username"`
	Password         *string                 `json:"-" hcl:"password"`
	Address          *string                 `json:"address,omitempty" hcl:"address"`
	DiscoverySrv     *string                 `json:"discoverySrv,omitempty" hcl:"discovery_srv"`
	DiscoverySrvName *string                 `json:"discoverySrvName,omitempty" hcl:"discovery_srv_name"`
	EtcdApi          *string                 `json:"etcdApi,omitempty" hcl:"etcd_api"`
	HaEnabled        *bool                   `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	Path             *string                 `json:"path,omitempty" hcl:"path"`
	Sync             *bool                   `json:"sync,omitempty" hcl:"sync"`
	Timeout          *StorageEtcdTimeoutSpec `json:"timeout,omitempty"`
	Max              *StorageEtcdMaxSpec     `json:"max,omitempty"`
	Credentials      *SecretSelector         `json:"credentials,omitempty"`
	Tls              *SecretSelector         `json:"tls,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageEtcdSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	secretMappings.Apply(&secret, "StorageCouchDBSpec.Credentials")
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageEtcdSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if s.Tls == nil {
		return []corev1.Volume{}, []corev1.VolumeMount{}, nil
	}
	if _, err := s.Tls.IsKind(c, ctx, vaultServer.Namespace, corev1.SecretTypeTLS); err != nil {
		return nil, nil, err
	}
	v := corev1.Volume{
		Name: "storage-etcd-tls",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Tls.SecretRef.Name,
			},
		},
	}
	m := corev1.VolumeMount{
		Name:      "storage-etcd-tls",
		ReadOnly:  true,
		MountPath: "/etcd",
	}
	return []corev1.Volume{v}, []corev1.VolumeMount{m}, nil
}

func (s *StorageEtcdSpec) Type() string {
	return "etcd"
}

func (s *StorageEtcdSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if s.Tls != nil {
		_m["tls_ca_file"] = strconv.Quote("/etcd/ca.crt")
		_m["tls_cert_file"] = strconv.Quote("/etcd/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/etcd/tls.key")
	}
	if s.Timeout != nil {
		if _s, err := s.Timeout.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	if s.Max != nil {
		if _s, err := s.Max.MapValue(); err != nil {
			return nil, err
		} else {
			maps.Copy(_m, _s)
		}
	}
	return _m, nil
}

type StorageGoogleCloudSpannerSpec struct {
	InternalStorageSpec `json:",inline"`

	Database    *string            `json:"database,omitempty" hcl:"database"`
	Table       *string            `json:"table,omitempty" hcl:"table"`
	HaEnabled   *bool              `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	HaTable     *string            `json:"haTable,omitempty" hcl:"ha_table"`
	Credentials *SecretKeySelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageGoogleCloudSpannerSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageGoogleCloudSpannerSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	_, err := s.Credentials.ContainsKey(c, ctx, vaultServer.Namespace)
	if err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-gcp-credentials",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Credentials.SecretRef.Name,
				Items: []corev1.KeyToPath{{
					Key:  s.Credentials.SecretRef.Key,
					Path: "credentials.json",
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-gcp-credentials",
		ReadOnly:  true,
		MountPath: "/gcp/credentials.json",
		SubPath:   "credentials.json",
	})
	return volumes, mounts, nil
}

func (s *StorageGoogleCloudSpannerSpec) Type() string {
	return "spanner"
}

func (s *StorageGoogleCloudSpannerSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageGoogleCloudStorageSpec struct {
	InternalStorageSpec `json:",inline"`

	Bucket      *string            `json:"bucket,omitempty" hcl:"bucket"`
	ChunkSize   *datasize.Size     `json:"chunkSize,omitempty" hcl:"chunk_size"`
	HaEnabled   *bool              `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	Credentials *SecretKeySelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageGoogleCloudStorageSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageGoogleCloudStorageSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	_, err := s.Credentials.ContainsKey(c, ctx, vaultServer.Namespace)
	if err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-gcp-credentials",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Credentials.SecretRef.Name,
				Items: []corev1.KeyToPath{{
					Key:  s.Credentials.SecretRef.Key,
					Path: "credentials.json",
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-gcp-credentials",
		ReadOnly:  true,
		MountPath: "/gcp/credentials.json",
		SubPath:   "credentials.json",
	})
	return volumes, mounts, nil
}

func (s *StorageGoogleCloudStorageSpec) Type() string {
	return "gcs"
}

func (s *StorageGoogleCloudStorageSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageFoundationDbTlsSpec struct {
	Password    *string         `json:"-" hcl:"password"`
	VerifyPeers *string         `json:"tlsVerifyPeers,omitempty" hcl:"tls_verify_peers"`
	Certificate *SecretSelector `json:"certificate,omitempty"`
}

func (s *StorageFoundationDbTlsSpec) Volumes(c *client.Client, ctx context.Context, namespace string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.Certificate == nil {
		return volumes, mounts, nil
	}
	if _, err := s.Certificate.IsKind(c, ctx, namespace, corev1.SecretTypeTLS); err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-foundationdb-tls",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.Certificate.SecretRef.Name,
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-foundationdb-tls",
		ReadOnly:  true,
		MountPath: "/foundationdb/tls",
	})
	return volumes, mounts, nil
}

func (s *StorageFoundationDbTlsSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if s.Certificate != nil {
		_m["tls_ca_file"] = strconv.Quote("/foundationdb/tls/ca.crt")
		_m["tls_cert_file"] = strconv.Quote("/foundationdb/tls/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/foundationdb/tls/tls.key")
	}
	return _m, nil
}

type StorageFoundationDbSpec struct {
	ApiVersion  *int32                      `json:"apiVersion,omitempty" hcl:"api_version"`
	ClusterFile *SecretKeySelector          `json:"clusterFile"`
	Tls         *StorageFoundationDbTlsSpec `json:"tls,omitempty"`
	Path        *string                     `json:"path,omitempty" hcl:"path"`
	HaEnabled   *bool                       `json:"haEnabled,omitempty" hcl:"ha_enabled"`
}

func (s *StorageFoundationDbSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	if s.Tls == nil || s.Tls.Password == nil {
		return nil
	}
	t := types.NamespacedName{Name: s.Tls.Certificate.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}

	secretMappings := utils.SecretKvMapping{
		"tls_password": {Dest: s.Tls.Password, Mandatory: false},
	}
	secretMappings.Apply(&secret, "StorageFoundationDbSpec.Tls.Password")
	return nil
}

func (s *StorageFoundationDbSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.Tls == nil {
		return volumes, mounts, nil
	}
	v, m, err := s.Tls.Volumes(c, ctx, vaultServer.Namespace)
	if err != nil {
		return nil, nil, err
	}
	volumes, mounts = append(volumes, v...), append(mounts, m...)

	if _, err := s.ClusterFile.IsKind(c, ctx, vaultServer.Namespace, corev1.SecretTypeOpaque); err != nil {
		return nil, nil, err
	}
	if _, err := s.ClusterFile.ContainsKey(c, ctx, vaultServer.Namespace); err != nil {
		return nil, nil, err
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-foundation-cluster-file",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: s.ClusterFile.SecretRef.Name,
				Items: []corev1.KeyToPath{{
					Key:  s.ClusterFile.SecretRef.Key,
					Path: "fdb.cluster",
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-foundation-cluster-file",
		ReadOnly:  true,
		MountPath: "/foundation/fdb.cluster",
		SubPath:   "fdb.cluster",
	})
	return volumes, mounts, nil
}

func (s *StorageFoundationDbSpec) Type() string {
	return "foundationdb"
}

func (s *StorageFoundationDbSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{
		"cluster_file": strconv.Quote("/foundationdb/vault.cluster"),
	}
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

type StorageAerospikeSpec struct {
	Username    *string         `json:"-" hcl:"username"`
	Password    *string         `json:"-" hcl:"password"`
	Hostname    *string         `json:"hostname,omitempty" hcl:"hostname"`
	Port        *int32          `json:"port,omitempty" hcl:"port"`
	HostList    []string        `json:"hostList,omitempty"`
	Namepace    *string         `json:"namespace,omitempty" hcl:"namespace"`
	Set         *string         `json:"set,omitempty" hcl:"set"`
	ClusterName *string         `json:"clusterName,omitempty" hcl:"cluster_name"`
	Timeout     *int32          `json:"timeout,omitempty" hcl:"timeout"`
	IdleTImeout *int32          `json:"idleTImeout,omitempty" hcl:"idle_t_imeout"`
	Credentials *SecretSelector `json:"credentials,omitempty"`

	// +kubebuilder:validation:Enum=INTERNAL;EXTERNAL
	AuthMode *string `json:"authMode,omitempty" hcl:"auth_mode"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageAerospikeSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	secretMappings.Apply(&secret, "StorageAerospikeSpec.Credentials")
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageAerospikeSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageAerospikeSpec) Type() string {
	return "aerospike"
}

func (s *StorageAerospikeSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StorageAlicloudOssSpec struct {
	InternalStorageSpec `json:",inline"`
	AccessKey           *string         `json:"-" hcl:"access_key"`
	SecretKey           *string         `json:"-" hcl:"secret_key"`
	Bucket              *string         `json:"bucket,omitempty" hcl:"bucket"`
	Endpoint            *string         `json:"endpoint,omitempty" hcl:"endpoint"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageAlicloudOssSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"access_key": {Dest: s.AccessKey, Mandatory: true},
		"secret_key": {Dest: s.SecretKey, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "StorageAlicloudOssSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageAlicloudOssSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageAlicloudOssSpec) Type() string {
	return "alicloudoss"
}

func (s *StorageAlicloudOssSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageAzureSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountKey          *string         `json:"-" hcl:"access_key"`
	AccountName         *string         `json:"accountName" hcl:"account_name"`
	Container           *string         `json:"container" hcl:"container"`
	Environment         *string         `json:"environment,omitempty" hcl:"environment"`
	ArmEndpoint         *string         `json:"armEndpoint,omitempty" hcl:"arm_endpoint"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageAzureSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"accountKey": {Dest: s.AccountKey, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "StorageAzureSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageAzureSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageAzureSpec) Type() string {
	return "alicloudoss"
}

func (s *StorageAzureSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type PemBundleCassandraSpec struct {
	// +kubebuilder:validation:Enum=bundle;json
	Type        string             `json:"type"`
	Credentials *SecretKeySelector `json:"file,omitempty"`
}

func (in *PemBundleCassandraSpec) Volumes(c *client.Client, ctx context.Context, namespace string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if _, err := in.Credentials.ContainsKey(c, ctx, namespace); err != nil {
		return nil, nil, err
	}
	target_file := "cert.json"
	if in.Type == "bundle" {
		target_file = "cert.pem"
	}
	volumes = append(volumes, corev1.Volume{
		Name: "storage-cassandra-pem-file",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: in.Credentials.SecretRef.Name,
				Items: []corev1.KeyToPath{{
					Key:  in.Credentials.SecretRef.Key,
					Path: target_file,
				}},
			},
		},
	})
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "storage-cassandra-pem-file",
		ReadOnly:  true,
		MountPath: fmt.Sprintf("/cassandra/%s", target_file),
		SubPath:   target_file,
	})
	return volumes, mounts, nil
}

type StorageCassandraSpec struct {
	InternalStorageSpec `json:",inline"`

	Username                 *string                 `json:"-" hcl:"username"`
	Password                 *string                 `json:"-" hcl:"password"`
	Keyspace                 *string                 `json:"keyspace,omitempty" hcl:"keyspace"`
	Table                    *string                 `json:"table,omitempty" hcl:"table"`
	ProtocolVersion          *int32                  `json:"protocolVersion,omitempty" hcl:"protocol_version"`
	DisableInitialHostLookup *bool                   `json:"disableInitialHostLookup,omitempty" hcl:"disable_initial_host_lookup"`
	InitialConnectionTimeout *int32                  `json:"initialConnectionTimeout,omitempty" hcl:"initial_connection_timeout"`
	ConnectionTimeout        *int32                  `json:"connectionTimeout,omitempty" hcl:"connection_timeout"`
	SimpleRetryPolicyRetries *int32                  `json:"simpleRetryPolicyRetries,omitempty" hcl:"simple_retry_policy_retries"`
	TlsSkipVerify            *int32                  `json:"tlsSkipVerify,omitempty" hcl:"tls_skip_verify"`
	Tls                      *int32                  `json:"tls,omitempty" hcl:"tls"`
	TlsMinVersion            *TLSVersion             `json:"tlsMinVersion,omitempty" hcl:"tls_min_version"`
	Pem                      *PemBundleCassandraSpec `json:"pem,omitempty"`
	Credentials              *SecretSelector         `json:"credentials,omitempty"`
	// +kubebuilder:validation:Enum=ANY;ONE;TWO;THREE;QUORUM;ALL;LOCAL_QUORUM;EACH_QUORUM;LOCAL_ONE
	Consistency string `json:"consistency,omitempty" hcl:"consistency"`
	// +kubebuilder:validation:MinItems=1
	Hosts []string `json:"hosts,omitempty"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageCassandraSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	secretMappings.Apply(&secret, "StorageCassandraSpec.Credentials")
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageCassandraSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	volumes, mounts := []corev1.Volume{}, []corev1.VolumeMount{}
	if s.Pem == nil {
		return volumes, mounts, nil
	}
	return s.Pem.Volumes(c, ctx, vaultServer.Namespace)
}

func (s *StorageCassandraSpec) Type() string {
	return "cassandra"
}

func (s *StorageCassandraSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}

	if s.Pem != nil && s.Pem.Type == "json" {
		_m["pem_json_file"] = strconv.Quote("/cassandra/cert.json")
	}
	if s.Pem != nil && s.Pem.Type == "json" {
		_m["pem_bundle_file"] = strconv.Quote("/cassandra/cert.pem")
	}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}

	if len(s.Hosts) > 0 {
		_m["hosts"] = strconv.Quote(strings.Join(s.Hosts, ","))
	}

	return _m, nil
}

type StorageCockroachDBSpec struct {
	InternalStorageSpec `json:",inline"`

	ConnectionUrl       *string            `json:"-" hcl:"connection_url"`
	ConnectionUrlSecret *SecretKeySelector `json:"connectionUrl"`
	Table               *string            `json:"table,omitempty" hcl:"table"`
	HaEnabled           *bool              `json:"haEnabled,omitempty" hcl:"ha_enabled"`
	HaTable             *string            `json:"haTable,omitempty" hcl:"ha_table"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageCockroachDBSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.ConnectionUrlSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"connection_url": {Dest: s.ConnectionUrl, Mandatory: true},
	}
	return secretMappings.Apply(&secret, "StorageCockroachDBSpec.Credentials")
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageCockroachDBSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageCockroachDBSpec) Type() string {
	return "cockroachdb"
}

func (s *StorageCockroachDBSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageCouchDBSpec struct {
	InternalStorageSpec `json:",inline"`

	Username    *string         `json:"-" hcl:"username"`
	Password    *string         `json:"-" hcl:"password"`
	Endpoint    *string         `json:"endpoint" hcl:"endpoint"`
	Credentials *SecretSelector `json:"credentials"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageCouchDBSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	t := types.NamespacedName{Name: s.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := (*c).Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := utils.SecretKvMapping{
		"username": {Dest: s.Username, Mandatory: true},
		"password": {Dest: s.Password, Mandatory: true},
	}
	secretMappings.Apply(&secret, "StorageCouchDBSpec.Credentials")
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageCouchDBSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageCouchDBSpec) Type() string {
	return "couchdb"
}

func (s *StorageCouchDBSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

type StorageConsulSpec struct {
	InternalStorageSpec `json:",inline"`
	ConsulSpec          `json:",inline"`

	Path         *string          `json:"path,omitempty" hcl:"path"`
	SessionTTL   *metav1.Duration `json:"sessionTtl,omitempty" hcl:"session_ttl"`
	LockWaitTime *metav1.Duration `json:"lockWaitTime,omitempty" hcl:"lock_wait_time"`
	// +kubebuilder:validation:Enum=default;strong
	ConsistencyMode *string `json:"consistencyMode,omitempty" hcl:"consistency_mode"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageConsulSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return s.ConsulSpec.Secrets(c, ctx, vaultServer)
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageConsulSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return s.ConsulSpec.Volumes(c, ctx, vaultServer)
}

func (s *StorageConsulSpec) Type() string {
	return "consul"
}

func (s *StorageConsulSpec) MapValue() (map[string]any, error) {
	_m := map[string]any{}
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	if _s, err := s.ConsulSpec.MapValue(); err != nil {
		return nil, err
	} else {
		maps.Copy(_m, _s)
	}
	return _m, nil
}

type StorageInMemSpec struct{}

// Secrets implements ConfigBuilderHelper.
func (s *StorageInMemSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageInMemSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageInMemSpec) Type() string {
	return "inmem"
}

func (s *StorageInMemSpec) MapValue() (map[string]any, error) {
	return map[string]any{}, nil
}

type StorageFileSystemSpec struct {
	Path string `json:"-" hcl:"path"`
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageFileSystemSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageFileSystemSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	mounts := []corev1.VolumeMount{
		{
			Name:      vaultServer.Spec.PersistentVolumeClaim.Spec.VolumeName,
			MountPath: "/data",
		},
	}
	return []corev1.Volume{}, mounts, nil
}

func (s *StorageFileSystemSpec) Type() string {
	return "file"
}

func (s *StorageFileSystemSpec) MapValue() (map[string]any, error) {
	return map[string]any{
		"path": strconv.Quote("/data/"),
	}, nil
}

type StorageMantaSpec struct {
	InternalStorageSpec `json:",inline"`
	Directory           *string `json:"directory" hcl:"directory"`
	User                *string `json:"user" hcl:"user"`
	KeyId               *string `json:"keyId" hcl:"key_id"`
	SubUser             *string `json:"subUser" hcl:"sub_user"`
	URL                 *string `json:"url" hcl:"url"`
	// TODO AddSSHKeyForAgent
}

// Secrets implements ConfigBuilderHelper.
func (s *StorageMantaSpec) Secrets(c *client.Client, ctx context.Context, vaultServer *VaultServer) error {
	return nil
}

// Volumes implements ConfigBuilderHelper.
func (s *StorageMantaSpec) Volumes(c *client.Client, ctx context.Context, vaultServer *VaultServer) ([]corev1.Volume, []corev1.VolumeMount, error) {
	return []corev1.Volume{}, []corev1.VolumeMount{}, nil
}

func (s *StorageMantaSpec) Type() string {
	return "manta"
}

func (s *StorageMantaSpec) MapValue() (map[string]any, error) {
	if _s, err := utils.HclExport(*s); err != nil {
		return nil, err
	} else {
		return _s, nil
	}
}

// #endregion
