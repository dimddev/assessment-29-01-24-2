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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg/datalogger"
)

// DataLoggerReconciler reconciles a DataLogger object
type DataLoggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DataLoggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var dataLogger appv1.DataLogger
	err := r.Client.Get(ctx, req.NamespacedName, &dataLogger)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to fetch DataLogger CRD")
		}
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 30 * time.Second,
		}, client.IgnoreNotFound(err)
	}

	dataLoggerReconciler := datalogger.Reconciler{
		Client:     r.Client,
		DataLogger: &dataLogger,
	}
	err = dataLoggerReconciler.Reconcile(ctx)
	if err != nil {
		log.Error(err, "unable to reconcile DataLogger CRD")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 30 * time.Second,
		}, err
	}

	log.Info("successfully reconciled DataLogger CRD")

	return ctrl.Result{
		Requeue: false,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataLoggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.DataLogger{}).
		Complete(r)
}
