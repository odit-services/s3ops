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

// S3UserSpec defines the desired state of S3User
type S3UserSpec struct {
	// +kubebuilder:validation:Required
	ServerRef ServerReference `json:"serverRef" yaml:"serverRef"`

	// +Optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	PolicyRefs []string `json:"policyRefs" yaml:"policyRefs"`
}

// S3UserStatus defines the observed state of S3User
type S3UserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CrStatus       `json:",inline" yaml:",inline"`
	Created        bool   `json:"created,omitempty"`
	SecretRef      string `json:"secretRef,omitempty"`
	ProviderMeta   string `json:"providerMeta,"`            // This can be used to store provider-specific metadata, like user ID or ARN
	UserIdentifier string `json:"userIdentifier,omitempty"` // This can be used to store a unique identifier for the user, like an email or username
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Created",type="boolean",JSONPath=".status.created",description="Whether the resource has been created"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The current state of the resource"
// +kubebuilder:printcolumn:name="LastAction",type="string",JSONPath=".status.lastAction",description="The last action taken on the resource"

// S3User is the Schema for the s3users API
type S3User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3UserSpec   `json:"spec,omitempty"`
	Status S3UserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// S3UserList contains a list of S3User
type S3UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3User{}, &S3UserList{})
}
