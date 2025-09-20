package controller

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jynolen/vault-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *VaultServerReconciler) resolveSecret(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	if vaultServer.Spec.Config.Telemetry != nil {
		r.telemetryCirconus(ctx, vaultServer, vaultServer.Spec.Config.Telemetry.Circonus)
	}
	return nil
}

func mapSecret(secret *corev1.Secret, name, src_key string, dest_key *string, mandatory ...bool) error {
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", name)
	}
	val, ok := secret.Data[src_key]
	if !ok {
		if len(mandatory) > 0 && mandatory[0] {
			return fmt.Errorf("%s does not contains key `%s`", name, src_key)
		}
		return nil
	}
	*dest_key = string(val)
	return nil
}

type kvMapping struct {
	Dest      *string
	Mandatory bool
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
	secretMappings := map[string]kvMapping{
		"api_token": kvMapping{Dest: circonus.ApiToken, Mandatory: true},
		"api_app":   kvMapping{Dest: circonus.ApiApp},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "TelemetryCirconus.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealAWSKMS(ctx context.Context, vaultServer *v1alpha1.VaultServer, awsKms *v1alpha1.SealAwsKmsSpec) error {
	t := types.NamespacedName{Name: awsKms.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]kvMapping{
		"access_key":    {Dest: awsKms.AccessKey, Mandatory: true},
		"secret_key":    {Dest: awsKms.SecretKey, Mandatory: true},
		"session_token": {Dest: awsKms.SessionToken},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "SealAwsKmsSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealAliCloudKMS(ctx context.Context, vaultServer *v1alpha1.VaultServer, aliCloudKms *v1alpha1.SealAliCloudKmsSpec) error {
	t := types.NamespacedName{Name: aliCloudKms.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]kvMapping{
		"access_key": {Dest: aliCloudKms.AccessKey, Mandatory: true},
		"secret_key": {Dest: aliCloudKms.SecretKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "SealAliCloudKmsSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealAzureKeyVault(ctx context.Context, vaultServer *v1alpha1.VaultServer, azureKeyVault *v1alpha1.SealAzureKeyVaultSpec) error {
	t := types.NamespacedName{Name: azureKeyVault.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]kvMapping{
		"client_id":     {Dest: azureKeyVault.ClientID, Mandatory: true},
		"client_secret": {Dest: azureKeyVault.ClientSecret, Mandatory: true},
		"tenant_id":     {Dest: azureKeyVault.TenantID, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "SealAzureKeyVaultSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealGCP(ctx context.Context, vaultServer *v1alpha1.VaultServer, azureKeyVault *v1alpha1.SealAzureKeyVaultSpec) error {
	t := types.NamespacedName{Name: azureKeyVault.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", secret.Name)
	}
	_, ok := secret.Data["credentials.json"]
	if !ok {
		return fmt.Errorf("%s does not contains key `credentials.json`", secret.Name)
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

func (r *VaultServerReconciler) sealPKCS11(ctx context.Context, vaultServer *v1alpha1.VaultServer, pkcs11 *v1alpha1.SealPKCS11Spec) error {
	t := types.NamespacedName{Name: pkcs11.PinSecret.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]kvMapping{
		"pin": {Dest: pkcs11.Pin, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "SealPKCS11Spec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}

func (r *VaultServerReconciler) sealTransit(ctx context.Context, vaultServer *v1alpha1.VaultServer, transit *v1alpha1.SealTransitSpec) error {
	t := types.NamespacedName{Name: transit.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	secretMappings := map[string]kvMapping{
		"token": {Dest: transit.Token, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "SealTransitSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
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
	secretMappings := map[string]kvMapping{
		"username": {Dest: aerospike.Username, Mandatory: true},
		"password": {Dest: aerospike.Password, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "StorageAerospikeSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
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
	secretMappings := map[string]kvMapping{
		"access_key": {Dest: aliCloudOss.AccessKey, Mandatory: true},
		"secret_key": {Dest: aliCloudOss.SecretKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "StorageAlicloudOssSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
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
	secretMappings := map[string]kvMapping{
		"accountKey": {Dest: azure.AccountKey, Mandatory: true},
	}
	for k, v := range secretMappings {
		if err := mapSecret(&secret, "StorageAzureSpec.Credentials", k, v.Dest, v.Mandatory); err != nil {
			return err
		}
	}
	return nil
}
