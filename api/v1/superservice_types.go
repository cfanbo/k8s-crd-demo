/*
Copyright 2023 sxf.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SuperServiceSpec defines the desired state of SuperService
type SuperServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SuperService. Edit superservice_types.go to remove/update
	//Foo string `json:"foo,omitempty"`

	// 字段定义
	Size  *int32               `json:"size"`
	Image string               `json:"image"`
	Envs  []corev1.EnvVar      `json:"envs,omitempty"`
	Ports []corev1.ServicePort `json:"ports,omitempty"`
}

// SuperServiceStatus defines the observed state of SuperService
type SuperServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	appsv1.DeploymentStatus `json:",inline"`
	Count                   int32 `json:"count"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SuperService is the Schema for the superservices API
type SuperService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SuperServiceSpec   `json:"spec,omitempty"`
	Status SuperServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SuperServiceList contains a list of SuperService
type SuperServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SuperService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SuperService{}, &SuperServiceList{})
}
