/*
Copyright 2019 tommylikehu@gmail.com.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodeServerSpec defines the desired state of CodeServer
type CodeServerSpec struct {
	// Specifies the storage size that will be used for code server
	// +kubebuilder:validation:Pattern="^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`

	// Specifies the storage class name for persistent volume claim
	StorageClassName string `json:"storageClassName,omitempty"`

	// Specifies the additional annotations for persistent volume claim
	StorageAnnotations map[string]string `json:"storageAnnotations,omitempty"`

	// VolumeName specifies the volume name for persistent volume claim
	VolumeName string `json:"volumeName,omitempty"`

	// Specifies the resource requirements for code server pod.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Specifies the period before controller suspend the resources (delete all resources except data).
	SuspendAfterSeconds *int64 `json:"suspendAfterSeconds,omitempty"`

	// Specifies the domain for code server
	Domain string `json:"domain,omitempty"`

	// Specifies the envs
	Envs []corev1.EnvVar `json:"envs,omitempty"`

	// Specifies the image used to running code server
	// +kubebuilder:default="ghcr.io/coder/code-server:latest"
	Image string `json:"image,omitempty"`

	// Specifies the init plugins that will be running to finish before code server running.
	InitPlugins map[string]map[string]string `json:"initPlugins,omitempty"`

	// Specifies the node selector for scheduling.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Specifies the terminal container port for connection, defaults in 19200.
	// +kubebuilder:default=19200
	ContainerPort int32 `json:"containerPort,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	IngressClassName string `json:"ingressClassName,omitempty"`
}

// CodeServerStatus defines the observed state of CodeServer
// +kubebuilder:validation:Enum=NotReady;Ready;Suspended
type CodeServerStatus string

const (
	CodeServerNotReady  CodeServerStatus = "NotReady"
	CodeServerReady     CodeServerStatus = "Ready"
	CodeServerSuspended CodeServerStatus = "Suspended"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="RUNTIME",type="string",JSONPath=".spec.runtime",description="Runtime type"
//+kubebuilder:printcolumn:name="STORAGE",type="string",JSONPath=".spec.storageSize",description="Storage size"

// CodeServer is the Schema for the codeservers API
type CodeServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodeServerSpec   `json:"spec,omitempty"`
	Status CodeServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CodeServerList contains a list of CodeServer
type CodeServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodeServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CodeServer{}, &CodeServerList{})
}
