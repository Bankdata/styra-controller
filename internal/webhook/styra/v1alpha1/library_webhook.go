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

// Package v1alpha1 contains webhook code for version v1alpha1
package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
)

// nolint:all
// log is for logging in this package.
var librarylog = logf.Log.WithName("library-resource")

// SetupLibraryWebhookWithManager registers the webhook for Library in the manager.
func SetupLibraryWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&styrav1alpha1.Library{}).
		WithValidator(&LibraryCustomValidator{}).
		WithDefaulter(&LibraryCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// nolint:all
// +kubebuilder:webhook:path=/mutate-styra-bankdata-dk-v1alpha1-library,mutating=true,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=libraries,verbs=create;update,versions=v1alpha1,name=mlibrary-v1alpha1.kb.io,admissionReviewVersions=v1

// LibraryCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Library when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type LibraryCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &LibraryCustomDefaulter{}

// nolint:all
// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Library.
func (d *LibraryCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	library, ok := obj.(*styrav1alpha1.Library)

	if !ok {
		return fmt.Errorf("expected an Library object but got %T", obj)
	}
	librarylog.Info("Defaulting for Library", "name", library.GetName())

	if library.Spec.SourceControl == nil || library.Spec.SourceControl.LibraryOrigin == nil {
		return nil
	}

	if library.Spec.SourceControl.LibraryOrigin.Commit != "" && library.Spec.SourceControl.LibraryOrigin.Reference != "" {
		library.Spec.SourceControl.LibraryOrigin.Reference = ""
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.

// nolint:all
// +kubebuilder:webhook:path=/validate-styra-bankdata-dk-v1alpha1-library,mutating=false,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=libraries,verbs=create;update,versions=v1alpha1,name=vlibrary-v1alpha1.kb.io,admissionReviewVersions=v1

// LibraryCustomValidator struct is responsible for validating the Library resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type LibraryCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &LibraryCustomValidator{}

// nolint:all
// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Library.
func (v *LibraryCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	library, ok := obj.(*styrav1alpha1.Library)
	if !ok {
		return nil, fmt.Errorf("expected a Library object but got %T", obj)
	}
	librarylog.Info("Validation for Library upon creation", "name", library.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// nolint:all
// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Library.
func (v *LibraryCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	library, ok := newObj.(*styrav1alpha1.Library)
	if !ok {
		return nil, fmt.Errorf("expected a Library object for the newObj but got %T", newObj)
	}
	librarylog.Info("Validation for Library upon update", "name", library.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// nolint:all
// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Library.
func (v *LibraryCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	library, ok := obj.(*styrav1alpha1.Library)
	if !ok {
		return nil, fmt.Errorf("expected a Library object but got %T", obj)
	}
	librarylog.Info("Validation for Library upon deletion", "name", library.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
