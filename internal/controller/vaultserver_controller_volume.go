package controller

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jynolen/vault-operator/api/v1alpha1"
	"github.com/zeebo/xxh3"
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

func (r *VaultServerReconciler) storageRaftLeaderTlsVolume(ctx context.Context, vaultServer *v1alpha1.VaultServer) ([]*corev1.Volume, []*corev1.VolumeMount, error) {
	mounts, volumes := []*corev1.VolumeMount{}, []*corev1.Volume{}

	if vaultServer.Spec.Config.Storage.Raft == nil {
		return volumes, mounts, nil
	}
	for _, retry := range vaultServer.Spec.Config.Storage.Raft.RetryJoin {
		if retry.LeaderTls == nil {
			continue
		}
		var secret corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Name: retry.LeaderTls.SecretRef.Name, Namespace: vaultServer.Namespace}, &secret); err != nil {
			return nil, nil, err
		}
		if secret.Type != corev1.SecretTypeTLS {
			return nil, nil, errors.New("ListenerTCP.TLS SecretType is not a kubernetes.io/tls")
		}
		hash := string(xxh3.HashString(*retry.LeaderApiAddr))
		volumes = append(volumes, &corev1.Volume{
			Name: fmt.Sprintf("storage-raft-leader-tls-%s", hash),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: retry.LeaderTls.SecretRef.Name,
				},
			},
		})
		mounts = append(mounts, &corev1.VolumeMount{
			Name:      fmt.Sprintf("storage-raft-leader-tls-%s", hash),
			ReadOnly:  true,
			MountPath: fmt.Sprintf("/raft/retry/%x", hash),
		})
	}
	return volumes, mounts, nil
}

func (r *VaultServerReconciler) sealTransitTcpTlsVolume(ctx context.Context, vaultServer *v1alpha1.VaultServer, transit *v1alpha1.TlsTransitSpec) ([]*corev1.Volume, []*corev1.VolumeMount, error) {
	mounts, volumes := []*corev1.VolumeMount{}, []*corev1.Volume{}

	if transit.ClientCert != nil {
		var secret corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Name: transit.ClientCert.SecretRef.Name, Namespace: vaultServer.Namespace}, &secret); err != nil {
			return nil, nil, err
		}
		if secret.Type != corev1.SecretTypeTLS {
			return nil, nil, errors.New("ListenerTCP.TLS SecretType is not a kubernetes.io/tls")
		}
		volumes = append(volumes, &corev1.Volume{
			Name: "seal-transit-client-tls",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: vaultServer.Spec.Config.ListenerTcp.TLS.Cert.SecretRef.Name,
				},
			},
		})
		mounts = append(mounts, &corev1.VolumeMount{
			Name:      "seal-transit-client-tls",
			ReadOnly:  true,
			MountPath: "/seal/transit",
		})
	}

	if transit.CaCert != nil {
		var cm corev1.ConfigMap
		if err := r.Get(ctx, types.NamespacedName{Name: transit.CaCert.ConfigMapRef.Name, Namespace: vaultServer.Namespace}, &cm); err != nil {
			return nil, nil, err
		}

		if _, ok := cm.Data["ca.crt"]; !ok {
			return nil, nil, fmt.Errorf("ConfigMap `%s` does not contains key `ca.crt", cm.Name)
		}
		volumes = append(volumes, &corev1.Volume{
			Name: "seal-transit-ca-tls",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: cm.Name},
				},
			},
		})
		mounts = append(mounts, &corev1.VolumeMount{
			Name:      "seal-transit-client-tls",
			ReadOnly:  true,
			MountPath: "/seal/transit/ca.crt",
			SubPath:   "ca.crt",
		})
	}
	return volumes, mounts, nil
}
