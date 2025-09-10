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

type InternalStorageSpec struct {
	MaxParallel int32 `json:"maxParallel,omitempty"`
}

type StorageRaftRetrySpec struct {
	LeaderApiAddr       string          `json:"leaderApiAddr,omitempty"`
	AutoJoin            string          `json:"autoJoin,omitempty"`
	AutoJoinPort        int32           `json:"autoJoinPort,omitempty"`
	LeaderTlsServername string          `json:"leaderTlsServername,omitempty"`
	LeaderTls           *SecretSelector `json:"leaderTls,omitempty"`
	// +kubebuilder:validation:Enum=http;https
	AutoJoinScheme string `json:"autoJoinScheme,omitempty"`
}
type StorageRaftSpec struct {
	Path                               string                 `json:"-"`
	NodeId                             string                 `json:"NodeId,omitempty"`
	PerformanceMultiplier              int32                  `json:"performanceMultiplier,omitempty"`
	TrailingLogs                       int32                  `json:"trailingLogs,omitempty"`
	SnapshotThreshold                  int32                  `json:"snapshotThreshold,omitempty"`
	SnapshotInterval                   int32                  `json:"snapshotInterval,omitempty"`
	RetryJoin                          []StorageRaftRetrySpec `json:"retryJoin,omitempty"`
	RetryJoinAsNonVoter                bool                   `json:"retryJoinAsNonVoter,omitempty"`
	MaxEntrySize                       int32                  `json:"maxEntrySize,omitempty"`
	MaxMountAndNamespaceTableEntrySize int32                  `json:"maxMountAndNamespaceTableEntrySize,omitempty"`
	AutopilotReconcileInterval         time.Duration          `json:"autopilotReconcileInterval,omitempty"`
	AutopilotUpdateInterval            time.Duration          `json:"autopilotUpdateInterval,omitempty"`
	AutopilotUpgradeVersion            string                 `json:"autopilotUpgradeVersion,omitempty"`
	AutopilotRedundancyZone            string                 `json:"autopilotRedundancyZone,omitempty"`
}

type StorageZooKeeperTLSpec struct {
	Enabled     bool           `json:"enabled,omitempty"`
	Certificate SecretSelector `json:"certificate,omitempty"`
	MinVersion  TLSVersion     `json:"minVersion,omitempty"`
	SkipVerify  bool           `json:"skipVerify,omitempty"`
	VerifyIP    bool           `json:"verifyIp,omitempty"`
}

type StorageZooKeeperSpec struct {
	Address    string                 `json:"address,omitempty"`
	Path       string                 `json:"path,omitempty"`
	ZnodeOwner string                 `json:"znodeOwner,omitempty"`
	AuthInfo   SecretKeySelector      `json:"authInfo,omitempty"`
	Tls        StorageZooKeeperTLSpec `json:"tls,omitempty"`
}

type StorageSwiftTenantSpec struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

type StorageSwiftDomainNameSpec struct {
	Project string `json:"project,omitempty"`
	User    string `json:"user,omitempty"`
}

type StorageSwiftSpec struct {
	InternalStorageSpec `json:",inline"`
	Container           string                     `json:"container,omitempty"`
	StorageUrl          bool                       `json:"storageUrl,omitempty"`
	Region              string                     `json:"region,omitempty"`
	Tenant              StorageSwiftTenantSpec     `json:"tenant,omitempty"`
	DomainName          StorageSwiftDomainNameSpec `json:"domainName,omitempty"`
	TrustId             string                     `json:"trustId,omitempty"`
	Credentials         SecretSelector             `json:"credentials,omitempty"`
}

