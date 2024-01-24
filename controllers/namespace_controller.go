// controllers/namespace_controller.go

package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"stackit.cloud/datalogger/pkg"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client pkg.ApiClientOperator
	ns     pkg.NamespaceOperator
	scheme *runtime.Scheme
}

func NewNamespaceReconciler(client pkg.ApiClientOperator, scheme *runtime.Scheme, ns pkg.NamespaceOperator) *NamespaceReconciler {
	return &NamespaceReconciler{client: client, scheme: scheme, ns: ns}
}

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	err := r.ns.Reconcile(ctx, req, r.client)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "unable to fetch dataLogger CRD namespace", req.Name, req.Namespace)
		}
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 30 * time.Second,
		}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
