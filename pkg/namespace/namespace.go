package namespace

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"stackit.cloud/datalogger/pkg"
)

type Namespace struct {
	logger pkg.LogOperator
}

func NewNamespaceReconciler(logger pkg.LogOperator) *Namespace {
	return &Namespace{logger: logger}
}

func (n Namespace) Reconcile(ctx context.Context, req ctrl.Request, apiClient pkg.ApiClientOperator) error {
	namespace := &corev1.Namespace{}

	err := apiClient.Get(ctx, req.NamespacedName, namespace)
	if err != nil {
		if !errors.IsNotFound(err) {
			n.logger.Error(err, "unable to fetch dataLogger CRD namespace", req.Name, req.Namespace)
		}

		return err
	}

	labels := namespace.GetLabels()

	err = n.createNamespaces(ctx, labels, apiClient)
	if err != nil {
		return err
	}

	return err
}

func (Namespace) createNamespaces(ctx context.Context, labels map[string]string, r pkg.ApiClientOperator) error {
	logger := log.FromContext(ctx)

	for key, value := range labels {
		// Assuming the label keys start with "namespaces"
		if key != "name" && strings.HasPrefix(key, "namespaces") {
			namespaceName := value

			// Check if the namespace already exists
			existingNamespace := &corev1.Namespace{}
			err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, existingNamespace)
			if err != nil && errors.IsNotFound(err) {
				// Namespace does not exist, create it
				namespace := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespaceName,
					},
				}

				errCreate := r.Create(ctx, namespace)
				if errCreate != nil {
					logger.Error(errCreate, "error creating namespace", "namespace", namespaceName)
					return errCreate
				}

				logger.Info("Namespace created successfully", "namespace", namespaceName)
			} else if err != nil {
				logger.Error(err, "error checking if namespace exists", "namespace", namespaceName)
				return err
			}
		}
	}
	return nil
}
