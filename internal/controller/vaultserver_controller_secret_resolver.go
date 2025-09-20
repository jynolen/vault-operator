package controller

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jynolen/vault-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *VaultServerReconciler) resolveSecret(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	if vaultServer.Spec.Config.Telemetry != nil && vaultServer.Spec.Config.Telemetry.Circonus != nil {
		r.telemetryCirconus(ctx, vaultServer)
	}
	return nil
}

func (r *VaultServerReconciler) telemetryCirconus(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	t := types.NamespacedName{Name: vaultServer.Spec.Config.Telemetry.Circonus.Credentials.SecretRef.Name, Namespace: vaultServer.Namespace}
	var secret corev1.Secret
	if err := r.Get(ctx, t, &secret); err != nil {
		return err
	}
	if secret.Type != corev1.SecretTypeOpaque {
		return errors.New("TelemetryCirconusCredentials is not Opaque")
	}
	val, ok := secret.Data["api_token"]
	if !ok {
		return errors.New("TelemetryCirconusCredentials does not contains key `api_token`")
	}
	strVal := string(val)
	vaultServer.Spec.Config.Telemetry.Circonus.ApiToken = &strVal

	val, ok = secret.Data["api_app"]
	if val, ok = secret.Data["api_app"]; ok {
		strVal = string(val)
		vaultServer.Spec.Config.Telemetry.Circonus.ApiToken = &strVal
	}
	return nil
}
