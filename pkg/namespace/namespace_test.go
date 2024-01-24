package namespace

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"stackit.cloud/datalogger/pkg"
)

func TestNamespaceReconcileWithNoErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)
	logger := pkg.NewMockLogOperator(mockCtrl)

	reconciler := NewNamespaceReconciler(logger)

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
		t.Run(fmt.Sprintf("datalogger-%s", test.name), func(t *testing.T) {
			t.Parallel()

			ns := &corev1.Namespace{}
			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}

			apiClient.EXPECT().Get(
				ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, ns, gomock.Any(),
			).Times(test.times).Do(
				func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...interface{}) error {
					ns1 := &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
							"namespaces1": test.namespace,
						}},
					}

					reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").
						FieldByName("Labels").
						Set(reflect.ValueOf(ns1.ObjectMeta.Labels))

					reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").SetString(test.namespace)

					return nil
				}).Return(test.errorValue1)

			ns1 := &corev1.Namespace{}
			apiClient.EXPECT().Get(
				ctx, client.ObjectKey{Name: test.namespace}, ns1).Times(1).Return(test.errorValue1)

			err := reconciler.Reconcile(ctx, req, apiClient)
			require.Nil(t, err)
		})
	}
}

func TestNamespaceReconcileWithErrors(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)

	apiClient := pkg.NewMockApiClientOperator(mockCtrl)
	logger := pkg.NewMockLogOperator(mockCtrl)

	reconciler := NewNamespaceReconciler(logger)

	tests := []struct {
		name        string
		namespace   string
		errorValue1 any
		errorValue2 any
		errorValue3 any
		times       int
		want        ctrl.Result
		notFound    *errors2.StatusError
	}{
		{
			name:        "DataLoggerController",
			namespace:   "my-namespace1",
			errorValue1: errors.New("unable to fetch dataLogger CRD namespace"),
			errorValue2: nil,
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			notFound:    nil,
		},
		{
			name:        "DataLoggerController",
			namespace:   "my-namespace2",
			errorValue1: nil,
			errorValue2: errors.New("unable to fetch dataLogger CRD namespace"),
			errorValue3: nil,
			times:       1,
			want:        ctrl.Result{Requeue: false},
			notFound:    errors2.NewNotFound(schema.GroupResource{Group: "", Resource: "resources"}, "resource-name"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("rental-%s", test.name), func(t *testing.T) {
			t.Parallel()

			ns := &corev1.Namespace{}
			reqType := types.NamespacedName{Namespace: test.namespace, Name: test.name}
			req := reconcile.Request{NamespacedName: reqType}

			if test.errorValue1 != nil && test.errorValue2 == nil && test.errorValue3 == nil {
				apiClient.EXPECT().Get(
					ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, ns, gomock.Any(),
				).Times(test.times).Return(test.errorValue1)

				logger.EXPECT().Error(
					errors.New("unable to fetch dataLogger CRD namespace"),
					"unable to fetch dataLogger CRD namespace",
					test.name,
					test.namespace).
					Times(1)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err, test.errorValue1)
			}

			if test.errorValue2 != nil && test.errorValue1 == nil && test.errorValue3 == nil {

				apiClient.EXPECT().Get(
					ctx, client.ObjectKey{Name: test.name, Namespace: test.namespace}, ns, gomock.Any(),
				).Times(test.times).Do(
					func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...interface{}) error {
						ns1 := &corev1.Namespace{
							ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
								"namespaces2": test.namespace,
							}},
						}

						reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").
							FieldByName("Labels").
							Set(reflect.ValueOf(ns1.ObjectMeta.Labels))

						reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").SetString(test.namespace)

						return nil
					}).Return(test.errorValue1)

				ns1 := &corev1.Namespace{}
				apiClient.EXPECT().Get(
					ctx, client.ObjectKey{Name: test.namespace}, ns1).Times(1).Return(test.notFound)

				namespace := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: test.namespace,
					},
				}

				apiClient.EXPECT().Create(
					ctx, namespace).Times(1).Return(test.errorValue2)

				err := reconciler.Reconcile(ctx, req, apiClient)
				require.EqualValues(t, err, test.errorValue2)
			}
		})
	}
}
