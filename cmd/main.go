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

// Package main is the main entrypoint used when running the controller.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/internal/config"
	controllers "github.com/bankdata/styra-controller/internal/controller/styra"
	"github.com/bankdata/styra-controller/internal/webhook"
	webhookstyrav1alpha1 "github.com/bankdata/styra-controller/internal/webhook/styra/v1alpha1"
	webhookstyrav1beta1 "github.com/bankdata/styra-controller/internal/webhook/styra/v1beta1"
	"github.com/bankdata/styra-controller/pkg/styra"
	//+kubebuilder:scaffold:imports
)

const (
	logLevelDebug = 1
)

var (
	scheme = runtime.NewScheme()

	// these are set by goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(styrav1alpha1.AddToScheme(scheme))
	utilruntime.Must(styrav1beta1.AddToScheme(scheme))
	utilruntime.Must(configv2alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		configFile   string
		printVersion bool
	)

	flag.StringVar(&configFile, "config", "/etc/styra-controller/config.yaml",
		"The controller will load its initial configuration from this file. ")
	flag.BoolVar(&printVersion, "version", false, "show version information")
	flag.Parse()

	if printVersion {
		fmt.Printf(
			"styra-controller (https://github.com/bankdata/styra-controller) version %s\ncommit %s\ndate %s\n",
			version, commit, date,
		)
		return
	}

	log := ctrl.Log.WithName("setup")
	logDebug := log.V(logLevelDebug)
	restCfg := ctrl.GetConfigOrDie()

	logDebug.Info(
		"rest config details",
		"timeout", restCfg.Timeout,
		"host", restCfg.Host,
		"apiPath", restCfg.APIPath,
	)

	ctrlConfig, err := config.Load(configFile, scheme)
	if err != nil {
		log.Error(err, "unable to load the config file")
		exit(err)
	}

	ctrl.SetLogger(zap.New(
		zap.UseDevMode(ctrlConfig.LogLevel >= logLevelDebug),
		zap.Level(zapcore.Level(-ctrlConfig.LogLevel)),
	))

	styraToken, err1 := config.TokenFromConfig(ctrlConfig)
	if err1 != nil {
		log.Error(err1, "Unable to load styra token")
		exit(err1)
	}

	options := config.OptionsFromConfig(ctrlConfig, scheme)

	if ctrlConfig.Sentry != nil {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         ctrlConfig.Sentry.DSN,
			Environment: ctrlConfig.Sentry.Environment,
			Release:     version,
			Debug:       ctrlConfig.Sentry.Debug,
			HTTPSProxy:  ctrlConfig.Sentry.HTTPSProxy,
		})
		if err != nil {
			log.Error(err, "failed to init sentry")
			exit(err)
		}
		defer sentry.Flush(2 * time.Second)
	}

	mgr, err := ctrl.NewManager(restCfg, options)
	if err != nil {
		log.Error(err, "unable to start manager")
		exit(err)
	}

	roles := make([]styra.Role, len(ctrlConfig.SystemUserRoles))
	for i, role := range ctrlConfig.SystemUserRoles {
		roles[i] = styra.Role(role)
	}

	styraHostURL := strings.TrimSuffix(ctrlConfig.Styra.Address, "/")
	styraClient := styra.New(styraHostURL, styraToken)

	if err := configureExporter(
		styraClient, ctrlConfig.DecisionsExporter, configv2alpha2.ExporterConfigTypeDecisions); err != nil {
		log.Error(err, fmt.Sprintf("unable to configure %s", configv2alpha2.ExporterConfigTypeDecisions))
	}

	if err := configureExporter(
		styraClient, ctrlConfig.ActivityExporter, configv2alpha2.ExporterConfigTypeActivity); err != nil {
		log.Error(err, fmt.Sprintf("unable to configure %s", configv2alpha2.ExporterConfigTypeActivity))
	}

	// System Controller
	systemReadyMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controller_system_status_ready",
			Help: "Show if a system is in status ready",
		},
		[]string{"system_name", "namespace", "system_id"},
	)

	if err := metrics.Registry.Register(systemReadyMetric); err != nil {
		err := errors.Wrap(err, "could not register controller_system_status_ready metric")
		log.Error(err, err.Error())
		exit(err)
	}

	reconcileSegmentTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "controller_system_reconcile_segment_seconds",
			Help:    "Time taken to perform one segment of reconciling a system",
			Buckets: prometheus.DefBuckets,
		}, []string{"segment"},
	)

	if err := metrics.Registry.Register(reconcileSegmentTimeMetric); err != nil {
		err := errors.Wrap(err, "could not register reconcileSegmentTimeMetric")
		log.Error(err, err.Error())
		exit(err)
	}

	reconcileTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "controller_system_reconcile_seconds",
			Help:    "Time taken to reconcile a system",
			Buckets: prometheus.DefBuckets,
		}, []string{"result"},
	)

	if err := metrics.Registry.Register(reconcileTimeMetric); err != nil {
		err := errors.Wrap(err, "could not register reconcileTimeMetric")
		log.Error(err, err.Error())
		exit(err)
	}

	systemMetrics := &controllers.SystemReconcilerMetrics{
		ControllerSystemStatusReady: systemReadyMetric,
		ReconcileSegmentTime:        reconcileSegmentTimeMetric,
		ReconcileTime:               reconcileTimeMetric,
	}

	r1 := &controllers.SystemReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Styra:    styraClient,
		Recorder: mgr.GetEventRecorderFor("system-controller"),
		Metrics:  systemMetrics,
		Config:   ctrlConfig,
	}

	if ctrlConfig.NotificationWebhooks != nil && ctrlConfig.NotificationWebhooks.SystemDatasourceChanged != "" {
		r1.WebhookClient = webhook.New(ctrlConfig.NotificationWebhooks.SystemDatasourceChanged, "")
	}

	if err = r1.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "System")
		exit(err)
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = webhookstyrav1beta1.SetupSystemWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "System")
			os.Exit(1)
		}
	}

	libraryReconciler := &controllers.LibraryReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: ctrlConfig,
		Styra:  styraClient,
	}

	if ctrlConfig.NotificationWebhooks != nil && ctrlConfig.NotificationWebhooks.LibraryDatasourceChanged != "" {
		libraryReconciler.WebhookClient = webhook.New("", ctrlConfig.NotificationWebhooks.LibraryDatasourceChanged)
	}

	if err = libraryReconciler.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Library")
		os.Exit(1)
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = webhookstyrav1alpha1.SetupLibraryWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "Library")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		exit(err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		exit(err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		exit(err)
	}
}

