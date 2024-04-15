/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"net/url"

	csv1alpha2 "github.com/walnuts1018/code-server-operator/api/v1alpha2"
	"github.com/walnuts1018/code-server-operator/internal/initplugins"
	initpluginsCommon "github.com/walnuts1018/code-server-operator/internal/initplugins/common"
	"github.com/walnuts1018/code-server-operator/util/random"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	networkingv1apply "k8s.io/client-go/applyconfigurations/networking/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CodeServer        = "code-server"
	CodeServerManager = "code-server-operator"
	MaxActiveSeconds  = 60 * 60 * 24
	MaxKeepSeconds    = 60 * 60 * 24 * 30
)

// CodeServerReconciler reconciles a CodeServer object
type CodeServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeservers/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *CodeServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var codeServer csv1alpha2.CodeServer

	// Fetch the CodeServer instance
	err := r.Client.Get(ctx, req.NamespacedName, &codeServer)
	if errors.IsNotFound(err) {
		logger.Info("CodeServer has been deleted. Trying to delete its related resources.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get CoderServer.", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, err
	}
	// Check if the CodeServer instance is marked for deletion
	if !codeServer.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if err := r.reconcileSecret(ctx, codeServer); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcilePVC(ctx, codeServer); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, codeServer); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, codeServer); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileIngress(ctx, codeServer); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CodeServerReconciler) reconcileSecret(ctx context.Context, codeServer csv1alpha2.CodeServer) error {
	logger := log.FromContext(ctx)

	secret := &corev1.Secret{}
	secret.SetName(codeServer.Name)
	secret.SetNamespace(codeServer.Namespace)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		if _, ok := secret.Data["password"]; !ok {
			pass, err := random.String(16, random.Alphanumeric)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}
			secret.Data["password"] = []byte(pass)
		}
		return ctrl.SetControllerReference(&codeServer, secret, r.Scheme)
	})

	if err != nil {
		return fmt.Errorf("failed to reconcile secret: %w", err)
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("Secret has been reconciled.", "name", codeServer.Name, "namespace", codeServer.Namespace)
	}

	return nil
}

func (r *CodeServerReconciler) reconcilePVC(ctx context.Context, codeServer csv1alpha2.CodeServer) error {
	logger := log.FromContext(ctx)

	pvc := &corev1.PersistentVolumeClaim{}
	pvc.SetName(codeServer.Name)
	pvc.SetNamespace(codeServer.Namespace)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if pvc.Labels == nil {
			pvc.Labels = make(map[string]string)
		}
		pvc.Labels["app.kubernetes.io/name"] = CodeServer
		pvc.Labels["app.kubernetes.io/instance"] = codeServer.Name
		pvc.Labels["app.kubernetes.io/created-by"] = CodeServerManager

		if pvc.Annotations == nil {
			pvc.Annotations = make(map[string]string)
		}
		for k, v := range codeServer.Annotations {
			pvc.Annotations[k] = v
		}

		if pvc.Spec.AccessModes == nil {
			pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		}

		if pvc.Spec.Resources.Requests == nil {
			pvc.Spec.Resources.Requests = make(corev1.ResourceList)
		}

		storageQuontity, err := resource.ParseQuantity(codeServer.Spec.StorageSize)
		if err != nil {
			return fmt.Errorf("failed to parse storage size: %w", err)
		}

		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageQuontity

		if codeServer.Spec.StorageClassName != "" {
			pvc.Spec.StorageClassName = &codeServer.Spec.StorageClassName
		}

		if codeServer.Spec.VolumeName != "" {
			pvc.Spec.VolumeName = codeServer.Spec.VolumeName
		}

		return ctrl.SetControllerReference(&codeServer, pvc, r.Scheme)
	})

	if err != nil {
		return fmt.Errorf("failed to reconcile PVC: %w", err)
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("PVC has been reconciled.", "name", codeServer.Name, "namespace", codeServer.Namespace)
	}

	return nil
}

