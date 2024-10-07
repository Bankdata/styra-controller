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
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrlpred "sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/api/styra/v1beta1"
	ctrlerr "github.com/bankdata/styra-controller/internal/errors"
	"github.com/bankdata/styra-controller/internal/fields"
	"github.com/bankdata/styra-controller/internal/finalizer"
	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/internal/labels"
	"github.com/bankdata/styra-controller/internal/predicate"
	"github.com/bankdata/styra-controller/internal/sentry"
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/bankdata/styra-controller/pkg/styra"
)

// SystemReconciler reconciles a System object
type SystemReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Styra         styra.ClientInterface
	WebhookClient webhook.Client
	Recorder      record.EventRecorder
	Metric        *prometheus.GaugeVec
	Config        *configv2alpha2.ProjectConfig
}

//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile implements renconcile.Renconciler and has responsibility of
// ensuring that the current state of the System resource renconciled
// towards the desired state.
func (r *SystemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciliation begins")

	log.Info("Fetching System")
	var system v1beta1.System
	if err := r.Get(ctx, req.NamespacedName, &system); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Could not find System in kubernetes")
			r.deleteMetric(req)
			return ctrl.Result{}, nil
		}
		r.deleteMetric(req)
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch System")
	}

	log = log.WithValues("systemID", system.Status.ID)

	if !labels.ControllerClassMatches(&system, r.Config.ControllerClass) {
		log.Info("This is not a System we are managing. Skipping reconciliation.")
		r.deleteMetric(req)
		return ctrl.Result{}, nil
	}

	var (
		res ctrl.Result
		err error
	)

	if system.ObjectMeta.DeletionTimestamp.IsZero() {
		res, err = r.reconcile(ctx, log, &system)
		r.updateMetric(req, system.Status.Ready)
	} else {
		res, err = r.reconcileDeletion(ctx, log, &system)
		if err != nil {
			r.updateMetric(req, system.Status.Ready)
		} else {
			r.deleteMetric(req)
		}
	}

	if err != nil {
		r.recordErrorEvent(&system, err)
		r.setSystemStatusError(&system, err)

		if err := r.Status().Update(ctx, &system); err != nil {
			return res, errors.Wrap(err, "could not set failure status on System")
		}
	}
	return res, err
}

func (r *SystemReconciler) setSystemStatusError(System *v1beta1.System, err error) {
	System.Status.FailureMessage = err.Error()
	System.Status.Phase = v1beta1.SystemPhaseFailed
	System.Status.Ready = false

	var rerr *ctrlerr.ReconcilerErr
	if errors.As(err, &rerr) {
		if rerr.ConditionType != "" {
			System.SetCondition(v1beta1.ConditionType(rerr.ConditionType), metav1.ConditionFalse)
		}
	}
}

func (r *SystemReconciler) updateMetric(req ctrl.Request, ready bool) {
	if r.Metric == nil {
		return
	}
	var value float64
	if ready {
		value = 1
	}
	r.Metric.WithLabelValues(req.Name, req.Namespace).Set(value)
}

func (r *SystemReconciler) deleteMetric(req ctrl.Request) {
	if r.Metric == nil {
		return
	}
	r.Metric.Delete(prometheus.Labels{"System": req.Name, "namespace": req.Namespace})
}

func (r *SystemReconciler) recordErrorEvent(system *v1beta1.System, err error) {
	var rerr *ctrlerr.ReconcilerErr
	if errors.As(err, &rerr) {
		if rerr.Event != "" {
			r.Recorder.Event(system, corev1.EventTypeWarning, rerr.Event, rerr.Error())
		}
	}
}

func (r *SystemReconciler) reconcileFinalizer(ctx context.Context, log logr.Logger, system *v1beta1.System) error {
	log.Info("Ensuring finalizer is present")
	finalizer.Add(system)
	if err := r.Update(ctx, system); err != nil {
		return ctrlerr.Wrap(err, "Could not set finalizer").
			WithEvent("ErrorSetFinalizer").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}
	return nil
}

