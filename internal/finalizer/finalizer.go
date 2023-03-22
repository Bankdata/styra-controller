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

// Package finalizer contains helpers for working with the controller finalizer.
package finalizer

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const name = "styra.bankdata.dk/finalizer"

// IsSet returns true if an object has the "styra.bankdata.dk/finalizer" finalizer.
func IsSet(o client.Object) bool {
	return controllerutil.ContainsFinalizer(o, name)
}

// Add adds the "styra.bankdata.dk/finalizer" finalizer to an object.
func Add(o client.Object) {
	controllerutil.AddFinalizer(o, name)
}

// Remove removes the "styra.bankdata.dk/finalizer" finalizer from an object.
func Remove(o client.Object) {
	controllerutil.RemoveFinalizer(o, name)
}
