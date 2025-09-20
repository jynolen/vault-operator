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
	"reflect"
	"strconv"
	"strings"

	"github.com/zeebo/xxh3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kythe.io/kythe/go/util/datasize"
)

// #region StorageSpec

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

func (s *StorageSpec) InternalStorage() HclHelper {
	v := reflect.ValueOf(*s)
	t := reflect.TypeOf(*s)

	for i := range t.NumField() {
		intf := v.Field(i)
		if !intf.IsNil() {
			return intf.Interface().(HclHelper)
		}
	}
	return nil
}

type InternalStorageSpec struct {
	MaxParallel *int32 `json:"maxParallel,omitempty"`
}

type StorageRaftRetrySpec struct {
	LeaderApiAddr       *string         `json:"leaderApiAddr,omitempty"`
	AutoJoin            *string         `json:"autoJoin,omitempty"`
	AutoJoinPort        *int32          `json:"autoJoinPort,omitempty"`
	LeaderTlsServername *string         `json:"leaderTlsServername,omitempty"`
	LeaderTls           *SecretSelector `json:"leaderTls,omitempty"`
	// +kubebuilder:validation:Enum=http;https
	AutoJoinScheme *string `json:"autoJoinScheme,omitempty"`
}

func (s *StorageRaftRetrySpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.LeaderApiAddr != nil {
		_m["leader_api_addr"] = strconv.Quote(*s.LeaderApiAddr)
	}
	if s.AutoJoin != nil {
		_m["auto_join"] = strconv.Quote(*s.AutoJoin)
	}
	if s.AutoJoinPort != nil {
		_m["auto_join_port"] = strconv.FormatInt(int64(*s.AutoJoinPort), 10)
	}
	if s.AutoJoinScheme != nil {
		_m["auto_join_scheme"] = strconv.Quote(*s.AutoJoinScheme)
	}
	if s.LeaderTlsServername != nil {
		_m["leader_tls_servername"] = strconv.Quote(*s.LeaderTlsServername)
	}
	if s.LeaderTls != nil {
		_tls := map[string]any{
			"leader_ca_cert_file":     strconv.Quote(fmt.Sprintf("/raft/retry/%x/ca.crt", xxh3.HashString(*s.LeaderApiAddr))),
			"leader_client_cert_file": strconv.Quote(fmt.Sprintf("/raft/retry/%x/tls.crt", xxh3.HashString(*s.LeaderApiAddr))),
			"leader_client_key_file":  strconv.Quote(fmt.Sprintf("/raft/retry/%x/tls.key", xxh3.HashString(*s.LeaderApiAddr))),
		}
		maps.Copy(_m, _tls)
	}
	return _m
}

type StorageRaftSpec struct {
	NodeId                             *string                `json:"nodeId,omitempty"`
	AutopilotUpgradeVersion            *string                `json:"autopilotUpgradeVersion,omitempty"`
	AutopilotRedundancyZone            *string                `json:"autopilotRedundancyZone,omitempty"`
	Experimental                       *map[string]string     `json:"experimental,omitempty"`
	RetryJoin                          []StorageRaftRetrySpec `json:"retryJoin,omitempty"`
	PerformanceMultiplier              *int32                 `json:"performanceMultiplier,omitempty"`
	TrailingLogs                       *int32                 `json:"trailingLogs,omitempty"`
	SnapshotThreshold                  *int32                 `json:"snapshotThreshold,omitempty"`
	SnapshotInterval                   *int32                 `json:"snapshotInterval,omitempty"`
	RetryJoinAsNonVoter                *bool                  `json:"retryJoinAsNonVoter,omitempty"`
	MaxEntrySize                       *int32                 `json:"maxEntrySize,omitempty"`
	MaxMountAndNamespaceTableEntrySize *int32                 `json:"maxMountAndNamespaceTableEntrySize,omitempty"`
	AutopilotReconcileInterval         *metav1.Duration       `json:"autopilotReconcileInterval,omitempty"`
	AutopilotUpdateInterval            *metav1.Duration       `json:"autopilotUpdateInterval,omitempty"`
}

func (s *StorageRaftSpec) Type() string {
	return "raft"
}

