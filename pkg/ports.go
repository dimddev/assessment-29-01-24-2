package pkg

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApiClientOperator interface {
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
	Reconcile(ctx context.Context, req reconcile.Request, r ApiClientOperator) error
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
