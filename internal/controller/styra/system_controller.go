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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
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
	"github.com/bankdata/styra-controller/internal/config"
	ctrlerr "github.com/bankdata/styra-controller/internal/errors"
	"github.com/bankdata/styra-controller/internal/fields"
	"github.com/bankdata/styra-controller/internal/finalizer"
	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/internal/labels"
	"github.com/bankdata/styra-controller/internal/predicate"
	"github.com/bankdata/styra-controller/internal/sentry"
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/s3"
	"github.com/bankdata/styra-controller/pkg/styra"
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
	Styra         styra.ClientInterface
	OCP           ocp.ClientInterface
	S3            s3.Client
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

	if !labels.ControllerClassMatches(&system, r.Config.ControllerClass) {
		log.Info("This is not a System we are managing. Skipping reconciliation.")
		r.deleteMetrics(req)
		return ctrl.Result{}, nil
	}

	var (
		res ctrl.Result
		err error
	)

	if system.ObjectMeta.DeletionTimestamp.IsZero() {
		res, err = r.reconcile(ctx, log, &system)
		if system.Labels[labels.LabelControlPlane] == labels.LabelValueControlPlaneOCP {
			r.updateMetric(req, system.Status.ID, system.Status.Ready, system.Labels[labels.LabelControlPlane])
		} else {
			r.updateMetric(req, system.Status.ID, system.Status.Ready, labels.LabelValueControlPlaneStyra)
		}

	} else {
		res, err = r.reconcileDeletion(ctx, log, &system)
		if err != nil {
			if system.Labels[labels.LabelControlPlane] == labels.LabelValueControlPlaneOCP {
				r.updateMetric(req, system.Status.ID, system.Status.Ready, system.Labels[labels.LabelControlPlane])
			} else {
				r.updateMetric(req, system.Status.ID, system.Status.Ready, labels.LabelValueControlPlaneStyra)
			}
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

	// If system switched control plane, we need to delete the old metric with the other control plane label
	var labelToDelete string
	if controlPlane == labels.LabelValueControlPlaneOCP {
		labelToDelete = labels.LabelValueControlPlaneStyra
	}
	if controlPlane == labels.LabelValueControlPlaneStyra {
		labelToDelete = labels.LabelValueControlPlaneOCP
	}

	// We cannot know if the metric exists before deleting it.
	// Therefore, we just delete and ignore the output of DeletePartialMatch.
	r.Metrics.ControllerSystemStatusReady.DeletePartialMatch(
		prometheus.Labels{
			"system_name":   req.Name,
			"namespace":     req.Namespace,
			"system_id":     systemID,
			"control_plane": labelToDelete,
		},
	)
}

func (r *SystemReconciler) deleteMetrics(req ctrl.Request) {
	if r.Metrics == nil || r.Metrics.ControllerSystemStatusReady == nil {
		return
	}
	if deleted := r.Metrics.ControllerSystemStatusReady.DeletePartialMatch(
		prometheus.Labels{"system_name": req.Name, "namespace": req.Namespace},
	); deleted != 1 {
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
	if !finalizer.IsSet(system) {
		return ctrl.Result{}, nil
	}

	systemConfiguredToUseStyra := system.Labels[labels.LabelControlPlane] == "" ||
		system.Labels[labels.LabelControlPlane] == labels.LabelValueControlPlaneStyra

	if r.Config.EnableStyraReconciliation && systemConfiguredToUseStyra {
		// finalizer is present so we need to ensure system is deleted in styra,
		// unless deletion protection is enabled or styra reconciliation is disabled
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
						WithEvent(v1beta1.EventErrorDeleteSystemInStyra)
				}
			}
		}
	}

	//OCP is also enabled if nothing is configured in controller config (everything is set to false)
	ocpEnabled := r.Config.EnableOPAControlPlaneReconciliation || r.Config.EnableOPAControlPlaneReconciliationTestData ||
		(!r.Config.EnableOPAControlPlaneReconciliation &&
			!r.Config.EnableOPAControlPlaneReconciliationTestData &&
			!r.Config.EnableStyraReconciliation)

	systemConfiguredToUseOCP := system.Labels[labels.LabelControlPlane] == labels.LabelValueControlPlaneOCP

	if ocpEnabled && systemConfiguredToUseOCP {
		// Delete associated bundle and source for system. Also delete datasources.
		// There is a safeguard in OCP that protects from deleting a source that is used by a bundle
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
				datasourceID := strings.ReplaceAll(datasource.Path, "/", "-")
				if err := r.OCP.DeleteSource(ctx, datasourceID); err != nil {
					var httpErr *httperror.HTTPError
					if errors.As(err, &httpErr) {
						if httpErr.StatusCode == http.StatusInternalServerError {
							// OCP returns 500 when deleting a source used by a bundle.
							// A datasource could be used by another system, so ignore error.
							continue
						}
					}
					return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete datasource source in OCP").
						WithEvent(v1beta1.EventErrorDeleteSourceInOCP)
				}
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

	if system.Labels[labels.LabelControlPlane] == labels.LabelValueControlPlaneOCP {
		if r.Config.EnableOPAControlPlaneReconciliation {
			log.Info("OPA Control Plane system reconcile starting")
			return r.ocpReconcile(ctx, log, system)
		}
		log.Info("OPA Control Plane Reconciliation have been disabled")
		return ctrl.Result{}, nil
	}

	if r.Config.EnableOPAControlPlaneReconciliationTestData {
		log.Info("OPA Control Plane Test Data flag is enabled - first lets do OCP reconciliation of test data")

		result, err := r.ocpReconcile(ctx, log, system)
		if err != nil {
			return result, err
		}
		log.Info("OPA Control Plane Test Data reconciliation completed - now do Styra DAS reconciliation")
	}

	if r.Config.EnableStyraReconciliation {
		log.Info("Styra DAS System reconcile starting")
		return r.styraReconcile(ctx, log, system)
	}
	log.Info("Styra DAS Reconciliation have been disabled")
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) ocpReconcile(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System) (ctrl.Result, error) {
	requirements := ocp.ToRequirements(r.Config.OPAControlPlaneConfig.DefaultRequirements)

	for _, datasource := range system.Spec.Datasources {
		datasource.Path = strings.ReplaceAll(datasource.Path, "/", "-")

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

	reconcileSystemBundleStart := time.Now()
	result, err = r.reconcileSystemBundle(ctx, uniqueName, requirements)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcileSystemBundleOcp").
		Observe(time.Since(reconcileSystemBundleStart).Seconds())
	if err != nil {
		return result, ctrlerr.Wrap(err, fmt.Sprintf("ocpReconcile: Could not reconcile system bundle: %s", uniqueName)).
			WithEvent(v1beta1.EventErrorUpdateBundle).
			WithSystemCondition(v1beta1.ConditionTypeSystemBundleUpdated)
	}
	system.SetCondition(v1beta1.ConditionTypeSystemBundleUpdated, metav1.ConditionTrue)

	// If the test data flag is enabled we skip the rest of the OCP reconciliation
	if r.Config.EnableOPAControlPlaneReconciliationTestData {
		log.Info("OCP reconciliation completed - skipping rest of OCP reconciliation due to test data flag")
		return ctrl.Result{}, nil
	}

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

	if system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) != nil &&
		*system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) == metav1.ConditionFalse {
		if r.Config.OPARestartEnabled() {
			log.Error(ctrlerr.New("Restarting OPA is not implemented yet"), "Error restarting OPA")
		}
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionTrue)
	}

	updateStatusStart := time.Now()
	err = r.Status().Update(ctx, system)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("updateStatusOcp").
		Observe(time.Since(updateStatusStart).Seconds())
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not change status.phase to Created").
			WithEvent(v1beta1.EventErrorPhaseToCreated)
	}

	msg := "OCP Reconciliation completed"
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

	bundleURL, err := url.JoinPath(r.Config.OPA.BundleServer.URL, r.Config.OPA.BundleServer.Path)
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Invalid OPA BundleServer URL or path").
			WithEvent(v1beta1.EventErrorConvertOPAConf).
			WithSystemCondition(v1beta1.ConditionTypeOPAConfigMapUpdated)
	}

	opaconf := ocp.OPAConfig{
		BundleService: &ocp.OPAServiceConfig{
			Name: "s3",
			URL:  bundleURL,
			Credentials: &ocp.ServiceCredentials{
				S3: &ocp.S3Signing{
					S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
				},
			},
		},
		LogService: &ocp.OPAServiceConfig{
			Name: "logs",
			URL:  r.Config.OPAControlPlaneConfig.DecisionAPIConfig.ServiceURL,
			Credentials: &ocp.ServiceCredentials{
				Bearer: &ocp.Bearer{
					TokenPath: "/run/secrets/kubernetes.io/serviceaccount/token",
				},
			},
		},
		DecisionLogReporting: r.Config.OPAControlPlaneConfig.DecisionAPIConfig.Reporting,
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

	reconcileS3CredentialsStart := time.Now()
	s3CredentialsRead, result, err := r.reconcileS3Credentials(
		ctx, log, system, uniqueName, secretName)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcileS3CredentialsOcp").
		Observe(time.Since(reconcileS3CredentialsStart).Seconds())
	if err != nil {
		return result, false, err
	}

	reconcilek8sOPASecret := time.Now()
	result, secretUpdated, err := r.reconcilek8sOPASecret(
		ctx, log, system, s3CredentialsRead, secretName)
	r.Metrics.ReconcileSegmentTime.
		WithLabelValues("reconcilek8sOPASecretOcp").
		Observe(time.Since(reconcilek8sOPASecret).Seconds())
	if err != nil {
		return result, false, err
	}
	log.Info("Reconciled OPA secret")
	return ctrl.Result{}, secretUpdated, nil
}

