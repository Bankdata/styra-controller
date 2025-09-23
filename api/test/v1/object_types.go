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

package v1

// +kubebuilder:skip

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Object is a very simple kubernetes object which doesn't have a spec or
// status. It only has type and object metadata. This type is useful for
// testing reconciliation predicates.
// +kubebuilder:object:root=true
type Object struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Object{})
}
