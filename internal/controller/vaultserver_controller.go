package controller

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	server "github.com/hashicorp/vault/command/server"
	"github.com/jynolen/vault-operator/api/v1alpha1"
	"github.com/jynolen/vault-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// typeAvailableMemcached represents the status of the Deployment reconciliation.
	typeAvailableVaultServer = "Available"
	// typeDegradedMemcached represents the status used when the custom resource is deleted and the finalizer operations are yet to occur.
	typeDegradedVaultServer = "Degraded"

	vaultOperatorFinalizer = "vault-operator.io/finalizer"
)

// VaultServerReconciler reconciles a VaultServer object.
type VaultServerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	logger   logr.Logger
}

// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=statefulset,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=service,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the VaultServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *VaultServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logf.FromContext(ctx)
	vaultServer := &v1alpha1.VaultServer{}
	err := r.Get(ctx, req.NamespacedName, vaultServer)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			r.logger.Info("vaultserver resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.logger.Error(err, "Failed to get vaultserver")
		return ctrl.Result{}, err
	}

	if len(vaultServer.Status.Conditions) == 0 {
		meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{Type: typeAvailableVaultServer, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, vaultServer); err != nil {
			r.logger.Error(err, "Failed to update VaultServer status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the vaultServer Custom Resource after updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		r.waitForStatus(ctx, vaultServer, req, metav1.ConditionUnknown, "Reconciling")
	}

	if !controllerutil.ContainsFinalizer(vaultServer, vaultOperatorFinalizer) {
		r.logger.Info("Adding Finalizer for VaultServer")

		if ok := controllerutil.AddFinalizer(vaultServer, vaultOperatorFinalizer); !ok {
			err = fmt.Errorf("finalizer for vaultServer was not added")
			r.logger.Error(err, "Failed to add finalizer for vaultServer")
			return ctrl.Result{}, err
		}

		if err = r.Update(ctx, vaultServer); err != nil {
			r.logger.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
		r.waitForFinalizer(ctx, vaultServer, req)
	}

	// Check if the VaultServer instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isVaultServerMarkedToBeDeleted := vaultServer.GetDeletionTimestamp() != nil
	if isVaultServerMarkedToBeDeleted {
		if !controllerutil.ContainsFinalizer(vaultServer, vaultOperatorFinalizer) {
			return ctrl.Result{}, nil
		}

		r.logger.Info("Performing Finalizer Operations for VaultServer before delete CR")

		// Let's add here a status "Downgrade" to reflect that this resource began its process to be terminated.
		meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{
			Type:   typeDegradedVaultServer,
			Status: metav1.ConditionUnknown, Reason: "Finalizing",
			Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", vaultServer.Name),
		})

		if err := r.Status().Update(ctx, vaultServer); err != nil {
			r.logger.Error(err, "Failed to update VaultServer status")
			return ctrl.Result{}, err
		}

		r.waitForStatus(ctx, vaultServer, req, metav1.ConditionUnknown, "Finalizing")

		// Perform all operations required before removing the finalizer and allow
		// the Kubernetes API to remove the custom resource.
		r.doFinalizerOperationsForVaultserver(vaultServer)

		// Re-fetch the vaultServer Custom Resource before updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		if err := r.Get(ctx, req.NamespacedName, vaultServer); err != nil {
			r.logger.Error(err, "Failed to re-fetch vaultServer")
			return ctrl.Result{}, err
		}

		meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{
			Type:   typeDegradedVaultServer,
			Status: metav1.ConditionTrue, Reason: "Finalizing",
			Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", vaultServer.Name),
		})

		if err := r.Status().Update(ctx, vaultServer); err != nil {
			r.logger.Error(err, "Failed to update VaultServer status")
			return ctrl.Result{}, err
		}

		r.waitForStatus(ctx, vaultServer, req, metav1.ConditionTrue, "Finalizing")

		r.logger.Info("Removing Finalizer for VaultServer after successfully perform the operations")
		if ok := controllerutil.RemoveFinalizer(vaultServer, vaultOperatorFinalizer); !ok {
			err = fmt.Errorf("finalizer for VaultServer was not removed")
			r.logger.Error(err, "Failed to remove finalizer for VaultServer")
			return ctrl.Result{}, err
		}

		if err := r.Update(ctx, vaultServer); err != nil {
			r.logger.Error(err, "Failed to remove finalizer for VaultServer")
			return ctrl.Result{}, err
		}
	}

	if err = r.reconcileSecret(ctx, vaultServer); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.reconcileStatefulSet(ctx, vaultServer); err != nil {
		return ctrl.Result{}, err
	}

	if err = r.reconcileService(ctx, vaultServer); err != nil {
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{
		Type:   typeAvailableVaultServer,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Resources for custom resource (%s) created successfully", vaultServer.Name),
	})

	if err := r.Status().Update(ctx, vaultServer); err != nil {
		vaultServer2 := &v1alpha1.VaultServer{}
		r.Get(ctx, req.NamespacedName, vaultServer2)
		r.logger.Error(err, "Failed to update VaultServer status")
		return ctrl.Result{}, err
	}

	r.waitForStatus(ctx, vaultServer, req, metav1.ConditionTrue, "Reconciling")
	return ctrl.Result{}, nil
}