func (r *SystemReconciler) reconcileDeletion(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("System deletion is in progress")
	if finalizer.IsSet(system) {
		// finalizer is present so we need to ensure system is deleted in
		// styra, unless deletion protection is enabled
		if system.Status.ID != "" {
			deletionProtected := false
			if system.Spec.DeletionProtection != nil {
				deletionProtected = *system.Spec.DeletionProtection
			} else {
				deletionProtected = r.Config.DeletionProtectionDefault
			}
			if !deletionProtected {
				log.Info("Deleting system in styra")
				_, err := r.Styra.DeleteSystem(ctx, system.Status.ID)
				if err != nil {
					return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete system in styra").
						WithEvent("ErrorDeleteSystemInStyra")
				}
			}
		}

		log.Info("Removing finalizer")
		finalizer.Remove(system)
		if err := r.Update(ctx, system); err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not remove finalizer").
				WithEvent("ErrorRemovingFinalizer")
		}
	}
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("Reconciling spec")

	if !finalizer.IsSet(system) {
		if err := r.reconcileFinalizer(ctx, log, system); err != nil {
			return ctrl.Result{}, err
		}
	}

	var (
		cfg *styra.SystemConfig
		err error
	)

	systemID := system.Status.ID
	migrationID := system.ObjectMeta.Annotations["styra-controller/migration-id"]

	if r.Config.EnableMigrations && systemID == "" && migrationID != "" {
		log.Info(fmt.Sprintf("Use migrationId(%s) to fetch system from Styra DAS", migrationID))

		cfg, err = r.getSystem(ctx, log, migrationID)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := r.reconcileID(ctx, log, system, migrationID); err != nil {
			return ctrl.Result{}, err
		}
	} else if systemID != "" {
		cfg, err = r.getSystem(ctx, log, systemID)
		if err != nil {
			var serr *styra.HTTPError
			if errors.As(err, &serr) && serr.StatusCode == http.StatusNotFound {
				res, err := r.createSystem(ctx, log, system)
				if err != nil {
					return ctrl.Result{}, err
				}
				if err := r.deleteDefaultPolicies(ctx, log, res.SystemConfig.ID); err != nil {
					return ctrl.Result{}, err
				}
				if err := r.reconcileID(ctx, log, system, res.SystemConfig.ID); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		}
	} else {
		displayName := system.DisplayName(r.Config.SystemPrefix, r.Config.SystemSuffix)

		cfg, err = r.getSystemByName(ctx, log, displayName)
		if err != nil {
			return ctrl.Result{}, err
		}
		if cfg != nil {
			if err := r.reconcileID(ctx, log, system, cfg.ID); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			res, err := r.createSystem(ctx, log, system)
			if err != nil {
				return ctrl.Result{}, err
			}
			if err := r.deleteDefaultPolicies(ctx, log, res.SystemConfig.ID); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.reconcileID(ctx, log, system, res.SystemConfig.ID); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	system.SetCondition(v1beta1.ConditionTypeCreatedInStyra, metav1.ConditionTrue)

	if result, err := r.reconcileCredentials(ctx, log, system); err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeGitCredentialsUpdated, metav1.ConditionTrue)

	if r.systemNeedsUpdate(log, system, cfg) {
		if cfg, err = r.updateSystem(ctx, log, system); err != nil {
			return ctrl.Result{}, err
		}
	}
	system.SetCondition(v1beta1.ConditionTypeSystemConfigUpdated, metav1.ConditionTrue)

	if result, err := r.reconcileSubjects(ctx, log, system); err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeSubjectsUpdated, metav1.ConditionTrue)

	if result, err := r.reconcileDatasources(ctx, log, system, cfg); err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeDatasourcesUpdated, metav1.ConditionTrue)

	opaConfig, err := r.Styra.GetOPAConfig(ctx, system.Status.ID)
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get OPA config from styra API").
			WithEvent("ErrorFetchOPAConfig").
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	result, updatedToken, err := r.reconcileOPAToken(ctx, log, system, opaConfig.Token)
	if err != nil {
		return result, err
	}
	if updatedToken {
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPATokenUpdated, metav1.ConditionTrue)

	result, updatedOPAConfigMap, err := r.reconcileOPAConfigMap(ctx, log, system, opaConfig)
	if err != nil {
		return result, err
	}
	if updatedOPAConfigMap {
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPAConfigMapUpdated, metav1.ConditionTrue)

	result, updatedSLPConfigMap, err := r.reconcileSLPConfigMap(ctx, log, system, opaConfig)
	if err != nil {
		return result, err
	}
	if updatedSLPConfigMap {
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeSLPConfigMapUpdated, metav1.ConditionTrue)

	system.Status.Ready = true
	system.Status.Phase = v1beta1.SystemPhaseCreated
	system.Status.FailureMessage = ""
	if err := r.Status().Update(ctx, system); err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not change status.phase to Created").
			WithEvent("ErrorPhaseToCreated")
	}

	msg := "Reconciliation completed"
	r.Recorder.Event(system, corev1.EventTypeNormal, "ReconciliationCompleted", msg)
	log.Info(msg)
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) getSystem(
	ctx context.Context,
	log logr.Logger,
	systemID string,
) (*styra.SystemConfig, error) {
	log.Info("Fetching system from Styra API")

	res, err := r.Styra.GetSystem(ctx, systemID)
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Could not fetch system from Styra API").
			WithEvent("ErrorFetchSystemFromStyra")
	}

	log.Info("Fetched system from Styra API")
	return res.SystemConfig, nil
}

