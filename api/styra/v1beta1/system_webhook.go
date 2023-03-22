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

package v1beta1

import (
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

// log is for logging in this package.
var systemlog = logf.Log.WithName("system-resource")

// SetupWebhookWithManager registers the System webhooks with the Manager.
func (s *System) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-styra-bankdata-dk-v1beta1-system,mutating=true,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=systems,verbs=create;update,versions=v1beta1,name=msystem.kb.io,admissionReviewVersions=v1

// Default implements webhook.Defaulter so that a webhook can be registered for the type.
var _ webhook.Defaulter = &System{}

// Default implements webhook.Defaulter so that a webhook can be registered for the type.
func (s *System) Default() {
	systemlog.Info("default", "name", s.Name)

	if s.Spec.SourceControl != nil {
		if s.Spec.SourceControl.Origin.Reference == "" && s.Spec.SourceControl.Origin.Commit == "" {
			s.Spec.SourceControl.Origin.Reference = "refs/heads/master"
		}
	}

	if s.Spec.DeletionProtection == nil {
		s.Spec.DeletionProtection = ptr.Bool(true)
	}
}

//+kubebuilder:webhook:path=/validate-styra-bankdata-dk-v1beta1-system,mutating=false,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=systems,verbs=create;update,versions=v1beta1,name=vsystem.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &System{}

// ValidateCreate implements webhook.Validator so that a webhook can be registered for the type.
func (s *System) ValidateCreate() error {
	systemlog.Info("validate create", "name", s.Name)

	return s.validate()
}

// ValidateUpdate implements webhook.Validator so that a webhook will be registered for the type.
func (s *System) ValidateUpdate(_ runtime.Object) error {
	systemlog.Info("validate update", "name", s.Name)

	return s.validate()
}

// ValidateDelete implements webhook.Validator so that a webhook will be registered for the type.
func (s *System) ValidateDelete() error {
	systemlog.Info("validate delete", "name", s.Name)

	return nil
}

func (s *System) validate() error {
	var errs field.ErrorList

	errs = append(errs, s.Spec.validate(field.NewPath("spec"))...)

	if len(errs) > 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: "System"},
			s.Name,
			errs,
		)
	}

	return nil
}

func (s *SystemSpec) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, s.validateDecisionMappings(path.Child("decisionMappings"))...)
	errs = append(errs, s.validateDatasources(path.Child("datasources"))...)

	return errs
}

func (s *SystemSpec) validateDecisionMappings(path *field.Path) field.ErrorList {
	var errs field.ErrorList

	for i, dm := range s.DecisionMappings {
		errs = append(errs, dm.validate(path.Index(i))...)
	}

	errs = append(errs, s.validateDecisionMappingNames(path)...)

	return errs
}

func (dm DecisionMapping) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if dm.Allowed != nil {
		errs = append(errs, dm.Allowed.validate(path.Child("allowed"))...)
	}
	errs = append(errs, dm.validateColumnKeys(path.Child("columns"))...)

	return errs
}

func (s *SystemSpec) validateDecisionMappingNames(path *field.Path) field.ErrorList {
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

func (dm DecisionMapping) validateColumnKeys(path *field.Path) field.ErrorList {
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

func (a *AllowedMapping) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if a.Expected != nil {
		errs = append(errs, a.Expected.validate(path.Child("expected"))...)
	}

	return errs
}

func (e Expected) validate(path *field.Path) field.ErrorList {
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

func (s *SystemSpec) validateDatasources(path *field.Path) field.ErrorList {
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