type StorageS3Spec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              string         `json:"bucket,omitempty"`
	Endpoint            string         `json:"endpoint,omitempty"`
	S3ForcePathStyle    bool           `json:"s3ForcePathStyle,omitempty"`
	DisableSsl          bool           `json:"disableSsl,omitempty"`
	KmsKeyId            string         `json:"kmsKeyId,omitempty"`
	Path                string         `json:"path,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageOCIObjectStorageSpec struct {
	Region         string         `json:"region,omitempty"`
	NamespaceName  string         `json:"namespaceName"`
	BucketName     string         `json:"bucketName"`
	HaEnabled      bool           `json:"haEnabled"`
	LockBucketName string         `json:"lockBucketName"`
	Credentials    SecretSelector `json:"credentials,omitempty"`
}

type StoragePostgreSqlAwsSpec struct {
	DbRegion string `json:"awsDbRegion,omitempty"`
}
type StoragePostgreSqlAzureSpec struct {
	ClientId string `json:"awsDbRegion,omitempty"`
}
type StoragePostgreSqlGcpSpec struct{}

// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type StoragePostgreSqlAutModeSpec struct {
	Aws   StoragePostgreSqlAwsSpec   `json:"aws,omitempty"`
	Azure StoragePostgreSqlAzureSpec `json:"azure,omitempty"`
	Gcp   StoragePostgreSqlGcpSpec   `json:"gcp,omitempty"`
}

type StoragePostgreSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       SecretKeySelector             `json:"connectionUrl,omitempty"`
	HaEnabled           bool                          `json:"haEnabled,omitempty"`
	HATable             string                        `json:"haTable,omitempty"`
	Table               string                        `json:"table,omitempty"`
	MaxIdleConnections  int32                         `json:"maxIdleConnections,omitempty"`
	AuthMode            *StoragePostgreSqlAutModeSpec `json:"authMode,omitempty"`
}

type StorageDynamoDBCapacitySpec struct {
	Read  int32 `json:"read,omitempty"`
	Write int32 `json:"write,omitempty"`
}

type StorageDynamoDBSpec struct {
	// +kubebuilder:validation:Enum=PROVISIONED;PAY_PER_REQUEST
	BillingMode          string                       `json:"billingMode,omitempty"`
	DynamodbAllowUpdates string                       `json:"dynamodbAllowUpdates,omitempty"`
	Endpoint             string                       `json:"endpoint,omitempty"`
	HaEnabled            bool                         `json:"haEnabled,omitempty"`
	MaxParallel          int32                        `json:"maxParallel,omitempty"`
	Region               int32                        `json:"region,omitempty"`
	Capacity             *StorageDynamoDBCapacitySpec `json:"capacity,omitempty"`
	Table                string                       `json:"table,omitempty"`
	Credentials          SecretSelector               `json:"credentials,omitempty"`
}

type StorageMySqlSpec struct {
	InternalStorageSpec        `json:",inline"`
	Address                    int32                 `json:"server"`
	Credentials                SecretSelector        `json:"credentials,omitempty"`
	Database                   string                `json:"database,omitempty"`
	Table                      string                `json:"table,omitempty"`
	Schema                     string                `json:"schema,omitempty"`
	MaxIdleConnections         int32                 `json:"maxIdleConnections,omitempty"`
	MaxConnectionLifetime      int32                 `json:"maxConnectionLifetime,omitempty"`
	PlaintextConnectionAllowed string                `json:"plaintextConnectionAllowed,omitempty"`
	TlsCa                      *ConfigMapKeySelector `json:"tlsCa,omitempty"`
	HaEnabled                  bool                  `json:"haEnabled,omitempty"`
	LockTable                  string                `json:"lockTable,omitempty"`
}

type StorageMsSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	Server              string         `json:"server"`
	Port                int32          `json:"port,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
	Database            string         `json:"database,omitempty"`
	Table               string         `json:"table,omitempty"`
	Schema              string         `json:"schema,omitempty"`
	ConnectionTimeout   int32          `json:"connectionTimeout,omitempty"`
	AppName             string         `json:"appName,omitempty"`

	// +kubebuilder:validation:Maximum=63
	// +kubebuilder:validation:Minimum=0
	LogLevel int32 `json:"logLevel,omitempty"`
}

type StorageEtcdTimeoutSpec struct {
	Request time.Duration `json:"request,omitempty"`
	Lock    time.Duration `json:"lock,omitempty"`
}

type StorageEtcdMaxSpec struct {
	ReceiveSize int32 `json:"receiveSize,omitempty"`
	SendSize    int32 `json:"sendSize,omitempty"`
}