func (r *SystemReconciler) getk8sOPASecret(
	ctx context.Context,
	system *v1beta1.System,
	secretName string) (corev1.Secret, error) {
	var secret corev1.Secret

	nsName := types.NamespacedName{Name: secretName, Namespace: system.Namespace}
	err := r.Get(ctx, nsName, &secret)
	if err != nil {
		return secret, err
	}
	return secret, err
}

func (r *SystemReconciler) reconcilek8sOPASecret(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	s3Credentials s3.Credentials,
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
			s.Data = map[string][]byte{
				s3.AWSSecretNameKeyID:     []byte(s3Credentials.AccessKeyID),
				s3.AWSSecretNameSecretKey: []byte(s3Credentials.SecretAccessKey),
				s3.AWSSecretNameRegion:    []byte(s3Credentials.Region),
			}
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

	if (string(s.Data[s3.AWSSecretNameKeyID]) != s3Credentials.AccessKeyID) ||
		(string(s.Data[s3.AWSSecretNameSecretKey]) != s3Credentials.SecretAccessKey) ||
		(string(s.Data[s3.AWSSecretNameRegion]) != s3Credentials.Region) {
		log.Info("Secret mismatch. Updating secret.")
		s.Data = map[string][]byte{
			s3.AWSSecretNameKeyID:     []byte(s3Credentials.AccessKeyID),
			s3.AWSSecretNameSecretKey: []byte(s3Credentials.SecretAccessKey),
			s3.AWSSecretNameRegion:    []byte(s3Credentials.Region),
		}
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

func (r *SystemReconciler) reconcileS3Credentials(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	uniqueName string,
	secretName string,
) (s3.Credentials, ctrl.Result, error) {
	s3Credentials := s3.Credentials{}
	s3Credentials.Region = r.Config.UserCredentialHandler.S3.Region
	s3Credentials.AccessKeyID = fmt.Sprintf("Access-Key-%s-%s", r.Config.UserCredentialHandler.S3.Bucket, uniqueName)

	userExist, err := r.S3.UserExists(ctx, s3Credentials.AccessKeyID)
	if err != nil {
		return s3Credentials, ctrl.Result{}, ctrlerr.Wrap(err, "reconcileS3Credentials: could not call S3")
	}

	if userExist {
		secret, err := r.getk8sOPASecret(ctx, system, secretName)
		if err != nil && !k8serrors.IsNotFound(err) {
			return s3Credentials, ctrl.Result{}, ctrlerr.Wrap(err, "reconcileS3Credentials: error while getting secret from k8s")
		}
		if secret.Data == nil ||
			len(secret.Data[s3.AWSSecretNameKeyID]) == 0 ||
			len(secret.Data[s3.AWSSecretNameSecretKey]) == 0 {
			log.Info(
				"AccessKey exists, but no secret in k8s-secret. We need to update secretKey for accessKey",
				"accessKey", s3Credentials.AccessKeyID,
			)

			s3Credentials.SecretAccessKey, err = r.S3.SetNewUserSecretKey(ctx, s3Credentials.AccessKeyID)
			if err != nil {
				log.Error(err, "failed to apply new secretKey for accessKey", "accessKey", s3Credentials.AccessKeyID)
				return s3Credentials, ctrl.Result{}, ctrlerr.Wrap(
					err, "reconcileS3Credentials: failed to apply new secretKey for accessKey",
				)
			}
			log.Info("SecretKey updated for existing accessKey", "accessKey", s3Credentials.AccessKeyID)
			return s3Credentials, ctrl.Result{}, nil
		}
		log.Info("User exists in MinIO and secret found in k8s. We must assume it is valid",
			"accessKey", s3Credentials.AccessKeyID)
		s3Credentials.SecretAccessKey = string(secret.Data[s3.AWSSecretNameSecretKey])
		return s3Credentials, ctrl.Result{}, nil
	}

	if !userExist {
		log.Info("AccessKey does not exist, creating new accessKey", "accessKey", s3Credentials.AccessKeyID)
		// create read only user for this bundle
		s3Credentials.SecretAccessKey, err = r.S3.CreateSystemBundleUser(
			ctx, s3Credentials.AccessKeyID, r.Config.UserCredentialHandler.S3.Bucket, uniqueName,
		)
		if err != nil {
			log.Error(err, "failed to create accessKey", "accessKey", s3Credentials.AccessKeyID)
			return s3Credentials, ctrl.Result{}, ctrlerr.Wrap(err, "reconcileS3Credentials: could not create accessKey")
		}
		log.Info("AccessKey created", "accessKey", s3Credentials.AccessKeyID)
	}

	return s3Credentials, ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcileSystemBundle(
	ctx context.Context,
	uniqueName string,
	requirements []ocp.Requirement) (ctrl.Result, error) {
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
		Requirements:  requirements,
	}
	err := r.OCP.PutBundle(ctx, bundle)

	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "ocpReconcile: could not create or update bundle in OCP")
	}
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) reconcileSystemSource(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	uniqueName string) (ctrl.Result, error) {

	if system.Spec.SourceControl == nil {
		return ctrl.Result{}, ctrlerr.New("reconcileSystemSource: no source control configured on system")
	}

	valid, err := isURLValid(system.Spec.SourceControl.Origin.URL)
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Error while validating URL")
	}
	if !valid {
		return ctrl.Result{}, ctrlerr.New("Invalid URL for source control")
	}

	gitConfig := &ocp.GitConfig{
		Repo:          system.Spec.SourceControl.Origin.URL,
		IncludedFiles: []string{"*.rego"},
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

	_, err = r.OCP.PutSource(ctx, uniqueName, &ocp.PutSourceRequest{
		Name: uniqueName,
		Git:  gitConfig,
	})
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "reconcileSystemSource: could not create or update source in OCP")
	}
	log.Info("OCP source upserted", "source", uniqueName)
	return ctrl.Result{}, nil
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

