package deployment

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

type OwnerCallback func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error

type Deployment struct {
	ownerCallback OwnerCallback
}

func NewDeployment(ownerCallback OwnerCallback) *Deployment {
	return &Deployment{ownerCallback: ownerCallback}
}

func (d Deployment) Reconcile(ctx context.Context, req ctrl.Request, r pkg.ApiClientOperator) error {
	dataLogger := &appv1.DataLogger{}
	if err := r.Get(ctx, req.NamespacedName, dataLogger); err != nil {
		return err
	}

	deployment := d.CreateDeployment(dataLogger)

	// Set owner reference to the dataLogger instance
	err := d.ownerCallback(dataLogger, deployment, r.Scheme())
	if err != nil {
		return err
	}

	// Create or update the Deployment
	err = d.CreateOrUpdate(ctx, deployment, r)
	if err != nil {
		return err
	}

	return nil
}

// CreateOrUpdate creates the resource if it doesn't exist, or updates it if it does.
func (d Deployment) CreateOrUpdate(ctx context.Context, obj *appsv1.Deployment, r pkg.ApiClientOperator) error {
	logger := log.FromContext(ctx)

	err := r.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, obj)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "error getting resource", obj.GetName(), obj.GetNamespace())
			return err
		}

		// Resource does not exist, create it
		err = r.Create(ctx, obj)
		if err != nil {
			logger.Error(err, "error creating resource", obj.GetName(), obj.GetNamespace())
			return err
		}

		logger.Info("Deployment was created successfully.", obj.GetName(), obj.GetNamespace())

		return nil
	}

	// Resource exists, update it
	err = r.Update(ctx, obj)
	if err != nil {
		logger.Error(err, "error updating resource", obj.GetName(), obj.GetNamespace())
		return err
	}

	return nil
}

func (d Deployment) CreateDeployment(dataLogger *appv1.DataLogger) *appsv1.Deployment {
	labels := map[string]string{
		"app.kubernetes.io/name":     dataLogger.ObjectMeta.Labels["app.kubernetes.io/name"],
		"app.kubernetes.io/instance": dataLogger.ObjectMeta.Labels["app.kubernetes.io/instance"],
		"app":                        dataLogger.Spec.CustomName,
	}

	selectorLabels := map[string]string{
		"app.kubernetes.io/name":     dataLogger.ObjectMeta.Labels["app.kubernetes.io/name"],
		"app.kubernetes.io/instance": dataLogger.ObjectMeta.Labels["app.kubernetes.io/instance"],
		"app":                        dataLogger.Spec.CustomName,
	}

	// Reconciliation logic: Create or update Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataLogger.Spec.CustomName,
			Namespace: dataLogger.ObjectMeta.Namespace, // Inherit namespace from dataLogger
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &dataLogger.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "datalogger-container",
							Image: "kennethreitz/httpbin", // Use the image from dataLogger
							Env: []corev1.EnvVar{
								{
									Name:  "CUSTOM_NAME",
									Value: dataLogger.Spec.CustomName,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}
