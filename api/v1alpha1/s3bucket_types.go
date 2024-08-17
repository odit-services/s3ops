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

// S3BucketSpec defines the desired state of S3Bucket
type S3BucketSpec struct {
	// +kubebuilder:validation:Required
	ServerRef ServerReference `json:"serverRef" yaml:"serverRef"`

	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Required
	Region string `json:"region" yaml:"region"`

	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	ObjectLocking bool `json:"objectLocking" yaml:"objectLocking"`

	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DisableNameGeneration bool `json:"disableNameGeneration" yaml:"disableNameGeneration"`

	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	SoftDelete bool `json:"softDelete" yaml:"softDelete"`

	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=none
	// +kubebuilder:validation:Enum=none;readwrite
	CreateUserFromTemplate string `json:"createUserFromTemplate" yaml:"createUserFromTemplate"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The current state of the resource"
// +kubebuilder:printcolumn:name="LastAction",type="string",JSONPath=".status.lastAction",description="The last action taken on the resource"

// S3Bucket is the Schema for the s3buckets API
type S3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketSpec   `json:"spec,omitempty"`
	Status S3BucketStatus `json:"status,omitempty"`
}

type S3BucketStatus struct {
	CrStatus `json:",inline" yaml:",inline"`
	Created  bool   `json:"created,omitempty" yaml:"created,omitempty"`
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
}

// +kubebuilder:object:root=true
// S3BucketList contains a list of S3Bucket
type S3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Bucket{}, &S3BucketList{})
}