func (r *SystemReconciler) getSystemByName(
	ctx context.Context,
	log logr.Logger,
	name string,
) (*styra.SystemConfig, error) {
	log.Info(fmt.Sprintf("Fetching system %v from Styra API if it exists", name))

	res, err := r.Styra.GetSystemByName(ctx, name)
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Could not fetch system from Styra API").
			WithEvent("ErrorFetchSystemFromStyra")
	}
	if res.SystemConfig != nil {
		log.Info(fmt.Sprintf("Fetched system %v from Styra API", name))
	} else {
		log.Info(fmt.Sprintf("System %v does not exist in Styra DAS.", name))
	}
	return res.SystemConfig, nil
}

func (r *SystemReconciler) createSystem(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (*styra.CreateSystemResponse, error) {
	log.Info("Creating system in Styra")

	cfg := r.specToSystemConfig(system)

	// We dont set the sourcecontrol settings on system creation, as we havent
	// created the git secret yet.
	cfg.SourceControl = nil

	// Styra does not seem to allow setting deltaBundles before the system is created
	cfg.BundleDownload = nil

	if log.V(1).Enabled() {
		log := log.V(1)
		bs, err := json.Marshal(cfg)
		if err != nil {
			log.Error(err, "Could not marshal request")
		}
		log.Info("Create system request", "request", string(bs))
	}

	res, err := r.Styra.CreateSystem(ctx, &styra.CreateSystemRequest{SystemConfig: cfg})
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Could not create system in Styra").
			WithEvent("ErrorCreateSystemInStyra").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	log.Info("Created system in Styra")
	return res, nil
}

func (r *SystemReconciler) reconcileCredentials(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("Reconciling credentials")

	if system.Spec.SourceControl == nil {
		log.Info("No source control settings defined. Skipping credentials reconciliation")
		return ctrl.Result{}, nil
	}

	var username string
	var password string
	if system.Spec.SourceControl.Origin.CredentialsSecretName == "" {
		gitCredential := r.Config.GetGitCredentialForRepo(system.Spec.SourceControl.Origin.URL)
		if gitCredential == nil {
			log.Info("Could not find matching credentials", "url", system.Spec.SourceControl.Origin.URL)
			return ctrl.Result{}, nil
		}
		username = gitCredential.User
		password = gitCredential.Password

	} else {
		secretName := system.Spec.SourceControl.Origin.CredentialsSecretName
		nsName := types.NamespacedName{
			Namespace: system.Namespace,
			Name:      secretName,
		}
		secret := &corev1.Secret{}
		if err := r.Get(ctx, nsName, secret); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not find credentials Secret").
					WithEvent("ErrorCredentialsSecretNotFound").
					WithSystemCondition(v1beta1.ConditionTypeGitCredentialsUpdated)
			}
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not fetch credentials Secret").
				WithEvent("ErrorCredentialsSecretCouldNotFetch").
				WithSystemCondition(v1beta1.ConditionTypeGitCredentialsUpdated)
		}
		if n, ok := secret.Data["name"]; ok {
			username = string(n)
		} else {
			log.Info("using deprecated username field from git credentials secret")
			username = string(secret.Data["username"])
		}
		if s, ok := secret.Data["secret"]; ok {
			password = string(s)
		} else {
			log.Info("using deprecated password field from git credentials secret")
			password = string(secret.Data["password"])
		}
	}

	if _, err := r.Styra.CreateUpdateSecret(
		ctx,
		system.GitSecretID(),
		&styra.CreateUpdateSecretsRequest{Name: username, Secret: password},
	); err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not create or update secret in Styra").
			WithEvent("ErrorCreateUpdateSecret").
			WithSystemCondition(v1beta1.ConditionTypeGitCredentialsUpdated)
	}

	log.Info("Reconciled credentials")
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) deleteDefaultPolicies(ctx context.Context, log logr.Logger, systemID string) error {
	log.Info("Deleting default policies")

	rulesPolicyName := fmt.Sprintf("systems/%s/rules", systemID)
	testPolicyName := fmt.Sprintf("systems/%s/test", systemID)

	if _, err := r.Styra.DeletePolicy(ctx, rulesPolicyName); err != nil {
		return ctrlerr.Wrap(err, "Could not delete default policy").
			WithEvent("ErrorDeleteDefaultPolicy").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	if _, err := r.Styra.DeletePolicy(ctx, testPolicyName); err != nil {
		return ctrlerr.Wrap(err, "Could not delete default policy").
			WithEvent("ErrorDeleteDefaultPolicy").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	return nil
}

func (r *SystemReconciler) reconcileID(ctx context.Context, log logr.Logger, system *v1beta1.System, id string) error {
	log.Info("Reconciling ID")

	if id == "" {
		return ctrlerr.New("ID is empty").
			WithEvent("ErrorReconcileIDEmpty").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	system.Status.ID = id

	if err := r.Status().Update(ctx, system); err != nil {
		return ctrlerr.Wrap(err, "Could not set system ID on System").
			WithEvent("ErrorSetID").
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	log.Info("Reconciled ID")
	return nil
}

func (r *SystemReconciler) reconcileSubjects(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("Reconciling subjects")

	for _, subject := range system.Spec.Subjects {
		if subject.IsUser() {
			log := log.WithValues("user", subject.Name)
			log.Info("Checking if user exists")
			res, err := r.Styra.GetUser(ctx, subject.Name)
			if err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get user from Styra API").
					WithEvent("ErrorGetUserFromStyra").
					WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
			}
			if res.StatusCode == http.StatusNotFound {
				log.Info("User does not exist in styra. Creating...")
				_, err := r.Styra.CreateInvitation(ctx, false, subject.Name)
				if err != nil {
					return ctrl.Result{}, ctrlerr.Wrap(err, "Could not create user in Styra").
						WithEvent("ErrorCreateInvitation").
						WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
				}
			}
		}
	}

	log.V(1).Info("Getting system rolebindings")
	res, err := r.Styra.ListRoleBindingsV2(ctx, &styra.ListRoleBindingsV2Params{
		ResourceKind: styra.RoleBindingKindSystem,
		ResourceID:   system.Status.ID,
	})
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get rolebindings for system in Styra").
			WithEvent("ErrorGetSystemRolebindings").
			WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
	}

	roleBindingsByRole := map[styra.Role]*styra.RoleBindingConfig{}
	for _, rb := range res.Rolebindings {
		roleBindingsByRole[rb.RoleID] = rb
	}
	systemUserRoles := make([]styra.Role, len(r.Config.SystemUserRoles))
	for i, role := range r.Config.SystemUserRoles {
		systemUserRoles[i] = styra.Role(role)
	}
	rolebindingSubjects := createRolebindingSubjects(
		system.Spec.Subjects,
		r.Config.SSO,
	)
	for _, role := range systemUserRoles {
		rb, ok := roleBindingsByRole[role]

		var subjects []*styra.Subject
		if ok {
			for _, subject := range rb.Subjects {
				if subject.Kind != styra.SubjectKindUser && subject.Kind != styra.SubjectKindClaim {
					subjects = append(subjects, subject)
				}
			}
		}

		subjects = append(subjects, rolebindingSubjects...)

		if !ok {
			if err := r.createRoleBinding(ctx, log, system, role, rolebindingSubjects); err != nil {
				return ctrl.Result{}, err
			}
		} else if !styra.SubjectsAreEqual(rb.Subjects, subjects) {
			if err := r.updateRoleBindingSubjects(ctx, log, rb, subjects); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Remove users and groups from all other rolebindings
	roles := map[styra.Role]struct{}{}
	for _, role := range systemUserRoles {
		roles[role] = struct{}{}
	}
	for _, rb := range res.Rolebindings {
		if _, ok := roles[rb.RoleID]; ok {
			continue
		}

		var subjects []*styra.Subject
		for _, subject := range rb.Subjects {
			if subject.Kind == styra.SubjectKindUser || subject.Kind == styra.SubjectKindClaim {
				continue
			}
			subjects = append(subjects, subject)
		}

		if !styra.SubjectsAreEqual(rb.Subjects, subjects) {
			if err := r.updateRoleBindingSubjects(ctx, log, rb, subjects); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	log.Info("Reconciled users")
	return ctrl.Result{}, nil
}

func createRolebindingSubjects(
	subjects []v1beta1.Subject,
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

		} else if subject.Kind == v1beta1.SubjectKindGroup && sso != nil {
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

func (r *SystemReconciler) createRoleBinding(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	role styra.Role,
	subjects []*styra.Subject,
) error {
	log = log.WithValues("role", role)
	log.Info("Creating rolebinding")

	if _, err := r.Styra.CreateRoleBinding(ctx, &styra.CreateRoleBindingRequest{
		ResourceFilter: &styra.ResourceFilter{
			ID:   system.Status.ID,
			Kind: styra.RoleBindingKindSystem,
		},
		RoleID:   role,
		Subjects: subjects,
	}); err != nil {
		return ctrlerr.Wrap(err, "Could not create rolebinding").
			WithEvent("ErrorCreateRolebinding").
			WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
	}

	log.Info("Created rolebinding")
	return nil
}

func (r *SystemReconciler) updateRoleBindingSubjects(
	ctx context.Context,
	log logr.Logger,
	roleBinding *styra.RoleBindingConfig,
	subjects []*styra.Subject,
) error {
	log = log.WithValues("role", roleBinding.RoleID)
	log.Info("Updating rolebinding")

	_, err := r.Styra.UpdateRoleBindingSubjects(
		ctx,
		roleBinding.ID,
		&styra.UpdateRoleBindingSubjectsRequest{Subjects: subjects},
	)
	if err != nil {
		return ctrlerr.Wrap(err, "Could not update Styra role binding").
			WithEvent("ErrorUpdateRolebinding").
			WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
	}

	log.Info("Updated rolebinding")
	return nil
}

func (r *SystemReconciler) reconcileDatasources(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	cfg *styra.SystemConfig,
) (ctrl.Result, error) {
	log.Info("Reconciling datasources")

	existingByID := map[string]*styra.DatasourceConfig{}
	if cfg != nil {
		for _, ds := range cfg.Datasources {
			if ds.ID != "" {
				existingByID[ds.ID] = ds
			}
		}
	}

	expectedByID := map[string]v1beta1.Datasource{}

	for _, ssds := range system.Spec.Datasources {
		id := path.Join("systems", system.Status.ID, ssds.Path)
		expectedByID[id] = ssds
		ds, ok := existingByID[id]
		if !ok || ds.Category != "rest" || ds.Description != ssds.Description {
			log := log.WithValues("datasourceID", id)
			log.Info("Creating or updating datasource")
			_, err := r.Styra.UpsertDatasource(ctx, id, &styra.UpsertDatasourceRequest{
				Category:    "rest",
				Description: ssds.Description,
			})
			if err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not create or update datasource").
					WithEvent("ErrorUpsertDatasource").
					WithSystemCondition(v1beta1.ConditionTypeDatasourcesUpdated)
			}

			if r.WebhookClient != nil {
				log.Info("Calling datasource changed webhook")
				if err := r.WebhookClient.SystemDatasourceChanged(ctx, log, system.Status.ID, id); err != nil {
					err = ctrlerr.Wrap(err, "Could not call datasource changed webhook").
						WithEvent("ErrorCallWebhook").
						WithSystemCondition(v1beta1.ConditionTypeDatasourcesUpdated)
					r.recordErrorEvent(system, err)
					log.Error(err, err.Error())
				}
			}
		}
	}

	if cfg == nil {
		return ctrl.Result{}, nil
	}

	for _, ds := range cfg.Datasources {
		if ds.ID == "" {
			continue
		}
		if ds.Optional {
			continue
		}
		if _, ok := expectedByID[ds.ID]; !ok {
			log := log.WithValues("datasourceID", ds.ID)
			log.Info("Deleting undeclared datasource")
			if _, err := r.Styra.DeleteDatasource(ctx, ds.ID); err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete datasource").
					WithEvent("ErrorDeleteDatasource").
					WithSystemCondition(v1beta1.ConditionTypeDatasourcesUpdated)
			}
		}
	}

	log.Info("Reconciled datasources")
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcileOPAToken(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	token string,
) (ctrl.Result, bool, error) {
	log.Info("Reconciling OPA token Secret")

	if token == "" {
		return ctrl.Result{}, false, ctrlerr.New("Cannot create token Secret without a token").
			WithEvent("ErrorOPATokenSecretNoToken").
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	var s corev1.Secret
	secretName := fmt.Sprintf("%s-opa-token", system.Name)
	nsName := types.NamespacedName{Name: secretName, Namespace: system.Namespace}
	if err := r.Get(ctx, nsName, &s); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Creating OPA token Secret")
			s.Name = nsName.Name
			s.Namespace = nsName.Namespace
			s.Labels = system.Labels
			if s.Labels == nil {
				s.Labels = map[string]string{}
			}
			labels.SetManagedBy(&s)
			s.Data = map[string][]byte{
				"token": []byte(token),
			}
			if err := controllerutil.SetControllerReference(system, &s, r.Scheme); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not set owner reference on Secret").
					WithEvent("ErrorOwnerRefOPATokenSecret").
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			if err := r.Create(ctx, &s); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create OPA token Secret").
					WithEvent("ErrorCreateOPATokenSecret").
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch OPA token Secret").
			WithEvent("ErrorFetchOPATokenSecret").
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&s, system) {
		return ctrl.Result{}, false, ctrlerr.New("Existing secret is not owned by controller. Skipping update").
			WithEvent("ErrorSecretNotOwnedByController").
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	if string(s.Data["token"]) != token {
		log.Info("Token mismatch. Updating secret.")
		s.Data = map[string][]byte{
			"token": []byte(token),
		}
		update = true
	}

	if update {
		if err := r.Update(ctx, &s); err != nil {
			return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not update OPA token Secret").
				WithEvent("ErrorUpdateOPATokenSecret").
				WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
		}
	}

	log.Info("Reconciled OPA token Secret")
	return ctrl.Result{}, update, nil
}

func (r *SystemReconciler) reconcileOPAConfigMap(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	opaconf styra.OPAConfig,
) (ctrl.Result, bool, error) {
	configmapName := fmt.Sprintf("%s-opa", system.Name)
	log = log.WithValues("configmapName", configmapName)
	log.Info("Reconciling OPA ConfigMap")

	var expectedOPAConfigMap corev1.ConfigMap
	var err error

	var customConfig map[string]interface{}
	if system.Spec.CustomOPAConfig != nil {
		err := yaml.Unmarshal(system.Spec.CustomOPAConfig.Raw, &customConfig)
		if err != nil {
			return ctrl.Result{}, false, err
		}
	}

	if system.Spec.LocalPlane == nil {
		log.Info("No styra local plane defined for System")
		expectedOPAConfigMap, err = k8sconv.OpaConfToK8sOPAConfigMapNoSLP(opaconf, r.Config.OPA, customConfig)
	} else {
		slpURL := fmt.Sprintf("http://%s/v1", system.Spec.LocalPlane.Name)
		expectedOPAConfigMap, err = k8sconv.OpaConfToK8sOPAConfigMap(opaconf, slpURL, r.Config.OPA, customConfig)
	}
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not convert OPA conf to ConfigMap").
			WithEvent("ErrorConvertOPAConf").
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	var cm corev1.ConfigMap
	nsName := types.NamespacedName{Name: configmapName, Namespace: system.Namespace}
	if err := r.Get(ctx, nsName, &cm); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Creating OPA ConfigMap")
			cm.Data = expectedOPAConfigMap.Data
			cm.Name = nsName.Name
			cm.Namespace = nsName.Namespace
			cm.Labels = system.Labels
			if cm.Labels == nil {
				cm.Labels = map[string]string{}
			}
			labels.SetManagedBy(&cm)
			if err := controllerutil.SetControllerReference(system, &cm, r.Scheme); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not set owner reference on OPA ConfigMap").
					WithEvent("ErrorOwnerRefOPAConfigMap").
					WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
			}
			if err := r.Create(ctx, &cm); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create OPA ConfigMap").
					WithEvent("ErrorCreateOPAConfigMap").
					WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch OPA ConfigMap").
			WithEvent("ErrorFetchOPAConfigMap").
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&cm, system) {
		return ctrl.Result{}, false, ctrlerr.New("ConfigMap already exists and is not owned by controller").
			WithEvent("ErrorConfigMapNotOwnedByController").
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	if cm.Data["opa-conf.yaml"] != expectedOPAConfigMap.Data["opa-conf.yaml"] {
		log.Info("Updating OPA ConfigMap")
		cm.Data = expectedOPAConfigMap.Data
		update = true
	}

	if update {
		if err := r.Update(ctx, &cm); err != nil {
			return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not update OPA ConfigMap").
				WithEvent("ErrorUpdateOPAConfigMap").
				WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
		}
	}

	log.Info("Reconciled OPA ConfigMap")
	return ctrl.Result{}, update, nil
}

func (r *SystemReconciler) reconcileSLPConfigMap(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	opaconf styra.OPAConfig,
) (ctrl.Result, bool, error) {
	log.Info("Reconciling SLP ConfigMap")

	if system.Spec.LocalPlane == nil {
		log.Info("No styra local plane defined for System")
		log.Info("Reconciled SLP ConfigMap")
		return ctrl.Result{}, false, nil
	}

	configmapName := fmt.Sprintf("%s-slp", system.Name)
	log = log.WithValues("configmapName", configmapName)

	expectedSLPConfigMap, err := k8sconv.OpaConfToK8sSLPConfigMap(opaconf)
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not convert OPA Conf to SLP ConfigMap").
			WithEvent("ErrorConvertOPAConf").
			WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
	}

	var cm corev1.ConfigMap
	nsName := types.NamespacedName{Name: configmapName, Namespace: system.Namespace}
	if err := r.Get(ctx, nsName, &cm); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Creating SLP ConfigMap")
			cm.Data = expectedSLPConfigMap.Data
			cm.Name = nsName.Name
			cm.Namespace = nsName.Namespace
			cm.Labels = system.Labels
			if cm.Labels == nil {
				cm.Labels = map[string]string{}
			}
			labels.SetManagedBy(&cm)
			if err := controllerutil.SetControllerReference(system, &cm, r.Scheme); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not set owner reference on SLP ConfigMap").
					WithEvent("ErrorOwnerRefSLPConfigMap").
					WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
			}
			if err := r.Create(ctx, &cm); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create SLP ConfigMap").
					WithEvent("ErrorCreateSLPConfigMap").
					WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch SLP ConfigMap").
			WithEvent("ErrorFetchSLPConfigMap").
			WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&cm, system) {
		return ctrl.Result{}, false, ctrlerr.New("ConfigMap already exists and is not owned by controller").
			WithEvent("ErrorFetchSLPConfigMap").
			WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
	}

	if cm.Data["slp.yaml"] != expectedSLPConfigMap.Data["slp.yaml"] {
		log.Info("Updating SLP ConfigMap")
		cm.Data = expectedSLPConfigMap.Data
		update = true
	}

	if update {
		if err := r.Update(ctx, &cm); err != nil {
			return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not update SLP ConfigMap").
				WithEvent("ErrorUpdateSLPConfigmap").
				WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
		}
	}

	log.Info("Reconciled SLP ConfigMap")
	return ctrl.Result{}, update, nil
}

func (r *SystemReconciler) updateSystem(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (*styra.SystemConfig, error) {
	log.Info("Updating system")
	cfg := r.specToSystemConfig(system)

	if log.V(1).Enabled() {
		log := log.V(1)
		bs, err := json.Marshal(cfg)
		if err != nil {
			log.Error(err, "could not marshal request")
		}
		log.Info("update req", "req", string(bs))
	}

	res, err := r.Styra.UpdateSystem(ctx, system.Status.ID, &styra.UpdateSystemRequest{SystemConfig: cfg})
	if err != nil {
		errMsg := "Could not update System"
		var styrahttperr *styra.HTTPError
		if errors.As(err, &styrahttperr) {
			errMsg = fmt.Sprintf("Could not update Styra system. Error %s", styrahttperr.Message)
		}
		return nil, ctrlerr.Wrap(err, errMsg).
			WithEvent("ErrorUpdateSystem").
			WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
	}

	log.Info("Updated system")
	return res.SystemConfig, nil
}

func (r *SystemReconciler) specToSystemConfig(system *v1beta1.System) *styra.SystemConfig {
	cfg := &styra.SystemConfig{
		Name:     system.DisplayName(r.Config.SystemPrefix, r.Config.SystemSuffix),
		Type:     "custom",
		ReadOnly: r.Config.ReadOnly,
	}

	enableDeltaBundles := true
	if r.Config.EnableDeltaBundlesDefault != nil {
		enableDeltaBundles = *r.Config.EnableDeltaBundlesDefault
	}
	if system.Spec.EnableDeltaBundles != nil {
		enableDeltaBundles = *system.Spec.EnableDeltaBundles
	}

	cfg.BundleDownload = &styra.BundleDownloadConfig{
		DeltaBundles: enableDeltaBundles,
	}

	if len(system.Spec.DecisionMappings) > 0 {
		cfg.DecisionMappings = map[string]styra.DecisionMapping{}
		for _, dm := range system.Spec.DecisionMappings {
			v1dm := styra.DecisionMapping{}

			if dm.Allowed != nil {
				v1dm.Allowed = &styra.DecisionMappingAllowed{}
				if dm.Allowed.Expected != nil {
					v1dm.Allowed.Expected = dm.Allowed.Expected.Value()
				}
				if dm.Allowed.Negated {
					v1dm.Allowed.Negated = dm.Allowed.Negated
				}
				v1dm.Allowed.Path = dm.Allowed.Path
			}

			if len(dm.Columns) > 0 {
				v1dm.Columns = make([]styra.DecisionMappingColumn, len(dm.Columns))
				for i := range dm.Columns {
					c := dm.Columns[i]
					v1dm.Columns[i] = styra.DecisionMappingColumn{Key: c.Key, Path: c.Path}
				}
			}

			if dm.Reason.Path != "" {
				v1dm.Reason = &styra.DecisionMappingReason{
					Path: dm.Reason.Path,
				}
			}

			cfg.DecisionMappings[dm.Name] = v1dm
		}
	}

	if system.Spec.SourceControl != nil {
		cfg.SourceControl = &styra.SourceControlConfig{
			Origin: styra.GitRepoConfig{
				Credentials: system.GitSecretID(),
				Path:        system.Spec.SourceControl.Origin.Path,
				URL:         system.Spec.SourceControl.Origin.URL,
			},
		}

		if system.Spec.SourceControl.Origin.Commit != "" {
			cfg.SourceControl.Origin.Commit = system.Spec.SourceControl.Origin.Commit
		} else if system.Spec.SourceControl.Origin.Reference != "" {
			cfg.SourceControl.Origin.Reference = system.Spec.SourceControl.Origin.Reference
		}
	}

	if system.Spec.DiscoveryOverrides != nil {
		if cfg.DeploymentParameters == nil {
			cfg.DeploymentParameters = &styra.DeploymentParameters{}
		}
		cfg.DeploymentParameters.Discovery = system.Spec.DiscoveryOverrides
	}

	return cfg
}

func (r *SystemReconciler) systemNeedsUpdate(log logr.Logger, system *v1beta1.System, cfg *styra.SystemConfig) bool {
	if cfg == nil {
		log.Info("System needs update: cfg is nil")
		return true
	}

	if cfg.ReadOnly != r.Config.ReadOnly {
		log.Info("System needs update: read only is not equal")
		return true
	}

	expectedModel := r.specToSystemConfig(system)

	if cfg.BundleDownload == nil || cfg.BundleDownload.DeltaBundles != expectedModel.BundleDownload.DeltaBundles {
		log.Info("System needs update: Deltabundle setting not equal")
		return true
	}

	if !reflect.DeepEqual(expectedModel.SourceControl, cfg.SourceControl) {
		log.Info("System needs update: source control is not equal")
		return true
	}

	namesAreEqual := expectedModel.Name == cfg.Name
	if !namesAreEqual {
		log.Info("System needs update: system names are not are not equal")
		return true
	}

	dmsAreEqual := styra.DecisionMappingsEquals(expectedModel.DecisionMappings, cfg.DecisionMappings)
	if !dmsAreEqual {
		log.Info("System needs update: decision mappings are not equal")
		return true
	}
	return false
}

// SetupWithManager registers the the System controller with the Manager.
func (r *SystemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// setup field indexes
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1beta1.System{},
		fields.SystemCredentialsSecretName,
		func(o client.Object) []string {
			System := o.(*v1beta1.System)
			if System.Spec.SourceControl == nil ||
				System.Spec.SourceControl.Origin.CredentialsSecretName == "" {
				return nil
			}
			return []string{System.Spec.SourceControl.Origin.CredentialsSecretName}
		},
	); err != nil {
		return errors.Wrap(err, "Could not create field index")
	}

	// Setup predicate which ensures that we only reconcile System changes
	// that match the controller class, and only for changes of the spec
	p, err := predicate.ControllerClass(r.Config.ControllerClass)
	if err != nil {
		return errors.Wrap(err, "Could not build predicate")
	}
	p = ctrlpred.And(p, ctrlpred.GenerationChangedPredicate{})

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.System{}, builder.WithPredicates(p)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findSystemsForSecret),
			builder.WithPredicates(ctrlpred.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findSystemsForConfigMap),
			builder.WithPredicates(ctrlpred.ResourceVersionChangedPredicate{}),
		).
		Complete(sentry.Decorate(r))
}

func (r *SystemReconciler) findSystemsForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	requests := r.findSystemsRefferingToSecret(ctx, secret)
	return append(requests, r.findSecretOwners(ctx, secret)...)
}

