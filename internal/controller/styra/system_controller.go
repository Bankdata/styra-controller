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
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

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
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/bankdata/styra-controller/pkg/ocp"
)

const (
	awsSecretNameKeyID     = "AWS_ACCESS_KEY_ID"
	awsSecretNameSecretKey = "AWS_SECRET_ACCESS_KEY"
	awsSecretNameRegion    = "AWS_REGION"
)

// SystemReconcilerMetrics holds the metrics for the SystemReconciller
type SystemReconcilerMetrics struct {
	ControllerSystemStatusReady *prometheus.GaugeVec
	ReconcileSegmentTime        *prometheus.HistogramVec
	ReconcileTime               *prometheus.HistogramVec
}

// SystemReconciler reconciles a System object
type SystemReconciler struct {
	client.Client
	APIReader     client.Reader
	Scheme        *runtime.Scheme
	OCP           ocp.ClientInterface
	WebhookClient webhook.Client
	Recorder      record.EventRecorder
	Metrics       *SystemReconcilerMetrics
	Config        *configv2alpha2.ProjectConfig
}

//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=styra.bankdata.dk,resources=systems/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;patch;
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile implements renconcile.Renconciler and has responsibility of
// ensuring that the current state of the System resource renconciled
// towards the desired state.
func (r *SystemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()
	log := log.FromContext(ctx)
	log.Info("Reconciliation begins")

	log.Info("Fetching System")
	var system v1beta1.System
	if err := r.APIReader.Get(ctx, req.NamespacedName, &system); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Could not find System in kubernetes")
			r.Metrics.ReconcileTime.WithLabelValues("delete").Observe(time.Since(start).Seconds())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch System")
	}

	log = log.WithValues("systemID", system.Status.ID)
	log = log.WithValues("controlPlane", system.Labels["styra-controller/control-plane"])
	log = log.WithValues("uniqueName", system.OCPUniqueName(r.Config.SystemPrefix, r.Config.SystemSuffix))

	if !r.isSystemNamespaceMatchingSelector(&system) {
		log.Info("Namespace is not in NamespaceSelector for reconciliation. Skipping reconciliation")
		r.deleteMetrics(req)
		return ctrl.Result{}, nil
	}

	var (
		res ctrl.Result
		err error
	)

	if system.ObjectMeta.DeletionTimestamp.IsZero() {
		res, err = r.reconcile(ctx, log, &system)
		r.updateMetric(req, system.Status.ID, system.Status.Ready, labels.LabelValueControlPlaneOCP)
	} else {
		res, err = r.reconcileDeletion(ctx, log, &system)
		if err != nil {
			r.updateMetric(req, system.Status.ID, system.Status.Ready, labels.LabelValueControlPlaneOCP)
		} else {
			r.deleteMetrics(req)
			r.Metrics.ReconcileTime.WithLabelValues("delete").Observe(time.Since(start).Seconds())
			return res, err
		}
	}

	if err != nil {
		log.Error(err, "Reconciliation failed")
		r.recordErrorEvent(&system, err)
		r.setSystemStatusError(&system, err)

		r.Metrics.ReconcileTime.WithLabelValues("error").Observe(time.Since(start).Seconds())
		if err := r.Status().Update(ctx, &system); err != nil {
			return res, errors.Wrap(err, "could not set failure status on System")
		}
	} else {
		r.Metrics.ReconcileTime.WithLabelValues("ok").Observe(time.Since(start).Seconds())
	}
	return res, err
}

func (r *SystemReconciler) isSystemNamespaceMatchingSelector(system *v1beta1.System) bool {
	if r.Config.NamespaceSelector == nil || len(r.Config.NamespaceSelector.MatchPatterns) == 0 {
		return true
	}

	for _, pattern := range r.Config.NamespaceSelector.MatchPatterns {
		if matched, _ := filepath.Match(pattern, system.Namespace); matched {
			return true
		}
	}

	return false
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

func (r *SystemReconciler) updateMetric(req ctrl.Request, systemID string, ready bool, controlPlane string) {
	if r.Metrics == nil || r.Metrics.ControllerSystemStatusReady == nil {
		return
	}

	var value float64
	if ready {
		value = 1
	}
	r.Metrics.ControllerSystemStatusReady.WithLabelValues(req.Name, req.Namespace, systemID, controlPlane).Set(value)
}

func (r *SystemReconciler) deleteMetrics(req ctrl.Request) {
	if r.Metrics == nil || r.Metrics.ControllerSystemStatusReady == nil {
		return
	}
	if deleted := r.Metrics.ControllerSystemStatusReady.DeletePartialMatch(
		prometheus.Labels{"system_name": req.Name, "namespace": req.Namespace},
	); deleted > 1 {
		log.Log.Error(errors.New("Failed to delete metric"), "Incorrect number of deleted metrics", "deleted", deleted)
	}
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
			WithEvent(v1beta1.EventErrorSetFinalizer).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInOcp)
	}
	return nil
}