func (s *StorageRaftSpec) MapValue() map[string]any {
	_m := map[string]any{
		"path": strconv.Quote("/raft/"),
	}
	if s.NodeId != nil {
		_m["node_id"] = strconv.Quote(*s.NodeId)
	}
	if s.PerformanceMultiplier != nil {
		_m["performance_multiplier"] = strconv.FormatInt(int64(*s.PerformanceMultiplier), 10)
	}
	if s.SnapshotThreshold != nil {
		_m["snapshot_threshold"] = strconv.FormatInt(int64(*s.SnapshotThreshold), 10)
	}
	if s.SnapshotInterval != nil {
		_m["snapshot_interval"] = strconv.FormatInt(int64(*s.SnapshotInterval), 10)
	}
	if s.RetryJoinAsNonVoter != nil {
		_m["retry_join_as_non_voter"] = strconv.FormatBool(*s.RetryJoinAsNonVoter)
	}
	if s.MaxEntrySize != nil {
		_m["max_entry_size"] = strconv.FormatInt(int64(*s.MaxEntrySize), 10)
	}
	if s.AutopilotReconcileInterval != nil {
		_m["autopilot_reconcile_interval"] = strconv.Quote(fmt.Sprintf("%s", s.AutopilotReconcileInterval.Duration))
	}
	if s.RetryJoinAsNonVoter != nil {
		_m["autopilot_update_interval"] = strconv.Quote(fmt.Sprintf("%s", s.AutopilotUpdateInterval.Duration))
	}

	if s.MaxMountAndNamespaceTableEntrySize != nil {
		_m["max_mount_and_namespace_table_entry_size"] = strconv.FormatInt(int64(*s.MaxMountAndNamespaceTableEntrySize), 10)
	}
	if s.AutopilotUpgradeVersion != nil {
		_m["autopilot_upgrade_version"] = strconv.Quote(*s.AutopilotUpgradeVersion)
	}
	if s.AutopilotRedundancyZone != nil {
		_m["autopilot_redundancy_zone"] = strconv.Quote(*s.AutopilotRedundancyZone)
	}
	if s.Experimental != nil {
		for k, v := range *s.Experimental {
			_m[k] = v
		}
	}
	return _m
}

type StorageZooKeeperTLSpec struct {
	Enabled     *bool           `json:"enabled,omitempty"`
	Certificate *SecretSelector `json:"certificate,omitempty"`
	MinVersion  *TLSVersion     `json:"minVersion,omitempty"`
	SkipVerify  *bool           `json:"skipVerify,omitempty"`
	VerifyIP    *bool           `json:"verifyIp,omitempty"`
}

type StorageZooKeeperSpec struct {
	Address      *string                 `json:"address,omitempty"`
	Path         *string                 `json:"path,omitempty"`
	RedirectAddr *string                 `json:"redirect_addr,omitempty"`
	ZnodeOwner   *SecretKeySelector      `json:"znodeOwner,omitempty"`
	AuthInfo     *SecretKeySelector      `json:"authInfo,omitempty"`
	Tls          *StorageZooKeeperTLSpec `json:"tls,omitempty"`
}

func (s *StorageZooKeeperSpec) Type() string {
	return "zookeeper"
}

func (s *StorageZooKeeperSpec) MapValue() map[string]any {
	_m := map[string]any{
		"tls_enabled": strconv.FormatBool(false),
	}
	if s.Address != nil {
		_m["address"] = strconv.Quote(*s.Address)
	}
	if s.Path != nil {
		_m["path"] = strconv.Quote(*s.Path)
	}
	if s.ZnodeOwner != nil {
		_m["znode_owner"] = strconv.Quote("TODO")
	}
	if s.AuthInfo != nil {
		_m["auth_info"] = strconv.Quote("TODO")
	}
	if s.AuthInfo != nil {
		_m["redirect_addr"] = strconv.Quote(*s.RedirectAddr)
	}
	if s.Tls != nil && *s.Tls.Enabled {
		tls := map[string]any{
			"tls_enabled":   strconv.FormatBool(true),
			"tls_ca_file":   strconv.Quote("/zookeeper/ca.crt"),
			"tls_cert_file": strconv.Quote("/zookeeper/tls.crt"),
			"tls_key_file":  strconv.Quote("/zookeeper/tls.key"),
		}
		if s.Tls.MinVersion != nil {
			tls["tls_min_version"] = strconv.Quote(string(*s.Tls.MinVersion))
		}
		if s.AuthInfo != nil {
			tls["tls_skip_verify"] = strconv.FormatBool(*s.Tls.SkipVerify)
		}
		if s.AuthInfo != nil {
			tls["tls_verify_ip"] = strconv.FormatBool(*s.Tls.VerifyIP)
		}
		maps.Copy(_m, tls)
	}
	return _m
}

type StorageSwiftTenantSpec struct {
	Name *string `json:"name,omitempty"`
	ID   *string `json:"id,omitempty"`
}

func (s *StorageSwiftTenantSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Name != nil {
		_m["tenant"] = strconv.Quote(*s.Name)
	}
	if s.ID != nil {
		_m["tenant_id"] = strconv.Quote(*s.ID)
	}
	return _m
}

type StorageSwiftDomainNameSpec struct {
	Project *string `json:"project,omitempty"`
	User    *string `json:"user,omitempty"`
}

func (s *StorageSwiftDomainNameSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Project != nil {
		_m["project-domain"] = strconv.Quote(*s.Project)
	}
	if s.User != nil {
		_m["domain"] = strconv.Quote(*s.User)
	}
	return _m
}

