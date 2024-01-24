package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

func TestServiceReconcileWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	controlledByTrue := func(obj metav1.Object, owner metav1.Object) bool {
		return true
	}

	reconciler := NewService(controlledByTrue)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		times       int
		want        ctrl.Result
	}{
		{
			name:        "DataLoggerController-1",
			namespace:   "my-namespace1",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
		},
		{
			name:        "DataLoggerController-2",
			namespace:   "my-namespace2",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("rental-%s", test.name), func(t *testing.T) {
			t.Parallel()

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}

			dataLogger := &appv1.DataLogger{}
			apiClient.EXPECT().Get(
				ctx,
				client.ObjectKey{Name: test.name, Namespace: test.namespace},
				dataLogger, gomock.Any(),
			).Times(1).Do(
				func(ctx context.Context, c client.ObjectKey, dataLogger *appv1.DataLogger, opts ...interface{}) error {
					reflect.ValueOf(dataLogger).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").SetString(test.namespace)
					reflect.ValueOf(dataLogger).Elem().FieldByName("Spec").FieldByName("CustomName").SetString(test.name)
					return nil
				},
			).Return(nil)

			deployment := &appsv1.Deployment{}
			apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, deployment).Times(1).Return(nil)

			service := &corev1.Service{}
			apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, service).Times(1).Return(nil)

			err := reconciler.Reconcile(ctx, req, apiClient)
			require.Nil(t, err)
		})
	}
}

func TestServiceReconcileWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	tests := []struct {
		name         string
		namespace    string
		errorValue1  any
		errorValue2  any
		times        int
		want         ctrl.Result
		notFound     *errors2.StatusError
		controlledBy func(obj metav1.Object, owner metav1.Object) bool
	}{
		{
			name:         "DataLoggerController-1",
			namespace:    "my-namespace1",
			errorValue1:  errors.New("get error 1"),
			errorValue2:  nil,
			times:        1,
			want:         ctrl.Result{Requeue: false},
			controlledBy: func(obj metav1.Object, owner metav1.Object) bool { return true },
		},
		{
			name:         "DataLoggerController-1.1",
			namespace:    "my-namespace1",
			errorValue1:  nil,
			errorValue2:  nil,
			times:        1,
			want:         ctrl.Result{Requeue: false},
			notFound:     errors2.NewNotFound(schema.GroupResource{Group: "", Resource: "resources"}, "resource-name"),
			controlledBy: func(obj metav1.Object, owner metav1.Object) bool { return true },
		},
		{
			name:         "DataLoggerController-2",
			namespace:    "my-namespace2",
			errorValue1:  nil,
			errorValue2:  errors.New("get error 2"),
			times:        1,
			want:         ctrl.Result{Requeue: false},
			controlledBy: func(obj metav1.Object, owner metav1.Object) bool { return true },
		},
		{
			name:         "DataLoggerController-3",
			namespace:    "my-namespace3",
			errorValue1:  nil,
			errorValue2:  nil,
			times:        1,
			want:         ctrl.Result{Requeue: false},
			notFound:     nil,
			controlledBy: func(obj metav1.Object, owner metav1.Object) bool { return false },
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			if test.errorValue1 != nil && test.errorValue2 == nil {
				reconciler := NewService(test.controlledBy)

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					dataLogger, gomock.Any(),
				).Times(1).Return(test.errorValue1)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err.Error(), "get error 1")
			}

			if test.errorValue1 == nil && test.errorValue2 == nil && test.notFound != nil {
				reconciler := NewService(test.controlledBy)

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					dataLogger, gomock.Any(),
				).Times(1).Return(test.notFound)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.Nil(t, err)

				return
			}

			if test.errorValue1 == nil && test.errorValue2 != nil && test.notFound == nil {
				reconciler := NewService(test.controlledBy)

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}
				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					dataLogger, gomock.Any(),
				).Times(1).Do(
					func(ctx context.Context, c client.ObjectKey, deps *appv1.DataLogger, opts ...interface{}) error {
						reflect.ValueOf(deps).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").SetString(test.namespace)
						reflect.ValueOf(deps).Elem().FieldByName("Spec").FieldByName("CustomName").SetString(test.name)
						return nil
					},
				).Return(nil)

				deployment := &appsv1.Deployment{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					deployment, gomock.Any(),
				).Times(1).Return(test.errorValue1)

				service := &corev1.Service{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					service, gomock.Any(),
				).Times(1).Return(test.errorValue2)

				service = reconciler.NewServiceForDataLogger(dataLogger)

				service.ObjectMeta.Name = test.name
				service.ObjectMeta.Namespace = test.namespace
				labels := map[string]string{
					"app": test.name,
				}

				service.ObjectMeta.Labels = labels
				service.Spec.Selector = labels

				apiClient.EXPECT().Create(ctx, gomock.Eq(service)).Times(1).Return(test.errorValue1)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.Nil(t, err)

				return
			}

			if !test.controlledBy(&corev1.Service{}, &appsv1.Deployment{}) {
				reconciler := NewService(test.controlledBy)

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}
				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					dataLogger, gomock.Any(),
				).Times(1).Do(
					func(ctx context.Context, c client.ObjectKey, deps *appv1.DataLogger, opts ...interface{}) error {
						reflect.ValueOf(deps).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").SetString(test.namespace)
						reflect.ValueOf(deps).Elem().FieldByName("Spec").FieldByName("CustomName").SetString(test.name)

						return nil
					},
				).Return(nil)

				deployment := &appsv1.Deployment{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					deployment, gomock.Any(),
				).Times(1).Do(func(ctx context.Context, c client.ObjectKey, deps *appsv1.Deployment, opts ...interface{}) error {
					labels := metav1.LabelSelector{
						MatchLabels:      map[string]string{"app": test.name},
						MatchExpressions: nil,
					}

					deps.Spec.Selector = &labels

					return nil
				}).Return(nil)

				service := &corev1.Service{}

				apiClient.EXPECT().Get(
					ctx,
					client.ObjectKey{Name: test.name, Namespace: test.namespace},
					service, gomock.Any(),
				).Times(1).Return(nil)

				dep := &appsv1.Deployment{}
				svc := &corev1.Service{}

				// labels := map[string]string{"app": test.name}
				labels := map[string]string{"app": test.name}

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
						"app": test.name,
					},
					Ports: []corev1.ServicePort{
						{
							Port:       8080,
							TargetPort: intstr.FromInt32(80),
							NodePort:   dataLogger.Spec.Port,
						},
					},
					Type: corev1.ServiceTypeNodePort, // Set the Service Type to NodePort
				}

				apiClient.EXPECT().Update(ctx, svc).Times(1).Return(nil)
				err := reconciler.Reconcile(ctx, req, apiClient)
				require.Nil(t, err)
			}
		})
	}
}
