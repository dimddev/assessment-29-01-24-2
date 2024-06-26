package pkg

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type APIClientOperator interface {
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	Scheme() *runtime.Scheme
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
}

type LogOperator interface {
	Info(msg string, keysAndValues ...any)
	Error(err error, msg string, keysAndValues ...any)
}

type ReconcileOperator interface {
	Reconcile(ctx context.Context, req reconcile.Request, r APIClientOperator) error
}

type DeploymentOperator interface {
	ReconcileOperator
}

type ServiceOperator interface {
	ReconcileOperator
}

type NamespaceOperator interface {
	ReconcileOperator
}

type DeploymentReferenceController interface {
	SetControllerReference(owner, controlled metav1.Object, scheme *runtime.Scheme) error
}

type ServiceReferenceController interface {
	IsControlledBy(obj metav1.Object, owner metav1.Object) bool
}
