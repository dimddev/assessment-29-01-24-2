/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataLoggerSpec defines the desired state of DataLogger
type DataLoggerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CustomName string `json:"custom-name"`
	Replicas   int32  `json:"replicas,omitempty"`
	Port       int32  `json:"port,omitempty"`
	NodePort   int32  `json:"node-port,omitempty"`
	TargetPort int32  `json:"target-port,omitempty"`
}

// DataLoggerStatus defines the observed state of DataLogger
type DataLoggerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type MetaDataLogger struct {
	metav1.TypeMeta `json:",inline"`
	Finalizers      []string `json:"finalizers,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DataLogger is the Schema for the dataloggers API
type DataLogger struct {
	MetaDataLogger
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataLoggerSpec   `json:"spec,omitempty"`
	Status DataLoggerStatus `json:"status,omitempty"`
}

// SetResourceVersion sets the resource version for the object
func (d *DataLogger) SetResourceVersion(version string) {
	d.ObjectMeta.ResourceVersion = version
}

// GetGenerateName returns the value of the generateName field
func (d *DataLogger) GetGenerateName() string {
	return d.ObjectMeta.GenerateName
}

func (d *DataLogger) GetCreationTimestamp() v1.Time {
	return d.ObjectMeta.CreationTimestamp
}

// GetLabels returns a copy of the labels associated with the object
func (d *DataLogger) GetLabels() map[string]string {
	if d.ObjectMeta.Labels == nil {
		return nil
	}

	labelsCopy := make(map[string]string, len(d.ObjectMeta.Labels))
	for key, value := range d.ObjectMeta.Labels {
		labelsCopy[key] = value
	}

	return labelsCopy
}

// GetNamespace returns the namespace of the object
func (d *DataLogger) GetNamespace() string {
	return d.ObjectMeta.Namespace
}

// SetNamespace sets the namespace of the object
func (d *DataLogger) SetNamespace(namespace string) {
	d.ObjectMeta.Namespace = namespace
}

// GetName returns the name of the object
func (d *DataLogger) GetName() string {
	return d.ObjectMeta.Name
}

// +kubebuilder:object:root=true

// DataLoggerList contains a list of DataLogger
type DataLoggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataLogger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataLogger{}, &DataLoggerList{})
}
