package service

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

type ControllerBy func(obj metav1.Object, owner metav1.Object) bool

type Service struct {
	controllerBy ControllerBy
}

func NewService(controllerBy ControllerBy) *Service {
	return &Service{controllerBy: controllerBy}
}

func (s Service) Reconcile(ctx context.Context, req ctrl.Request, r pkg.APIClientOperator) error {
	logger := log.FromContext(ctx)

	dataLogger := &appv1.DataLogger{}

	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, dataLogger)
	// Ignore not found error, reconcile will be called again when the resource is created.
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	// Fetch the corresponding Deployment
	deployment := &appsv1.Deployment{}

	err = r.Get(ctx, client.ObjectKey{Name: dataLogger.Spec.CustomName, Namespace: dataLogger.Namespace}, deployment)
	if err != nil {
		return err
	}

	// Fetch or create the corresponding Service
	service := &corev1.Service{}

	err = r.Get(ctx, client.ObjectKey{
		Name:      dataLogger.Spec.CustomName,
		Namespace: dataLogger.ObjectMeta.Namespace,
	}, service)
	if err != nil {
		// Service not found, create a new one
		service = s.NewServiceForDataLogger(dataLogger)

		err = r.Create(ctx, service)
		if err != nil {
			return err
		}

		logger.Info("Service was created for dataLogger", service.Name, service.Namespace)

		return nil
	}

	// Reconcile the Service's selector to match the Deployment's Pods
	if !s.controllerBy(service, deployment) {
		service = s.UpdateService(service, deployment, dataLogger)

		err = r.Update(ctx, service)
		if err != nil {
			return err
		}

		logger.Info("Service was updated to match Deployment", service.Name, deployment.Name)
	}

	return nil
}

func (Service) UpdateService(svc *corev1.Service, dep *appsv1.Deployment, dLog *appv1.DataLogger) *corev1.Service {
	// Ensure that the Deployment's Pods have a specific label for the Service to select
	labels := map[string]string{"app": dLog.Spec.CustomName}

	// The Service is not controlled by the Deployment, update it
	svc.ObjectMeta.Labels = labels
	svc.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(dep, schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}),
	}

	svc.Spec = corev1.ServiceSpec{
		Selector: map[string]string{
			"app": dLog.Spec.CustomName,
		},
		Ports: []corev1.ServicePort{
			{
				Port:       dLog.Spec.Port,
				TargetPort: intstr.FromInt32(dLog.Spec.TargetPort),
				NodePort:   dLog.Spec.NodePort,
			},
		},
		Type: corev1.ServiceTypeNodePort, // Set the Service Type to NodePort
	}

	return svc
}

// NewServiceForDataLogger creates a new Service for the given dataLogger CR
func (Service) NewServiceForDataLogger(dataLogger *appv1.DataLogger) *corev1.Service {
	labels := map[string]string{
		"app": dataLogger.Spec.CustomName,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataLogger.Spec.CustomName,
			Namespace: dataLogger.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(dataLogger, appsv1.SchemeGroupVersion.WithKind("dataLogger")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       dataLogger.Spec.Port,
					TargetPort: intstr.FromInt32(dataLogger.Spec.TargetPort),
					NodePort:   dataLogger.Spec.NodePort,
				},
			},
			Type: corev1.ServiceTypeNodePort, // Set the Service Type to NodePort
		},
	}
}
