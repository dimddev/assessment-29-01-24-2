package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appv1 "stackit.cloud/datalogger/api/v1"
	"stackit.cloud/datalogger/controllers/mock"
	"stackit.cloud/datalogger/pkg"
)

func TestDataLoggerControllerWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	operator := mock.NewMockDataLoggerReconcileOperator(mockCtrl)
	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	schema := &runtime.Scheme{}

	reconciler := NewDataLoggerReconciler(apiClient, operator, schema)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		times       int
		want        ctrl.Result
	}{
		{
			name:        "DataLoggerController",
			namespace:   "my-namespace1",
			errorValue1: nil,
			errorValue2: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
		},
		{
			name:        "DataLoggerController",
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

			crdObject := &appv1.DataLogger{}

			apiClient.EXPECT().Get(
				ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, crdObject,
			).Times(test.times).Return(test.errorValue1)

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}

			operator.EXPECT().Reconcile(ctx, req, crdObject).Times(test.times).Return(test.errorValue2)

			actual, err := reconciler.Reconcile(ctx, req)
			require.Nil(t, err)

			reflect.DeepEqual(test.want, actual)
		})
	}
}

func TestDataLoggerControllerWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	operator := mock.NewMockDataLoggerReconcileOperator(mockCtrl)
	apiClient := pkg.NewMockApiClientOperator(mockCtrl)

	schema := &runtime.Scheme{}

	reconciler := NewDataLoggerReconciler(apiClient, operator, schema)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		times       int
		want        ctrl.Result
	}{
		{
			name:        "DataLoggerController",
			namespace:   "my-namespace2",
			errorValue1: errors.New("unknown api error"),
			errorValue2: nil,
			times:       1,
			want:        ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second},
		},
		{
			name:        "DataLoggerController",
			namespace:   "my-namespace3",
			errorValue1: nil,
			errorValue2: errors.New("unknown api error"),
			times:       1,
			want:        ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("rental-%s", test.name), func(t *testing.T) {
			t.Parallel()

			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}
			crdObject := &appv1.DataLogger{}

			apiClient.EXPECT().Get(
				ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, crdObject,
			).Times(test.times).Return(test.errorValue1)

			if test.errorValue2 != nil {
				operator.EXPECT().Reconcile(ctx, req, crdObject).Times(test.times).Return(test.errorValue2)
			}

			actual, err := reconciler.Reconcile(ctx, req)
			require.Error(t, err, "unknown api error")

			reflect.DeepEqual(test.want, actual)
		})
	}
}