type StorageEtcdSpec struct {
	Address          string                  `json:"address,omitempty"`
	DiscoverySrv     string                  `json:"discoverySrv,omitempty"`
	DiscoverySrvName string                  `json:"discoverySrvName,omitempty"`
	EtcdApi          string                  `json:"etcdApi,omitempty"`
	HaEnabled        bool                    `json:"haEnabled,omitempty"`
	Path             string                  `json:"path,omitempty"`
	Sync             bool                    `json:"sync,omitempty"`
	Timeout          *StorageEtcdTimeoutSpec `json:"timeout,omitempty"`
	Max              *StorageEtcdMaxSpec     `json:"max,omitempty"`
	Credentials      SecretSelector          `json:"credentials,omitempty"`
	TLS              *SecretSelector         `json:"tls,omitempty"`
}

type StorageGoogleCloudSpannerSpec struct {
	InternalStorageSpec `json:",inline"`
	Database            string          `json:"database,omitempty"`
	Table               string          `json:"table,omitempty"`
	HaEnabled           bool            `json:"haEnabled,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

type StorageGoogleCloudStorageSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              string          `json:"bucket,omitempty"`
	ChunkSize           int32           `json:"chunkSize,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

type StorageFoundationDbSpec struct {
	ApiVersion  int32                `json:"apiVersion,omitempty"`
	ClusterFile ConfigMapKeySelector `json:"clusterFile,omitempty"`
	Tls         *SecretSelector      `json:"tls,omitempty"`
	Path        string               `json:"path,omitempty"`
	HaEnabled   bool                 `json:"haEnabled,omitempty"`
}

type StorageAerospikeSpec struct {
	Hostname    string          `json:"hostname,omitempty"`
	Port        int32           `json:"port,omitempty"`
	HostList    []string        `json:"hostList,omitempty"`
	Namepace    string          `json:"namespace,omitempty"`
	Set         string          `json:"set,omitempty"`
	ClusterName string          `json:"clusterName,omitempty"`
	Timeout     int32           `json:"timeout,omitempty"`
	IdleTImeout int32           `json:"idleTImeout,omitempty"`
	Credentials *SecretSelector `json:"credentials,omitempty"`

	// +kubebuilder:validation:Enum=INTERNAL;EXTERNAL
	AuthMode string `json:"authMode,omitempty"`
}

type StorageAlicloudOssSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              string          `json:"bucket,omitempty"`
	Endpoint            string          `json:"endpoint,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

type StorageAzureSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountName         string          `json:"accountName,omitempty"`
	Container           string          `json:"container,omitempty"`
	Environment         string          `json:"environment,omitempty"`
	ArmEndpoint         string          `json:"armEndpoint,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

type StorageCassandraSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountName         string          `json:"accountName,omitempty"`
	Container           string          `json:"container,omitempty"`
	Environment         string          `json:"environment,omitempty"`
	ArmEndpoint         string          `json:"armEndpoint,omitempty"`
	Credentials         *SecretSelector `json:"credentials,omitempty"`
}

type StorageCockroachDBSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       *SecretKeySelector `json:"connectionUrl"`
	Table               string             `json:"table,omitempty"`
	HaEnabled           bool               `json:"haEnabled,omitempty"`
	HATable             string             `json:"haTable,omitempty"`
}

type StorageCouchDBSpec struct {
	InternalStorageSpec `json:",inline"`
	Endpoint            string         `json:"endpoint"`
	Credentials         SecretSelector `json:"credentials"`
}

type StorageConsulSpec struct {
	InternalStorageSpec `json:",inline"`
	ConsulSpec          `json:",inline"`
	Path                string        `json:"path,omitempty"`
	SessionTTL          time.Duration `json:"sessionTtl,omitempty"`
	LockWaitTime        time.Duration `json:"lockWaitTime,omitempty"`
	// +kubebuilder:validation:Enum=default;strong
	ConsistencyMode string `json:"consistencyMode,omitempty"`
}

type StorageInMemSpec struct{}

type StorageFileSystemSpec struct {
	Path string `json:"-"`
}

type StorageMantaSpec struct {
	Directory string `json:"directory"`
	User      string `json:"user"`
	KeyId     string `json:"keyId"`
	SubUser   string `json:"subUser"`
	URL       string `json:"url"`
}

// #endregion
