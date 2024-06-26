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
	"reflect"

	csv1alpha2 "github.com/walnuts1018/code-server-operator/api/v1alpha2"
	"github.com/walnuts1018/code-server-operator/util/random"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CodeServerDeploymentReconciler reconciles a CodeServerDeployment object
type CodeServerDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserverdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserverdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserverdeployments/finalizers,verbs=update

//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserver,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserver/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cs.walnuts.dev,resources=codeserver/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CodeServerDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *CodeServerDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var codeServerDeployments csv1alpha2.CodeServerDeployment

	err := r.Client.Get(ctx, req.NamespacedName, &codeServerDeployments)
	if errors.IsNotFound(err) {
		logger.Info("CodeServerDeployment resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "Failed to get CodeServerDeployment")
		return ctrl.Result{}, err
	}

	if !codeServerDeployments.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("CodeServerDeployment is being deleted")
		return ctrl.Result{}, nil
	}

	if err := r.reconcileCodeServer(ctx, &codeServerDeployments); err != nil {
		logger.Error(err, "Failed to reconcile CodeServer")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CodeServerDeploymentReconciler) reconcileCodeServer(ctx context.Context, codeServerDeployments *csv1alpha2.CodeServerDeployment) error {
	logger := log.FromContext(ctx)

	codeServers := csv1alpha2.CodeServerList{}
	err := r.Client.List(ctx, &codeServers, &client.ListOptions{
		Namespace: codeServerDeployments.Namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/name":              CodeServer,
			"cs.walnuts.dev/codeserverdeployment": codeServerDeployments.Name,
		}),
	})

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list CodeServer: %w", err)
	}

	// 既存のCodeServerのSpecとCodeServerDeploymentのSpecが一致しているか確認
	for _, codeServer := range codeServers.Items {
		if !reflect.DeepEqual(codeServer.Spec, codeServerDeployments.Spec.Template.Spec) {
			patch := &unstructured.Unstructured{}
			patch.SetGroupVersionKind(csv1alpha2.GroupVersion.WithKind("CodeServer"))
			patch.SetNamespace(codeServerDeployments.Namespace)
			patch.SetName(codeServer.Name)
			patch.SetLabels(map[string]string{
				"app.kubernetes.io/name":              CodeServer,
				"app.kubernetes.io/instance":          codeServer.Name,
				"app.kubernetes.io/created-by":        CodeServerManager,
				"cs.walnuts.dev/codeserverdeployment": codeServerDeployments.Name,
			})
			patch.UnstructuredContent()["spec"] = codeServerDeployments.Spec.Template.Spec
			patch.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion:         codeServerDeployments.APIVersion,
					Kind:               codeServerDeployments.Kind,
					Name:               codeServerDeployments.Name,
					UID:                codeServerDeployments.UID,
					Controller:         func(b bool) *bool { return &b }(true),
					BlockOwnerDeletion: func(b bool) *bool { return &b }(true),
				},
			})

			if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{FieldManager: CodeServerManager, Force: ptr.To(true)}); err != nil {
				return fmt.Errorf("failed to apply CodeServer: %w", err)
			}
			logger.Info("Patched CodeServer", "Name", codeServer.Name)
		}
	}

	nowReplicas := int32(len(codeServers.Items))
	if nowReplicas == codeServerDeployments.Spec.Replicas {
		return nil
	}

	for i := nowReplicas; i > codeServerDeployments.Spec.Replicas; i-- {
		codeServer := &codeServers.Items[i-1]
		if err := r.Client.Delete(ctx, codeServer); err != nil {
			return fmt.Errorf("failed to delete CodeServer: %w", err)
		}
	}

	for i := nowReplicas; i < codeServerDeployments.Spec.Replicas; i++ {
		suffix, err := random.String(6, random.LowerLetters)
		if err != nil {
			logger.Error(err, "Failed to generate random string")
			return err
		}

		codeServer := &csv1alpha2.CodeServer{}
		codeServer.Name = codeServerDeployments.Name + "-" + suffix
		codeServer.Namespace = codeServerDeployments.Namespace

		op, err := ctrl.CreateOrUpdate(ctx, r.Client, codeServer, func() error {
			codeServer.Spec = codeServerDeployments.Spec.Template.Spec

			if codeServer.Labels == nil {
				codeServer.Labels = make(map[string]string)
			}

			codeServer.Labels["app.kubernetes.io/name"] = CodeServer
			codeServer.Labels["app.kubernetes.io/instance"] = codeServer.Name
			codeServer.Labels["app.kubernetes.io/created-by"] = CodeServerManager
			codeServer.Labels["cs.walnuts.dev/codeserverdeployment"] = codeServerDeployments.Name

			return ctrl.SetControllerReference(codeServerDeployments, codeServer, r.Scheme)
		})

		if err != nil {
			return fmt.Errorf("failed to reconcile CodeServer: %w", err)
		}

		if op != controllerutil.OperationResultNone {
			logger.Info("Reconciled CodeServer", "operation", op)
		}

		continue
	}

	logger.Info("Reconcile CodeServer successfully")
	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *CodeServerDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csv1alpha2.CodeServerDeployment{}).
		Owns(&csv1alpha2.CodeServer{}).
		Complete(r)
}
