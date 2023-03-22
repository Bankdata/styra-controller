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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/bankdata/styra-controller/pkg/ptr"
)

var globalDatasourceWebhookLog = logf.Log.WithName("globaldatasource-resource")

// SetupWebhookWithManager sets up the GlobalDatasource webhooks with the Manager.
func (r *GlobalDatasource) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-styra-bankdata-dk-v1alpha1-globaldatasource,mutating=true,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=globaldatasources,verbs=create;update,versions=v1alpha1,name=mglobaldatasource.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &GlobalDatasource{}

// Default implements webhook.Defaulter so that a webhook can be registered for the type.
func (r *GlobalDatasource) Default() {
	globalDatasourceWebhookLog.Info("default", "name", r.Name)

	if r.Spec.DeletionProtection == nil {
		r.Spec.DeletionProtection = ptr.Bool(true)
	}
	if r.Spec.Enabled == nil {
		r.Spec.Enabled = ptr.Bool(true)
	}

	switch r.Spec.Category {
	case GlobalDatasourceCategoryGitRego:
		r.defaultGit()
	}
}

func (r *GlobalDatasource) defaultGit() {
	if r.Spec.Commit == "" && r.Spec.Reference == "" {
		r.Spec.Reference = "refs/heads/master"
	}
}

//+kubebuilder:webhook:path=/validate-styra-bankdata-dk-v1alpha1-globaldatasource,mutating=false,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=globaldatasources,verbs=create;update,versions=v1alpha1,name=vglobaldatasource.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GlobalDatasource{}

// ValidateCreate implements webhook.Validator so that a webhook can be registered for the type.
func (r *GlobalDatasource) ValidateCreate() error {
	globalDatasourceWebhookLog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so that a webhook can be registered for the type.
func (r *GlobalDatasource) ValidateUpdate(_ runtime.Object) error {
	globalDatasourceWebhookLog.Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so that a webhook can be registered for the type.
func (r *GlobalDatasource) ValidateDelete() error {
	globalDatasourceWebhookLog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *GlobalDatasource) validate() error {
	var errs field.ErrorList

	switch r.Spec.Category {
	case GlobalDatasourceCategoryGitRego:
		errs = append(errs, r.Spec.validateGitRegoFields()...)
	}

	if len(errs) > 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: "GlobalDatasource"},
			r.Name,
			errs,
		)
	}

	return nil
}

func (s *GlobalDatasourceSpec) validateGitRegoFields() field.ErrorList {
	var errs field.ErrorList

	if s.Commit != "" && s.Reference != "" {
		errs = append(
			errs,
			field.Invalid(
				field.NewPath("spec").Child("reference"),
				s.Reference,
				"reference can not be set when commit is specified",
			),
		)
	}

	if s.URL == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("url"), "category git/rego requires url"))
	}

	return errs
}
