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

	apiClient := pkg.NewMockAPIClientOperator(mockCtrl)

	mockedReference := pkg.NewMockDeploymentReferenceController(mockCtrl)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		times       int
	}{
		{
			name:        "DataLoggerController-1",
			namespace:   "my-namespace1",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
		},
		{
			name:        "DataLoggerController-2",
			namespace:   "my-namespace2",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			reconciler := NewDeployment(mockedReference)

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

			apiClient.EXPECT().Scheme().Times(1).Return(test.errorValue1)
			mockedReference.EXPECT().SetControllerReference(dataLogger, deployment, apiClient.Scheme())

			apiClient.EXPECT().Update(ctx, deployment).Times(1).Return(test.errorValue2)

			err := reconciler.Reconcile(ctx, req, apiClient)
			require.Nil(t, err)
		})
	}
}

func TestDeploymentReconcileWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockAPIClientOperator(mockCtrl)

	mockedReference := pkg.NewMockDeploymentReferenceController(mockCtrl)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		times       int
		notFound    *errors2.StatusError
	}{
		{
			name:        "DataLoggerController-1",
			namespace:   "my-namespace1",
			errorValue1: errors.New("get error 1"),
			errorValue2: nil,
			times:       1,
		},
		{
			name:        "DataLoggerController-2",
			namespace:   "my-namespace2",
			errorValue1: nil,
			errorValue2: errors.New("get error 2"),
			times:       1,
		},
		{
			name:        "DataLoggerController-2.2",
			namespace:   "my-namespace2.2",
			errorValue1: nil,
			errorValue2: nil,
			notFound:    errors2.NewNotFound(schema.GroupResource{Group: "", Resource: "resources"}, "resource-name"),
			times:       1,
		},
		{
			name:        "DataLoggerController-3",
			namespace:   "my-namespace3",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			if test.errorValue1 != nil && test.errorValue2 == nil && test.notFound == nil {
				reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
				req := reconcile.Request{NamespacedName: reqType}
				dataLogger := &appv1.DataLogger{}

				apiClient.EXPECT().Get(ctx, reqType, dataLogger, gomock.Any()).Times(1).Return(test.errorValue1)

				reconciler := NewDeployment(mockedReference)

				err := reconciler.Reconcile(ctx, req, apiClient)

				require.EqualValues(t, err.Error(), "get error 1")
			}

			if test.errorValue1 == nil && test.errorValue2 != nil && test.notFound == nil {
				reconciler := NewDeployment(mockedReference)

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

				apiClient.EXPECT().Scheme().Times(1).Return(nil)

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

				apiClient.EXPECT().Scheme().Times(1).Return(nil)

				mockedReference.EXPECT().SetControllerReference(dataLogger, deployment, apiClient.Scheme()).Times(1).Return(nil)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err.Error(), "get error 2")
			}

			if test.errorValue1 == nil && test.errorValue2 == nil && test.notFound == nil {
				reconciler := NewDeployment(mockedReference)

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

				apiClient.EXPECT().Scheme().Times(2).Return(test.errorValue1)

				dataLogger = &appv1.DataLogger{}

				dataLogger.ObjectMeta.Labels = labels
				dataLogger.Spec.CustomName = test.name
				dataLogger.ObjectMeta.Name = test.name
				dataLogger.ObjectMeta.Namespace = test.namespace

				deployment := reconciler.CreateDeployment(dataLogger)

				mockedReference.EXPECT().SetControllerReference(dataLogger, deployment, apiClient.Scheme()).Times(1).Return(errors.New("owner error"))

				err := reconciler.Reconcile(ctx, req, apiClient)

				require.EqualValues(t, err.Error(), "owner error")
			}

			// not found
			if test.errorValue1 == nil && test.errorValue2 == nil && test.notFound != nil {
				reconciler := NewDeployment(mockedReference)

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

				apiClient.EXPECT().Scheme().Times(1).Return(nil)

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

				apiClient.EXPECT().Get(ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, deployment).Times(1).Return(test.notFound)
				apiClient.EXPECT().Scheme().Times(1).Return(nil)
				mockedReference.EXPECT().SetControllerReference(dataLogger, deployment, apiClient.Scheme()).Return(nil)

				apiClient.EXPECT().Create(ctx, deployment).Times(1).Return(nil)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.Nil(t, err)
			}
		})
	}
}