func (r *CodeServerReconciler) reconcileDeployment(ctx context.Context, codeServer csv1alpha2.CodeServer) error {
	logger := log.FromContext(ctx)

	owner, err := controllerReference(codeServer, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to create controller reference: %w", err)
	}

	const volumeName = "home"
	initContainers, err := initplugins.CreatePlugin(codeServer.Spec.InitPlugins, initpluginsCommon.CommonFields{
		Image:      codeServer.Spec.Image,
		VolumeName: volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to create init plugins: %w", err)
	}

	envs := make([]*corev1apply.EnvVarApplyConfiguration, 0, len(codeServer.Spec.Envs))

	envs = append(envs, corev1apply.EnvVar().
		WithName("PASSWORD").
		WithValueFrom(corev1apply.EnvVarSource().
			WithSecretKeyRef(corev1apply.SecretKeySelector().
				WithName(codeServer.Name).
				WithKey("password"),
			),
		),
	)
	for _, env := range codeServer.Spec.Envs {
		envs = append(envs, corev1apply.EnvVar().
			WithName(env.Name).
			WithValue(env.Value),
		)
	}

	resourceRequirements := corev1apply.ResourceRequirements().
		WithLimits(corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		})
	if codeServer.Spec.Resources.Limits != nil {
		cpu, ok := codeServer.Spec.Resources.Limits[corev1.ResourceCPU]
		if ok {
			resourceRequirements = resourceRequirements.WithLimits(corev1.ResourceList{
				corev1.ResourceCPU: cpu,
			})
		}
		memory, ok := codeServer.Spec.Resources.Limits[corev1.ResourceMemory]
		if ok {
			resourceRequirements = resourceRequirements.WithLimits(corev1.ResourceList{
				corev1.ResourceMemory: memory,
			})
		}
	}
	if codeServer.Spec.Resources.Requests != nil {
		cpu, ok := codeServer.Spec.Resources.Requests[corev1.ResourceCPU]
		if ok {
			resourceRequirements = resourceRequirements.WithRequests(corev1.ResourceList{
				corev1.ResourceCPU: cpu,
			})
		}
		memory, ok := codeServer.Spec.Resources.Requests[corev1.ResourceMemory]
		if ok {
			resourceRequirements = resourceRequirements.WithRequests(corev1.ResourceList{
				corev1.ResourceMemory: memory,
			})
		}
	}

	imagePullSecrets := make([]*corev1apply.LocalObjectReferenceApplyConfiguration, 0, len(codeServer.Spec.ImagePullSecrets))
	for _, secret := range codeServer.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, corev1apply.LocalObjectReference().
			WithName(secret.Name),
		)
	}

	command := fmt.Sprintf("%s && /usr/bin/entrypoint.sh --bind-addr 0.0.0.0:%d", codeServer.Spec.InitCommand, codeServer.Spec.ContainerPort)
	if _, ok := codeServer.Spec.InitPlugins["git"]; ok {
		command = fmt.Sprintf("%s /home/coder/work", command)
	}
	deployment := appsv1apply.Deployment(codeServer.Name, codeServer.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       CodeServer,
			"app.kubernetes.io/instance":   codeServer.Name,
			"app.kubernetes.io/created-by": CodeServerManager,
		}).
		WithOwnerReferences(owner).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(1).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{
				"app.kubernetes.io/name":       CodeServer,
				"app.kubernetes.io/instance":   codeServer.Name,
				"app.kubernetes.io/created-by": CodeServerManager,
			})).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(map[string]string{
					"app.kubernetes.io/name":       CodeServer,
					"app.kubernetes.io/instance":   codeServer.Name,
					"app.kubernetes.io/created-by": CodeServerManager,
				}).
				WithSpec(corev1apply.PodSpec().
					WithSecurityContext(corev1apply.PodSecurityContext().
						WithFSGroup(1000).
						WithFSGroupChangePolicy(corev1.FSGroupChangeOnRootMismatch).
						WithRunAsUser(1000).
						WithRunAsGroup(1000),
					).
					WithImagePullSecrets(imagePullSecrets...).
					WithInitContainers(initContainers...).
					WithContainers(corev1apply.Container().
						WithName(CodeServer).
						WithImage(codeServer.Spec.Image).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPorts(corev1apply.ContainerPort().
							WithName("http").
							WithProtocol(corev1.ProtocolTCP).
							WithContainerPort(codeServer.Spec.ContainerPort),
						).
						WithEnv(envs...).
						WithVolumeMounts(corev1apply.VolumeMount().
							WithName(volumeName).
							WithMountPath("/home/coder"),
						).
						WithResources(resourceRequirements).
						WithCommand(
							"/bin/sh",
							"-c",
							command,
						),
					).
					WithVolumes(corev1apply.Volume().
						WithName(volumeName).
						WithPersistentVolumeClaim(corev1apply.PersistentVolumeClaimVolumeSource().
							WithClaimName(codeServer.Name),
						),
					).
					WithNodeSelector(codeServer.Spec.NodeSelector),
				),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployment)
	if err != nil {
		return fmt.Errorf("failed to convert deployment to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: codeServer.Namespace, Name: codeServer.Name}, &current)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	currentApplyConfig, err := appsv1apply.ExtractDeployment(&current, CodeServerManager)
	if err != nil {
		return fmt.Errorf("failed to extract apply configuration from deployment: %w", err)
	}

	if equality.Semantic.DeepEqual(deployment, currentApplyConfig) {
		return nil
	}

	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{FieldManager: CodeServerManager, Force: ptr.To(true)}); err != nil {
		return fmt.Errorf("failed to apply deployment: %w", err)
	}

	logger.Info("Deployment has been reconciled.", "name", codeServer.Name, "namespace", codeServer.Namespace)

	return nil
}

