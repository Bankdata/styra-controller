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

package styra

import (
	"context"
	"fmt"
	"net/http"
	"path"

	ctrlerr "github.com/bankdata/styra-controller/internal/errors"
	"github.com/bankdata/styra-controller/internal/predicate"
	"github.com/bankdata/styra-controller/internal/sentry"
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/bankdata/styra-controller/pkg/styra"
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
)

// LibraryReconciler reconciles a Library object
type LibraryReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Styra         styra.ClientInterface
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
	log.Info("Reconciliation begins")

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

	log.Info("Reconciling git credentials from default credentials")
	gitCredential := r.Config.GetGitCredentialForRepo(k8sLib.Spec.SourceControl.LibraryOrigin.URL)
	if gitCredential == nil {
		log.Info("Could not find matching credentials", "url", k8sLib.Spec.SourceControl.LibraryOrigin.URL)
	} else {
		_, err := r.Styra.CreateUpdateSecret(
			ctx,
			path.Join("libraries", k8sLib.Spec.Name, "git"),
			&styra.CreateUpdateSecretsRequest{
				Name:   gitCredential.User,
				Secret: gitCredential.Password,
			},
		)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not update Styra secret")
		}
	}

	log.Info("Reconciling Library")
	update := false
	libResp, err := r.Styra.GetLibrary(ctx, k8sLib.Spec.Name)
	if err != nil {
		var httpErr *styra.HTTPError
		if errors.As(err, &httpErr) {
			if httpErr.StatusCode == http.StatusNotFound {
				update = true
			}
		}
	} else if r.needsUpdate(&k8sLib, libResp.LibraryEntityExpanded) {
		update = true
	}

	if update {
		log.Info("UpsertLibrary")
		_, err := r.Styra.UpsertLibrary(ctx, k8sLib.Spec.Name, r.specToUpdate(&k8sLib))
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not upsert library")
		}
	}

	if libResp == nil {
		// This is often the case when a library has just been created.
		log.Info(fmt.Sprint("Library ", k8sLib.Spec.Name, " could not be fetched. Requeueing..."))
		return ctrl.Result{Requeue: true}, nil
	}

	if result, err := r.reconcileDatasources(ctx, log, &k8sLib, libResp.LibraryEntityExpanded); err != nil {
		return result, ctrlerr.Wrap(err, "Could not reconcile library datasources")
	}

	if result, err := r.reconcileSubjects(ctx, log, &k8sLib); err != nil {
		return result, err
	}

	log.Info("Reconciliation completed")
	return ctrl.Result{}, nil
}

func (r *LibraryReconciler) specToUpdate(k8sLib *styrav1alpha1.Library) *styra.UpsertLibraryRequest {
	if k8sLib == nil {
		return nil
	}
	specs := k8sLib.Spec
	k8sSourceControl := specs.SourceControl.LibraryOrigin

	sourceControl := styra.LibraryGitRepoConfig{
		Commit:      k8sSourceControl.Commit,
		Credentials: path.Join("libraries", specs.Name, "git"),
		Path:        k8sSourceControl.Path,
		Reference:   k8sSourceControl.Reference,
		URL:         k8sSourceControl.URL,
	}

	req := styra.UpsertLibraryRequest{
		Description:   specs.Description,
		ReadOnly:      true,
		SourceControl: &styra.LibrarySourceControlConfig{LibraryOrigin: &sourceControl},
	}

	return &req
}

func (r *LibraryReconciler) needsUpdate(k8sLib *styrav1alpha1.Library, styraLib *styra.LibraryEntityExpanded) bool {
	if k8sLib == nil {
		return false
	}

	specs := k8sLib.Spec
	if styraLib == nil ||
		specs.Name != styraLib.ID ||
		specs.Description != styraLib.Description ||
		!styraLib.ReadOnly ||
		!sameSourceControl(specs.SourceControl, styraLib.SourceControl) {
		return true
	}

	if styraLib.SourceControl.LibraryOrigin.Credentials != path.Join("libraries", k8sLib.Spec.Name, "git") {
		if r.Config.GetGitCredentialForRepo(specs.SourceControl.LibraryOrigin.URL) != nil {
			return true
		}
	}

	return false
}

func sameSourceControl(k8sLib *styrav1alpha1.SourceControl, styraLib *styra.LibrarySourceControlConfig) bool {
	return k8sLib.LibraryOrigin.Path == styraLib.LibraryOrigin.Path &&
		k8sLib.LibraryOrigin.Reference == styraLib.LibraryOrigin.Reference &&
		k8sLib.LibraryOrigin.Commit == styraLib.LibraryOrigin.Commit &&
		k8sLib.LibraryOrigin.URL == styraLib.LibraryOrigin.URL
}