func (r *SystemReconciler) reconcileDeletion(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("System deletion is in progress")
	if !finalizer.IsSet(system) {
		return ctrl.Result{}, nil
	}

	deletionProtected := false
	if system.Spec.DeletionProtection != nil {
		deletionProtected = *system.Spec.DeletionProtection
	} else {
		deletionProtected = r.Config.DeletionProtectionDefault
	}
	if !deletionProtected {
		log.Info("Deleting bundle and source for system in OCP")
		uniqueName := system.OCPUniqueName(r.Config.SystemPrefix, r.Config.SystemSuffix)
		if err := r.OCP.DeleteBundle(ctx, uniqueName); err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete bundle in OCP").
				WithEvent(v1beta1.EventErrorDeleteBundleInOCP)
		}
		if err := r.OCP.DeleteSource(ctx, uniqueName); err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete source in OCP").
				WithEvent(v1beta1.EventErrorDeleteSourceInOCP)
		}

		for _, datasource := range system.Spec.Datasources {
			datasourceID := strings.ToLower(strings.ReplaceAll(datasource.Path, "/", "-"))
			if err := r.OCP.DeleteSource(ctx, datasourceID); err != nil {
				var httpErr *httperror.HTTPError
				if errors.As(err, &httpErr) {
					if httpErr.StatusCode == http.StatusInternalServerError {
						continue
					}
				}
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete datasource source in OCP").
					WithEvent(v1beta1.EventErrorDeleteSourceInOCP)
			}
		}
	}

	log.Info("Removing finalizer")
	finalizer.Remove(system)
	if err := r.Update(ctx, system); err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not remove finalizer").
			WithEvent(v1beta1.EventErrorRemovingFinalizer)
	}
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	log.Info("Reconciling system spec")

	if !finalizer.IsSet(system) {
		if err := r.reconcileFinalizer(ctx, log, system); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("OPA Control Plane system reconcile starting")
	return r.ocpReconcile(ctx, log, system)
}