func (r *SystemReconciler) styraReconcile(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System) (ctrl.Result, error) {
	var (
		cfg *styra.SystemConfig
		err error
	)

	systemID := system.Status.ID
	migrationID := system.ObjectMeta.Annotations["styra-controller/migration-id"]
	if r.Config.EnableMigrations && systemID == "" && migrationID != "" {
		log.Info(fmt.Sprintf("Use migrationId(%s) to fetch system from Styra DAS", migrationID))
		getSystemStart := time.Now()
		cfg, err = r.getSystem(ctx, log, migrationID)
		r.Metrics.ReconcileSegmentTime.WithLabelValues("getSystem").Observe(time.Since(getSystemStart).Seconds())

		if err != nil {
			return ctrl.Result{}, err
		}
		if err := r.reconcileID(ctx, log, system, migrationID); err != nil {
			return ctrl.Result{}, err
		}
	} else if systemID != "" {
		getSystemStart := time.Now()
		cfg, err = r.getSystem(ctx, log, systemID)
		r.Metrics.ReconcileSegmentTime.WithLabelValues("getSystem").Observe(time.Since(getSystemStart).Seconds())

		if err != nil {
			var serr *httperror.HTTPError
			if errors.As(err, &serr) && serr.StatusCode == http.StatusNotFound {
				createSystemStart := time.Now()
				res, err := r.createSystemWithID(ctx, log, system, systemID)
				if err != nil {
					if errors.As(err, &serr) && serr.StatusCode == http.StatusConflict {
						log.Info("System still found in Styra Cache - creating new system")
						res, err := r.createSystem(ctx, log, system)
						r.Metrics.ReconcileSegmentTime.WithLabelValues("createSystemBadId").
							Observe(time.Since(createSystemStart).Seconds())

						if err != nil {
							return ctrl.Result{}, err
						}

						err = r.postSystemCreation(ctx, log, res.SystemConfig.ID, system)
						if err != nil {
							return ctrl.Result{}, err
						}
					} else {
						return ctrl.Result{}, err
					}
				} else {
					r.Metrics.ReconcileSegmentTime.WithLabelValues("createSystemWithId").
						Observe(time.Since(createSystemStart).Seconds())
					err = r.postSystemCreation(ctx, log, res.SystemConfig.ID, system)
					if err != nil {
						return ctrl.Result{}, err
					}
				}
			} else {
				return ctrl.Result{}, err
			}
		}
	} else {
		displayName := system.DisplayName(r.Config.SystemPrefix, r.Config.SystemSuffix)

		getSystemByNameStart := time.Now()
		cfg, err = r.getSystemByName(ctx, log, displayName)
		r.Metrics.ReconcileSegmentTime.WithLabelValues("getSystemByName").
			Observe(time.Since(getSystemByNameStart).Seconds())
		if err != nil {
			return ctrl.Result{}, err
		}
		if cfg != nil {
			reconcileIDStart := time.Now()
			err = r.reconcileID(ctx, log, system, cfg.ID)
			r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileID").
				Observe(time.Since(reconcileIDStart).Seconds())
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			createSystemStart := time.Now()
			res, err := r.createSystem(ctx, log, system)
			r.Metrics.ReconcileSegmentTime.WithLabelValues("createSystem").
				Observe(time.Since(createSystemStart).Seconds())
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.postSystemCreation(ctx, log, res.SystemConfig.ID, system)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	system.SetCondition(v1beta1.ConditionTypeCreatedInStyra, metav1.ConditionTrue)

	reconcileCredentialsStart := time.Now()
	result, err := r.reconcileCredentials(ctx, log, system)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileCredentials").
		Observe(time.Since(reconcileCredentialsStart).Seconds())
	if err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeGitCredentialsUpdated, metav1.ConditionTrue)

	needsUpdate, err := r.systemNeedsUpdate(log, system, cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	if needsUpdate {
		updateSystemStart := time.Now()
		cfg, err = r.updateSystem(ctx, log, system)
		r.Metrics.ReconcileSegmentTime.WithLabelValues("updateSystem").
			Observe(time.Since(updateSystemStart).Seconds())
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	system.SetCondition(v1beta1.ConditionTypeSystemConfigUpdated, metav1.ConditionTrue)

	reconcileSubjectsStart := time.Now()
	result, err = r.reconcileSubjects(ctx, log, system)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileSubjects").
		Observe(time.Since(reconcileSubjectsStart).Seconds())
	if err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeSubjectsUpdated, metav1.ConditionTrue)

	reconcileDatasourcesStart := time.Now()
	result, err = r.reconcileDatasources(ctx, log, system, cfg)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileDatasources").
		Observe(time.Since(reconcileDatasourcesStart).Seconds())
	if err != nil {
		return result, err
	}
	system.SetCondition(v1beta1.ConditionTypeDatasourcesUpdated, metav1.ConditionTrue)

	getOPAConfigStart := time.Now()
	opaConfig, err := r.Styra.GetOPAConfig(ctx, system.Status.ID)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("getOPAConfig").
		Observe(time.Since(getOPAConfigStart).Seconds())
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get OPA config from styra API").
			WithEvent(v1beta1.EventErrorFetchOPAConfig).
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	reconcileOPATokenStart := time.Now()
	result, updatedToken, err := r.reconcileOPAToken(ctx, log, system, opaConfig.Token)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileOPAToken").
		Observe(time.Since(reconcileOPATokenStart).Seconds())
	if err != nil {
		return result, err
	}

	if updatedToken {
		var conditionType v1beta1.ConditionType
		if system.Spec.LocalPlane == nil {
			conditionType = v1beta1.ConditionTypeOPAUpToDate
		} else {
			conditionType = v1beta1.ConditionTypeSLPUpToDate
		}
		system.SetCondition(conditionType, metav1.ConditionFalse)
		err = r.Status().Update(ctx, system)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err,
				"Could not update system to reflect that Styra token in pods is outdated").
				WithEvent(v1beta1.EventErrorUpdateStatus).
				WithSystemCondition(conditionType)
		}
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPATokenUpdated, metav1.ConditionTrue)

	reconcileOPAConfigMapStart := time.Now()
	result, updatedOPAConfigMap, err := r.reconcileOPAConfigMap(ctx, log, system, opaConfig)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileOPAConfigMap").
		Observe(time.Since(reconcileOPAConfigMapStart).Seconds())
	if err != nil {
		return result, err
	}
	if updatedOPAConfigMap {
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionFalse)

		err = r.Status().Update(ctx, system)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err,
				"Could not update system to reflect that Styra token in pods is outdated").
				WithEvent(v1beta1.EventErrorUpdateStatus).
				WithSystemCondition(v1beta1.ConditionTypeOPAUpToDate)
		}
		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeOPAConfigMapUpdated, metav1.ConditionTrue)

	reconcileSLPConfigMapStart := time.Now()
	result, updatedSLPConfigMap, err := r.reconcileSLPConfigMap(ctx, log, system, opaConfig)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileSLPConfigMap").
		Observe(time.Since(reconcileSLPConfigMapStart).Seconds())
	if err != nil {
		return result, err
	}
	if updatedSLPConfigMap {
		system.SetCondition(v1beta1.ConditionTypeSLPUpToDate, metav1.ConditionFalse)

		err = r.Status().Update(ctx, system)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err,
				"Could not update system to reflect that Styra token in pods is outdated").
				WithEvent(v1beta1.EventErrorUpdateStatus).
				WithSystemCondition(v1beta1.ConditionTypeSLPUpToDate)
		}

		return result, nil
	}
	system.SetCondition(v1beta1.ConditionTypeSLPConfigMapUpdated, metav1.ConditionTrue)

	system.Status.Ready = true
	system.Status.Phase = v1beta1.SystemPhaseCreated
	system.Status.FailureMessage = ""

	if system.GetCondition(v1beta1.ConditionTypeSLPUpToDate) != nil &&
		*system.GetCondition(v1beta1.ConditionTypeSLPUpToDate) == metav1.ConditionFalse {
		if r.Config.SLPRestartEnabled() {
			res, err := r.restartSLPs(ctx, log, system)
			if err != nil {
				log.Error(err, "Error restarting SLPs")
				return res, ctrlerr.Wrap(err, "Error restarting SLPs").
					WithEvent(v1beta1.EventErrorRestartSLPs).
					WithSystemCondition(v1beta1.ConditionTypeSLPUpToDate)
			}
		}
		system.SetCondition(v1beta1.ConditionTypeSLPUpToDate, metav1.ConditionTrue)
	}

	if system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) != nil &&
		*system.GetCondition(v1beta1.ConditionTypeOPAUpToDate) == metav1.ConditionFalse {
		if r.Config.OPARestartEnabled() {
			log.Error(errors.New("Restarting OPA is not implemented yet"), "Error restarting OPA")
		}
		system.SetCondition(v1beta1.ConditionTypeOPAUpToDate, metav1.ConditionTrue)
	}

	updateStatusStart := time.Now()
	err = r.Status().Update(ctx, system)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("updateStatus").Observe(time.Since(updateStatusStart).Seconds())
	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not change status.phase to Created").
			WithEvent(v1beta1.EventErrorPhaseToCreated)
	}

	msg := "Styra Reconciliation completed"
	r.Recorder.Event(system, corev1.EventTypeNormal, "ReconciliationCompleted", msg)
	log.Info(msg)
	return ctrl.Result{}, nil
}