func (r *LibraryReconciler) reconcileDatasources(ctx context.Context, log logr.Logger, k8sLib *styrav1alpha1.Library,
	styraLib *styra.LibraryEntityExpanded) (ctrl.Result, error) {
	log.Info("Reconciling library datasources")

	existingByID := map[string]styra.LibraryDatasourceConfig{}
	if styraLib != nil {
		for _, ds := range styraLib.DataSources {
			if ds.ID != "" {
				existingByID[ds.ID] = ds
			}
		}
	}

	expectedByID := map[string]styrav1alpha1.LibraryDatasource{}
	if k8sLib.Spec.Datasources != nil {
		for _, ds := range k8sLib.Spec.Datasources {
			id := path.Join("libraries", k8sLib.Spec.Name, ds.Path)
			expectedByID[id] = ds
		}
	}

	// Create the missing datasources
	for id := range expectedByID {
		ds, exists := existingByID[id]
		if !exists || ds.Category != "rest" {
			log := log.WithValues("datasourceID", id)
			log.Info("Creating or updating datasource")
			request := &styra.UpsertDatasourceRequest{
				Category: "rest",
				Enabled:  true,
			}
			_, err := r.Styra.UpsertDatasource(ctx, id, request)
			if err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not create or update library datasource")
			}

			if r.WebhookClient != nil {
				log.Info("Calling library datasource changed webhook")
				if err := r.WebhookClient.DatasourceChanged(ctx, log, "jwt-library", ""); err != nil {
					err = ctrlerr.Wrap(err, "could not call 'library datasource changed' webhook")
					log.Error(err, err.Error())
				}
			}
		}
	}

	// Delete the unexpected datasources
	if styraLib == nil {
		return ctrl.Result{}, nil
	}

	for _, ds := range styraLib.DataSources {
		if ds.ID == "" {
			log.Info("There exists some datasource without id?")
			continue
		}

		if _, expected := expectedByID[ds.ID]; !expected {
			log.Info("Deleting undeclared datasource")

			if _, err := r.Styra.DeleteDatasource(ctx, ds.ID); err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete library datasource")
			}
		}
	}

	log.Info("Reconciled datasources")
	return ctrl.Result{}, nil
}

func (r *LibraryReconciler) reconcileSubjects(
	ctx context.Context,
	log logr.Logger,
	k8sLib *styrav1alpha1.Library,
) (ctrl.Result, error) {
	log.Info("Reconciling subjects for library")

	// Make sure all users already exist in Styra, otherwise create them
	if err := r.createUsersIfMissing(ctx, log, k8sLib); err != nil {
		return ctrl.Result{}, err
	}

	// Delete Rolebindings to other roles then LibraryViewer
	if err := r.deleteIncorrectRoleBindings(ctx, log, k8sLib); err != nil {
		return ctrl.Result{}, err
	}

	// Create rolebinding for LibraryViewer if it is missing
	if err := r.createRoleBindingIfMissing(ctx, log, styra.RoleLibraryViewer, k8sLib); err != nil {
		return ctrl.Result{}, err
	}

	// Make sure the LibraryViewer rolebinding contains correct subjects
	if err := r.updateRoleBindingIfNeeded(ctx, log, styra.RoleLibraryViewer, k8sLib); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LibraryReconciler) createUsersIfMissing(
	ctx context.Context,
	log logr.Logger,
	k8sLib *styrav1alpha1.Library) error {
	for _, subject := range k8sLib.Spec.Subjects {
		if subject.IsUser() {
			log := log.WithValues("user", subject.Name)
			log.Info("Checking if user exists")
			res, err := r.Styra.GetUser(ctx, subject.Name)
			if err != nil {
				return ctrlerr.Wrap(err, "Could not get user from Styra API")
			}

			if res.StatusCode == http.StatusNotFound {
				log.Info("User does not exist in styra. Creating...")
				_, err := r.Styra.CreateInvitation(ctx, false, subject.Name)
				if err != nil {
					return ctrlerr.Wrap(err, "Could not create user in Styra")
				}
			}
		}
	}
	return nil
}