func (r *SystemReconciler) ocpReconcile(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System) (ctrl.Result, error) {
	var requirements []ocp.Requirement

	for _, datasource := range system.Spec.Datasources {
		datasource.Path = strings.ToLower(strings.ReplaceAll(datasource.Path, "/", "-"))

		created, err := r.createSourceIfNotExists(ctx, log, datasource)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err,
				fmt.Sprintf("ocpReconcile: Could not ensure datasource/source exists: %s", datasource.Path),
			).WithEvent(v1beta1.EventErrorUpdateSource).
				WithSystemCondition(v1beta1.ConditionTypeRequirementsUpdated)
		}

		if created && r.WebhookClient != nil {
			log.Info("Calling datasource changed webhook")
			if err := r.WebhookClient.SystemDatasourceChangedOCP(ctx, log, datasource.Path); err != nil {
				err = ctrlerr.Wrap(err, "Could not call datasource changed webhook").
					WithEvent(v1beta1.EventErrorCallWebhook).
					WithSystemCondition(v1beta1.ConditionTypeRequirementsUpdated)
				r.recordErrorEvent(system, err)
				log.Error(err, err.Error())
			}
		}

		requirements = append(requirements, ocp.NewRequirement(datasource.Path))
	}
	system.SetCondition(v1beta1.ConditionTypeRequirementsUpdated, metav1.ConditionTrue)

	reconcileSystemSourceStart := time.Now()
	uniqueName := system.OCPUniqueName(r.Config.SystemPrefix, r.Config.SystemSuffix)
	result, err := r.reconcileSystemSource(ctx, log, system, uniqueName)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcileSystemSourceOcp").
		Observe(time.Since(reconcileSystemSourceStart).Seconds())
	if err != nil {
		return result, ctrlerr.Wrap(
			err, fmt.Sprintf("ocpReconcile: Could not reconcile system source: %s", uniqueName)).
			WithEvent(v1beta1.EventErrorUpdateSource).
			WithSystemCondition(v1beta1.ConditionTypeSystemSourceUpdated)
	}
	requirements = append(requirements, ocp.NewRequirement(uniqueName))
	system.SetCondition(v1beta1.ConditionTypeSystemSourceUpdated, metav1.ConditionTrue)

	defaultRequirements := ocp.ToRequirements(r.Config.OPAControlPlaneConfig.DefaultRequirements)

	reconcileSystemBundleStart := time.Now()
	result, err = r.reconcileSystemBundle(ctx, uniqueName, requirements, defaultRequirements)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcileSystemBundleOcp").
		Observe(time.Since(reconcileSystemBundleStart).Seconds())
	if err != nil {
		return result, ctrlerr.Wrap(err, fmt.Sprintf("ocpReconcile: Could not reconcile system bundle: %s", uniqueName)).
			WithEvent(v1beta1.EventErrorUpdateBundle).
			WithSystemCondition(v1beta1.ConditionTypeSystemBundleUpdated)
	}
	system.SetCondition(v1beta1.ConditionTypeSystemBundleUpdated, metav1.ConditionTrue)

	secretName := fmt.Sprintf("%s-opa-secret", system.Name)
	result, secretUpdated, err := r.reconcileOPASecret(ctx, log, system, uniqueName, secretName)
	if err != nil {
		return result, ctrlerr.Wrap(err, fmt.Sprintf("ocpReconcile: Could not reconcile OPA Secret: %s", secretName)).
			WithEvent(v1beta1.EventErrorUpdateOPASecret).
			WithSystemCondition(v1beta1.ConditionTypeOPASecretUpdated)
	}
	if secretUpdated {
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionFalse)
		err = r.Status().Update(ctx, system)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not update system to reflect that secret is outdated").
				WithEvent(v1beta1.EventErrorUpdateStatus).
				WithSystemCondition(v1beta1.ConditionTypeOPAUpToDate)
		}
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPASecretUpdated, metav1.ConditionTrue)

	configmapName := fmt.Sprintf("%s-opa-config", system.Name)
	result, updatedOPAConfigMap, err := r.reconcileOPAConfigMapForOCP(ctx, log, system, uniqueName, configmapName)
	if err != nil {
		return result, ctrlerr.Wrap(err, fmt.Sprintf("ocpReconcile: Could not reconcile OPA ConfigMap: %s", configmapName)).
			WithEvent(v1beta1.EventErrorUpdateOPAConfigMap).
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}
	if updatedOPAConfigMap {
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionFalse)
		err = r.Status().Update(ctx, system)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not update system to reflect that configmap is outdated").
				WithEvent(v1beta1.EventErrorUpdateStatus).
				WithSystemCondition(v1beta1.ConditionTypeOPAUpToDate)
		}
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPAConfigMapUpdated, metav1.ConditionTrue)

	system.Status.Ready = true
	system.Status.Phase = v1beta1.SystemPhaseCreated
	system.Status.FailureMessage = ""

	if system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) == nil ||
		*system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) != metav1.ConditionTrue {
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionTrue)
	}

	updateStatusStart := time.Now()
	err = r.Status().Update(ctx, system)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("updateStatusOcp").Observe(time.Since(updateStatusStart).Seconds())
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not change status.phase to Created").
			WithEvent(v1beta1.EventErrorPhaseToCreated)
	}

	msg := "OPA Control Plane reconciliation completed"
	r.Recorder.Event(system, corev1.EventTypeNormal, "ReconciliationCompleted", msg)
	log.Info(msg)
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcileOPAConfigMapForOCP(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	uniqueName string,
	configmapName string,
) (ctrl.Result, bool, error) {
	log.Info("Reconciling OPA ConfigMap")

	var expectedOPAConfigMap corev1.ConfigMap
	var customConfig map[string]interface{}
	if system.Spec.CustomOPAConfig != nil {
		err := yaml.Unmarshal(system.Spec.CustomOPAConfig.Raw, &customConfig)
		if err != nil {
			return ctrl.Result{}, false, err
		}
	}

	bundleURL, err := url.JoinPath(r.Config.OPA.BundleServer.URL, r.Config.OPA.BundleServer.Path)
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Invalid OPA BundleServer URL or path").
			WithEvent(v1beta1.EventErrorConvertOPAConf).
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	bundleServiceCredentials := &ocp.ServiceCredentials{
		S3: &ocp.S3Signing{
			S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
		},
	}
	if r.Config.OPA.BundleServer.TokenPath != "" {
		bundleServiceCredentials = &ocp.ServiceCredentials{
			Bearer: &ocp.Bearer{
				TokenPath: r.Config.OPA.BundleServer.TokenPath,
			},
		}
	}

	opaconf := ocp.OPAConfig{
		BundleService: &ocp.OPAServiceConfig{
			Name:        r.Config.OPA.BundleServer.Name,
			URL:         bundleURL,
			Credentials: bundleServiceCredentials,
		},
		LogService: &ocp.OPAServiceConfig{
			Name: r.Config.OPA.DecisionAPIConfig.Name,
			URL:  r.Config.OPA.DecisionAPIConfig.ServiceURL,
			Credentials: &ocp.ServiceCredentials{
				Bearer: &ocp.Bearer{
					TokenPath: r.Config.OPA.DecisionAPIConfig.TokenPath,
				},
			},
		},
		DecisionLogReporting: r.Config.OPA.DecisionAPIConfig.Reporting,
		BundleResource:       fmt.Sprintf("bundles/%s/bundle.tar.gz", uniqueName),
		UniqueName:           uniqueName,
		Namespace:            system.Namespace,
	}

	expectedOPAConfigMap, err = k8sconv.OPAConfToK8sOPAConfigMapforOCP(opaconf, r.Config.OPA, customConfig, log)
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not convert OPA conf to ConfigMap").
			WithEvent(v1beta1.EventErrorConvertOPAConf).
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
					WithEvent(v1beta1.EventErrorOwnerRefOPAConfigMap).
					WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
			}
			if err := r.Create(ctx, &cm); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create OPA ConfigMap").
					WithEvent(v1beta1.EventErrorCreateOPAConfigMap).
					WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch OPA ConfigMap").
			WithEvent(v1beta1.EventErrorFetchOPAConfigMap).
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&cm, system) {
		return ctrl.Result{}, false, ctrlerr.New("ConfigMap already exists and is not owned by controller").
			WithEvent(v1beta1.EventErrorConfigMapNotOwnedByController).
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
				WithEvent(v1beta1.EventErrorUpdateOPAConfigMap).
				WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
		}
	}

	log.Info("Reconciled OPA ConfigMap")
	return ctrl.Result{}, update, nil
}

