/*
Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

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

// GlobalDatasourceCategory represents a datasource category.
// +kubebuilder:validation:Enum=git/rego
type GlobalDatasourceCategory string

const (
	// GlobalDatasourceCategoryGitRego represents the git/rego datasource category.
	GlobalDatasourceCategoryGitRego GlobalDatasourceCategory = "git/rego"
)

// GlobalDatasourceSpec is the specification of the GlobalDatasource.
type GlobalDatasourceSpec struct {
	// Category is the datasource category. For more information about
	// supported categories see the available GlobalDatasourceCategory
	// constants in the package.
	Category GlobalDatasourceCategory `json:"category"`
	// Description describes the datasource.
	Description string `json:"description,omitempty"`
	// Enabled toggles whether or not the datasource should be enabled.
	Enabled *bool `json:"enabled,omitempty"`
	// PollingInterval sets the interval for when the datasource should be refreshed.
	PollingInterval string `json:"pollingInterval,omitempty"`
	// Commit is a commit SHA for the git/xx datasources. If `Reference` and this
	// is set, this takes precedence.
	Commit string `json:"commit,omitempty"`
	// CredentialsSecretRef references a secret with keys `name` and `secret`
	// which will be used for authentication.
	CredentialsSecretRef *GlobalDatasourceSecretRef `json:"credentialsSecretRef,omitempty"`
	// Reference is a git reference for the git/xx datasources.
	Reference string `json:"reference,omitempty"`
	// URL is used in http and git/xx datasources.
	URL string `json:"url,omitempty"`
	// DeletionProtection skips deleting the datasource in Styra when the
	// `GlobalDatasource` resource is deleted.
	DeletionProtection *bool `json:"deletionProtection,omitempty"`
	// Path is the path in git in git/xx datasources.
	Path string `json:"path,omitempty"`
}

// GlobalDatasourceSecretRef represents a reference to a secret.
type GlobalDatasourceSecretRef struct {
	// Namespace is the namespace where the secret resides.
	Namespace string `json:"namespace"`
	// Name is the name of the secret.
	Name string `json:"name"`
}

// GlobalDatasourceStatus holds the status of the GlobalDatasource resource.
type GlobalDatasourceStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// GlobalDatasource is a resource used for creating global datasources in
// Styra. These datasources are available across systems and can be used for
// shared data. GlobalDatasource can also be used to create libraries by using
// the GlobalDatasourceCategoryGitRego category.
type GlobalDatasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalDatasourceSpec   `json:"spec,omitempty"`
	Status GlobalDatasourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GlobalDatasourceList represents a list of GlobalDatasource resources.
type GlobalDatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalDatasource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalDatasource{}, &GlobalDatasourceList{})
}
