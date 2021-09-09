/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImageSpec defines the information needed to spawn a custom Job Container
type ImageSpec struct {
	// Image name to be spawn as a job within the cluster
	Name string `json:"name,omitempty"`

	// Version of the desired image
	Version string `json:"version,omitempty"`
}

// ParameterSpec define the information needed to run commands within the container
type ParameterSpec struct {
	// This field will fill the command section within the pod spec. Replace the ENTRYPOINT dockerfile directive
	Command string `json:"command,omitempty"`

	// This field will fill the args section within the pod spec. Replace the CMD dockerfile directive
	Args string `json:"args,omitempty"`
}

// ContainerJobSpec defines the desired state of ContainerJob
type ContainerJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// How many instances should we run for a singe image name
	JobId string `json:"id,omitempty"`

	// Information of the image being created
	ImageSpec ImageSpec `json:"image,omitempty"`

	// Information about additional commands to be run within the container
	ParameterSpec ParameterSpec `json:"parameters,omitempty"`
}

// ContainerJobStatus defines the observed state of ContainerJob
type ContainerJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ContainerJob is the Schema for the containerjobs API
type ContainerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerJobSpec   `json:"spec,omitempty"`
	Status ContainerJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContainerJobList contains a list of ContainerJob
type ContainerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerJob{}, &ContainerJobList{})
}