func (r *SystemReconciler) reconcileOPASecret(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	uniqueName string,
	secretName string,
) (ctrl.Result, bool, error) {
	log.Info("Reconciling OPA secret")
	_ = uniqueName
	log.Info("Direct S3 credential management is disabled, creating OPA secret without generated credentials")

	reconcilek8sOPASecret := time.Now()
	result, secretUpdated, err := r.reconcilek8sOPASecret(ctx, log, system, secretName)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcilek8sOPASecretOcp").
		Observe(time.Since(reconcilek8sOPASecret).Seconds())
	if err != nil {
		return result, false, err
	}
	log.Info("Reconciled OPA secret")
	return ctrl.Result{}, secretUpdated, nil
}

func (r *SystemReconciler) reconcilek8sOPASecret(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	secretName string,
) (ctrl.Result, bool, error) {

	var s corev1.Secret
	nsName := types.NamespacedName{Name: secretName, Namespace: system.Namespace}
	if err := r.Get(ctx, nsName, &s); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Creating OPA Secret")
			s.Name = nsName.Name
			s.Namespace = nsName.Namespace
			s.Labels = system.Labels
			if s.Labels == nil {
				s.Labels = map[string]string{}
			}
			labels.SetManagedBy(&s)
			s.Data = map[string][]byte{}
			if err := controllerutil.SetControllerReference(system, &s, r.Scheme); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not set owner reference on Secret").
					WithEvent(v1beta1.EventErrorOwnerRefOPATokenSecret).
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			if err := r.Create(ctx, &s); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create OPA Secret").
					WithEvent(v1beta1.EventErrorCreateOPATokenSecret).
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch OPA Secret").
			WithEvent(v1beta1.EventErrorFetchOPATokenSecret).
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&s, system) {
		return ctrl.Result{}, false, ctrlerr.New("Existing secret is not owned by controller. Skipping update").
			WithEvent(v1beta1.EventErrorSecretNotOwnedByController).
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	if _, ok := s.Data[awsSecretNameKeyID]; ok {
		log.Info("Removing controller-managed S3 access key from OPA Secret")
		delete(s.Data, awsSecretNameKeyID)
		update = true
	}
	if _, ok := s.Data[awsSecretNameSecretKey]; ok {
		log.Info("Removing controller-managed S3 secret key from OPA Secret")
		delete(s.Data, awsSecretNameSecretKey)
		update = true
	}
	if _, ok := s.Data[awsSecretNameRegion]; ok {
		log.Info("Removing controller-managed S3 region from OPA Secret")
		delete(s.Data, awsSecretNameRegion)
		update = true
	}

	if update {
		if err := r.Update(ctx, &s); err != nil {
			return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not update OPA secret").
				WithEvent(v1beta1.EventErrorUpdateOPATokenSecret).
				WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
		}
	}

	return ctrl.Result{}, update, nil
}

