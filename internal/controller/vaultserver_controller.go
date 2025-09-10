package controller

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"maps"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/jynolen/vault-operator/api/v1alpha1"
	"github.com/jynolen/vault-operator/internal/utils"
	"github.com/zeebo/xxh3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// typeAvailableMemcached represents the status of the Deployment reconciliation
	typeAvailableVaultServer = "Available"
	// typeDegradedMemcached represents the status used when the custom resource is deleted and the finalizer operations are yet to occur.
	typeDegradedVaultServer = "Degraded"

	vaultOperatorFinalizer = "vault-operator.io/finalizer"
)

//go:embed template/config.gotpl
var vaultConfigTemplateStr string

var VaultConfigTemplate *template.Template = template.Must(template.New("configMapGenerator").Funcs(sprig.FuncMap()).Parse(vaultConfigTemplateStr))

// VaultServerReconciler reconciles a VaultServer object
type VaultServerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault-operator.io,resources=vaultservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=statefulset,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *VaultServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	vaultServer := &v1alpha1.VaultServer{}
	err := r.Get(ctx, req.NamespacedName, vaultServer)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("vaultserver resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get vaultserver")
		return ctrl.Result{}, err
	}

	if len(vaultServer.Status.Conditions) == 0 {
		meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{Type: typeAvailableVaultServer, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, vaultServer); err != nil {
			log.Error(err, "Failed to update VaultServer status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the vaultServer Custom Resource after updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, vaultServer); err != nil {
			log.Error(err, "Failed to re-fetch vaultServer")
			return ctrl.Result{}, err
		}
	}

	if !controllerutil.ContainsFinalizer(vaultServer, vaultOperatorFinalizer) {
		log.Info("Adding Finalizer for VautlServer")
		if ok := controllerutil.AddFinalizer(vaultServer, vaultOperatorFinalizer); !ok {
			err = fmt.Errorf("finalizer for vaultServer was not added")
			log.Error(err, "Failed to add finalizer for vaultServer")
			return ctrl.Result{}, err
		}

		if err = r.Update(ctx, vaultServer); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	// Check if the VaultServer instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isVaultServerMarkedToBeDeleted := vaultServer.GetDeletionTimestamp() != nil
	if isVaultServerMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(vaultServer, vaultOperatorFinalizer) {
			log.Info("Performing Finalizer Operations for VaultServer before delete CR")

			// Let's add here a status "Downgrade" to reflect that this resource began its process to be terminated.
			meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{Type: typeDegradedVaultServer,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", vaultServer.Name)})

			if err := r.Status().Update(ctx, vaultServer); err != nil {
				log.Error(err, "Failed to update VaultServer status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before removing the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			r.doFinalizerOperationsForVaultserver(vaultServer)

			// TODO(user): If you add operations to the doFinalizerOperationsForMemcached method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the vaultServer Custom Resource before updating the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raising the error "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, vaultServer); err != nil {
				log.Error(err, "Failed to re-fetch vaultServer")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{Type: typeDegradedVaultServer,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", vaultServer.Name)})

			if err := r.Status().Update(ctx, vaultServer); err != nil {
				log.Error(err, "Failed to update VaultServer status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for VaultServer after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(vaultServer, typeDegradedVaultServer); !ok {
				err = fmt.Errorf("finalizer for VaultServer was not removed")
				log.Error(err, "Failed to remove finalizer for VaultServer")
				return ctrl.Result{}, err
			}

			if err := r.Update(ctx, vaultServer); err != nil {
				log.Error(err, "Failed to remove finalizer for VaultServer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Check if the configMap already exists, if not create a new one
	if err = r.reconcileConfigMap(ctx, vaultServer); err != nil {
		return ctrl.Result{}, err
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
	meta.SetStatusCondition(&vaultServer.Status.Conditions, metav1.Condition{Type: typeAvailableVaultServer,
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Resources for custom resource (%s) created successfully", vaultServer.Name)})

	if err := r.Status().Update(ctx, vaultServer); err != nil {
		log.Error(err, "Failed to update VaultServer status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *VaultServerReconciler) reconcileService(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	return nil
}

func (r *VaultServerReconciler) reconcileStatefulSet(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	return nil
}

func (r *VaultServerReconciler) reconcileConfigMap(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
	log := logf.FromContext(ctx, "ConfigMap.Namespace", vaultServer.GetNamespace(), "ConfigMap.Name", vaultServer.GetConfigMapNameForVaultConfig())
	cm, err := r.configMapForVaultServer(vaultServer)
	if err != nil {
		log.Error(err, "Failed to template vault config, template error")
		return err
	}

	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: vaultServer.Name, Namespace: vaultServer.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new ConfigMap")
		if err = r.Create(ctx, cm); err != nil {
			log.Error(err, "Failed to create new configMap")
			return err
		}
		return nil
	}
	if found.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)] != cm.Annotations[configChecksumAnnotationName(vaultServer.TypeMeta)] {
		log.Info("Generated Vault-Config file diverge, updating the ConfigMap")
		if err = r.Update(ctx, cm); err != nil {
			log.Error(err, "Failed to create new configMap")
			return err
		}
		return nil
	}
	return nil
}

func configChecksumAnnotationName(obj metav1.TypeMeta) string {
	return fmt.Sprintf("%s/vault-config-checksum", utils.GetFullCrdFQDN(&obj))
}

func (r *VaultServerReconciler) configMapForVaultServer(vaultServer *v1alpha1.VaultServer) (*corev1.ConfigMap, error) {
	var buf bytes.Buffer
	cmData := make(map[string]string)
	cmAnnotations := make(map[string]string)
	if err := VaultConfigTemplate.Execute(&buf, vaultServer.Spec.Config); err != nil {
		return nil, err
	}

	cmData["vault.hcl"] = buf.String()
	if vaultServer.Spec.ConfigMapOverride != nil {
		maps.Copy(cmData, vaultServer.Spec.ConfigMapOverride.Data)
	}
	cmAnnotations[configChecksumAnnotationName(vaultServer.TypeMeta)] = fmt.Sprintf("0x%x", xxh3.HashString(cmData["vault.hcl"]))
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        vaultServer.GetConfigMapNameForVaultConfig(),
			Namespace:   vaultServer.GetNamespace(),
			Labels:      vaultServer.GetLabels(),
			Annotations: cmAnnotations,
		},
		Data: cmData,
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(vaultServer, cm, r.Scheme); err != nil {
		return nil, err
	}
	return cm, nil
}

func (r *VaultServerReconciler) reconcileSecret(ctx context.Context, vaultServer *v1alpha1.VaultServer) error {
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

// statefulSetForVaultServer returns a VaultServer Deployment object
func (r *VaultServerReconciler) statefulSetForVaultServer(
	vaultServer *v1alpha1.VaultServer) (*appsv1.StatefulSet, error) {
	replicas := vaultServer.Spec.Size
	containerPort, err := strconv.ParseInt(vaultServer.Spec.Config.ClusterAddr.Port(), 10, 32)
	if err != nil {
		return nil, err
	}
	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vaultServer.Name,
			Namespace: vaultServer.Namespace,
			Labels:    vaultServer.Spec.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: vaultServer.Spec.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: vaultServer.Spec.Labels,
				},
				Spec: corev1.PodSpec{
					// TODO(user): Uncomment the following code to configure the nodeAffinity expression
					// according to the platforms which are supported by your solution. It is considered
					// best practice to support multiple architectures. build your manager image using the
					// makefile target docker-buildx. Also, you can use docker manifest inspect <image>
					// to check what are the platforms supported.
					// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
					// Affinity: &corev1.Affinity{
					//	 NodeAffinity: &corev1.NodeAffinity{
					//		 RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					//			 NodeSelectorTerms: []corev1.NodeSelectorTerm{
					//				 {
					//					 MatchExpressions: []corev1.NodeSelectorRequirement{
					//						 {
					//							 Key:      "kubernetes.io/arch",
					//							 Operator: "In",
					//							 Values:   []string{"amd64", "arm64", "ppc64le", "s390x"},
					//						 },
					//						 {
					//							 Key:      "kubernetes.io/os",
					//							 Operator: "In",
					//							 Values:   []string{"linux"},
					//						 },
					//					 },
					//				 },
					//		 	 },
					//		 },
					//	 },
					// },
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						// The vaultServer image does not use a non-zero numeric user as the default user.
						// Due to RunAsNonRoot field being set to true, we need to force the user in the
						// container to a non-zero numeric user. We do this using the RunAsUser field.
						// However, if you are looking to provide solution for K8s vendors like OpenShift
						// be aware that you cannot run under its restricted-v2 SCC if you set this value.
						// IMPORTANT: seccomProfile was introduced with Kubernetes 1.19
						// If you are looking for to produce solutions to be supported
						// on lower versions you must remove this option.
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Image:           vaultServer.Spec.Image,
						Name:            "vaultServer",
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
							// "RunAsUser" fields empty.
							RunAsNonRoot:             ptr.To(true),
							RunAsUser:                ptr.To(int64(1001)),
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(containerPort),
							Name:          "vaultServer",
						}},
						Command: []string{"vaultServer", "-m=64", "-o", "modern", "-v"},
					}},
				},
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
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