func (r *VaultServerReconciler) waitForFinalizer(ctx context.Context, vaultServer *v1alpha1.VaultServer, req ctrl.Request) error {
	for {
		r.logger.Info("Re-fetching to update VaultServer Finalizer")
		if err := r.Get(ctx, req.NamespacedName, vaultServer); err != nil {
			r.logger.Error(err, "Failed to re-fetch vaultServer")
			return err
		}
		if len(vaultServer.ObjectMeta.Finalizers) == 1 && vaultServer.ObjectMeta.Finalizers[0] == vaultOperatorFinalizer {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (r *VaultServerReconciler) waitForStatus(ctx context.Context, vaultServer *v1alpha1.VaultServer, req ctrl.Request, condition metav1.ConditionStatus, reason string) error {
	for {
		r.logger.Info("Re-fetching to update VaultServer Status")
		if err := r.Get(ctx, req.NamespacedName, vaultServer); err != nil {
			r.logger.Error(err, "Failed to re-fetch vaultServer")
			return err
		}
		if len(vaultServer.Status.Conditions) == 0 {
			continue
		}
		for _, objectCondition := range vaultServer.Status.Conditions {
			if objectCondition.Status == condition && objectCondition.Reason == reason {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (r *VaultServerReconciler) reconcileService(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	log := r.logger.WithValues("Service.Namespace", vaultServer.GetNamespace(), "Service.Name", vaultServer.Name)
	svc := r.serviceForVaultServer(vaultServer)
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: vaultServer.Name, Namespace: vaultServer.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new service")
		if err = r.Create(ctx, svc); err != nil {
			log.Error(err, "Failed to create new service")
			return err
		}
		return nil
	}
	log.Info("Updating service")
	found.Spec = svc.Spec
	if err = r.Update(ctx, found); err != nil {
		log.Error(err, "Failed to update service")
		return err
	}

	return nil
}

func (r *VaultServerReconciler) serviceForVaultServer(vaultServer *v1alpha1.VaultServer) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vaultServer.Name,
			Namespace: vaultServer.Namespace,
			Labels:    vaultServer.Labels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: vaultServer.Labels(),
			Type:     vaultServer.Spec.Service.Type,
			Ports: []corev1.ServicePort{
				{
					Name:       "api",
					Port:       8200,
					TargetPort: intstr.FromInt(8200),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "cluster",
					Port:       8201,
					TargetPort: intstr.FromInt(8201),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func (r *VaultServerReconciler) reconcileStatefulSet(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	log := r.logger.WithValues("StatefulSet.Namespace", vaultServer.GetNamespace(), "StatefulSet.Name", vaultServer.Name)

	sts, err := r.statefulSetForVaultServer(ctx, vaultServer)
	if err != nil {
		log.Error(err, "Failed to template vault config, template error")
		return err
	}
	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: vaultServer.Name, Namespace: vaultServer.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new statefulset")
		if err = r.Create(ctx, sts); err != nil {
			log.Error(err, "Failed to create new statefulset")
			return err
		}
		return nil
	}

	if r.isStsDiverging(sts, found) {
		log.Info("Generated StatefulSet diverge, updating the StatefulSet")

		maps.Copy(found.Labels, sts.Labels)
		found.Spec = sts.Spec
		if err = r.Update(ctx, found); err != nil {
			log.Error(err, "Failed to update statefulset")
			return err
		}
		return nil
	}
	return nil
}

func (r *VaultServerReconciler) isStsDiverging(new, old *appsv1.StatefulSet) bool {
	if diff := cmp.Diff(new.Spec, old.Spec); diff != "" {
		r.logger.V(8).Info(diff)
		return true
	}
	if diff := cmp.Diff(new.ObjectMeta.Annotations, old.ObjectMeta.Annotations); diff != "" {
		r.logger.V(8).Info(diff)
		return true
	}
	if diff := cmp.Diff(new.ObjectMeta.Labels, old.ObjectMeta.Labels); diff != "" {
		r.logger.V(8).Info(diff)
		return true
	}
	return false
}

func configChecksumAnnotationName(obj metav1.TypeMeta) string {
	return fmt.Sprintf("%s/vault-config-checksum", utils.GetFullCrdFQDN(&obj))
}

func (r *VaultServerReconciler) secretForVaultServer(vaultServer *v1alpha1.VaultServer) (*corev1.Secret, error) {
	secretData := make(map[string][]byte)
	secretAnnotations := make(map[string]string)
	re := regexp.MustCompile(`\n\s*\n`)
	renderedConfig, err := vaultServer.HclRender()
	if err != nil {
		return nil, err
	}
	config := re.ReplaceAllString(renderedConfig, "\n")
	if _, err := server.ParseConfig(config, ""); err != nil {
		r.logger.V(8).Info(fmt.Sprintf("Generated Vault-Config value %s\n", config))
		r.logger.Error(err, "")
		return nil, err
	}

	if vaultServer.Spec.SecretOverride != nil {
		maps.Copy(secretData, vaultServer.Spec.SecretOverride.Data)
	}
	secretData["vault.hcl"] = ([]byte)(config)
	secretAnnotations[configChecksumAnnotationName(vaultServer.TypeMeta)] = vaultServer.ConfigHash()
	cm := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        vaultServer.GetSecretNameForVaultConfig(),
			Namespace:   vaultServer.GetNamespace(),
			Labels:      vaultServer.GetLabels(),
			Annotations: secretAnnotations,
		},
		Data: secretData,
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(vaultServer, cm, r.Scheme); err != nil {
		return nil, err
	}
	return cm, nil
}

func (r *VaultServerReconciler) reconcileSecret(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	log := r.logger.WithValues("Secret.Namespace", vaultServer.GetNamespace(), "Secret.Name", vaultServer.GetSecretNameForVaultConfig())

	if err := vaultServer.ResolveSecret(&r.Client, ctx); err != nil {
		return err
	}
	secret, err := r.secretForVaultServer(vaultServer)
	if err != nil {
		log.Error(err, "Failed to template vault config, template error")
		return err
	}
	found := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: vaultServer.GetSecretNameForVaultConfig(), Namespace: vaultServer.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new secret")
		if err = r.Create(ctx, secret); err != nil {
			log.Error(err, "Failed to create new secret")
			return err
		}
		return nil
	}
	log.Info(fmt.Sprintf("Generated Vault-Config value %s\n", secret.Data[vaultServer.ServerConfigFileName()]))

	if found.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)] != secret.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)] {
		log.Info("Generated Vault-Config file diverge, updating the Secret")

		maps.Copy(found.Labels, secret.Labels)
		found.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)] = secret.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)]
		found.Data = secret.Data

		if err = r.Update(ctx, found); err != nil {
			log.Error(err, "Failed to create new secret")
			return err
		}
		return nil
	}
	return nil
}

