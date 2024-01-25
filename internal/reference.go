package internal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DeploymentReference struct{}

func NewDeploymentReference() *DeploymentReference {
	return &DeploymentReference{}
}

func (r DeploymentReference) SetControllerReference(owner, controlled metav1.Object, scheme *runtime.Scheme) error {
	return controllerutil.SetControllerReference(owner, controlled, scheme)
}

type ServiceReference struct{}

func NewServiceReference() *ServiceReference {
	return &ServiceReference{}
}

func (s ServiceReference) IsControlledBy(obj metav1.Object, owner metav1.Object) bool {
	return metav1.IsControlledBy(obj, owner)
}
