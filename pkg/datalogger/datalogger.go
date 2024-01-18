package datalogger

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	dataloggerv1 "stackit.cloud/datalogger/api/v1"
)

// ClusterFinalizer is the name used for our finalizer in the DataLogger resource
// that will be used to handle the deletion of the deployed resources (e.g. namespace)
// accordingly
const ClusterFinalizer = "finalizer.stackit.cloud/datalogger"

type Reconciler struct {
	Client     client.Client
	DataLogger *dataloggerv1.DataLogger
}

func (r *Reconciler) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)

	// add finalizer or handle deletion
	err := r.handleFinalizer(ctx)
	if err != nil {
		return err
	}

	if !r.DataLogger.DeletionTimestamp.IsZero() {
		log.Info("data logger will be deleted, skipping reconciliation")
		return nil
	}

	// reconcile necessary resources
	// - namespace
	// - deployment (with our data-logger image: kennethreitz/httpbin)
	// - service

	// TODO: reconcile namespace

	// TODO: reconcile deployment

	// TODO: reconcile service

	return nil
}

func (r *Reconciler) handleFinalizer(ctx context.Context) error {
	_ = log.FromContext(ctx)

	// TODO: handle deletion of cluster using the finalizer approach of Kubernetes

	return nil
}