func (r *LibraryReconciler) deleteIncorrectRoleBindings(
	ctx context.Context,
	log logr.Logger,
	k8sLib *styrav1alpha1.Library) error {
	log.Info("Deleting Rolebindings to roles that are not LibraryViewers")
	res, err := r.Styra.ListRoleBindingsV2(ctx, &styra.ListRoleBindingsV2Params{
		ResourceKind: styra.RoleBindingKindLibrary,
		ResourceID:   k8sLib.Spec.Name,
	})
	if err != nil {
		return ctrlerr.Wrap(err, "Could not get rolebindings for Library in Styra")
	}

	for _, styraRB := range res.Rolebindings {
		if styraRB.RoleID != styra.Role("LibraryViewer") {
			if _, err := r.Styra.DeleteRoleBindingV2(ctx, styraRB.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *LibraryReconciler) createRoleBindingIfMissing(
	ctx context.Context,
	log logr.Logger,
	role styra.Role,
	k8sLib *styrav1alpha1.Library) error {
	res, err := r.Styra.ListRoleBindingsV2(ctx, &styra.ListRoleBindingsV2Params{
		ResourceKind: styra.RoleBindingKindLibrary,
		ResourceID:   k8sLib.Spec.Name,
	})
	if err != nil {
		return ctrlerr.Wrap(err, "Could not get rolebindings for Library in Styra")
	}

	if len(res.Rolebindings) == 0 {
		log.Info(fmt.Sprintf("No rolebindings exist for the %s Library. Creating rolebinding.", k8sLib.Spec.Name))

		k8sRolebindingSubjects := createLibraryRolebindingSubjects(
			k8sLib.Spec.Subjects,
			r.Config.SSO,
		)

		err := r.createRoleBinding(ctx, log, k8sLib, role, k8sRolebindingSubjects)
		if err != nil {
			return ctrlerr.Wrap(err, "Could not create rolebinding in Styra")
		}
	}
	return nil
}

func (r *LibraryReconciler) updateRoleBindingIfNeeded(
	ctx context.Context,
	log logr.Logger,
	role styra.Role,
	k8sLib *styrav1alpha1.Library) error {
	res, err := r.Styra.ListRoleBindingsV2(ctx, &styra.ListRoleBindingsV2Params{
		ResourceKind: styra.RoleBindingKindLibrary,
		ResourceID:   k8sLib.Spec.Name,
	})
	if err != nil {
		return ctrlerr.Wrap(err, "Could not get rolebindings for Library in Styra")
	}

	k8sRolebindingSubjects := createLibraryRolebindingSubjects(
		k8sLib.Spec.Subjects,
		r.Config.SSO,
	)

	// res.Rolebindings should only contain one rolebinding, for the "LibraryViewer" role
	// Also, should contain nothing else after deleteIncorrectRoleBindings
	for _, rb := range res.Rolebindings {
		if rb.RoleID == role {
			if !styra.SubjectsAreEqual(k8sRolebindingSubjects, rb.Subjects) {
				if err := r.updateRoleBindingSubjects(ctx, log, rb.ID, k8sRolebindingSubjects); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *LibraryReconciler) updateRoleBindingSubjects(
	ctx context.Context,
	log logr.Logger,
	roleBindingID string,
	subjects []*styra.Subject,
) error {
	log.Info("Updating rolebinding")

	_, err := r.Styra.UpdateRoleBindingSubjects(
		ctx,
		roleBindingID,
		&styra.UpdateRoleBindingSubjectsRequest{Subjects: subjects},
	)
	if err != nil {
		return ctrlerr.Wrap(err, "Could not update Styra role binding")
	}

	log.Info("Updated rolebinding")
	return nil
}

func createLibraryRolebindingSubjects(
	subjects []styrav1alpha1.LibrarySubject,
	sso *configv2alpha2.SSOConfig,
) []*styra.Subject {
	styraSubjectsByUserID := map[string]struct{}{}
	styraSubjectsByClaimValue := map[string]struct{}{}

	styraSubjects := []*styra.Subject{}

	for _, subject := range subjects {
		if subject.IsUser() {
			if _, ok := styraSubjectsByUserID[subject.Name]; ok {
				continue
			}
			styraSubjects = append(styraSubjects, &styra.Subject{
				Kind: styra.SubjectKindUser,
				ID:   subject.Name,
			})
			styraSubjectsByUserID[subject.Name] = struct{}{}

		} else if subject.Kind == styrav1alpha1.LibrarySubjectKindGroup && sso != nil {
			if _, ok := styraSubjectsByClaimValue[subject.Name]; ok {
				continue
			}

			styraSubjects = append(styraSubjects, &styra.Subject{
				Kind: styra.SubjectKindClaim,
				ClaimConfig: &styra.ClaimConfig{
					IdentityProvider: sso.IdentityProvider,
					Key:              sso.JWTGroupsClaim,
					Value:            subject.Name,
				},
			})
			styraSubjectsByClaimValue[subject.Name] = struct{}{}
		}
	}

	return styraSubjects
}

func (r *LibraryReconciler) createRoleBinding(
	ctx context.Context,
	log logr.Logger,
	k8sLib *styrav1alpha1.Library,
	role styra.Role,
	subjects []*styra.Subject,
) error {
	log.Info("Creating rolebinding")

	if _, err := r.Styra.CreateRoleBinding(ctx, &styra.CreateRoleBindingRequest{
		ResourceFilter: &styra.ResourceFilter{
			ID:   k8sLib.Spec.Name,
			Kind: styra.RoleBindingKindLibrary,
		},
		RoleID:   role,
		Subjects: subjects,
	}); err != nil {
		return ctrlerr.Wrap(err, "Could not create rolebinding")
	}

	log.Info("Created rolebinding")
	return nil
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
