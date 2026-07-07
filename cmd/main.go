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

// Package main is the main entrypoint used when running the controller.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

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
	"github.com/bankdata/styra-controller/pkg/ocp"
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
		configFiles  config.StringSlice
		printVersion bool
	)

	flag.Var(&configFiles, "config",
		"Config file to load. Can be specified multiple times; files are deep-merged in order. "+
			"(default /etc/styra-controller/config.yaml)")
	flag.BoolVar(&printVersion, "version", false, "show version information")
	flag.Parse()

	if len(configFiles) == 0 {
		configFiles = config.StringSlice{"/etc/styra-controller/config.yaml"}
	}

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

	ctrlConfig, err := config.Load(configFiles, scheme)
	if err != nil {
		exit(errors.Wrap(err, "unable to load the config file(s)"))
	}

	ctrl.SetLogger(zap.New(
		zap.UseDevMode(ctrlConfig.LogLevel >= logLevelDebug),
		zap.Level(zapcore.Level(-ctrlConfig.LogLevel)),
	))

	options := config.OptionsFromConfig(ctrlConfig, scheme)

	mgr, err := ctrl.NewManager(restCfg, options)
	if err != nil {
		exit(errors.Wrap(err, "unable to start manager"))
	}

	var opaControlPlaneClient ocp.ClientInterface
	if ctrlConfig.OPAControlPlaneConfig == nil ||
		ctrlConfig.OPAControlPlaneConfig.Address == "" ||
		ctrlConfig.OPAControlPlaneConfig.Token == "" {
		err := errors.New(
			"missing OPA Control Plane configuration: address and token are required",
		)
		exit(errors.Wrap(err, "unable to start manager"))
	}

	if ctrlConfig.OPAControlPlaneConfig.BundleObjectStorage == nil {
		err := errors.New("missing OPA Control Plane bundle object storage configuration")
		exit(errors.Wrap(err, "unable to start manager"))
	}

	ocpHostURL := strings.TrimSuffix(ctrlConfig.OPAControlPlaneConfig.Address, "/")
	opaControlPlaneClient = ocp.New(ocpHostURL, ctrlConfig.OPAControlPlaneConfig.Token)

	// System Controller
	systemReadyMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controller_system_status_ready",
			Help: "Show if a system is in status ready",
		},
		[]string{"system_name", "namespace", "system_id", "control_plane"},
	)

	if err := metrics.Registry.Register(systemReadyMetric); err != nil {
		err := errors.Wrap(err, "could not register controller_system_status_ready metric")
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
		exit(err)
	}

	systemMetrics := &controllers.SystemReconcilerMetrics{
		ControllerSystemStatusReady: systemReadyMetric,
		ReconcileSegmentTime:        reconcileSegmentTimeMetric,
		ReconcileTime:               reconcileTimeMetric,
	}

	r1 := &controllers.SystemReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorder("system-controller"),
		Metrics:   systemMetrics,
		Config:    ctrlConfig,
		APIReader: mgr.GetAPIReader(),
	}

	r1.OCP = opaControlPlaneClient

	r1.WebhookClient = webhook.New(
		ctrlConfig.OPAControlPlaneConfig.SystemDatasourceChanged,
		ctrlConfig.OPAControlPlaneConfig.LibraryDatasourceChanged)

	if err = r1.SetupWithManager(mgr, "styra-controller"); err != nil {
		exit(errors.Wrap(err, "unable to create System controller"))
	}

	if err = r1.CreateDefaultRequirements(context.Background(), log); err != nil {
		exit(errors.Wrap(err, "unable to create default requirements"))
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = webhookstyrav1beta1.SetupSystemWebhookWithManager(mgr); err != nil {
			exit(errors.Wrap(err, "unable to create System webhook"))
		}
	}

	libraryReconciler := &controllers.LibraryReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: ctrlConfig,
	}

	libraryReconciler.OCP = opaControlPlaneClient

	libraryReconciler.WebhookClient = webhook.New(
		ctrlConfig.OPAControlPlaneConfig.SystemDatasourceChanged,
		ctrlConfig.OPAControlPlaneConfig.LibraryDatasourceChanged)

	if err = libraryReconciler.SetupWithManager(mgr); err != nil {
		exit(errors.Wrap(err, "unable to create Library controller"))
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = webhookstyrav1alpha1.SetupLibraryWebhookWithManager(mgr); err != nil {
			exit(errors.Wrap(err, "unable to create Library webhook"))
		}
	}
	//+kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		exit(errors.Wrap(err, "unable to set up health check"))
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		exit(errors.Wrap(err, "unable to set up ready check"))
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		exit(errors.Wrap(err, "problem running manager"))
	}
}

func exit(err error) {
	fmt.Printf("Exiting due to error: %v\n", err)
	time.Sleep(10 * time.Second)
	os.Exit(1)
}
