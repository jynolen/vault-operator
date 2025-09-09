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

	url "github.com/jynolen/vault-operator/internal/url"

	k8s "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvFromSource represents the source of a set of ConfigMaps or Secrets

type DataReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type DatakeyReference struct {
	DataReference `json:",inline"`
	Key           string `json:"key"`
}

type SecretKeySelector struct {
	SecretRef DatakeyReference `json:"secretKeyRef"`
}

type SecretSelector struct {
	SecretRef DataReference `json:"secretRef"`
}

type ConfigMapSelector struct {
	ConfigMapRef DataReference `json:"configMapRef"`
}

type ConfigMapKeySelector struct {
	ConfigMapRef DatakeyReference `json:"configMapRef"`
}

// +kubebuilder:validation:Enum=tls10;tls11;tls12;tls13
type TLSVersion string

type ConsulTlsSpec struct {
	CaCert     ConfigMapKeySelector `json:"caCert,omitempty"`
	ClientCert SecretSelector       `json:"clientCert"`
	MinVersion TLSVersion           `json:"minVersion,omitempty"`
	SkipVerify bool                 `json:"skipVerify,omitempty"`
}

type ConsulSpec struct {
	Address             string            `json:"address,omitempty"`
	CheckTimeout        time.Duration     `json:"checkTimeout,omitempty"`
	DisableRegistration bool              `json:"disableRegistration,omitempty"`
	Service             string            `json:"service,omitempty"`
	ServiceTags         []string          `json:"serviceTags,omitempty"`
	ServiceMeta         map[string]string `json:"serviceMeta,omitempty"`
	ServiceAddress      string            `json:"serviceAddress,omitempty"`
	Token               SecretKeySelector `json:"clientCert"`
	Tls                 ConsulTlsSpec     `json:"tls,omitempty"`

	// +kubebuilder:validation:Enum=http;https
	Scheme string `json:"scheme,omitempty"`
}

// VaultServer is the Schema for the vaultservers API.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultServerSpec   `json:"spec,omitempty"`
	Status VaultServerStatus `json:"status,omitempty"`
}

type VaultServerSpec struct {
	Size   int32             `json:"size"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels,omitempty"`

	Config VaultServerConfigSpec `json:"config"`
}

type VaultServerConfigSpec struct {
	// Operator Managed Config
	ClusterAddr           url.URL `json:"-"`
	ApiAddr               url.URL `json:"-"`
	PluginDirectory       string  `json:"-"`
	PluginTmpdir          string  `json:"-"`
	PluginFileUid         int32   `json:"-"`
	PluginFilePermissions string  `json:"-"`
	PidFile               string  `json:"-"`
	LogFile               string  `json:"-"`

	// User Managed Config
	Ui                        bool          `json:"ui,omitempty"`
	DisableMLock              bool          `json:"disableMLock,omitempty"`
	ClusterName               string        `json:"clusterName,omitempty"`
	CacheSize                 int32         `json:"cacheSize,omitempty"`
	DisableSize               bool          `json:"disableCache,omitempty"`
	DefaultLeaseTTL           time.Duration `json:"defaultLeaseTTL,omitempty"`
	MaxLeaseTTL               time.Duration `json:"maxLeaseTTL,omitempty"`
	DefaultMaxRequestDuration time.Duration `json:"defaultMaxRequestDuration,omitempty"`

	// +kubebuilder:validation:Enum=statelock;quotas;expiration
	DetectDeadlocks                string `json:"detectDeadlocks,omitempty"`
	RawStorageEndpoint             string `json:"rawStorageEndpoint,omitempty"`
	IntrospectionEndpoint          string `json:"introspectionEndpoint,omitempty"`
	EnableResponseHeaderHostname   bool   `json:"enableResponseHeaderHostname,omitempty"`
	EnableResponseHeaderRaftNodeId bool   `json:"enableResponseHeaderRaftNodeId,omitempty"`

	// +kubebuilder:validation:Enum=trace;debug;info;warn;error
	LogLevel string `json:"logLevel,omitempty"`
	// +kubebuilder:validation:Enum=standard;json
	LogFormat   string   `json:"logFormat,omitempty"`
	Experiments []string `json:"experiments,omitempty"`

	ImpreciseLeaseRoleTracking bool `json:"impreciseLeaseRoleTracking,omitempty"`
	EnablePostUnsealTrace      bool `json:"enablePostUnsealTrace,omitempty"`

	DisableClustering           bool              `json:"disableClustering,omitempty"`
	DisableSealwrap             bool              `json:"disableSealwrap,omitempty"`
	DisablePerformanceStandby   bool              `json:"disablePerformanceStandby,omitempty"`
	License                     SecretKeySelector `json:"license,omitempty"`
	AdministrativeNamespacePath string            `json:"administrativeNamespacePath,omitempty"`
	RemoveIrrevocableLeaseAfter time.Duration     `json:"removeIrrevocableLeaseAfter,omitempty"`

	// OSS features stanza
	Listener            ListenerSpec            `json:"listener,omitempty"`
	Telemetry           TelemetrySpec           `json:"telemetry,omitempty"`
	UserLockout         []UserLockoutSpec       `json:"userLockout,omitempty"`
	Seal                SealSpec                `json:"seal,omitempty"`
	ServiceRegistration ServiceRegistrationSpec `json:"serviceRegistration,omitempty"`
	Storage             StorageSpec             `json:"storage,omitempty"`

	// Enterprise features stanza
	KMSLibrary                 KmsLibrarySpec                 `json:"kmsLibrary,omitempty"`
	Replication                ReplicationSpec                `json:"replication,omitempty"`
	Reporting                  ReportingSpec                  `json:"reporting,omitempty"`
	SentinelSpec               SentinelSpec                   `json:"sentinel,omitempty"`
	AdaptiveOverloadProtection AdaptiveOverloadProtectionSpec `json:"adaptiveOverloadProtection,omitempty"`
}