func (r *SystemReconciler) reconcileSystemBundle(
	ctx context.Context,
	uniqueName string,
	requirements []ocp.Requirement,
	defaultRequirements []ocp.Requirement) (ctrl.Result, error) {
	if r.Config.OPAControlPlaneConfig.BundleObjectStorage.S3 == nil {
		return ctrl.Result{}, ctrlerr.New("reconcileSystemBundle: no object storage configured")
	}

	objectStorage := ocp.ObjectStorage{
		AmazonS3: &ocp.AmazonS3{
			Bucket:      r.Config.OPAControlPlaneConfig.BundleObjectStorage.S3.Bucket,
			Key:         fmt.Sprintf("bundles/%s/bundle.tar.gz", uniqueName),
			Region:      r.Config.OPAControlPlaneConfig.BundleObjectStorage.S3.Region,
			URL:         r.Config.OPAControlPlaneConfig.BundleObjectStorage.S3.URL,
			Credentials: r.Config.OPAControlPlaneConfig.BundleObjectStorage.S3.OCPConfigSecretName,
		},
	}

	bundle := &ocp.PutBundleRequest{
		Name:          uniqueName,
		ObjectStorage: objectStorage,
		Requirements:  append(requirements, defaultRequirements...),
		Revision:      bundleRevision(uniqueName, defaultRequirements, requirements),
	}
	err := r.OCP.PutBundle(ctx, bundle)

	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "ocpReconcile: could not create or update bundle in OCP")
	}
	return ctrl.Result{}, nil
}

// bundleRevision produces a Rego template string containing:
// - data: sha256 hash of all datasource SQL hashes
// - git-sha: the git commit for the system's unique source
// - libraries: sha256 hash of all default requirement git commits
// for example "data:sha256,git-sha:commitsha,libraries:sha256"
func bundleRevision(uniqueName string, defaultRequirements []ocp.Requirement, requirements []ocp.Requirement) string {
	sqlHashes := make([]string, len(defaultRequirements)+len(requirements))
	gitCommits := make([]string, len(defaultRequirements)+len(requirements))
	requirementList := append(requirements, defaultRequirements...)
	for i, req := range requirementList {
		sqlHashes[i] = fmt.Sprintf(`input.sources["%s"].sql.hash`, req.Source)
		gitCommits[i] = fmt.Sprintf(`input.sources["%s"].git.commit`, req.Source)
	}
	sqlHashSet := strings.Join(sqlHashes, ", ")
	gitCommitSet := strings.Join(gitCommits, ", ")

	return fmt.Sprintf(
		`$"data:{crypto.sha256(concat("", [%s]))},`+
			`git-sha:{input.sources["%s"].git.commit},`+ // git sha for the system source
			`libraries:{crypto.sha256(concat("", [%s]))}"`,
		sqlHashSet, uniqueName, gitCommitSet,
	)
}

