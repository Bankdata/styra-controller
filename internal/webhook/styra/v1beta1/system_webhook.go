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

// Package v1beta1 contains webhook code for version v1beta1
package v1beta1

import (
	"context"
	"fmt"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
)

// nolint:all
// log is for logging in this package.
var systemlog = logf.Log.WithName("system-resource")

// SetupSystemWebhookWithManager registers the webhook for System in the manager.
func SetupSystemWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&styrav1beta1.System{}).
		WithValidator(&SystemCustomValidator{}).
		WithDefaulter(&SystemCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// nolint:all
// +kubebuilder:webhook:path=/mutate-styra-bankdata-dk-v1beta1-system,mutating=true,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=systems,verbs=create;update,versions=v1beta1,name=msystem-v1beta1.kb.io,admissionReviewVersions=v1

// SystemCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind System when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type SystemCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &SystemCustomDefaulter{}

// nolint:all
// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind System.
func (d *SystemCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	system, ok := obj.(*styrav1beta1.System)

	if !ok {
		return fmt.Errorf("expected an System object but got %T", obj)
	}
	systemlog.Info("Defaulting for System", "name", system.GetName())

	if system.Spec.SourceControl != nil {
		if system.Spec.SourceControl.Origin.Reference == "" && system.Spec.SourceControl.Origin.Commit == "" {
			system.Spec.SourceControl.Origin.Reference = "refs/heads/master"
		}
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.

// nolint:all
// +kubebuilder:webhook:path=/validate-styra-bankdata-dk-v1beta1-system,mutating=false,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=systems,verbs=create;update,versions=v1beta1,name=vsystem-v1beta1.kb.io,admissionReviewVersions=v1

// SystemCustomValidator struct is responsible for validating the System resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SystemCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &SystemCustomValidator{}

// nolint:all
// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type System.
func (v *SystemCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	system, ok := obj.(*styrav1beta1.System)
	if !ok {
		return nil, fmt.Errorf("expected a System object but got %T", obj)
	}
	systemlog.Info("Validation for System upon creation", "name", system.GetName())

	return validateSystem(system)
}

// nolint:all
// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type System.
func (v *SystemCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	system, ok := newObj.(*styrav1beta1.System)
	if !ok {
		return nil, fmt.Errorf("expected a System object for the newObj but got %T", newObj)
	}
	systemlog.Info("Validation for System upon update", "name", system.GetName())

	return validateSystem(system)
}

// nolint:all
// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type System.
func (v *SystemCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	system, ok := obj.(*styrav1beta1.System)
	if !ok {
		return nil, fmt.Errorf("expected a System object but got %T", obj)
	}
	systemlog.Info("Validation for System upon deletion", "name", system.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

func validateSystem(s *styrav1beta1.System) (admission.Warnings, error) {
	var errs field.ErrorList

	errs = append(errs, validateSystemSpec(&s.Spec, field.NewPath("spec"))...)

	if len(errs) > 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: styrav1beta1.GroupVersion.Group, Kind: "System"},
			s.Name,
			errs,
		)
	}

	return nil, nil
}

func validateSystemSpec(s *styrav1beta1.SystemSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, validateDecisionMappings(s, path.Child("decisionMappings"))...)
	errs = append(errs, validateDatasources(s, path.Child("datasources"))...)

	return errs
}

func validateDecisionMappings(s *styrav1beta1.SystemSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	for i, dm := range s.DecisionMappings {
		errs = append(errs, validateDecisionMapping(&dm, path.Index(i))...)
	}

	errs = append(errs, validateDecisionMappingNames(s, path)...)

	return errs
}

func validateDecisionMapping(dm *styrav1beta1.DecisionMapping, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if dm.Allowed != nil {
		errs = append(errs, validateDMAllowed(dm.Allowed, path.Child("allowed"))...)
	}
	errs = append(errs, validateColumnKeys(dm, path.Child("columns"))...)

	return errs
}

func validateDecisionMappingNames(s *styrav1beta1.SystemSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	idxsByName := map[string][]int{}

	for i, dm := range s.DecisionMappings {
		idxsByName[dm.Name] = append(idxsByName[dm.Name], i)
	}

	names := make([]string, 0, len(idxsByName))
	for name := range idxsByName {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		idxs := idxsByName[name]
		if len(idxs) > 1 {
			for _, idx := range idxs {
				errs = append(errs, field.Duplicate(path.Index(idx).Child("name"), name))
			}
		}
	}

	return errs
}

func validateColumnKeys(dm *styrav1beta1.DecisionMapping, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	idxsByKey := map[string][]int{}
	for i, col := range dm.Columns {
		idxsByKey[col.Key] = append(idxsByKey[col.Key], i)
	}

	keys := make([]string, 0, len(idxsByKey))
	for key := range idxsByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		idxs := idxsByKey[key]
		if len(idxs) > 1 {
			for _, idx := range idxs {
				errs = append(errs, field.Duplicate(path.Index(idx).Child("key"), key))
			}
		}
	}

	return errs
}

func validateDMAllowed(a *styrav1beta1.AllowedMapping, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if a.Expected != nil {
		errs = append(errs, validateExpected(a.Expected, path.Child("expected"))...)
	}

	return errs
}

func validateExpected(e *styrav1beta1.Expected, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if e.Boolean == nil && e.Integer == nil && e.String == nil ||
		e.Boolean != nil && e.Integer == nil && e.String == nil ||
		e.Boolean == nil && e.Integer != nil && e.String == nil ||
		e.Boolean == nil && e.Integer == nil && e.String != nil {
		return nil
	}

	if e.Boolean != nil {
		errs = append(errs, field.Forbidden(path.Child("boolean"), "only one of boolean, string or int should be set"))
	}
	if e.Integer != nil {
		errs = append(errs, field.Forbidden(path.Child("integer"), "only one of boolean, string or int should be set"))
	}
	if e.String != nil {
		errs = append(errs, field.Forbidden(path.Child("string"), "only one of boolean, string or int should be set"))
	}

	return errs
}

func validateDatasources(s *styrav1beta1.SystemSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	idxsByPath := map[string][]int{}

	for i, ds := range s.Datasources {
		idxsByPath[ds.Path] = append(idxsByPath[ds.Path], i)
	}

	paths := make([]string, 0, len(idxsByPath))
	for name := range idxsByPath {
		paths = append(paths, name)
	}
	sort.Strings(paths)

	for _, name := range paths {
		idxs := idxsByPath[name]
		if len(idxs) > 1 {
			for _, idx := range idxs {
				errs = append(errs, field.Duplicate(path.Index(idx).Child("path"), name))
			}
		}
	}

	return errs
}
