package controllers

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"stackit.cloud/datalogger/pkg"
)

func TestNamespaceControllerWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	operator := pkg.NewMockNamespaceOperator(mockCtrl)
	apiClient := pkg.NewMockAPIClientOperator(mockCtrl)

	schema := &runtime.Scheme{}

	reconciler := NewNamespaceReconciler(apiClient, schema, operator)

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

			operator.EXPECT().Reconcile(ctx, req, apiClient).Times(1).Return(nil)

			actual, err := reconciler.Reconcile(ctx, req)
			require.Nil(t, err)

			reflect.DeepEqual(test.want, actual)
		})
	}
}

func TestNamespaceControllerWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	operator := pkg.NewMockNamespaceOperator(mockCtrl)
	apiClient := pkg.NewMockAPIClientOperator(mockCtrl)

	schema := &runtime.Scheme{}

	reconciler := NewNamespaceReconciler(apiClient, schema, operator)

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
			errorValue1: errors.New("reconcile error 1"),
			errorValue2: nil,
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

			operator.EXPECT().Reconcile(ctx, req, apiClient).Times(1).Return(test.errorValue1)

			actual, err := reconciler.Reconcile(ctx, req)
			reflect.DeepEqual(test.want, actual)

			if test.errorValue1 != nil {
				require.Error(t, err, "reconcile error 1")
			} else {
				require.Nil(t, err)
			}

		})
	}
}
