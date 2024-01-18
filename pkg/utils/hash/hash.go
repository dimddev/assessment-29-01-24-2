// Package hash contains the hash utility's
package hash

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	v1 "stackit.cloud/datalogger/api/v1"
)

var annotation = fmt.Sprintf("%s/last-applied-hash", v1.GroupVersion.Group)

// ComputeToAnnotation creates a hash for an interface and return an annotation map
func ComputeToAnnotation(obj client.Object) (map[string]string, error) {
	buffer := make(map[string]string)

	// if there are no annotations yet, make sure the map is initialized
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(buffer)
	}

	// remove the annotation, as we don't want to hash the annotation
	delete(obj.GetAnnotations(), annotation)

	hash, err := hashJSONObject(obj)
	if err != nil {
		return nil, err
	}
	buffer[annotation] = hash

	return buffer, nil
}

// hashJSONObject will create a SHA256 hash for the passed object
// using json.Marshal
func hashJSONObject(obj any) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// Equal checks if the computed hash inside the annotations of the passed
// objects is equal. If the desired object has no hash annotation, it will
// be computed for the comparison.
func Equal(ctx context.Context, desired, current client.Object) bool {
	log := log.FromContext(ctx)

	// check if there is already a hash annotation for the desired object
	// and use it or compute a new one
	hashA, ok := desired.GetAnnotations()[annotation]
	if !ok {
		log.Info("desired object has no hash annotation, computing one")
		// As our want objects in our tests don't include a valid hash, these
		// objects need to be hashed before the comparison. Our objects in the
		// cluster are already hashed using the build() method inside the reconciler.
		annotations, err := ComputeToAnnotation(desired)
		if err != nil {
			log.Error(err, "failed to compute hash for desired object")
			return false
		}
		hashA = annotations[annotation]
	}
	// use the already computed hash for the got object
	hashB, ok := current.GetAnnotations()[annotation]
	if !ok {
		hashB = ""
	}

	return hashA == hashB
}
