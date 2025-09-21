package controller

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jynolen/vault-operator/api/v1alpha1"
	"github.com/jynolen/vault-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *VaultServerReconciler) resolveSecret(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	if vaultServer.Spec.Config.Telemetry != nil {
		r.telemetryCirconus(ctx, vaultServer, vaultServer.Spec.Config.Telemetry.Circonus)
	}
	return nil
}

func (r *VaultServerReconciler) telemetryCirconus(ctx context.Context, vaultServer *v1alpha1.VaultServer, circonus *v1alpha1.CirconusSpec) error {
	if circonus == nil {
		return nil
	}
	t := types.NamespacedName{Name: circonus.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"api_token": utils.KvMapping{Dest: circonus.ApiToken, Mandatory: true},
		"api_app":   utils.KvMapping{Dest: circonus.ApiApp},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "TelemetryCirconus.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealOCI(ctx context.Context, vaultServer *v1alpha1.VaultServer, azureKeyVault *v1alpha1.SealAzureKeyVaultSpec) error {
	t := types.NamespacedName{Name: azureKeyVault.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", secret.Name)
	}
	_, ok := secret.Data["config"]
	if !ok {
		return fmt.Errorf("%s does not contains key `config`", secret.Name)
	}
	return nil
}

func (r *VaultServerReconciler) sealTransit(ctx context.Context, vaultServer *v1alpha1.VaultServer, transit *v1alpha1.SealTransitSpec) error {
	t := types.NamespacedName{Name: transit.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"token": {Dest: transit.Token, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "SealTransitSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageAerospikeSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, aerospike *v1alpha1.StorageAerospikeSpec) error {
	if aerospike == nil {
		return nil
	}
	t := types.NamespacedName{Name: aerospike.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"username": {Dest: aerospike.Username, Mandatory: true},
		"password": {Dest: aerospike.Password, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageAerospikeSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageCassandraSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, cassandra *v1alpha1.StorageCassandraSpec) error {
	if cassandra == nil {
		return nil
	}
	t := types.NamespacedName{Name: cassandra.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"username": {Dest: cassandra.Username, Mandatory: true},
		"password": {Dest: cassandra.Password, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageCassandraSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageCockroachDBSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, coachroachdb *v1alpha1.StorageCockroachDBSpec) error {
	if coachroachdb == nil {
		return nil
	}
	t := types.NamespacedName{Name: coachroachdb.ConnectionUrlSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"connection_url": {Dest: coachroachdb.ConnectionUrl, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageCockroachDBSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageConsulSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, consul *v1alpha1.StorageConsulSpec) error {
	if consul == nil {
		return nil
	}
	t := types.NamespacedName{Name: consul.TokenSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"token": {Dest: consul.Token, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageConsulSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageCouchDBSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, couchdb *v1alpha1.StorageCouchDBSpec) error {
	if couchdb == nil {
		return nil
	}
	t := types.NamespacedName{Name: couchdb.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"username": {Dest: couchdb.Username, Mandatory: true},
		"password": {Dest: couchdb.Password, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageCouchDBSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageEtcdSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, etcd *v1alpha1.StorageEtcdSpec) error {
	if etcd == nil {
		return nil
	}
	t := types.NamespacedName{Name: etcd.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"username": {Dest: etcd.Username, Mandatory: true},
		"password": {Dest: etcd.Password, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageEtcdSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) StorageFoundationDbSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, foundationdb *v1alpha1.StorageFoundationDbSpec) error {
	if foundationdb == nil || foundationdb.Tls.Certificate == nil {
		return nil
	}
	t := types.NamespacedName{Name: foundationdb.Tls.Certificate.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}

	secretMappings := map[string]utils.KvMapping{
		"tls_password": {Dest: foundationdb.Tls.Password, Mandatory: false},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageFoundationDbSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageDynamoDBSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, dynamodb *v1alpha1.StorageDynamoDBSpec) error {
	if dynamodb == nil {
		return nil
	}
	t := types.NamespacedName{Name: dynamodb.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"access_key":    {Dest: dynamodb.AccessKey, Mandatory: true},
		"secret_key":    {Dest: dynamodb.SecretKey, Mandatory: true},
		"session_token": {Dest: dynamodb.SessionToken},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageDynamoDBSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageAlicloudOss(ctx context.Context, vaultServer *v1alpha1.VaultServer, aliCloudOss *v1alpha1.StorageAlicloudOssSpec) error {
	if aliCloudOss == nil {
		return nil
	}
	t := types.NamespacedName{Name: aliCloudOss.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"access_key": {Dest: aliCloudOss.AccessKey, Mandatory: true},
		"secret_key": {Dest: aliCloudOss.SecretKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageAlicloudOssSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) storageAzureSpec(ctx context.Context, vaultServer *v1alpha1.VaultServer, azure *v1alpha1.StorageAzureSpec) error {
	if azure == nil {
		return nil
	}
	t := types.NamespacedName{Name: azure.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]utils.KvMapping{
		"accountKey": {Dest: azure.AccountKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := utils.MapSecret(&secret, "StorageAzureSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}
