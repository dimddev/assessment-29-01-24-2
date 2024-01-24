package deployment

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

func TestDeploymentReconcileWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	tests := []struct {
		name          string
		namespace     string
		errorValue1   any
		errorValue2   any
		times         int
		ownerCallback OwnerCallback
	}{
		{
			name:          "DataLoggerController-1",
			namespace:     "my-namespace1",
			errorValue1:   nil,
			errorValue2:   nil,
			times:         1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error { return nil },
		},
		{
			name:          "DataLoggerController-2",
			namespace:     "my-namespace2",
			errorValue1:   nil,
			errorValue2:   nil,
			times:         1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error { return nil },
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			reconciler := NewDeployment(OwnerCallback(test.ownerCallback))

			labels := map[string]string{
				"app.kubernetes.io/name":     "app.kubernetes.io/name",
				"app.kubernetes.io/instance": "app.kubernetes.io/instance",
				"app":                        test.name,
			}

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}

			dataLogger := &appv1.DataLogger{}
			apiClient.EXPECT().Get(
				ctx,
				reqType,
				dataLogger, gomock.Any(),
			).Times(1).Do(func(ctx context.Context, key client.ObjectKey, obj *appv1.DataLogger, opts ...interface{}) error {
				obj.ObjectMeta.Namespace = test.namespace
				obj.ObjectMeta.Name = test.name
				obj.ObjectMeta.Labels = labels
				obj.Spec.CustomName = test.name

				return nil
			}).Return(test.errorValue1)

			apiClient.EXPECT().Scheme().Times(1).Return(test.errorValue1)

			dataLogger = &appv1.DataLogger{}

			dataLogger.ObjectMeta.Labels = labels
			dataLogger.Spec.CustomName = test.name
			dataLogger.ObjectMeta.Name = test.name
			dataLogger.ObjectMeta.Namespace = test.namespace
			//
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.name,
					Namespace: test.namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &dataLogger.Spec.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
				},
			}

			deployment = reconciler.CreateDeployment(dataLogger)
			apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, deployment).Times(1).Return(test.errorValue2)

			apiClient.EXPECT().Update(ctx, deployment).Times(1).Return(test.errorValue2)

			err := reconciler.Reconcile(ctx, req, apiClient)
			require.Nil(t, err)
		})
	}
}

func TestDeploymentReconcileWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	tests := []struct {
		name          string
		namespace     string
		errorValue1   any
		errorValue2   any
		times         int
		ownerCallback OwnerCallback
		notFound      *errors2.StatusError
	}{
		{
			name:          "DataLoggerController-1",
			namespace:     "my-namespace1",
			errorValue1:   errors.New("get error 1"),
			errorValue2:   nil,
			times:         1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error { return nil },
		},
		{
			name:          "DataLoggerController-2",
			namespace:     "my-namespace2",
			errorValue1:   nil,
			errorValue2:   errors.New("get error 2"),
			times:         1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error { return nil },
		},
		{
			name:          "DataLoggerController-2.2",
			namespace:     "my-namespace2.2",
			errorValue1:   nil,
			errorValue2:   nil,
			notFound:      errors2.NewNotFound(schema.GroupResource{Group: "", Resource: "resources"}, "resource-name"),
			times:         1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error { return nil },
		},
		{
			name:        "DataLoggerController-3",
			namespace:   "my-namespace3",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
			ownerCallback: func(owner metav1.Object, controlled metav1.Object, scheme *runtime.Scheme) error {
				return errors.New("owner error")
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			if test.errorValue1 != nil && test.errorValue2 == nil {
				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}
				dataLogger := &appv1.DataLogger{}

				apiClient.EXPECT().Get(ctx, reqType, dataLogger, gomock.Any()).Times(1).Return(test.errorValue1)

				reconciler := NewDeployment(test.ownerCallback)
				err := reconciler.Reconcile(ctx, req, apiClient)

				require.EqualValues(t, err.Error(), "get error 1")
			}

			if test.errorValue1 == nil && test.errorValue2 != nil {
				labels := map[string]string{
					"app.kubernetes.io/name":     "app.kubernetes.io/name",
					"app.kubernetes.io/instance": "app.kubernetes.io/instance",
					"app":                        test.name,
				}

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}
				apiClient.EXPECT().Get(
					ctx,
					reqType,
					dataLogger, gomock.Any(),
				).Times(1).Do(func(ctx context.Context, key client.ObjectKey, obj *appv1.DataLogger, opts ...interface{}) error {
					obj.ObjectMeta.Namespace = test.namespace
					obj.ObjectMeta.Name = test.name
					obj.ObjectMeta.Labels = labels
					obj.Spec.CustomName = test.name

					return nil
				}).Return(test.errorValue1)

				apiClient.EXPECT().Scheme().Times(1).Return(test.errorValue1)

				dataLogger = &appv1.DataLogger{}

				dataLogger.ObjectMeta.Labels = labels
				dataLogger.Spec.CustomName = test.name
				dataLogger.ObjectMeta.Name = test.name
				dataLogger.ObjectMeta.Namespace = test.namespace
				//
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.name,
						Namespace: test.namespace,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &dataLogger.Spec.Replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				}

				reconciler := NewDeployment(test.ownerCallback)
				deployment = reconciler.CreateDeployment(dataLogger)

				apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, deployment).Times(1).Return(test.errorValue2)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err.Error(), "get error 2")
			}

			if test.errorValue1 == nil && test.errorValue2 == nil && test.notFound == nil {
				reconciler := NewDeployment(OwnerCallback(test.ownerCallback))

				labels := map[string]string{
					"app.kubernetes.io/name":     "app.kubernetes.io/name",
					"app.kubernetes.io/instance": "app.kubernetes.io/instance",
					"app":                        test.name,
				}

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}
				apiClient.EXPECT().Get(
					ctx,
					reqType,
					dataLogger, gomock.Any(),
				).Times(1).Do(func(ctx context.Context, key client.ObjectKey, obj *appv1.DataLogger, opts ...interface{}) error {
					obj.ObjectMeta.Namespace = test.namespace
					obj.ObjectMeta.Name = test.name
					obj.ObjectMeta.Labels = labels
					obj.Spec.CustomName = test.name

					return nil
				}).Return(test.errorValue1)

				apiClient.EXPECT().Scheme().Times(1).Return(test.errorValue1)

				dataLogger = &appv1.DataLogger{}

				dataLogger.ObjectMeta.Labels = labels
				dataLogger.Spec.CustomName = test.name
				dataLogger.ObjectMeta.Name = test.name
				dataLogger.ObjectMeta.Namespace = test.namespace
				//

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err.Error(), "owner error")
			}

			// not found
			if test.errorValue1 == nil && test.errorValue2 == nil && test.notFound != nil {
				labels := map[string]string{
					"app.kubernetes.io/name":     "app.kubernetes.io/name",
					"app.kubernetes.io/instance": "app.kubernetes.io/instance",
					"app":                        test.name,
				}

				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}

				dataLogger := &appv1.DataLogger{}
				apiClient.EXPECT().Get(
					ctx,
					reqType,
					dataLogger, gomock.Any(),
				).Times(1).Do(func(ctx context.Context, key client.ObjectKey, obj *appv1.DataLogger, opts ...interface{}) error {
					obj.ObjectMeta.Namespace = test.namespace
					obj.ObjectMeta.Name = test.name
					obj.ObjectMeta.Labels = labels
					obj.Spec.CustomName = test.name

					return nil
				}).Return(test.errorValue1)

				apiClient.EXPECT().Scheme().Times(1).Return(test.errorValue1)

				dataLogger = &appv1.DataLogger{}

				dataLogger.ObjectMeta.Labels = labels
				dataLogger.Spec.CustomName = test.name
				dataLogger.ObjectMeta.Name = test.name
				dataLogger.ObjectMeta.Namespace = test.namespace
				//
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.name,
						Namespace: test.namespace,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &dataLogger.Spec.Replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				}

				reconciler := NewDeployment(test.ownerCallback)
				deployment = reconciler.CreateDeployment(dataLogger)

				apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, deployment).Times(1).Return(test.notFound)

				apiClient.EXPECT().Create(ctx, deployment).Times(1).Return(nil)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.Nil(t, err)
			}

		})
	}
}