func (r *SystemReconciler) postSystemCreation(
	ctx context.Context,
	log logr.Logger,
	id string,
	system *v1beta1.System,
) error {
	deleteDefaultPolicyStart := time.Now()
	err := r.deleteDefaultPolicies(ctx, log, id)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("deleteDefaultPolicies").
		Observe(time.Since(deleteDefaultPolicyStart).Seconds())
	if err != nil {
		return err
	}

	reconcileIDStart := time.Now()
	err = r.reconcileID(ctx, log, system, id)
	r.Metrics.ReconcileSegmentTime.WithLabelValues("reconcileID").
		Observe(time.Since(reconcileIDStart).Seconds())

	if err != nil {
		return err
	}
	return nil
}

func (r *SystemReconciler) restartSLPs(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (ctrl.Result, error) {
	if strings.ToLower(r.Config.PodRestart.SLPRestart.DeploymentType) != "statefulset" {
		log.Info("Restarting SLPs is not supported for this deployment type",
			"deploymentType",
			r.Config.PodRestart.SLPRestart.DeploymentType,
		)
		return ctrl.Result{}, nil
	}

	log.Info("Restarting SLPs")
	nsName := types.NamespacedName{Name: system.Spec.LocalPlane.Name, Namespace: system.Namespace}
	var sts appsv1.StatefulSet
	if err := r.Get(ctx, nsName, &sts); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("System has SLP but no SLP found with that name")
			return ctrl.Result{}, ctrlerr.Wrap(err, "SLP statefulset not found for system with SLP enabled").
				WithEvent(v1beta1.EventErrorStatefulSetNotFound).
				WithSystemCondition(v1beta1.ConditionTypeSLPUpToDate)
		}
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get StatefulSet").
			WithEvent(v1beta1.EventErrorGetStatefulSet).
			WithSystemCondition(v1beta1.ConditionTypeSLPUpToDate)
	}

	patch := []byte(fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`,
		time.Now().Format(time.RFC3339),
	))
	if err := r.Patch(ctx, &sts, client.RawPatch(types.StrategicMergePatchType, patch)); err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not patch StatefulSet").
			WithEvent(v1beta1.EventErrorPatchStatefulSet).
			WithSystemCondition(v1beta1.ConditionTypeSLPUpToDate)
	}

	log.Info("Restarted SLPs")
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
			WithEvent(v1beta1.EventErrorFetchSystemFromStyra)
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
			WithEvent(v1beta1.EventErrorFetchSystemFromStyra)
	}
	if res.SystemConfig != nil {
		log.Info(fmt.Sprintf("Fetched system %v from Styra API", name))
	} else {
		log.Info(fmt.Sprintf("System %v does not exist in Styra DAS.", name))
	}
	return res.SystemConfig, nil
}

func (r *SystemReconciler) createSystemWithID(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
	id string,
) (*styra.PutSystemResponse, error) {
	log.Info("Creating system in Styra with ID")
	cfg, err := r.specToSystemConfig(system)
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Error while reading system spec").
			WithEvent(v1beta1.EventErrorCreateSystemInStyra).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

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
		log.Info("PUT system request", "request", string(bs))
	}

	headers := map[string]string{"If-None-Match": "*"}
	res, err := r.Styra.PutSystem(ctx, &styra.PutSystemRequest{SystemConfig: cfg}, id, headers)
	if err != nil {
		return nil, ctrlerr.Wrap(err, fmt.Sprintf("Could not create system in Styra with id %s", id)).
			WithEvent(v1beta1.EventErrorCreateSystemInStyra).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}
	return res, nil
}

func (r *SystemReconciler) createSystem(
	ctx context.Context,
	log logr.Logger,
	system *v1beta1.System,
) (*styra.CreateSystemResponse, error) {
	log.Info("Creating system in Styra")

	cfg, err := r.specToSystemConfig(system)
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Error while reading system spec").
			WithEvent(v1beta1.EventErrorCreateSystemInStyra).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

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
			WithEvent(v1beta1.EventErrorCreateSystemInStyra).
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
					WithEvent(v1beta1.EventErrorCredentialsSecretNotFound).
					WithSystemCondition(v1beta1.ConditionTypeGitCredentialsUpdated)
			}
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not fetch credentials Secret").
				WithEvent(v1beta1.EventErrorCredentialsSecretCouldNotFetch).
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
			WithEvent(v1beta1.EventErrorCreateUpdateSecret).
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
			WithEvent(v1beta1.EventErrorDeleteDefaultPolicy).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	if _, err := r.Styra.DeletePolicy(ctx, testPolicyName); err != nil {
		return ctrlerr.Wrap(err, "Could not delete default policy").
			WithEvent(v1beta1.EventErrorDeleteDefaultPolicy).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	return nil
}

func (r *SystemReconciler) reconcileID(ctx context.Context, log logr.Logger, system *v1beta1.System, id string) error {
	log.Info("Reconciling ID")

	if id == "" {
		return ctrlerr.New("ID is empty").
			WithEvent(v1beta1.EventErrorReconcileID).
			WithSystemCondition(v1beta1.ConditionTypeCreatedInStyra)
	}

	system.Status.ID = id

	if err := r.Status().Update(ctx, system); err != nil {
		return ctrlerr.Wrap(err, "Could not set system ID on System").
			WithEvent(v1beta1.EventErrorReconcileID).
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

	usersResponse, fromCache, err := r.Styra.GetUsers(ctx)
	if fromCache {
		log.Info("Users response from cache")
	} else {
		log.Info("Users response from Styra API - cache updated")
	}

	if err != nil {
		return ctrl.Result{}, ctrlerr.Wrap(err, "Could not get users from Styra API").
			WithEvent(v1beta1.EventErrorGetUsersFromStyra).
			WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
	}

	for _, subject := range system.Spec.Subjects {
		if subject.IsUser() {
			log := log.WithValues("user", subject.Name)
			log.Info("Checking if user exists")

			found := false
			for _, user := range usersResponse.Users {
				if user.ID == subject.Name {
					found = true
					break
				}
			}

			if !found {
				log.Info("User does not exist in Styra. Creating user and invalidate user cache")

				// Invalidate cache to ensure we get the new user in the next reconciliation
				r.Styra.InvalidateCache()

				_, err := r.Styra.CreateInvitation(ctx, false, subject.Name)
				if err != nil {
					return ctrl.Result{}, ctrlerr.Wrap(err, "Could not create user in Styra").
						WithEvent(v1beta1.EventErrorCreateInvitation).
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
			WithEvent(v1beta1.EventErrorGetSystemRolebindings).
			WithSystemCondition(v1beta1.ConditionTypeSubjectsUpdated)
	}

	styraRoleBindingsByRole := map[styra.Role]*styra.RoleBindingConfig{}
	for _, rb := range res.Rolebindings {
		styraRoleBindingsByRole[rb.RoleID] = rb
	}
	controllerSystemUserRoles := make([]styra.Role, len(r.Config.SystemUserRoles))
	for i, role := range r.Config.SystemUserRoles {
		controllerSystemUserRoles[i] = styra.Role(role)
	}
	systemRolebindingSubjects := createRolebindingSubjects(
		system.Spec.Subjects,
		r.Config.SSO,
	)
	for _, role := range controllerSystemUserRoles {
		rb, ok := styraRoleBindingsByRole[role]

		var subjects []*styra.Subject
		if ok {
			// We only want want the controller to manage user and claim subjects
			for _, subject := range rb.Subjects {
				if subject.Kind != styra.SubjectKindUser && subject.Kind != styra.SubjectKindClaim {
					subjects = append(subjects, subject)
				}
			}
		}

		subjects = append(subjects, systemRolebindingSubjects...)

		if !ok {
			if err := r.createRoleBinding(ctx, log, system, role, systemRolebindingSubjects); err != nil {
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
	for _, role := range controllerSystemUserRoles {
		roles[role] = struct{}{}
	}
	for _, rb := range res.Rolebindings {
		if _, ok := roles[rb.RoleID]; ok {
			continue
		}

		var subjects []*styra.Subject

		// We only want want the controller to manage user and claim subjects
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

	log.Info("Reconciled subjects")
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
			WithEvent(v1beta1.EventErrorCreateRolebinding).
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
			WithEvent(v1beta1.EventErrorUpdateRolebinding).
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
					WithEvent(v1beta1.EventErrorUpsertDatasource).
					WithSystemCondition(v1beta1.ConditionTypeDatasourcesUpdated)
			}

			if r.WebhookClient != nil {
				log.Info("Calling datasource changed webhook")
				if err := r.WebhookClient.SystemDatasourceChanged(ctx, log, system.Status.ID, id); err != nil {
					err = ctrlerr.Wrap(err, "Could not call datasource changed webhook").
						WithEvent(v1beta1.EventErrorCallWebhook).
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

		ignore, err := config.MatchesIgnorePattern(r.Config.DatasourceIgnorePatterns, ds.ID)
		if err != nil {
			return ctrl.Result{}, ctrlerr.Wrap(err, "Could not check if system datasource should be ignored")
		}

		if ignore {
			log.Info("Datasource is ignored", "id", ds.ID)
			continue
		}

		if _, ok := expectedByID[ds.ID]; !ok {
			log := log.WithValues("datasourceID", ds.ID)
			log.Info("Deleting undeclared datasource", "id", ds.ID)
			if _, err := r.Styra.DeleteDatasource(ctx, ds.ID); err != nil {
				return ctrl.Result{}, ctrlerr.Wrap(err, "Could not delete datasource").
					WithEvent(v1beta1.EventErrorDeleteDatasource).
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
			WithEvent(v1beta1.EventErrorOPATokenSecretNoToken).
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
					WithEvent(v1beta1.EventErrorOwnerRefOPATokenSecret).
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			if err := r.Create(ctx, &s); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create OPA token Secret").
					WithEvent(v1beta1.EventErrorCreateOPATokenSecret).
					WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch OPA token Secret").
			WithEvent(v1beta1.EventErrorFetchOPATokenSecret).
			WithSystemCondition(v1beta1.ConditionTypeOPATokenUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&s, system) {
		return ctrl.Result{}, false, ctrlerr.New("Existing secret is not owned by controller. Skipping update").
			WithEvent(v1beta1.EventErrorSecretNotOwnedByController).
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
				WithEvent(v1beta1.EventErrorUpdateOPATokenSecret).
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
		expectedOPAConfigMap, err = k8sconv.OPAConfToK8sOPAConfigMapNoSLP(opaconf, r.Config.OPA, customConfig)
	} else {
		slpURL := fmt.Sprintf("http://%s/v1", system.Spec.LocalPlane.Name)
		expectedOPAConfigMap, err = k8sconv.OPAConfToK8sOPAConfigMap(opaconf, slpURL, r.Config.OPA, customConfig)
	}
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

	expectedSLPConfigMap, err := k8sconv.OPAConfToK8sSLPConfigMap(opaconf)
	if err != nil {
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not convert OPA Conf to SLP ConfigMap").
			WithEvent(v1beta1.EventErrorConvertOPAConf).
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
					WithEvent(v1beta1.EventErrorOwnerRefSLPConfigMap).
					WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
			}
			if err := r.Create(ctx, &cm); err != nil {
				return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not create SLP ConfigMap").
					WithEvent(v1beta1.EventErrorCreateSLPConfigMap).
					WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
			}
			return ctrl.Result{}, true, nil
		}
		return ctrl.Result{}, false, ctrlerr.Wrap(err, "Could not fetch SLP ConfigMap").
			WithEvent(v1beta1.EventErrorFetchSLPConfigMap).
			WithSystemCondition(v1beta1.ConditionTypeSLPConfigMapUpdated)
	}

	update := false

	if !metav1.IsControlledBy(&cm, system) {
		return ctrl.Result{}, false, ctrlerr.New("ConfigMap already exists and is not owned by controller").
			WithEvent(v1beta1.EventErrorFetchSLPConfigMap).
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
				WithEvent(v1beta1.EventErrorUpdateSLPConfigmap).
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
	cfg, err := r.specToSystemConfig(system)
	if err != nil {
		return nil, ctrlerr.Wrap(err, "Error while reading system spec").
			WithEvent(v1beta1.EventErrorUpdateSystem).
			WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
	}

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
		var styrahttperr *httperror.HTTPError
		if errors.As(err, &styrahttperr) {
			errMsg = fmt.Sprintf("Could not update Styra system. Error %s", styrahttperr.Error())
		}
		return nil, ctrlerr.Wrap(err, errMsg).
			WithEvent(v1beta1.EventErrorUpdateSystem).
			WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
	}

	log.Info("Updated system")
	return res.SystemConfig, nil
}

func (r *SystemReconciler) specToSystemConfig(system *v1beta1.System) (*styra.SystemConfig, error) {
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
		valid, err := isURLValid(system.Spec.SourceControl.Origin.URL)
		if err != nil {
			return nil, ctrlerr.Wrap(err, "Error while validating URL").
				WithEvent(v1beta1.EventErrorUpdateSystem).
				WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
		}

		if !valid {
			return nil, ctrlerr.New("Invalid URL for source control").
				WithEvent(v1beta1.EventErrorUpdateSystem).
				WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
		}

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

	return cfg, nil
}

func isURLValid(u string) (bool, error) {
	//empty url is valid
	if strings.TrimSpace(u) == "" {
		return true, nil
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		return false, err
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false, nil
	}

	// Regular expression to match valid URL path characters
	valid, err := regexp.MatchString(`^[a-zA-Z0-9\-._~!$&'()*+,;=:@/]*$`, parsedURL.Path)
	if err != nil {
		return false, err
	}

	return valid, nil
}

func (r *SystemReconciler) systemNeedsUpdate(
	log logr.Logger,
	system *v1beta1.System,
	cfg *styra.SystemConfig,
) (bool, error) {
	if cfg == nil {
		log.Info("System needs update: cfg is nil")
		return true, nil
	}

	if cfg.ReadOnly != r.Config.ReadOnly {
		log.Info("System needs update: read only is not equal")
		return true, nil
	}

	expectedModel, err := r.specToSystemConfig(system)
	if err != nil {
		return true, ctrlerr.Wrap(err, "Error while reading system spec").
			WithEvent(v1beta1.EventErrorUpdateSystem).
			WithSystemCondition(v1beta1.ConditionTypeSystemConfigUpdated)
	}

	if cfg.BundleDownload == nil || cfg.BundleDownload.DeltaBundles != expectedModel.BundleDownload.DeltaBundles {
		log.Info("System needs update: Deltabundle setting not equal")
		return true, nil
	}

	if !reflect.DeepEqual(expectedModel.SourceControl, cfg.SourceControl) {
		log.Info("System needs update: source control is not equal")
		return true, nil
	}

	namesAreEqual := expectedModel.Name == cfg.Name
	if !namesAreEqual {
		log.Info("System needs update: system names are not are not equal")
		return true, nil
	}

	dmsAreEqual := styra.DecisionMappingsEquals(expectedModel.DecisionMappings, cfg.DecisionMappings)
	if !dmsAreEqual {
		log.Info("System needs update: decision mappings are not equal")
		return true, nil
	}
	return false, nil
}

// CreateDefaultRequirements creates all the configured default sources in OCP
func (r *SystemReconciler) CreateDefaultRequirements(ctx context.Context, log logr.Logger) error {
	if r.Config.EnableOPAControlPlaneReconciliation || r.Config.EnableOPAControlPlaneReconciliationTestData {
		log.Info("Creating ocp default requirements")
		for _, defaultRequirement := range r.Config.OPAControlPlaneConfig.DefaultRequirements {
			_, err := r.createSourceIfNotExists(ctx, log, v1beta1.Datasource{Path: defaultRequirement})
			if err != nil {
				return err
			}
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
