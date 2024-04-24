/*
Copyright 2024.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodeServerDeploymentSpec defines the desired state of CodeServerDeployment
type CodeServerDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template CodeServersTemplate `json:"template"`
	Replicas int32               `json:"replicas"`
}

type CodeServersTemplate struct {
	Spec CodeServerSpec `json:"spec"`
}

// CodeServerDeploymentStatus defines the observed state of CodeServerDeployment
type CodeServerDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="REPLICAS",type="integer",JSONPath=".spec.replicas",description="Number of replicas"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// CodeServerDeployment is the Schema for the codeserverdeployments API
type CodeServerDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodeServerDeploymentSpec   `json:"spec,omitempty"`
	Status CodeServerDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CodeServerDeploymentList contains a list of CodeServerDeployment
type CodeServerDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodeServerDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CodeServerDeployment{}, &CodeServerDeploymentList{})
}
