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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TdirectorySpec defines the desired state of Tdirectory
type TdirectorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TdirectoryApp TdirectoryApp `json:"tdirectoryApp,omitempty"`
	TdirectoryDB  TdirectoryDB  `json:"tdirectoryDB,omitempty"`
}

type TdirectoryApp struct {
	Repository      string             `json:"repository,omitempty"`
	Tag             string             `json:"tag,omitempty"`
	ImagePullPolicy corev1.PullPolicy  `json:"imagePullPolicy,omitempty"`
	Replicas        int32              `json:"replicas,omitempty"`
	Port            int32              `json:"port,omitempty"`
	TargetPort      int                `json:"targetPort,omitempty"`
	ServiceType     corev1.ServiceType `json:"serviceType,omitempty"`
}

type TdirectoryDB struct {
	Repository      string            `json:"repository,omitempty"`
	Tag             string            `json:"tag,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	Replicas        int32             `json:"replicas,omitempty"`
	Port            int32             `json:"port,omitempty"`
	DBSize          resource.Quantity `json:"dbSize,omitempty"`
}

// TdirectoryStatus defines the observed state of Tdirectory
type TdirectoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tdirectory is the Schema for the tdirectories API
type Tdirectory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TdirectorySpec   `json:"spec,omitempty"`
	Status TdirectoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TdirectoryList contains a list of Tdirectory
type TdirectoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tdirectory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tdirectory{}, &TdirectoryList{})
}