// finalizeMemcached will perform the required operations before delete the CR.
func (r *VaultServerReconciler) doFinalizerOperationsForVaultserver(cr *v1alpha1.VaultServer) {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	// Note: It is not recommended to use finalizers with the purpose of deleting resources which are
	// created and managed in the reconciliation. These ones, such as the Deployment created on this reconcile,
	// are defined as dependent of the custom resource. See that we use the method ctrl.SetControllerReference.
	// to set the ownerRef which means that the Deployment will be deleted by the Kubernetes API.
	// More info: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/

	// The following implementation will raise an event
	r.Recorder.Event(cr, "Warning", "Deleting",
		fmt.Sprintf("Custom Resource %s is being deleted from the namespace %s",
			cr.Name,
			cr.Namespace))
}

// statefulSetForVaultServer returns a VaultServer Deployment object.
func (r *VaultServerReconciler) statefulSetForVaultServer(
	ctx context.Context,
	vaultServer *v1alpha1.VaultServer,
) (*appsv1.StatefulSet, error) {
	replicas := vaultServer.Spec.Replicas
	addressPort, err := strconv.ParseInt(vaultServer.Spec.Config.ListenerTcp.Address().Port(), 10, 32)
	if err != nil {
		return nil, err
	}
	clusterPort, err := strconv.ParseInt(vaultServer.Spec.Config.ListenerTcp.ClusterAddress().Port(), 10, 32)
	if err != nil {
		return nil, err
	}
	stsVolumes, stsVolumesMount, err := vaultServer.Volumes(&r.Client, ctx)
	if err != nil {
		return nil, err
	}
	partition := int32(0)

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vaultServer.Name,
			Namespace: vaultServer.Namespace,
			Labels:    vaultServer.Spec.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:             &replicas,
			PodManagementPolicy:  appsv1.OrderedReadyPodManagement,
			RevisionHistoryLimit: &vaultServer.Spec.RevisionHistoryLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: vaultServer.Labels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: vaultServer.Labels(),
					Annotations: map[string]string{
						configChecksumAnnotationName(vaultServer.TypeMeta): vaultServer.ConfigHash(),
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
					},
					Containers: []corev1.Container{{
						Image:           vaultServer.Spec.Image,
						Name:            "vault-server",
						ImagePullPolicy: corev1.PullIfNotPresent,
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             ptr.To(true),
							RunAsUser:                ptr.To(int64(1000)),
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities: &corev1.Capabilities{
								Add:  []corev1.Capability{"IPC_LOCK"},
								Drop: []corev1.Capability{"ALL"},
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(addressPort),
							Name:          "api-port",
							Protocol:      corev1.ProtocolTCP,
						}, {
							ContainerPort: int32(clusterPort),
							Name:          "cluster-port",
							Protocol:      corev1.ProtocolTCP,
						}},
						LivenessProbe: &corev1.Probe{
							InitialDelaySeconds: 10,
							SuccessThreshold:    1,
							FailureThreshold:    3,
							PeriodSeconds:       10,
							TimeoutSeconds:      1,
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Port:   intstr.FromInt(int(addressPort)),
									Path:   "/v1/sys/health?standbyok=true&sealedcode=201&uninitcode=201",
									Scheme: corev1.URIScheme(vaultServer.Spec.Config.ListenerTcp.Scheme()),
								},
							},
						},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: "File",
						Command:                  []string{"vault", "server", "--config", vaultServer.ServerConfigFilePath()},
						VolumeMounts:             stsVolumesMount,
					}},
					Volumes:                       stsVolumes,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &vaultServer.Spec.TerminationGracePeriodSeconds,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
				},
			},
			VolumeClaimTemplates: vaultServer.Spec.PersistentVolumeClaim,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: &partition,
				},
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(vaultServer, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VaultServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VaultServer{}).
		Named("vaultserver").
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