func exit(err error) {
	sentry.CaptureException(err)
	sentry.Flush(2 * time.Second)
	os.Exit(1)
}

func configureExporter(
	styraClient styra.ClientInterface,
	exporterConfig *configv2alpha2.ExporterConfig,
	exporterType configv2alpha2.ExporterConfigType) error {
	if exporterConfig == nil {
		ctrl.Log.Info(fmt.Sprintf("no exporter configuration found for %s", exporterType))
		return nil
	}
	clientCertName := exporterConfig.Kafka.TLS.ClientCertificateName

	if !exporterConfig.Enabled {
		ctrl.Log.Info(fmt.Sprintf("Now removing exporter %s", exporterType))

		_, err := styraClient.DeleteSecret(context.Background(), clientCertName)
		if err != nil {
			ctrl.Log.Error(err, fmt.Sprintf("failed to delete secret %s for %s", clientCertName, exporterType))
			return err
		}

		if exporterType == configv2alpha2.ExporterConfigTypeActivity {
			rawJSON := json.RawMessage("{\"activity_exporter\": null}")
			_, err = styraClient.UpdateWorkspaceRaw(context.Background(), rawJSON)
		} else if exporterType == configv2alpha2.ExporterConfigTypeDecisions {
			rawJSON := json.RawMessage("{\"decisions_exporter\": null}")
			_, err = styraClient.UpdateWorkspaceRaw(context.Background(), rawJSON)
		}
		if err != nil {
			ctrl.Log.Error(err, fmt.Sprintf("could not remove %s", exporterType))
			return err
		}
		ctrl.Log.Info(fmt.Sprintf("%s removed", exporterType))

		return nil
	}

	ctrl.Log.Info(fmt.Sprintf("configuring %s", exporterType))

	_, err := styraClient.CreateUpdateSecret(context.Background(), clientCertName, &styra.CreateUpdateSecretsRequest{
		Description: "Client certificate for Kafka",
		// Secret name should be client cert and secret should be client key
		Name:   strings.TrimSuffix(exporterConfig.Kafka.TLS.ClientCertificate, "\n"),
		Secret: strings.TrimSuffix(exporterConfig.Kafka.TLS.ClientKey, "\n"),
	},
	)
	if err != nil {
		ctrl.Log.Error(err, fmt.Sprintf("failed to upload secret %s for %s", clientCertName, exporterType))
		return err
	}

	exportConfig := &styra.ExporterConfig{
		Interval: exporterConfig.Interval,
		Kafka: &styra.KafkaConfig{
			Authentication: "TLS",
			Brokers:        exporterConfig.Kafka.Brokers,
			RequredAcks:    exporterConfig.Kafka.RequiredAcks,
			Topic:          exporterConfig.Kafka.Topic,
			TLS: &styra.KafkaTLS{
				ClientCert:         clientCertName,
				RootCA:             strings.TrimSuffix(exporterConfig.Kafka.TLS.RootCA, "\n"),
				InsecureSkipVerify: exporterConfig.Kafka.TLS.InsecureSkipVerify,
			},
		},
	}

	if exporterType == "ActivityExporter" {
		_, err = styraClient.UpdateWorkspace(context.Background(), &styra.UpdateWorkspaceRequest{
			ActivityExporter: exportConfig,
		})
	} else if exporterType == "DecisionsExporter" {
		_, err = styraClient.UpdateWorkspace(context.Background(), &styra.UpdateWorkspaceRequest{
			DecisionsExporter: exportConfig,
		})
	}

	if err != nil {
		ctrl.Log.Error(err, fmt.Sprintf("could not update workspace configuration for %s", exporterType))
		return err
	}

	ctrl.Log.Info(fmt.Sprintf("successfully configured %s", exporterType))
	return nil
}
