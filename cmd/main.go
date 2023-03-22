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
	"flag"
	"os"
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

	configv1 "github.com/bankdata/styra-controller/api/config/v1"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	controllers "github.com/bankdata/styra-controller/internal/controller/styra"
	"github.com/bankdata/styra-controller/internal/webhook"
	"github.com/bankdata/styra-controller/pkg/styra"
	//+kubebuilder:scaffold:imports
)

const (
	logLevelDebug = -1
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(styrav1alpha1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(styrav1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

	flag.Parse()

	ctrlConfig := configv1.ProjectConfig{}
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		var err error
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			exit(err)
		}
	}
	opts := zap.Options{}

	if ctrlConfig.LogLevel >= logLevelDebug {
		opts.Development = true
	}

	opts.Level = zapcore.Level(ctrlConfig.LogLevel)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if ctrlConfig.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         ctrlConfig.SentryDSN,
			Environment: ctrlConfig.Environment,
			Release:     os.Getenv("STYRACTRL_VERSION"),
			Debug:       ctrlConfig.SentryDebug,
			HTTPSProxy:  ctrlConfig.SentryHTTPSProxy,
		})
		if err != nil {
			setupLog.Error(err, "failed to init sentry")
			exit(err)
		}
		defer sentry.Flush(2 * time.Second)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		exit(err)
	}

	roles := make([]styra.Role, len(ctrlConfig.StyraSystemUserRoles))
	for i, role := range ctrlConfig.StyraSystemUserRoles {
		roles[i] = styra.Role(role)
	}

	styraClient := styra.New(ctrlConfig.StyraAddress, ctrlConfig.StyraToken)

	// System Controller
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "controller_system_status_ready",
			Help: "Show if a system is in status ready",
		},
		[]string{"system", "namespace"},
	)

	if err := metrics.Registry.Register(metric); err != nil {
		err := errors.Wrap(err, "could not register controller_system_status_ready metric")
		setupLog.Error(err, err.Error())
		exit(err)
	}

	r1 := &controllers.SystemReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Styra:    styra.New(ctrlConfig.StyraAddress, ctrlConfig.StyraToken),
		Recorder: mgr.GetEventRecorderFor("system-controller"),
		Metric:   metric,
		Config:   &ctrlConfig,
	}

	if ctrlConfig.DatasourceWebhookAddress != "" {
		r1.WebhookClient = webhook.New(ctrlConfig.DatasourceWebhookAddress)
	}

	if err = r1.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "System")
		exit(err)
	}

	if !ctrlConfig.WebhooksDisabled {
		if err = (&styrav1beta1.System{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "System")
			os.Exit(1)
		}
	}

	if err = (&controllers.GlobalDatasourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: &ctrlConfig,
		Styra:  styraClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GlobalDatasource")
		os.Exit(1)
	}

	if !ctrlConfig.WebhooksDisabled {
		if err = (&styrav1alpha1.GlobalDatasource{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GlobalDatasource")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		exit(err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		exit(err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		exit(err)
	}
}

func exit(err error) {
	sentry.CaptureException(err)
	sentry.Flush(2 * time.Second)
	os.Exit(1)
}
