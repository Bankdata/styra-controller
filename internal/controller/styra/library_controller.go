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

package styra

import (
	"context"
	"fmt"
	"strings"

	"github.com/bankdata/styra-controller/internal/predicate"
	"github.com/bankdata/styra-controller/internal/sentry"
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/pkg/ocp"
)

// LibraryReconciler reconciles a Library object
type LibraryReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	OCP           ocp.ClientInterface
	Config        *configv2alpha2.ProjectConfig
	WebhookClient webhook.Client
}

//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=libraries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=libraries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=libraries/finalizers,verbs=update

// Reconcile implements renconcile.Renconciler and has responsibility of
// ensuring that the current state of the Library resource renconciled
// towards the desired state.
func (r *LibraryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciliation of libraries begins")

	var k8sLib styrav1alpha1.Library
	if err := r.Get(ctx, req.NamespacedName, &k8sLib); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Could not find Library")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "Could not get Library")
	}

	if !k8sLib.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info(fmt.Sprintf("Library %s is under deletion in k8s. Ignoring it.", k8sLib.Spec.Name))
		return ctrl.Result{}, nil
	}

	log.Info("OPA Control Plane library reconcile starting")
	return r.ocpReconcile(ctx, log, k8sLib)
}

func (r *LibraryReconciler) ocpReconcile(
	ctx context.Context,
	log logr.Logger,
	k8sLib styrav1alpha1.Library) (ctrl.Result, error) {
	log.Info("Reconciling Library")

	reconcileLibrarySourceResult, err := r.reconcileLibrarySource(ctx, log, k8sLib)
	if err != nil {
		return reconcileLibrarySourceResult, err
	}

	log.Info("Reconciliation completed")
	return ctrl.Result{}, nil
}

func (r *LibraryReconciler) reconcileLibrarySource(
	ctx context.Context,
	log logr.Logger,
	k8sLib styrav1alpha1.Library) (ctrl.Result, error) {
	gitConfig := &ocp.GitConfig{
		Repo:          k8sLib.Spec.SourceControl.LibraryOrigin.URL,
		IncludedFiles: []string{"*.rego"},
		ExcludedFiles: []string{"*_test.rego"},
		Path:          ".",
	}
	if k8sLib.Spec.SourceControl.LibraryOrigin.Path != "" {
		gitConfig.Path = k8sLib.Spec.SourceControl.LibraryOrigin.Path
	}
	if k8sLib.Spec.SourceControl.LibraryOrigin.Commit != "" {
		gitConfig.Commit = k8sLib.Spec.SourceControl.LibraryOrigin.Commit
	}
	if k8sLib.Spec.SourceControl.LibraryOrigin.Reference != "" {
		gitConfig.Reference = k8sLib.Spec.SourceControl.LibraryOrigin.Reference
	}

	gitCredentialFound := false
	for _, cred := range r.Config.OPAControlPlaneConfig.GitCredentials {
		if strings.Contains(k8sLib.Spec.SourceControl.LibraryOrigin.URL, cred.RepoPrefix) {
			gitConfig.CredentialID = cred.ID
			gitCredentialFound = true
			break
		}
	}
	if !gitCredentialFound {
		return ctrl.Result{}, fmt.Errorf(
			"createLibrarySource: Unsupported git repository: %s",
			k8sLib.Spec.SourceControl.LibraryOrigin.URL,
		)
	}

	_, err := r.OCP.PutSource(ctx, k8sLib.Spec.Name, &ocp.PutSourceRequest{
		Name: k8sLib.Spec.Name,
		Git:  gitConfig,
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "createLibrarySource: could not create or update source in OCP")
	}
	log.Info("OCP source upserted", "source", k8sLib.Spec.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LibraryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	p, err := predicate.ControllerClass(r.Config.ControllerClass)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&styrav1alpha1.Library{}, builder.WithPredicates(p)).
		Complete(sentry.Decorate(r))
}
