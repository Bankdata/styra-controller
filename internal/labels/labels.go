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

// Package labels contains helpers for working with labels.
package labels

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Constants for labels configured on System resources.
const (
	labelControllerClass = "styra-controller/class"
	labelManagedBy       = "app.kubernetes.io/managed-by"
	labelValueManagedBy  = "styra-controller"
	LabelControlPlane    = "styra-controller/control-plane"
)

// ControllerClassLabelSelector creates a metav1.LabelSelector which selects
// objects that has the "styra-controller/class" label with the value `class`.
func ControllerClassLabelSelector(class string) metav1.LabelSelector {
	var selector metav1.LabelSelector
	if class != "" {
		selector = metav1.LabelSelector{
			MatchLabels: map[string]string{
				labelControllerClass: class,
			},
		}
	} else {
		selector = metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{{
				Key:      labelControllerClass,
				Operator: metav1.LabelSelectorOpDoesNotExist,
			}},
		}
	}
	return selector
}

// ControllerClassLabelSelectorAsSelector creates a labels.Selecter which
// selects objects that has the "styra-controller/class" label with the value `class`.
func ControllerClassLabelSelectorAsSelector(class string) (labels.Selector, error) {
	ls := ControllerClassLabelSelector(class)
	return metav1.LabelSelectorAsSelector(&ls)
}

// SetManagedBy sets the `app.kubernetes.io/managed-by` label to
// styra-controller.
func SetManagedBy(o client.Object) {
	labels := o.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[labelManagedBy] = labelValueManagedBy
	o.SetLabels(labels)
}

// HasManagedBy checks if the object has the label `app.kubernetes.io/managed-by`
// set to styra-controller
func HasManagedBy(o client.Object) bool {
	managedBy, ok := o.GetLabels()[labelManagedBy]
	return ok && managedBy == labelValueManagedBy
}

// ControllerClassMatches checks if the object has the `styra-controller/class` label
// with the value `class`.
func ControllerClassMatches(o client.Object, class string) bool {
	labels := o.GetLabels()
	if labels == nil {
		return class == ""
	}
	return labels[labelControllerClass] == class
}