// #region ListenerSpec
// Resource Specification Specific to Listener

type ListenerSpec struct {
	Type                                  string                    `json:"-"`
	MaxRequestLimit                       MaxRequestLimitSpec       `json:"maxRequestLimit,omitempty"`
	RequireRequestHeader                  bool                      `json:"requireRequestHeader,omitempty"`
	TLS                                   ListenerTlsSpec           `json:"tls,omitempty"`
	HTTPimeout                            HTTPTimeoutSpec           `json:"httpTimeout,omitempty"`
	ProxyProtocol                         ProxyProtocolSpec         `json:"proxyProcotol,omitempty"`
	XForwarded                            XForwardedForSpec         `json:"xForwardedFor,omitempty"`
	UnauthenticatedMetricsAccess          bool                      `json:"unauthenticatedMetricsAccess,omitempty"`
	UnauthenticatedPprofAccess            bool                      `json:"unauthenticatedPprofAccess,omitempty"`
	UnauthenticatedInFlightRequestsAccess bool                      `json:"unauthenticatedInFlightRequestsAccess,omitempty"`
	Cors                                  CorsSpec                  `json:"cors,omitempty"`
	CustomResponseHeaders                 CustomResponseHeadersSpec `json:"customResponseHeaders,omitempty"`
	ChrootNamespace                       string                    `json:"chrootNamespace,omitempty"`
	Redact                                RedactSpec                `json:"redact,omitempty"`
	DisableReplicationStatusEndpoints     bool                      `json:"disableReplicationStatusEndpoints,omitempty"`
	DisableRequestLimiter                 bool                      `json:"disableRequestLimiter,omitempty"`
	CustomMaxJson                         CustomMaxJsonSpec         `json:"customMaxJson,omitempty"`
}

type MaxRequestLimitSpec struct {
	Size     resource.Quantity `json:"size,omitempty"`
	Duration time.Duration     `json:"duration,omitempty"`
}

