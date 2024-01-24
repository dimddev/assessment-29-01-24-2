/*
Copyright 2023.

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

package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"stackit.cloud/datalogger/pkg"

	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "stackit.cloud/datalogger/api/v1"
)

// DataLoggerReconciler reconciles a dataLogger object
type DataLoggerReconciler struct {
	apiClient pkg.APIClientOperator
	operator  DataLoggerReconcileOperator
	scheme    *runtime.Scheme
}

func NewDataLoggerReconciler(
	apiClient pkg.APIClientOperator,
	operator DataLoggerReconcileOperator,
	scheme *runtime.Scheme,
) *DataLoggerReconciler {
	return &DataLoggerReconciler{apiClient: apiClient, scheme: scheme, operator: operator}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DataLoggerReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	dataLogger := &appv1.DataLogger{}

	err := r.apiClient.Get(
		ctx,
		client.ObjectKey{Name: req.NamespacedName.Name, Namespace: req.NamespacedName.Namespace},
		dataLogger,
	)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "unable to fetch dataLogger CRD", "datalogger controller", req.Namespace)
		}

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: RequeueTime * time.Second,
		}, client.IgnoreNotFound(err)
	}

	err = r.operator.Reconcile(ctx, req, dataLogger)
	if err != nil {
		logger.Error(err, "unable to reconcile dataLogger CRD", "namespace", req.Namespace)

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: RequeueTime * time.Second,
		}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataLoggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.DataLogger{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
