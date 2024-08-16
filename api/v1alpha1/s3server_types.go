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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3ServerSpec defines the desired state of S3Server
type S3ServerSpec struct {

	// +kubebuilder:validation:Enum=minio
	// +kubebuilder:validation:Required
	// +kubebuilder:default=minio
	Type string `json:"type" yaml:"type"`

	// +kubebuilder:default=true
	TLS bool `json:"tls" yaml:"tls"`

	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// +kubebuilder:validation:Required
	Auth S3ServerAuthSpec `json:"auth" yaml:"auth"`
}

type S3ServerAuthSpec struct {
	// +kubebuilder:validation:Required
	AccessKey string `json:"accessKey" yaml:"accessKey"`

	// +kubebuilder:validation:Required
	SecretKey string `json:"secretKey" yaml:"secretKey"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// S3Server is the Schema for the s3servers API
type S3Server struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata"`

	Spec   S3ServerSpec   `json:"spec,omitempty" yaml:"spec"`
	Status S3ServerStatus `json:"status,omitempty" yaml:"status"`
}

// +kubebuilder:object:root=true
type S3ServerStatus struct {
	Online     bool               `json:"online" yaml:"online"`
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// S3ServerList contains a list of S3Server
type S3ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Server `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Server{}, &S3ServerList{})
}
