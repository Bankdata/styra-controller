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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var librarylog = logf.Log.WithName("library-resource")

// SetupWebhookWithManager sets up the Library webhooks with the Manager.
func (r *Library) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-styra-bankdata-dk-v1alpha1-library,mutating=true,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=libraries,verbs=create;update,versions=v1alpha1,name=mlibrary.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Library{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Library) Default() {
	librarylog.Info("default", "name", r.Name)

	if r.Spec.SourceControl.LibraryOrigin.Commit != "" && r.Spec.SourceControl.LibraryOrigin.Reference != "" {
		r.Spec.SourceControl.LibraryOrigin.Reference = ""
	}

	if r.Spec.Datasources == nil {
		r.Spec.Datasources = []LibraryDatasource{}
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-styra-bankdata-dk-v1alpha1-library,mutating=false,failurePolicy=fail,sideEffects=None,groups=styra.bankdata.dk,resources=libraries,verbs=create;update,versions=v1alpha1,name=vlibrary.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Library{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateCreate() (admission.Warnings, error) {
	librarylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	librarylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Library) ValidateDelete() (admission.Warnings, error) {
	librarylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