// findSystemsRefferingToSecret detects if modified secret is the secret containing Git credentials for a System.
func (r *SystemReconciler) findSystemsRefferingToSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	var systemsWithCredentialsRef v1beta1.SystemList
	ls, err := labels.ControllerClassLabelSelectorAsSelector(r.Config.ControllerClass)
	if err != nil {
		panic(err)
	}

	if err := r.List(ctx, &systemsWithCredentialsRef, &client.ListOptions{
		FieldSelector: fields.SystemCredentialsSecretNameSelector(secret.GetName()),
		LabelSelector: ls,
		Namespace:     secret.GetNamespace(),
	}); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(systemsWithCredentialsRef.Items))
	for i, item := range systemsWithCredentialsRef.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}

// findSecretOwners detects if modified secret is the secret containing the opa token for a System
func (r *SystemReconciler) findSecretOwners(ctx context.Context, secret client.Object) []reconcile.Request {
	var requests []reconcile.Request

	for _, owner := range secret.GetOwnerReferences() {
		if !ownerIsSystem(owner) {
			continue
		}
		s := &v1beta1.System{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      owner.Name,
			Namespace: secret.GetNamespace(),
		}, s); err == nil {
			if !labels.ControllerClassMatches(s, r.Config.ControllerClass) {
				continue
			}
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      s.GetName(),
					Namespace: s.GetNamespace(),
				},
			})
		}
	}

	return requests
}

func ownerIsSystem(owner metav1.OwnerReference) bool {
	return owner.APIVersion == v1beta1.GroupVersion.String() &&
		owner.Kind == "System"
}

// findSystemsForConfigMap detects if modified configmap is the configmap containing opa/slp config.
func (r *SystemReconciler) findSystemsForConfigMap(ctx context.Context, configmap client.Object) []reconcile.Request {
	var requests []reconcile.Request

	for _, owner := range configmap.GetOwnerReferences() {
		if !ownerIsSystem(owner) {
			continue
		}
		s := &v1beta1.System{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      owner.Name,
			Namespace: configmap.GetNamespace(),
		}, s); err == nil {
			if !labels.ControllerClassMatches(s, r.Config.ControllerClass) {
				continue
			}
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      s.GetName(),
					Namespace: s.GetNamespace(),
				},
			})
		}
	}

	return requests
}
