/*
Copyright (C) 2025 Bankdata (bankdata@bankdata.dk)

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

// LibrarySpec defines the desired state of Library
type LibrarySpec struct {

	// Name is the name the Library will have in Styra DAS
	Name string `json:"name"`

	// Description is the description of the Library
	Description string `json:"description"`

	// Subjects is the list of subjects which should have access to the system.
	Subjects []LibrarySubject `json:"subjects,omitempty"`

	// SourceControl is the sourcecontrol configuration for the Library
	SourceControl *SourceControl `json:"sourceControl,omitempty"`

	// Datasources is the list of datasources in the Library
	Datasources []LibraryDatasource `json:"datasources,omitempty"`
}

// LibraryDatasource contains metadata of a datasource, stored in a library
type LibraryDatasource struct {
	// Path is the path within the system where the datasource should reside.
	Path string `json:"path"`

	// Description is a description of the datasource
	Description string `json:"description,omitempty"`
}

// SourceControl is a struct from styra where we only use a single field
// but kept for clarity when comparing to the API
type SourceControl struct {
	LibraryOrigin *GitRepo `json:"libraryOrigin"`
}

// LibrarySubjectKind represents a kind of a subject.
type LibrarySubjectKind string

const (
	// LibrarySubjectKindUser is the subject kind user.
	LibrarySubjectKindUser LibrarySubjectKind = "user"

	// LibrarySubjectKindGroup is the subject kind group.
	LibrarySubjectKindGroup LibrarySubjectKind = "group"
)

// LibrarySubject represents a subject which has been granted access to the Library.
// The subject is assigned to the LibraryViewer role.
type LibrarySubject struct {
	// Kind is the LibrarySubjectKind of the subject.
	//+kubebuilder:validation:Enum=user;group
	Kind LibrarySubjectKind `json:"kind,omitempty"`

	// Name is the name of the subject. The meaning of this field depends on the
	// SubjectKind.
	Name string `json:"name"`
}

// IsUser returns whether or not the kind of the subject is a user.
func (subject LibrarySubject) IsUser() bool {
	return subject.Kind == LibrarySubjectKindUser || subject.Kind == ""
}

// GitRepo defines the Git configurations a library can be defined by
type GitRepo struct {
	// Path is the path in the git repo where the policies are located.
	Path string `json:"path,omitempty"`

	// Reference is used to point to a tag or branch. This will be ignored if
	// `Commit` is specified.
	Reference string `json:"reference,omitempty"`

	// Commit is used to point to a specific commit SHA. This takes precedence
	// over `Reference` if both are specified.
	Commit string `json:"commit,omitempty"`

	// URL is the URL of the git repo.
	URL string `json:"url"`
}

// LibrarySecretRef defines how to access a k8s secret for the library.
type LibrarySecretRef struct {
	// Namespace is the namespace where the secret resides.
	Namespace string `json:"namespace"`
	// Name is the name of the secret.
	Name string `json:"name"`
}

// LibraryStatus defines the observed state of Library
type LibraryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Library is the Schema for the libraries API
type Library struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LibrarySpec   `json:"spec,omitempty"`
	Status LibraryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LibraryList contains a list of Library
type LibraryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Library `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Library{}, &LibraryList{})
}