type StorageSwiftSpec struct {
	InternalStorageSpec `json:",inline"`
	AuthUrl             *string                     `json:"authUrl,omitempty"`
	Container           *string                     `json:"container,omitempty"`
	StorageUrl          *string                     `json:"storageUrl,omitempty"`
	Region              *string                     `json:"region,omitempty"`
	Tenant              *StorageSwiftTenantSpec     `json:"tenant,omitempty"`
	DomainName          *StorageSwiftDomainNameSpec `json:"domainName,omitempty"`
	TrustId             *string                     `json:"trustId,omitempty"`
	Credentials         *SecretSelector             `json:"credentials,omitempty"`
}

func (s *StorageSwiftSpec) Type() string {
	return "swift"
}

func (s *StorageSwiftSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.AuthUrl != nil {
		_m["auth_url"] = strconv.Quote(*s.AuthUrl)
	}
	if s.Container != nil {
		_m["container"] = strconv.Quote(*s.Container)
	}
	if s.StorageUrl != nil {
		_m["storage_url"] = strconv.Quote(*s.StorageUrl)
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.TrustId != nil {
		_m["trust_id"] = strconv.Quote(*s.TrustId)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Tenant != nil {
		maps.Copy(_m, s.Tenant.MapValue())
	}
	if s.DomainName != nil {
		maps.Copy(_m, s.DomainName.MapValue())
	}

	return _m
}

type StorageS3Spec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              *string         `json:"bucket"`
	KmsKeyId            *string         `json:"kmsKeyId,omitempty"`
	Path                *string         `json:"path,omitempty"`
	Endpoint            *string         `json:"endpoint,omitempty"`
	Region              *string         `json:"region,omitempty"`
	S3ForcePathStyle    *bool           `json:"s3ForcePathStyle,omitempty"`
	DisableSsl          *bool           `json:"disableSsl,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageS3Spec) Type() string {
	return "s3"
}

func (s *StorageS3Spec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Bucket != nil {
		_m["bucket"] = strconv.Quote(*s.Bucket)
	}
	if s.Endpoint != nil {
		_m["endpoint"] = strconv.Quote(*s.Endpoint)
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.S3ForcePathStyle != nil {
		_m["s3_force_path_style"] = strconv.FormatBool(*s.S3ForcePathStyle)
	}
	if s.DisableSsl != nil {
		_m["disable_ssl"] = strconv.FormatBool(*s.DisableSsl)
	}
	if s.KmsKeyId != nil {
		_m["kms_key_id"] = strconv.Quote(*s.KmsKeyId)
	}
	if s.Path != nil {
		_m["path"] = strconv.Quote(*s.Path)
	}
	return _m
}

type StorageOCIObjectStorageSpec struct {
	Region         *string         `json:"region,omitempty"`
	NamespaceName  *string         `json:"namespaceName"`
	BucketName     *string         `json:"bucketName"`
	HaEnabled      *bool           `json:"haEnabled"`
	LockBucketName *string         `json:"lockBucketName"`
	Credentials    *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageOCIObjectStorageSpec) Type() string {
	return "oci"
}

func (s *StorageOCIObjectStorageSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.NamespaceName != nil {
		_m["namesapce_name"] = strconv.Quote(*s.NamespaceName)
	}
	if s.BucketName != nil {
		_m["bucket_name"] = strconv.Quote(*s.BucketName)
	}
	if s.LockBucketName != nil {
		_m["lock_bucket_name"] = strconv.Quote(*s.LockBucketName)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	return _m
}

type StoragePostgreSqlAwsSpec struct {
	DbRegion *string `json:"awsDbRegion,omitempty"`
}

func (s *StoragePostgreSqlAwsSpec) MapValue() map[string]any {
	_m := map[string]any{
		"auth_mode": strconv.Quote("aws_iam"),
	}
	if s.DbRegion != nil {
		_m["lock_bucket_name"] = strconv.Quote(*s.DbRegion)
	}
	return _m
}

type StoragePostgreSqlAzureSpec struct {
	ClientId *string `json:"clientId,omitempty"`
}

func (s *StoragePostgreSqlAzureSpec) MapValue() map[string]any {
	_m := map[string]any{
		"auth_mode": strconv.Quote("azure_msi"),
	}
	if s.ClientId != nil {
		_m["client_id"] = strconv.Quote(*s.ClientId)
	}
	return _m
}

type StoragePostgreSqlGcpSpec struct{}

func (s *StoragePostgreSqlGcpSpec) MapValue() map[string]any {
	return map[string]any{
		"auth_mode": strconv.Quote("gcp_iam"),
	}
}

type StoragePostgreSqlStandardSpec struct{}

func (s *StoragePostgreSqlStandardSpec) MapValue() map[string]any {
	return map[string]any{
		"auth_mode": strconv.Quote("standard"),
	}
}

// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type StoragePostgreSqlAutModeSpec struct {
	Aws      *StoragePostgreSqlAwsSpec      `json:"aws,omitempty"`
	Azure    *StoragePostgreSqlAzureSpec    `json:"azure,omitempty"`
	Gcp      *StoragePostgreSqlGcpSpec      `json:"gcp,omitempty"`
	Standard *StoragePostgreSqlStandardSpec `json:"standard,omitempty"`
}

func (s *StoragePostgreSqlAutModeSpec) MapValue() map[string]any {
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
	return map[string]any{}
}

type StoragePostgreSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       *SecretKeySelector            `json:"connectionUrl,omitempty"`
	AuthMode            *StoragePostgreSqlAutModeSpec `json:"authMode,omitempty"`
	HaEnabled           *bool                         `json:"haEnabled,omitempty"`
	HATable             *string                       `json:"haTable,omitempty"`
	Table               *string                       `json:"table,omitempty"`
	MaxIdleConnections  *int32                        `json:"maxIdleConnections,omitempty"`
}

func (s *StoragePostgreSqlSpec) Type() string {
	return "postgresql"
}

func (s *StoragePostgreSqlSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.MaxIdleConnections != nil {
		_m["max_idle_connections"] = strconv.FormatInt(int64(*s.MaxIdleConnections), 10)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.HATable != nil {
		_m["ha_table"] = strconv.Quote(*s.HATable)
	}

	if s.AuthMode != nil {
		maps.Copy(_m, s.AuthMode.MapValue())
	}
	return _m
}

type StorageDynamoDBCapacitySpec struct {
	Read  *int32 `json:"read,omitempty"`
	Write *int32 `json:"write,omitempty"`
}

func (s *StorageDynamoDBCapacitySpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.Read != nil {
		_m["read_capacity"] = strconv.FormatInt(int64(*s.Read), 10)
	}
	if s.Write != nil {
		_m["write_capacity"] = strconv.FormatInt(int64(*s.Write), 10)
	}
	return _m
}

type StorageDynamoDBSpec struct {
	InternalStorageSpec  `json:",inline"`
	DynamodbAllowUpdates *string                      `json:"dynamodbAllowUpdates,omitempty"`
	Credentials          *SecretSelector              `json:"credentials,omitempty"`
	Capacity             *StorageDynamoDBCapacitySpec `json:"capacity,omitempty"`
	Endpoint             *string                      `json:"endpoint,omitempty"`
	HaEnabled            *bool                        `json:"haEnabled,omitempty"`
	Region               *string                      `json:"region,omitempty"`
	Table                *string                      `json:"table,omitempty"`
	// +kubebuilder:validation:Enum=PROVISIONED;PAY_PER_REQUEST
	BillingMode *string `json:"billingMode,omitempty"`
}

func (s *StorageDynamoDBSpec) Type() string {
	return "dynamodb"
}

func (s *StorageDynamoDBSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.BillingMode != nil {
		_m["billing_mode"] = strconv.Quote(*s.BillingMode)
	}
	if s.DynamodbAllowUpdates != nil {
		_m["dynamodb_allow_updates"] = strconv.Quote(*s.DynamodbAllowUpdates)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Endpoint != nil {
		_m["endpoint"] = strconv.Quote(*s.Endpoint)
	}
	if s.Region != nil {
		_m["region"] = strconv.Quote(*s.Region)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}

	if s.Capacity != nil {
		maps.Copy(_m, s.Capacity.MapValue())
	}
	return _m
}

type StorageMySqlSpec struct {
	InternalStorageSpec        `json:",inline"`
	Address                    *string               `json:"address"`
	Credentials                *SecretSelector       `json:"credentials,omitempty"`
	Database                   *string               `json:"database,omitempty"`
	Table                      *string               `json:"table,omitempty"`
	MaxIdleConnections         *int32                `json:"maxIdleConnections,omitempty"`
	MaxConnectionLifetime      *int32                `json:"maxConnectionLifetime,omitempty"`
	PlaintextConnectionAllowed *string               `json:"plaintextConnectionAllowed,omitempty"`
	TlsCa                      *ConfigMapKeySelector `json:"tlsCa,omitempty"`
	HaEnabled                  *bool                 `json:"haEnabled,omitempty"`
	LockTable                  *string               `json:"lockTable,omitempty"`
}

func (s *StorageMySqlSpec) Type() string {
	return "mysql"
}

func (s *StorageMySqlSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Address != nil {
		_m["address"] = strconv.Quote(*s.Address)
	}
	if s.Database != nil {
		_m["database"] = strconv.Quote(*s.Database)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.MaxIdleConnections != nil {
		_m["max_idle_connections"] = strconv.FormatInt(int64(*s.MaxIdleConnections), 10)
	}
	if s.MaxConnectionLifetime != nil {
		_m["max_connection_lifetime"] = strconv.FormatInt(int64(*s.MaxConnectionLifetime), 10)
	}
	if s.LockTable != nil {
		_m["lock_table"] = strconv.Quote(*s.LockTable)
	}
	if s.TlsCa != nil {
		_m["tls_ca_file"] = strconv.Quote("/mysql/ca.crt")
	}
	return _m
}

type StorageMsSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	Server              *string         `json:"server"`
	Port                *int32          `json:"port,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
	Database            *string         `json:"database,omitempty"`
	Table               *string         `json:"table,omitempty"`
	Schema              *string         `json:"schema,omitempty"`
	ConnectionTimeout   *int32          `json:"connectionTimeout,omitempty"`
	AppName             *string         `json:"appName,omitempty"`

	// +kubebuilder:validation:Maximum=63
	// +kubebuilder:validation:Minimum=0
	LogLevel *int32 `json:"logLevel,omitempty"`
}

func (s *StorageMsSqlSpec) Type() string {
	return "mssql"
}

func (s *StorageMsSqlSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Server != nil {
		_m["server"] = strconv.Quote(*s.Server)
	}
	if s.Port != nil {
		_m["port"] = strconv.FormatInt(int64(*s.Port), 10)
	}
	if s.Database != nil {
		_m["database"] = strconv.Quote(*s.Database)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.Schema != nil {
		_m["schema"] = strconv.Quote(*s.Schema)
	}
	if s.AppName != nil {
		_m["appname"] = strconv.Quote(*s.AppName)
	}
	if s.ConnectionTimeout != nil {
		_m["connectionTimeout"] = strconv.FormatInt(int64(*s.ConnectionTimeout), 10)
	}
	if s.LogLevel != nil {
		_m["logLevel"] = strconv.FormatInt(int64(*s.LogLevel), 10)
	}
	return _m
}

type StorageEtcdTimeoutSpec struct {
	Request *metav1.Duration `json:"request,omitempty"`
	Lock    *metav1.Duration `json:"lock,omitempty"`
}

func (s *StorageEtcdTimeoutSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Request != nil {
		_m["request_timeout"] = strconv.Quote(fmt.Sprintf("%s", s.Request.Duration))
	}
	if s.Lock != nil {
		_m["lock_timeout"] = strconv.Quote(fmt.Sprintf("%s", s.Lock.Duration))
	}
	return _m
}

type StorageEtcdMaxSpec struct {
	ReceiveSize *datasize.Size `json:"receiveSize,omitempty"`
	SendSize    *datasize.Size `json:"sendSize,omitempty"`
}

func (s *StorageEtcdMaxSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.ReceiveSize != nil {
		_m["max_receive_size"] = strconv.Quote(fmt.Sprintf("%d", s.ReceiveSize.Bytes()))
	}
	if s.SendSize != nil {
		_m["max_send_size"] = strconv.Quote(fmt.Sprintf("%d", s.SendSize.Bytes()))
	}
	return _m
}

type StorageEtcdSpec struct {
	Address          *string                 `json:"address,omitempty"`
	DiscoverySrv     *string                 `json:"discoverySrv,omitempty"`
	DiscoverySrvName *string                 `json:"discoverySrvName,omitempty"`
	EtcdApi          *string                 `json:"etcdApi,omitempty"`
	HaEnabled        *bool                   `json:"haEnabled,omitempty"`
	Path             *string                 `json:"path,omitempty"`
	Sync             *bool                   `json:"sync,omitempty"`
	Timeout          *StorageEtcdTimeoutSpec `json:"timeout,omitempty"`
	Max              *StorageEtcdMaxSpec     `json:"max,omitempty"`
	Credentials      *SecretSelector         `json:"credentials,omitempty"`
	Tls              *SecretSelector         `json:"tls,omitempty"`
}

func (s *StorageEtcdSpec) Type() string {
	return "etcd"
}

func (s *StorageEtcdSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Address != nil {
		_m["address"] = strconv.Quote(*s.Address)
	}
	if s.DiscoverySrv != nil {
		_m["discovery_srv"] = strconv.Quote(*s.DiscoverySrv)
	}
	if s.DiscoverySrvName != nil {
		_m["discovery_srv_name"] = strconv.Quote(*s.DiscoverySrvName)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.EtcdApi != nil {
		_m["etcd_api"] = strconv.Quote(*s.EtcdApi)
	}
	if s.Path != nil {
		_m["path"] = strconv.Quote(*s.Path)
	}
	if s.Sync != nil {
		_m["sync"] = strconv.FormatBool(*s.Sync)
	}

	if s.Tls != nil {
		_m["tls_ca_file"] = strconv.Quote("/etcd/ca.crt")
		_m["tls_cert_file"] = strconv.Quote("/etcd/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/etcd/tls.key")
	}
	if s.Timeout != nil {
		maps.Copy(_m, s.Timeout.MapValue())
	}
	if s.Max != nil {
		maps.Copy(_m, s.Max.MapValue())
	}
	return _m
}

type StorageGoogleCloudSpannerSpec struct {
	InternalStorageSpec `json:",inline"`
	Database            *string         `json:"database,omitempty"`
	Table               *string         `json:"table,omitempty"`
	HaEnabled           *bool           `json:"haEnabled,omitempty"`
	HaTable             *string         `json:"haTable,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageGoogleCloudSpannerSpec) Type() string {
	return "spanner"
}

func (s *StorageGoogleCloudSpannerSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Database != nil {
		_m["database"] = strconv.Quote(*s.Database)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.HaTable != nil {
		_m["ha_table"] = strconv.Quote(*s.HaTable)
	}
	return _m
}

type StorageGoogleCloudStorageSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              *string         `json:"bucket,omitempty"`
	ChunkSize           *datasize.Size  `json:"chunkSize,omitempty"`
	HaEnabled           *bool           `json:"haEnabled,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageGoogleCloudStorageSpec) Type() string {
	return "gcs"
}

func (s *StorageGoogleCloudStorageSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Bucket != nil {
		_m["bucket"] = strconv.Quote(*s.Bucket)
	}
	if s.ChunkSize != nil {
		_m["chunk_size"] = strconv.FormatInt(int64(s.ChunkSize.Kilobytes()), 10)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	return _m
}

type StorageFoundationDbTlsSpec struct {
	VerifyPeers *string         `json:"tlsVerifyPeers,omitempty"`
	Certificate *SecretSelector `json:"certificate,omitempty"`
}

func (s *StorageFoundationDbTlsSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.VerifyPeers != nil {
		_m["tls_verify_peers"] = strconv.Quote(*s.VerifyPeers)
	}
	if s.Certificate != nil {
		_m["tls_ca_file"] = strconv.Quote("/foundationdb/tls/ca.crt")
		_m["tls_cert_file"] = strconv.Quote("/foundationdb/tls/tls.crt")
		_m["tls_key_file"] = strconv.Quote("/foundationdb/tls/tls.key")
	}
	return _m
}

type StorageFoundationDbSpec struct {
	ApiVersion  *int32                      `json:"apiVersion,omitempty"`
	ClusterFile *ConfigMapKeySelector       `json:"clusterFile,omitempty"`
	Tls         *StorageFoundationDbTlsSpec `json:"tls,omitempty"`
	Path        *string                     `json:"path,omitempty"`
	HaEnabled   *bool                       `json:"haEnabled,omitempty"`
}

func (s *StorageFoundationDbSpec) Type() string {
	return "foundationdb"
}

func (s *StorageFoundationDbSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.ClusterFile != nil {
		_m["bucket"] = strconv.Quote("/foundationdb/vault.cluster")
	}
	if s.Path != nil {
		_m["path"] = strconv.Quote(*s.Path)
	}
	if s.ApiVersion != nil {
		_m["api_version"] = strconv.FormatInt(int64(*s.ApiVersion), 10)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.Tls != nil {
		maps.Copy(_m, s.Tls.MapValue())
	}
	return _m
}

type StorageAerospikeSpec struct {
	Hostname    *string         `json:"hostname,omitempty"`
	Port        *int32          `json:"port,omitempty"`
	HostList    []string        `json:"hostList,omitempty"`
	Namepace    *string         `json:"namespace,omitempty"`
	Set         *string         `json:"set,omitempty"`
	ClusterName *string         `json:"clusterName,omitempty"`
	Timeout     *int32          `json:"timeout,omitempty"`
	IdleTImeout *int32          `json:"idleTImeout,omitempty"`
	Credentials *SecretSelector `json:"credentials,omitempty"`

	// +kubebuilder:validation:Enum=INTERNAL;EXTERNAL
	AuthMode *string `json:"authMode,omitempty"`
}

func (s *StorageAerospikeSpec) Type() string {
	return "aerospike"
}

func (s *StorageAerospikeSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Hostname != nil {
		_m["hostname"] = strconv.Quote(*s.Hostname)
	}
	if s.Port != nil {
		_m["port"] = strconv.Quote(strconv.FormatInt(int64(*s.Port), 10))
	}
	if len(s.HostList) > 0 {
		_m["hostlist"] = strconv.Quote(strings.Join(s.HostList, ","))
	}
	if s.ClusterName != nil {
		_m["cluster_name"] = strconv.Quote(*s.ClusterName)
	}
	if s.AuthMode != nil {
		_m["auth_mode"] = strconv.Quote(*s.AuthMode)
	}
	if s.Timeout != nil {
		_m["timeout"] = strconv.FormatInt(int64(*s.Timeout), 10)
	}
	if s.IdleTImeout != nil {
		_m["idle_timeout"] = strconv.FormatInt(int64(*s.IdleTImeout), 10)
	}
	if s.Credentials != nil {
		_m["username"] = strconv.Quote("TODO")
		_m["password"] = strconv.Quote("TODO")
	}
	return _m
}

type StorageAlicloudOssSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              *string         `json:"bucket,omitempty"`
	Endpoint            *string         `json:"endpoint,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageAlicloudOssSpec) Type() string {
	return "alicloudoss"
}

func (s *StorageAlicloudOssSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.Endpoint != nil {
		_m["endpoint"] = strconv.Quote(*s.Endpoint)
	}
	if s.Bucket != nil {
		_m["bucket"] = strconv.Quote(*s.Bucket)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	return _m
}

type StorageAzureSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountName         *string         `json:"accountName"`
	Container           *string         `json:"container"`
	Environment         *string         `json:"environment,omitempty"`
	ArmEndpoint         *string         `json:"armEndpoint,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

func (s *StorageAzureSpec) Type() string {
	return "alicloudoss"
}

func (s *StorageAzureSpec) MapValue() map[string]any {
	_m := map[string]any{}
	if s.AccountName != nil {
		_m["accountName"] = strconv.Quote(*s.AccountName)
	}
	if s.Container != nil {
		_m["container"] = strconv.Quote(*s.Container)
	}
	if s.Environment != nil {
		_m["environment"] = strconv.Quote(*s.Environment)
	}
	if s.ArmEndpoint != nil {
		_m["arm_endpoint"] = strconv.Quote(*s.ArmEndpoint)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	return _m
}

type PemBundleCassandraSpec struct {
	// +kubebuilder:validation:Enum=bundle;json
	Type        string             `json:"type"`
	Credentials *SecretKeySelector `json:"file,omitempty"`
}

type StorageCassandraSpec struct {
	InternalStorageSpec      `json:",inline"`
	Keyspace                 *string                 `json:"keyspace,omitempty"`
	Table                    *string                 `json:"table,omitempty"`
	ProtocolVersion          *int32                  `json:"protocolVersion,omitempty"`
	DisableInitialHostLookup *bool                   `json:"disableInitialHostLookup,omitempty"`
	InitialConnectionTimeout *int32                  `json:"initialConnectionTimeout,omitempty"`
	ConnectionTimeout        *int32                  `json:"connectionTimeout,omitempty"`
	SimpleRetryPolicyRetries *int32                  `json:"simpleRetryPolicyRetries,omitempty"`
	TlsSkipVerify            *int32                  `json:"tlsSkipVerify,omitempty"`
	Tls                      *int32                  `json:"tls,omitempty"`
	TlsMinVersion            *TLSVersion             `json:"tlsMinVersion,omitempty"`
	Pem                      *PemBundleCassandraSpec `json:"pem,omitempty"`
	Credentials              *SecretSelector         `json:"credentials,omitempty"`
	// +kubebuilder:validation:Enum=ANY;ONE;TWO;THREE;QUORUM;ALL;LOCAL_QUORUM;EACH_QUORUM;LOCAL_ONE
	Consistency string `json:"consistency,omitempty"`
	// +kubebuilder:validation:MinItems=1
	Hosts []string `json:"hosts,omitempty"`
}

func (s *StorageCassandraSpec) Type() string {
	return "cassandra"
}

func (s *StorageCassandraSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if len(s.Hosts) > 0 {
		_m["hosts"] = strconv.Quote(strings.Join(s.Hosts, ","))
	}
	if s.Keyspace != nil {
		_m["keyspace"] = strconv.Quote(*s.Keyspace)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.DisableInitialHostLookup != nil {
		_m["disable_initial_host_lookup"] = strconv.FormatBool(*s.DisableInitialHostLookup)
	}
	if s.InitialConnectionTimeout != nil {
		_m["initial_connection_timeout"] = strconv.FormatInt(int64(*s.InitialConnectionTimeout), 10)
	}
	if s.ConnectionTimeout != nil {
		_m["connection_timeout"] = strconv.FormatInt(int64(*s.ConnectionTimeout), 10)
	}
	if s.SimpleRetryPolicyRetries != nil {
		_m["simple_retry_policy_retries"] = strconv.FormatInt(int64(*s.SimpleRetryPolicyRetries), 10)
	}
	if s.Tls != nil {
		_m["tls"] = strconv.FormatInt(int64(*s.Tls), 10)
	}
	if s.TlsSkipVerify != nil {
		_m["tls_skip_verify"] = strconv.FormatInt(int64(*s.TlsSkipVerify), 10)
	}
	if s.TlsMinVersion != nil {
		_m["tls_min_version"] = strconv.Quote(string(*s.TlsMinVersion))
	}
	if s.Credentials != nil {
		_m["username"] = strconv.Quote("TODO")
		_m["password"] = strconv.Quote("TODO")
	}

	if s.Pem.Type == "json" {
		_m["pem_json_file"] = strconv.Quote("/cassandra/cert.json")
	}
	if s.Pem.Type == "json" {
		_m["pem_bundle_file"] = strconv.Quote("/cassandra/cert.pem")
	}
	return _m
}

type StorageCockroachDBSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       *SecretKeySelector `json:"connectionUrl"`
	Table               *string            `json:"table,omitempty"`
	HaEnabled           *bool              `json:"haEnabled,omitempty"`
	HaTable             *string            `json:"haTable,omitempty"`
}

func (s *StorageCockroachDBSpec) Type() string {
	return "cockroachdb"
}

func (s *StorageCockroachDBSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.MaxParallel != nil {
		_m["hosts"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Table != nil {
		_m["table"] = strconv.Quote(*s.Table)
	}
	if s.HaEnabled != nil {
		_m["ha_enabled"] = strconv.FormatBool(*s.HaEnabled)
	}
	if s.HaTable != nil {
		_m["ha_table"] = strconv.Quote(*s.HaTable)
	}
	if s.ConnectionUrl != nil {
		_m["connection_url"] = strconv.Quote("TODO")
	}
	return _m
}

type StorageCouchDBSpec struct {
	InternalStorageSpec `json:",inline"`
	Endpoint            *string         `json:"endpoint"`
	Credentials         *SecretSelector `json:"credentials"`
}

func (s *StorageCouchDBSpec) Type() string {
	return "couchdb"
}

func (s *StorageCouchDBSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.Endpoint != nil {
		_m["endpoint"] = strconv.Quote(*s.Endpoint)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Credentials != nil {
		_m["username"] = strconv.Quote("TODO")
		_m["password"] = strconv.Quote("TODO")
	}
	return _m
}

type StorageConsulSpec struct {
	InternalStorageSpec `json:",inline"`
	ConsulSpec          `json:",inline"`
	Path                *string          `json:"path,omitempty"`
	SessionTTL          *metav1.Duration `json:"sessionTtl,omitempty"`
	LockWaitTime        *metav1.Duration `json:"lockWaitTime,omitempty"`
	// +kubebuilder:validation:Enum=default;strong
	ConsistencyMode *string `json:"consistencyMode,omitempty"`
}

func (s *StorageConsulSpec) Type() string {
	return "consul"
}

func (s *StorageConsulSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.ConsistencyMode != nil {
		_m["consistency_mode"] = strconv.Quote(*s.ConsistencyMode)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	if s.Path != nil {
		_m["path"] = strconv.Quote(*s.Path)
	}
	if s.SessionTTL != nil {
		_m["session_ttl"] = strconv.Quote(fmt.Sprintf("%d", s.SessionTTL.Duration))
	}
	if s.LockWaitTime != nil {
		_m["lock_wait_time"] = strconv.Quote(fmt.Sprintf("%ds", s.LockWaitTime.Duration))
	}
	maps.Copy(_m, s.ConsulSpec.MapValue())
	return _m
}

type StorageInMemSpec struct{}

func (s *StorageInMemSpec) Type() string {
	return "inmem"
}

func (s *StorageInMemSpec) MapValue() map[string]any {
	return map[string]any{}
}

type StorageFileSystemSpec struct {
	Path string `json:"-"`
}

func (s *StorageFileSystemSpec) Type() string {
	return "file"
}

func (s *StorageFileSystemSpec) MapValue() map[string]any {
	return map[string]any{
		"path": strconv.Quote("/data/"),
	}
}

type StorageMantaSpec struct {
	InternalStorageSpec `json:",inline"`
	Directory           *string `json:"directory"`
	User                *string `json:"user"`
	KeyId               *string `json:"keyId"`
	SubUser             *string `json:"subUser"`
	URL                 *string `json:"url"`
	//TODO:AddSSHKeyForAgent
}

func (s *StorageMantaSpec) Type() string {
	return "manta"
}

func (s *StorageMantaSpec) MapValue() map[string]any {
	_m := map[string]any{}

	if s.Directory != nil {
		_m["directory"] = strconv.Quote(*s.Directory)
	}
	if s.User != nil {
		_m["user"] = strconv.Quote(*s.User)
	}
	if s.KeyId != nil {
		_m["key_id"] = strconv.Quote(*s.KeyId)
	}
	if s.SubUser != nil {
		_m["subuser"] = strconv.Quote(*s.SubUser)
	}
	if s.URL != nil {
		_m["url"] = strconv.Quote(*s.URL)
	}
	if s.MaxParallel != nil {
		_m["max_parallel"] = strconv.FormatInt(int64(*s.MaxParallel), 10)
	}
	return _m
}

// #endregion
