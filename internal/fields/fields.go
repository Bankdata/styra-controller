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

// Package fields contains helpers for working with fields in the CRDs. These
// are mostly used when setting up field indexers.
package fields

import "k8s.io/apimachinery/pkg/fields"

const (
	// SystemCredentialsSecretName is the path to the credential secret name.
	SystemCredentialsSecretName = ".spec.sourceControl.origin.credentialsSecretName"
)

// SystemCredentialsSecretNameSelector returns a field selector for finding the
// Systems referencing a secret.
func SystemCredentialsSecretNameSelector(name string) fields.Selector {
	return fields.OneTermEqualSelector(SystemCredentialsSecretName, name)
}