func (r *SystemReconciler) reconcileSystemSource(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	uniqueName string) (ctrl.Result, error) {

	if system.Spec.SourceControl == nil {
		return ctrl.Result{}, ctrlerr.New("reconcileSystemSource: no source control configured on system")
	}

	if !isURLValid(system.Spec.SourceControl.Origin.URL) {
		return ctrl.Result{}, ctrlerr.New("Invalid URL for source control")
	}

	gitConfig := &ocp.GitConfig{
		Repo:          system.Spec.SourceControl.Origin.URL,
		IncludedFiles: []string{"*.rego"},
		ExcludedFiles: []string{"*_test.rego"},
	}
	if system.Spec.SourceControl.Origin.Commit != "" {
		gitConfig.Commit = system.Spec.SourceControl.Origin.Commit
	} else if system.Spec.SourceControl.Origin.Reference != "" {
		gitConfig.Reference = system.Spec.SourceControl.Origin.Reference
	}
	if system.Spec.SourceControl.Origin.Path != "" {
		gitConfig.Path = system.Spec.SourceControl.Origin.Path
	}
	gitCredentialFound := false
	for _, cred := range r.Config.OPAControlPlaneConfig.GitCredentials {
		if strings.Contains(system.Spec.SourceControl.Origin.URL, cred.RepoPrefix) {
			gitConfig.CredentialID = cred.ID
			gitCredentialFound = true
			break
		}
	}
	if !gitCredentialFound {
		return ctrl.Result{}, fmt.Errorf(
			"reconcileSystemSource: Unsupported git repository: %s",
			system.Spec.SourceControl.Origin.URL)
	}

	_, err := r.OCP.PutSource(ctx, uniqueName, &ocp.PutSourceRequest{
		Name: uniqueName,
		Git:  gitConfig,
	})
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "reconcileSystemSource: could not create or update source in OCP")
	}
	log.Info("OCP source upserted", "source", uniqueName)
	return ctrl.Result{}, nil
}

func isURLValid(rawURL string) bool {
	if rawURL == "" {
		return true
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// URL must have a scheme (http or https) and a host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// Reject URLs with brackets in the path (invalid for repository URLs)
	if strings.Contains(parsedURL.Path, "[") || strings.Contains(parsedURL.Path, "]") {
		return false
	}

	return true
}

func (r *SystemReconciler) createSourceIfNotExists(
	ctx context.Context,
	log logr.Logger,
	source v1beta1.Datasource) (bool, error) {
	_, err := r.OCP.GetSource(ctx, source.Path)
	if err == nil {
		log.Info("Source already exists", "source", source.Path)
		return false, nil
	}

	var httpErr *httperror.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode != http.StatusNotFound {
			return false, ctrlerr.Wrap(err, "GetSource in createSourceIfNotExists failed")
		}
	} else {
		return false, ctrlerr.Wrap(err, "GetSource in createSourceIfNotExists failed")
	}

	log.Info("Creating source", "source", source.Path)
	_, err = r.OCP.PutSource(ctx, source.Path, &ocp.PutSourceRequest{
		Name: source.Path,
	})
	if err != nil {
		return false, ctrlerr.Wrap(err, "PutSource in createSourceIfNotExists failed")
	}
	log.Info("Source created", "source", source.Path)

	return true, nil
}

// CreateDefaultRequirements creates all the configured default sources in OCP.
func (r *SystemReconciler) CreateDefaultRequirements(ctx context.Context, log logr.Logger) error {
	log.Info("Creating OCP default requirements")
	for _, defaultRequirement := range r.Config.OPAControlPlaneConfig.DefaultRequirements {
		_, err := r.createSourceIfNotExists(ctx, log, v1beta1.Datasource{Path: defaultRequirement})
		if err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager registers the the System controller with the Manager.
func (r *SystemReconciler) SetupWithManager(mgr ctrl.Manager, name string) error {
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

	updatedPred := ctrlpred.Or(ctrlpred.GenerationChangedPredicate{}, ctrlpred.LabelChangedPredicate{})

	p = ctrlpred.And(p, updatedPred)

	return ctrl.NewControllerManagedBy(mgr).Named(name).
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
		Complete(r)
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
