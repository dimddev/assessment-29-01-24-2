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

package main

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"stackit.cloud/datalogger/internal"

	"stackit.cloud/datalogger/pkg/datalogger"
	"stackit.cloud/datalogger/pkg/deployment"
	"stackit.cloud/datalogger/pkg/namespace"
	"stackit.cloud/datalogger/pkg/service"

	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// CustomOptions extends ctrl.Options with MetricsBindAddress and Port
type CustomOptions struct {
	ctrl.Options

	MetricsBindAddress string
	Port               int
}

func main() {
	var customOpts CustomOptions
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&customOpts.MetricsBindAddress, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.IntVar(&customOpts.Port, "port", 9443, "The port the controller manager serves on.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	customOpts.Options = ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "dfc9ea2a.stackit.cloud",
	}

	// Load kubeconfig from the specified path or the default path
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load kubeconfig")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(kubeconfig, customOpts.Options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	deploymentReference := internal.NewDeploymentReference()

	newDeployment := deployment.NewDeployment(deploymentReference)

	serviceReference := internal.NewServiceReference()

	newService := service.NewService(serviceReference)

	dataLoggerReconciler := datalogger.NewReconciler(mgr.GetClient(), newDeployment, newService)

	err = controllers.NewDataLoggerReconciler(
		mgr.GetClient(), dataLoggerReconciler, mgr.GetScheme()).SetupWithManager(mgr)

	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "dataLogger")
		os.Exit(1)
	}

	namespaceOperator := namespace.NewNamespaceReconciler(setupLog)

	err = controllers.NewNamespaceReconciler(mgr.GetClient(), mgr.GetScheme(), namespaceOperator).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