func (r *CodeServerReconciler) reconcileService(ctx context.Context, codeServer csv1alpha2.CodeServer) error {
	logger := log.FromContext(ctx)

	owner, err := controllerReference(codeServer, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to create controller reference: %w", err)
	}

	ports := []*corev1apply.ServicePortApplyConfiguration{
		corev1apply.ServicePort().
			WithName("http").
			WithProtocol(corev1.ProtocolTCP).
			WithPort(codeServer.Spec.ContainerPort).
			WithTargetPort(intstr.FromInt32(codeServer.Spec.ContainerPort)),
	}

	for _, port := range codeServer.Spec.PublicProxyPorts {
		ports = append(ports, corev1apply.ServicePort().
			WithName(fmt.Sprintf("http-%d", port)).
			WithProtocol(corev1.ProtocolTCP).
			WithPort(port).
			WithTargetPort(intstr.FromInt32(port)),
		)
	}

	service := corev1apply.Service(codeServer.Name, codeServer.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       CodeServer,
			"app.kubernetes.io/instance":   codeServer.Name,
			"app.kubernetes.io/created-by": CodeServerManager,
		}).
		WithOwnerReferences(owner).
		WithSpec(corev1apply.ServiceSpec().
			WithType(corev1.ServiceTypeClusterIP).
			WithPorts(ports...,
			).
			WithSelector(map[string]string{
				"app.kubernetes.io/name":       CodeServer,
				"app.kubernetes.io/instance":   codeServer.Name,
				"app.kubernetes.io/created-by": CodeServerManager,
			}),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(service)
	if err != nil {
		return fmt.Errorf("failed to convert service to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current corev1.Service
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: codeServer.Namespace, Name: codeServer.Name}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get service: %w", err)
	}

	currentApplyConfig, err := corev1apply.ExtractService(&current, CodeServerManager)
	if err != nil {
		return fmt.Errorf("failed to extract apply configuration from service: %w", err)
	}

	if equality.Semantic.DeepEqual(service, currentApplyConfig) {
		return nil
	}

	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{FieldManager: CodeServerManager, Force: ptr.To(true)}); err != nil {
		return fmt.Errorf("failed to apply service: %w", err)
	}

	logger.Info("Service has been reconciled.", "name", codeServer.Name, "namespace", codeServer.Namespace)

	return nil
}

func (r *CodeServerReconciler) reconcileIngress(ctx context.Context, codeServer csv1alpha2.CodeServer) error {
	logger := log.FromContext(ctx)

	owner, err := controllerReference(codeServer, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to create controller reference: %w", err)
	}

	url, err := url.Parse(codeServer.Spec.Domain)
	if err != nil {
		return fmt.Errorf("failed to parse domain: %w", err)
	}
	host := fmt.Sprintf("%s.%s", codeServer.Name, url.String())

	paths := []*networkingv1apply.HTTPIngressPathApplyConfiguration{networkingv1apply.HTTPIngressPath().
		WithPath("/").
		WithPathType(networkingv1.PathTypePrefix).
		WithBackend(networkingv1apply.IngressBackend().
			WithService(networkingv1apply.IngressServiceBackend().
				WithName(codeServer.Name).
				WithPort(networkingv1apply.ServiceBackendPort().
					WithName("http"),
				),
			),
		),
	}

	for _, port := range codeServer.Spec.PublicProxyPorts {
		paths = append(paths, networkingv1apply.HTTPIngressPath().
			WithPath(fmt.Sprintf("/proxy/%d", port)).
			WithPathType(networkingv1.PathTypePrefix).
			WithBackend(networkingv1apply.IngressBackend().
				WithService(networkingv1apply.IngressServiceBackend().
					WithName(codeServer.Name).
					WithPort(networkingv1apply.ServiceBackendPort().
						WithName(fmt.Sprintf("http-%d", port)),
					),
				),
			),
		)
	}

	spec := networkingv1apply.IngressSpec().
		WithRules(networkingv1apply.IngressRule().
			WithHost(host).
			WithHTTP(networkingv1apply.HTTPIngressRuleValue().
				WithPaths(
					paths...,
				),
			),
		)

	if codeServer.Spec.IngressClassName != "" {
		spec = spec.WithIngressClassName(codeServer.Spec.IngressClassName)
	}

	ingress := networkingv1apply.Ingress(codeServer.Name, codeServer.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       CodeServer,
			"app.kubernetes.io/instance":   codeServer.Name,
			"app.kubernetes.io/created-by": CodeServerManager,
		}).
		WithOwnerReferences(owner).
		WithSpec(spec)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ingress)
	if err != nil {
		return fmt.Errorf("failed to convert ingress to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current networkingv1.Ingress
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: codeServer.Namespace, Name: codeServer.Name}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get ingress: %w", err)
	}

	currentApplyConfig, err := networkingv1apply.ExtractIngress(&current, CodeServerManager)
	if err != nil {
		return fmt.Errorf("failed to extract apply configuration from ingress: %w", err)
	}

	if equality.Semantic.DeepEqual(ingress, currentApplyConfig) {
		return nil
	}

	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{FieldManager: CodeServerManager, Force: ptr.To(true)}); err != nil {
		return fmt.Errorf("failed to apply ingress: %w", err)
	}

	logger.Info("Ingress has been reconciled.", "name", codeServer.Name, "namespace", codeServer.Namespace)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CodeServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csv1alpha2.CodeServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}

func controllerReference(codeServer csv1alpha2.CodeServer, scheme *runtime.Scheme) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(&codeServer, scheme)
	if err != nil {
		return nil, err
	}
	ref := metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(codeServer.Name).
		WithUID(codeServer.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)
	return ref, nil
}
