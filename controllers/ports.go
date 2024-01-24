package controllers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	appv1 "stackit.cloud/datalogger/api/v1"
)

type DataLoggerReconcileOperator interface {
	Reconcile(ctx context.Context, req ctrl.Request, dataLogger *appv1.DataLogger) error
}

type NamespaceReconcileOperator interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}
