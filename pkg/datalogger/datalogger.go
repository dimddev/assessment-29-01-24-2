package datalogger

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

// ClusterFinalizer is the name used for our finalizer in the dataLogger resource
// that will be used to handle the deletion of the deployed resources (e.g. namespace)
// accordingly
const ClusterFinalizer = "finalizer.stackit.cloud/datalogger"

type Reconciler struct {
	apiClient  pkg.APIClientOperator
	deployment pkg.DeploymentOperator
	service    pkg.ServiceOperator
}

func NewReconciler(
	apiClient pkg.APIClientOperator,
	deployment pkg.DeploymentOperator,
	service pkg.ServiceOperator,
) *Reconciler {
	return &Reconciler{apiClient: apiClient, deployment: deployment, service: service}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request, dataLogger *appv1.DataLogger) error {
	if !dataLogger.ObjectMeta.DeletionTimestamp.IsZero() && len(dataLogger.ObjectMeta.Finalizers) > 0 {
		if dataLogger.ObjectMeta.Finalizers[0] == ClusterFinalizer {
			dataLogger.ObjectMeta.Finalizers = nil

			err := r.Finalize(ctx, dataLogger, req)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := r.deployment.Reconcile(ctx, req, r.apiClient)
	if err != nil {
		return err
	}

	err = r.service.Reconcile(ctx, req, r.apiClient)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) Finalize(ctx context.Context, dataLogger *appv1.DataLogger, req ctrl.Request) error {
	logger := log.FromContext(ctx)

	ns := &corev1.Namespace{}

	err := r.GetResource(ctx, ns, dataLogger.Spec.CustomName, req.Namespace, logger)
	if err != nil {
		return err
	}

	err = r.DeleteResource(ctx, ns, logger)
	if err != nil {
		return err
	}

	if err = r.apiClient.Update(ctx, dataLogger); err != nil {
		logger.Error(
			err,
			"unable to update dataLogger CR instance",
			dataLogger.Spec.CustomName,
			req.Namespace,
		)

		return err
	}

	logger.Info("Resource was updated successfully", dataLogger.Spec.CustomName, dataLogger.GetNamespace())

	return nil
}

func (r *Reconciler) DeleteResource(
	ctx context.Context,
	obj client.Object,
	logger pkg.LogOperator,
) error {
	err := r.apiClient.Delete(ctx, obj)
	if err != nil {
		logger.Error(
			err,
			"unable to delete dataLogger CD instance",
			obj.GetName(), obj.GetNamespace(),
			obj.GetNamespace(),
		)

		return err
	}

	logger.Info("Resource was deleted successfully", obj.GetName(), obj.GetNamespace())

	return nil
}

func (r *Reconciler) GetResource(
	ctx context.Context,
	obj client.Object,
	name string,
	namespace string,
	logger pkg.LogOperator,
) error {
	err := r.apiClient.Get(ctx, client.ObjectKey{Name: namespace}, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(
				err,
				"unable to fetch dataLogger CR instance",
				name,
				namespace,
			)
		}

		return err
	}

	return nil
}