type ListenerTlsSpec struct {
	Disable bool           `json:"disable,omitempty"`
	Cert    SecretSelector `json:"certificate,omitempty"`

	MinVersion TLSVersion `json:"minVersion,omitempty"`
	MaxVersion TLSVersion `json:"maxVersion,omitempty"`

	// +kubebuilder:validation:Enum=TLS_RSA_WITH_RC4_128_SHA;TLS_RSA_WITH_3DES_EDE_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA;TLS_RSA_WITH_AES_256_CBC_SHA;TLS_RSA_WITH_AES_128_CBC_SHA256;TLS_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_RC4_128_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_RSA_WITH_RC4_128_SHA;TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA;TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA;TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256;TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256;TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305;TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256;TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256;TLS_AES_128_GCM_SHA256;TLS_AES_256_GCM_SHA384;TLS_CHACHA20_POLY1305_SHA256
	CipherSuites string `json:"cipherSuites,omitempty"`

	ClientCa ConfigMapKeySelector `json:"license,omitempty"`

	RequireAndVerifyClientCert bool `json:"requireAndVerifyClientCert,omitempty"`
	DisableClientCerts         bool `json:"disableClientCerts,omitempty"`
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

// +kubebuilder:validation:Pattern=[12345]\d\d
type SpecificStatusCodeSpec string

// +kubebuilder:validation:Pattern=[12345x][x\d][x\d]
type CollectiveStatusCodeSpec string

// #endregion

// #region TelemetrySpec
// Resource Specification Specific to Telemetry

type TelemetrySpec struct {
	DisableHostname                  bool          `json:"disableHostname,omitempty"`
	EnableHostnameLabel              bool          `json:"enable_hostname_label,omitempty"`
	MetricsPrefix                    string        `json:"metrics_prefix,omitempty"`
	LeaseMetricsEpsilon              time.Duration `json:"lease_metrics_epsilon,omitempty"`
	NumLeaseMetricsTimeBuckets       int32         `json:"num_lease_metrics_buckets,omitempty"`
	LeaseMetricsNameSpaceLabels      bool          `json:"add_lease_metrics_namespace_label,omitempty"`
	RollbackMetricsIncludeMountPoint bool          `json:"add_mount_point_rollback_metrics,omitempty"`
	FilterDefault                    bool          `json:"filter_default,omitempty"`
	PrefixFilter                     []string      `json:"prefix_filter,omitempty"`
	MaximumGaugeCardinality          int32         `json:"maximum_gauge_cardinality,omitempty"`
	UsageGaugePeriod                 time.Duration `json:"usage_gauge_period,omitempty"`

	Statsite    StatSiteSpec    `json:"statsite,omitempty"`
	Statsd      StatsdSpec      `json:"statsd,omitempty"`
	Circonus    CirconusSpec    `json:"circonus,omitempty"`
	DogStatsD   DogStatsDSpec   `json:"dogStatsD,omitempty"`
	StackDriver StackDriverSpec `json:"stackDriver,omitempty"`
	Prometheus  PrometheusSpec  `json:"prometheus,omitempty"`
}

type StatSiteSpec struct {
	Addr string `json:"address,omitempty"`
}

type StatsdSpec struct {
	Addr string `json:"address,omitempty"`
}

type CirconusSpec struct {
	Credentials                SecretSelector `json:"credentials"`
	ApiURL                     string         `json:"apiUrl"`
	SubmissionInterval         string         `json:"submissionInterval"`
	CheckSubmissionURL         string         `json:"submissionUrl"`
	CheckID                    string         `json:"checkId"`
	CheckForceMetricActivation string         `json:"checkForceMetricActivation"`
	CheckInstanceID            string         `json:"checkInstanceId"`
	CheckSearchTag             string         `json:"checkSearchTag"`
	CheckTags                  string         `json:"checkTags"`
	CheckDisplayName           string         `json:"checkDisplayName"`
	BrokerID                   string         `json:"brokerId"`
	BrokerSelectTag            string         `json:"brokerSelectTag"`
}

type DogStatsDSpec struct {
	Addr string   `json:"addr,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type PrometheusSpec struct {
	RetentionTime time.Duration `json:"retentionTime,omitempty"`
}

type StackDriverSpec struct {
	ProjectID string `json:"projectId,omitempty"`
	Location  string `json:"location,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	DebugLogs bool   `json:"debugLogs,omitempty"`
}

// #endregion

// #region UserLockoutSpec
// Resource Specification Specific to UserLockout

type InternalUserLockoutSpec struct {
	Threshold      int32         `json:"threshold,omitempty"`
	Duration       time.Duration `json:"duration,omitempty"`
	CounterReset   time.Duration `json:"counterReset,omitempty"`
	DisableLockout bool          `json:"disableLockout,omitempty"`
}

type UserLockoutSpec struct {
	All      InternalUserLockoutSpec `json:"all,omitempty"`
	UserPass InternalUserLockoutSpec `json:"userpass,omitempty"`
	LDAP     InternalUserLockoutSpec `json:"ldap,omitempty"`
	AppRole  InternalUserLockoutSpec `json:"appRole,omitempty"`
}

// #endregion

// #region SealSpec
// Resource Specification Specific to Seal

type SealSpec struct {
	AliCloudKMS   AliCloudKMSSealSpec   `json:"aliCloudKms,omitempty"`
	AWSKms        AWSKmsSealSpec        `json:"awsKms,omitempty"`
	AzureKeyVault AzureKeyVaultSealSpec `json:"azureKeyVault,omitempty"`
	GCPKms        GcpKmsSealSpec        `json:"gcpKms,omitempty"`
	OCIKms        OCIKmsSealSpec        `json:"ociKms,omitempty"`
	PKCS11        PKCS11SealSpec        `json:"pkcs11,omitempty"`
	Transit       TransitSealSpec       `json:"transit,omitempty"`
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

// #region ServiceRegistrationSpec
// Resource Specification Specific to ServiceRegistration

type ServiceRegistrationKubernetesSpec struct {
	Nmaespace string `json:"namesspace,omitempty"`
	PodName   string `json:"podname,omitempty"`
}

type ServiceRegistrationSpec struct {
	Consul     ConsulSpec                        `json:"consul"`
	Kubernetes ServiceRegistrationKubernetesSpec `json:"kubernetes"`
}

// #endregion

// #region StorageSpec
// Resource Specification Specific to Storage

type StorageSpec struct {
	Aerospike          StorageAerospikeSpec          `json:"aerospike,omitempty"`
	AlicloudOss        StorageAlicloudOssSpec        `json:"alicloudOss,omitempty"`
	InMem              StorageInMemSpec              `json:"inMem,omitempty"`
	FileSystem         StorageFileSystemSpec         `json:"fileSystem,omitempty"`
	Azure              StorageAzureSpec              `json:"azure,omitempty"`
	Cassandra          StorageCassandraSpec          `json:"cassandra,omitempty"`
	CockroachDB        StorageCockroachDBSpec        `json:"cockroachDb,omitempty"`
	Consul             StorageConsulSpec             `json:"consul,omitempty"`
	CouchDB            StorageCouchDBSpec            `json:"couchDb,omitempty"`
	DynamoDB           StorageDynamoDBSpec           `json:"dynamoDb,omitempty"`
	Etcd               StorageEtcdSpec               `json:"etcd,omitempty"`
	FoundationDb       StorageFoundationDbSpec       `json:"foundationDb,omitempty"`
	GoogleCloudSpanner StorageGoogleCloudSpannerSpec `json:"googleCloudSpanner,omitempty"`
	GoogleCloudStorage StorageGoogleCloudStorageSpec `json:"googleCloudStorage,omitempty"`
	Raft               StorageRaftSpec               `json:"raft,omitempty"`
	Manta              StorageMantaSpec              `json:"manta,omitempty"`
	MsSql              StorageMsSqlSpec              `json:"msSql,omitempty"`
	MySql              StorageMySqlSpec              `json:"mySql,omitempty"`
	OCIObjectStorage   StorageOCIObjectStorageSpec   `json:"ociObjectStorage,omitempty"`
	PostgreSql         StoragePostgreSqlSpec         `json:"postgreSql,omitempty"`
	S3                 StorageS3Spec                 `json:"s3,omitempty"`
	Swift              StorageSwiftSpec              `json:"swift,omitempty"`
	ZooKeeper          StorageZooKeeperSpec          `json:"zooKeeper,omitempty"`
}

type InternalStorageSpec struct {
	MaxParallel int32 `json:"maxParallel,omitempty"`
}

type StorageRaftRetrySpec struct {
	LeaderApiAddr       string         `json:"leaderApiAddr,omitempty"`
	AutoJoin            string         `json:"autoJoin,omitempty"`
	AutoJoinPort        int32          `json:"autoJoinPort,omitempty"`
	LeaderTlsServername string         `json:"leaderTlsServername,omitempty"`
	LeaderTls           SecretSelector `json:"leaderTls,omitempty"`
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

type StoragePostgreSqlAutModeSpec struct {
	Aws   StoragePostgreSqlAwsSpec   `json:"aws,omitempty"`
	Azure StoragePostgreSqlAzureSpec `json:"azure,omitempty"`
	Gcp   StoragePostgreSqlGcpSpec   `json:"gcp,omitempty"`
}

type StoragePostgreSqlSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       SecretKeySelector            `json:"connectionUrl,omitempty"`
	HaEnabled           bool                         `json:"haEnabled,omitempty"`
	HATable             string                       `json:"haTable,omitempty"`
	Table               string                       `json:"table,omitempty"`
	MaxIdleConnections  int32                        `json:"maxIdleConnections,omitempty"`
	AuthMode            StoragePostgreSqlAutModeSpec `json:"authMode,omitempty"`
}

type StorageDynamoDBCapacitySpec struct {
	Read  int32 `json:"read,omitempty"`
	Write int32 `json:"write,omitempty"`
}

type StorageDynamoDBSpec struct {
	// +kubebuilder:validation:Enum=PROVISIONED;PAY_PER_REQUEST
	BillingMode          string                      `json:"billingMode,omitempty"`
	DynamodbAllowUpdates string                      `json:"dynamodbAllowUpdates,omitempty"`
	Endpoint             string                      `json:"endpoint,omitempty"`
	HaEnabled            bool                        `json:"haEnabled,omitempty"`
	MaxParallel          int32                       `json:"maxParallel,omitempty"`
	Region               int32                       `json:"region,omitempty"`
	Capacity             StorageDynamoDBCapacitySpec `json:"capacity,omitempty"`
	Table                string                      `json:"table,omitempty"`
	Credentials          SecretSelector              `json:"credentials,omitempty"`
}

type StorageMySqlSpec struct {
	InternalStorageSpec        `json:",inline"`
	Address                    int32                `json:"server"`
	Credentials                SecretSelector       `json:"credentials,omitempty"`
	Database                   string               `json:"database,omitempty"`
	Table                      string               `json:"table,omitempty"`
	Schema                     string               `json:"schema,omitempty"`
	MaxIdleConnections         int32                `json:"maxIdleConnections,omitempty"`
	MaxConnectionLifetime      int32                `json:"maxConnectionLifetime,omitempty"`
	PlaintextConnectionAllowed string               `json:"plaintextConnectionAllowed,omitempty"`
	TlsCa                      ConfigMapKeySelector `json:"tlsCa,omitempty"`
	HaEnabled                  bool                 `json:"haEnabled,omitempty"`
	LockTable                  string               `json:"lockTable,omitempty"`
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
	Address          string                 `json:"address,omitempty"`
	DiscoverySrv     string                 `json:"discoverySrv,omitempty"`
	DiscoverySrvName string                 `json:"discoverySrvName,omitempty"`
	EtcdApi          string                 `json:"etcdApi,omitempty"`
	HaEnabled        bool                   `json:"haEnabled,omitempty"`
	Path             string                 `json:"path,omitempty"`
	Sync             bool                   `json:"sync,omitempty"`
	Timeout          StorageEtcdTimeoutSpec `json:"timeout,omitempty"`
	Max              StorageEtcdMaxSpec     `json:"max,omitempty"`
	Credentials      SecretSelector         `json:"credentials,omitempty"`
	TLS              SecretSelector         `json:"tls,omitempty"`
}

type StorageGoogleCloudSpannerSpec struct {
	InternalStorageSpec `json:",inline"`
	Database            string         `json:"database,omitempty"`
	Table               string         `json:"table,omitempty"`
	HaEnabled           bool           `json:"haEnabled,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageGoogleCloudStorageSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              string         `json:"bucket,omitempty"`
	ChunkSize           int32          `json:"chunkSize,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageFoundationDbSpec struct {
	ApiVersion  int32                `json:"apiVersion,omitempty"`
	ClusterFile ConfigMapKeySelector `json:"clusterFile,omitempty"`
	Tls         SecretSelector       `json:"tls,omitempty"`
	Path        string               `json:"path,omitempty"`
	HaEnabled   bool                 `json:"haEnabled,omitempty"`
}

type StorageAerospikeSpec struct {
	Hostname    string         `json:"hostname,omitempty"`
	Port        int32          `json:"port,omitempty"`
	HostList    []string       `json:"hostList,omitempty"`
	Namepace    string         `json:"namespace,omitempty"`
	Set         string         `json:"set,omitempty"`
	ClusterName string         `json:"clusterName,omitempty"`
	Timeout     int32          `json:"timeout,omitempty"`
	IdleTImeout int32          `json:"idleTImeout,omitempty"`
	Credentials SecretSelector `json:"credentials,omitempty"`

	// +kubebuilder:validation:Enum=INTERNAL;EXTERNAL
	AuthMode string `json:"authMode,omitempty"`
}

type StorageAlicloudOssSpec struct {
	InternalStorageSpec `json:",inline"`
	Bucket              string         `json:"bucket,omitempty"`
	Endpoint            string         `json:"endpoint,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageAzureSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountName         string         `json:"accountName,omitempty"`
	Container           string         `json:"container,omitempty"`
	Environment         string         `json:"environment,omitempty"`
	ArmEndpoint         string         `json:"armEndpoint,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageCassandraSpec struct {
	InternalStorageSpec `json:",inline"`
	AccountName         string         `json:"accountName,omitempty"`
	Container           string         `json:"container,omitempty"`
	Environment         string         `json:"environment,omitempty"`
	ArmEndpoint         string         `json:"armEndpoint,omitempty"`
	Credentials         SecretSelector `json:"credentials,omitempty"`
}

type StorageCockroachDBSpec struct {
	InternalStorageSpec `json:",inline"`
	ConnectionUrl       SecretKeySelector `json:"connectionUrl"`
	Table               string            `json:"table,omitempty"`
	HaEnabled           bool              `json:"haEnabled,omitempty"`
	HATable             string            `json:"haTable,omitempty"`
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

// #region KmsLibrarySpec
// Resource Specification Specific to KmsLibrary

type KmsLibrarySpec struct {
	// +kubebuilder:validation:Enum=pkcs11
	Type    string                   `json:"type,omitempty"`
	Name    string                   `json:"name,omitempty"`
	Library k8s.ConfigMapKeySelector `json:"library,omitempty"`
}

// #endregion

// #region ReplicationSpec
// Resource Specification Specific to Replication

type ReplicationSpec struct {
	ResolverDiscoverServers   bool              `json:"resolverDiscoverServers,omitempty"`
	LogshipperBufferLength    int32             `json:"logshipperBufferLength,omitempty"`
	LogshipperBufferSize      resource.Quantity `json:"logshipperBufferSize,omitempty"`
	AllowForwardingViaHeader  bool              `json:"allowForwardingViaHeader,omitempty"`
	BestEffortWalWaitDuration time.Duration     `json:"bestEffortWalWaitDuratione,omitempty"`
}

// #endregion

// #region ReportingSpec
// Resource Specification Specific to Reporting

type ReportingLicenseSpec struct {
	Enabled               bool  `json:"enabled,omitempty"`
	BillingStartTimestamp int32 `json:"billingStartTimestamp,omitempty"`
	DevelopmentCluster    bool  `json:"developmentCluster,omitempty"`
}

type ReportingSpec struct {
	SnapshotRetentionTime        time.Duration        `json:"snapshotRetentionTime,omitempty"`
	DisableProductUsageReporting bool                 `json:"disableProductUsageReporting,omitempty"`
	License                      ReportingLicenseSpec `json:"license,omitempty"`
}

// #endregion

// #region SentinelSpec
// Resource Specification Specific to Sentinel

type SentinelSpec struct {
	AdditionalEnabledModules []string `json:"additionalEnabledModules,omitempty"`
}

// #endregion

// #region AdaptiveOverloadProtectionSpec
// Resource Specification Specific to AdaptiveOverloadProtection

type AdaptiveOverloadProtectionSpec struct {
	DisableWriteController bool `json:"disableWriteController,omitempty"`
}

// #endregion

// VaultServerSpec defines the desired state of VaultServer.

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

func init() {
	SchemeBuilder.Register(&VaultServer{}, &VaultServerList{})
}
