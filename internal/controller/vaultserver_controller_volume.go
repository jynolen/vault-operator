package controller

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jynolen/vault-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *VaultServerReconciler) listenerTcpTlsVolume(ctx context.Context, vaultServer *v1alpha1.VaultServer) (*corev1.Volume, *corev1.VolumeMount, error) {
	if vaultServer.Spec.Config.ListenerTcp.TLS.Cert == nil {
		return nil, nil, nil
	}
	var secret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: vaultServer.Spec.Config.ListenerTcp.TLS.Cert.SecretRef.Name, Namespace: vaultServer.Namespace}, &secret); err != nil {
		return nil, nil, err
	}
	if secret.Type != corev1.SecretTypeTLS {
		return nil, nil, errors.New("ListenerTCP.TLS SecretType is not a kubernetes.io/tls")
	}
	volume := &corev1.Volume{
		Name: "listenertcp-tls",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: vaultServer.Spec.Config.ListenerTcp.TLS.Cert.SecretRef.Name,
			},
		},
	}
	volumeMount := &corev1.VolumeMount{
		Name:      "listenertcp-tls",
		ReadOnly:  true,
		MountPath: "/tls",
	}
	return volume, volumeMount, nil
}
