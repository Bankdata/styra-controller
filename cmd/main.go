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
	"fmt"
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
	configv2alpha1 "github.com/bankdata/styra-controller/api/config/v2alpha1"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/bankdata/styra-controller/internal/config"
	controllers "github.com/bankdata/styra-controller/internal/controller/styra"
	"github.com/bankdata/styra-controller/internal/webhook"
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
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(styrav1beta1.AddToScheme(scheme))
	utilruntime.Must(configv2alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		configFile   string
		printVersion bool
	)

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

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

	ctrlConfig := &configv2alpha1.ProjectConfig{}
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		var err error
		ctrlConfig, err = config.Load(configFile, scheme)
		if err != nil {
			log.Error(err, "unable to load the config file")
			exit(err)
		}

		//nolint:staticcheck // issue https://github.com/Bankdata/styra-controller/issues/82
		options, err = options.AndFrom(ctrlConfig)
		if err != nil {
			log.Error(err, "could not load options from config")
			exit(err)
		}
	}

	ctrl.SetLogger(zap.New(
		zap.UseDevMode(ctrlConfig.LogLevel >= logLevelDebug),
		zap.Level(zapcore.Level(-ctrlConfig.LogLevel)),
	))

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

	restCfg := ctrl.GetConfigOrDie()

	logDebug.Info(
		"rest config details",
		"timeout", restCfg.Timeout,
		"host", restCfg.Host,
		"apiPath", restCfg.APIPath,
	)

	mgr, err := ctrl.NewManager(restCfg, options)
	if err != nil {
		log.Error(err, "unable to start manager")
		exit(err)
	}

	roles := make([]styra.Role, len(ctrlConfig.SystemUserRoles))
	for i, role := range ctrlConfig.SystemUserRoles {
		roles[i] = styra.Role(role)
	}

	styraClient := styra.New(ctrlConfig.Styra.Address, ctrlConfig.Styra.Token)

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
		log.Error(err, err.Error())
		exit(err)
	}

	r1 := &controllers.SystemReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Styra:    styra.New(ctrlConfig.Styra.Address, ctrlConfig.Styra.Token),
		Recorder: mgr.GetEventRecorderFor("system-controller"),
		Metric:   metric,
		Config:   ctrlConfig,
	}

	if ctrlConfig.NotificationWebhook != nil {
		r1.WebhookClient = webhook.New(ctrlConfig.NotificationWebhook.Address)
	}

	if err = r1.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "System")
		exit(err)
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = (&styrav1beta1.System{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "System")
			os.Exit(1)
		}
	}

	if err = (&controllers.GlobalDatasourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: ctrlConfig,
		Styra:  styraClient,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "GlobalDatasource")
		os.Exit(1)
	}

	if !ctrlConfig.DisableCRDWebhooks {
		if err = (&styrav1alpha1.GlobalDatasource{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "GlobalDatasource")
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
