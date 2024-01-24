package datalogger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/pkg"
)

func TestDataLoggerReconcilerFinalizerWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	mockedApiClient := pkg.NewMockApiClientOperator(mockCtrl)

	mockedDeployment := pkg.NewMockDeploymentOperator(mockCtrl)
	mockedService := pkg.NewMockServiceOperator(mockCtrl)
	reconciler := NewReconciler(mockedApiClient, mockedDeployment, mockedService)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		errorValue3 any
		times       int
		want        ctrl.Result
		crdObject   *appv1.DataLogger
	}{
		{
			name:        "DataLoggerController1",
			namespace:   "my-namespace1",
			errorValue1: nil,
			errorValue2: nil,
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: nil, Finalizers: []string{},
				},
			},
		},
		{
			name:        "DataLoggerController2",
			namespace:   "my-namespace2",
			errorValue1: nil,
			errorValue2: nil,
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now().Add(10 * time.Second)}, Finalizers: []string{ClusterFinalizer},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			ns := &corev1.Namespace{}

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.namespace}
			req := reconcile.Request{NamespacedName: reqType}

			if test.crdObject.DeletionTimestamp != nil {
				mockedApiClient.EXPECT().Get(
					ctx, client.ObjectKey{Name: test.namespace}, ns,
				).Times(test.times).Return(test.errorValue1)

				mockedApiClient.EXPECT().Delete(ctx, ns).Times(test.times).Return(test.errorValue2)

				mockedApiClient.EXPECT().Update(ctx, test.crdObject).Times(test.times).Return(test.errorValue3)
			} else {

				mockedDeployment.EXPECT().Reconcile(ctx, req, mockedApiClient).Times(test.times).Return(test.errorValue1)
				mockedService.EXPECT().Reconcile(ctx, req, mockedApiClient).Times(test.times).Return(test.errorValue1)
			}

			err := reconciler.Reconcile(ctx, req, test.crdObject)
			require.Nil(t, err)
		})
	}
}

func TestDataLoggerReconcilerFinalizerWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	mockedApiClient := pkg.NewMockApiClientOperator(mockCtrl)

	mockedDeployment := pkg.NewMockDeploymentOperator(mockCtrl)
	mockedService := pkg.NewMockServiceOperator(mockCtrl)
	reconciler := NewReconciler(mockedApiClient, mockedDeployment, mockedService)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 error
		errorValue2 error
		errorValue3 error
		times       int
		want        ctrl.Result
		crdObject   *appv1.DataLogger
	}{
		{
			name:        "DataLoggerController0",
			namespace:   "my-namespace0",
			errorValue1: errors.New("get method api error 0"),
			errorValue2: nil,
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: nil, Finalizers: []string{},
				},
			},
		},
		{
			name:        "DataLoggerController1",
			namespace:   "my-namespace1",
			errorValue1: nil,
			errorValue2: errors.New("get method api error 2"),
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: nil, Finalizers: []string{},
				},
			},
		},
		{
			name:        "DataLoggerController2",
			namespace:   "my-namespace2",
			errorValue1: errors.New("get method api error 1"),
			errorValue2: nil,
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now().Add(10 * time.Second)}, Finalizers: []string{ClusterFinalizer},
				},
			},
		},
		{
			name:        "DataLoggerController3",
			namespace:   "my-namespace3",
			errorValue1: nil,
			errorValue2: errors.New("get method api error 3"),
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now().Add(10 * time.Second)}, Finalizers: []string{ClusterFinalizer},
				},
			},
		},
		{
			name:        "DataLoggerController4",
			namespace:   "my-namespace4",
			errorValue1: nil,
			errorValue2: nil,
			errorValue3: errors.New("get method api error 4"),
			times:       1,
			want:        ctrl.Result{Requeue: false},
			crdObject: &appv1.DataLogger{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now().Add(10 * time.Second)}, Finalizers: []string{ClusterFinalizer},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			ns := &corev1.Namespace{}

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.namespace}
			req := reconcile.Request{NamespacedName: reqType}

			if test.crdObject.DeletionTimestamp != nil {
				if test.errorValue1 != nil && (test.errorValue2 == nil && test.errorValue3 == nil) {
					mockedApiClient.EXPECT().Get(
						ctx, client.ObjectKey{Name: test.namespace}, ns,
					).Times(test.times).Return(test.errorValue1)

					err := reconciler.Reconcile(ctx, req, test.crdObject)
					require.EqualValues(t, err.Error(), test.errorValue1.Error())
				}

				if test.errorValue2 != nil && (test.errorValue1 == nil && test.errorValue2 == nil) {
					mockedApiClient.EXPECT().Get(
						ctx, client.ObjectKey{Name: test.namespace}, ns,
					).Times(test.times).Return(test.errorValue1)

					mockedApiClient.EXPECT().Delete(
						ctx, ns,
					).Times(test.times).Return(test.errorValue2)

					err := reconciler.Reconcile(ctx, req, test.crdObject)
					require.EqualValues(t, err.Error(), test.errorValue2.Error())
				}

				if test.errorValue3 != nil && (test.errorValue2 == nil && test.errorValue1 == nil) {
					mockedApiClient.EXPECT().Get(
						ctx, client.ObjectKey{Name: test.namespace}, ns,
					).Times(test.times).Return(test.errorValue1)

					mockedApiClient.EXPECT().Delete(
						ctx, ns,
					).Times(test.times).Return(test.errorValue2)

					mockedApiClient.EXPECT().Update(
						ctx, test.crdObject,
					).Times(test.times).Return(test.errorValue3)

					err := reconciler.Reconcile(ctx, req, test.crdObject)
					require.EqualValues(t, err.Error(), test.errorValue3.Error())
				}
			} else {
				if test.errorValue1 != nil {
					mockedDeployment.EXPECT().Reconcile(ctx, req, mockedApiClient).Times(test.times).Return(test.errorValue1)

					err := reconciler.Reconcile(ctx, req, test.crdObject)
					require.EqualValues(t, err.Error(), test.errorValue1.Error())
				}

				if test.errorValue2 != nil {
					mockedDeployment.EXPECT().Reconcile(ctx, req, mockedApiClient).Times(test.times).Return(test.errorValue1)
					mockedService.EXPECT().Reconcile(ctx, req, mockedApiClient).Times(1).Return(test.errorValue2)

					err := reconciler.Reconcile(ctx, req, test.crdObject)

					require.EqualValues(t, err.Error(), test.errorValue2.Error())
				}
			}
		})
	}
}
