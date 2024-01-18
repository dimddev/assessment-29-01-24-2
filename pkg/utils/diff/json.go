// Package diff contains the diff utility's
package diff

import (
	"encoding/json"

	"github.com/wI2L/jsondiff"
	"k8s.io/apimachinery/pkg/runtime"
)

// JSON will return a kubernetes runtime.Object marshaled into a JSON string,
// that is wrapped in []byte representation
func JSON(a, b runtime.Object) ([]byte, error) {
	patch, err := jsondiff.Compare(a, b)
	if err != nil {
		return nil, err
	}
	_ = jsondiff.Factorize()
	buffer, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	return buffer, err
}
