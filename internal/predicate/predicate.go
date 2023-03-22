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

// Package predicate contains predicates used by the controllers.
package predicate

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/bankdata/styra-controller/internal/labels"
)

// ControllerClass creates a predicate which ensures that we only reconcile
// resources that match the controller class label selector
// labels.ControllerClassLabelSelector.
func ControllerClass(class string) (predicate.Predicate, error) {
	labelSelector := labels.ControllerClassLabelSelector(class)
	predicate, err := predicate.LabelSelectorPredicate(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "could not create LabelSelectorPredicate")
	}
	return predicate, nil
}
