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
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	styrav1alpha1 "github.com/bankdata/styra-controller/api/styra/v1alpha1"
	styrav1beta1 "github.com/bankdata/styra-controller/api/styra/v1beta1"
	styractrls "github.com/bankdata/styra-controller/internal/controller/styra"
	webhookmocks "github.com/bankdata/styra-controller/internal/webhook/mocks"
	ocpclientmock "github.com/bankdata/styra-controller/pkg/ocp/mocks"
	"github.com/bankdata/styra-controller/pkg/ptr"
	s3clientmock "github.com/bankdata/styra-controller/pkg/s3/mocks"
	"github.com/bankdata/styra-controller/pkg/styra"
	styraclientmock "github.com/bankdata/styra-controller/pkg/styra/mocks"
	//+kubebuilder:scaffold:imports
)

var (
	k8sClient               client.Client
	testEnv                 *envtest.Environment
	managerCtx              context.Context
	managerCancel           context.CancelFunc
	managerCtxPodRestart    context.Context
	managerCancelPodRestart context.CancelFunc
	styraClientMock         *styraclientmock.ClientInterface
	ocpClientMock           *ocpclientmock.ClientInterface
	s3ClientMock            *s3clientmock.Client
	webhookMock             *webhookmocks.Client
)

const (
	timeout  = 30 * time.Second
	interval = time.Second
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecs(t, "test/integration/controller")
}

func resetMock(m *mock.Mock) {
	m.Calls = nil
	m.ExpectedCalls = nil
}

var _ = ginkgo.BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(ginkgo.GinkgoWriter), zap.UseDevMode(true)))

	if !ginkgo.Label("integration").MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		return
	}

	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(cfg).NotTo(gomega.BeNil())

	err = styrav1alpha1.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	err = styrav1beta1.AddToScheme(scheme.Scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(k8sClient).NotTo(gomega.BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme.Scheme,
		LeaderElection: false,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	styraClientMock = &styraclientmock.ClientInterface{}
	ocpClientMock = &ocpclientmock.ClientInterface{}
	webhookMock = &webhookmocks.Client{}
	s3ClientMock = &s3clientmock.Client{}
	systemReconciler := styractrls.SystemReconciler{
		Client:        k8sClient,
		APIReader:     k8sManager.GetAPIReader(),
		Scheme:        k8sManager.GetScheme(),
		Styra:         styraClientMock,
		OCP:           ocpClientMock,
		S3:            s3ClientMock,
		WebhookClient: webhookMock,
		Recorder:      k8sManager.GetEventRecorderFor("system-controller"),
		Config: &configv2alpha2.ProjectConfig{
			SystemUserRoles: []string{string(styra.RoleSystemViewer)},
			SSO: &configv2alpha2.SSOConfig{
				IdentityProvider: "AzureAD Bankdata",
				JWTGroupsClaim:   "groups",
			},
			DatasourceIgnorePatterns:            []string{"^.*/ignore$"},
			ReadOnly:                            true,
			EnableDeltaBundlesDefault:           ptr.Bool(false),
			EnableStyraReconciliation:           true,
			EnableOPAControlPlaneReconciliation: true,
			DefaultRequirements:                 []string{"library1"},
			ObjectStorage: &configv2alpha2.ObjectStorage{
				AWS: &configv2alpha2.AWSObjectStorage{
					Bucket:              "test-bucket",
					Region:              "eu-west-1",
					URL:                 "s3-url",
					OCPConfigSecretName: "s3-credentials",
					AdminCredentials: &configv2alpha2.AWSCredentials{
						AccessKeyID:     "access-key-id",
						SecretAccessKey: "secret-access-key",
					},
				},
			},
			OPAControlPlaneConfig: &configv2alpha2.OPAControlPlaneConfig{
				Address: "ocp-url",
				Token:   "ocp-token",
				GitCredentials: []*configv2alpha2.GitCredentials{&configv2alpha2.GitCredentials{
					ID:         "github-credentials",
					RepoPrefix: "https://github",
				}},
			},
		},

		Metrics: &styractrls.SystemReconcilerMetrics{
			ControllerSystemStatusReady: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "controller_system_status_ready",
					Help: "Show if a system is in status ready",
				},
				[]string{"system_name", "namespace", "system_id"},
			),
			ReconcileSegmentTime: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "controller_system_reconcile_segment_seconds",
					Help:    "Time taken to perform one segment of reconciling a system",
					Buckets: prometheus.DefBuckets,
				}, []string{"segment"},
			),
			ReconcileTime: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "controller_system_reconcile_seconds",
					Help:    "Time taken to reconcile a system",
					Buckets: prometheus.DefBuckets,
				}, []string{"result"},
			),
		},
	}

	err = systemReconciler.SetupWithManager(k8sManager, "styra-controller")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	libraryReconciler := &styractrls.LibraryReconciler{
		Config: &configv2alpha2.ProjectConfig{
			SSO: &configv2alpha2.SSOConfig{
				IdentityProvider: "AzureAD Bankdata",
				JWTGroupsClaim:   "groups",
			},
			DatasourceIgnorePatterns: []string{"^.*/ignore$"},
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-secret"},
			},
			EnableStyraReconciliation: true,
		},
		Client:        k8sClient,
		Styra:         styraClientMock,
		OCP:           ocpClientMock,
		WebhookClient: webhookMock,
	}

	err = libraryReconciler.SetupWithManager(k8sManager)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	managerCtx, managerCancel = context.WithCancel(context.Background())

	go func() {
		defer ginkgo.GinkgoRecover()
		err = k8sManager.Start(managerCtx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}()

	// Test setup for systemReconcilerPodRestart that deploys a system with an ID and a Statefulset for a SLP
	k8sManagerPodRestart, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme.Scheme,
		LeaderElection: false,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// Controller with PodRestart enabled for SLPs.
	systemReconcilerPodRestart := styractrls.SystemReconciler{
		Client:        k8sClient,
		APIReader:     k8sManagerPodRestart.GetAPIReader(),
		Scheme:        k8sManagerPodRestart.GetScheme(),
		Styra:         styraClientMock,
		OCP:           ocpClientMock,
		WebhookClient: webhookMock,
		Recorder:      k8sManagerPodRestart.GetEventRecorderFor("system-controller"),
		Config: &configv2alpha2.ProjectConfig{
			ControllerClass: "styra-controller-pod-restart",
			SystemUserRoles: []string{string(styra.RoleSystemViewer)},
			SSO: &configv2alpha2.SSOConfig{
				IdentityProvider: "AzureAD Bankdata",
				JWTGroupsClaim:   "groups",
			},
			DatasourceIgnorePatterns:  []string{"^.*/ignore$"},
			ReadOnly:                  true,
			EnableDeltaBundlesDefault: ptr.Bool(false),
			PodRestart: &configv2alpha2.PodRestartConfig{
				SLPRestart: &configv2alpha2.SLPRestartConfig{
					Enabled:        true,
					DeploymentType: "statefulset",
				},
			},
			EnableStyraReconciliation: true,
		},
		Metrics: &styractrls.SystemReconcilerMetrics{
			ControllerSystemStatusReady: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "controller_system_status_ready",
					Help: "Show if a system is in status ready",
				},
				[]string{"system_name", "namespace", "system_id"},
			),
			ReconcileSegmentTime: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "controller_system_reconcile_segment_seconds",
					Help:    "Time taken to perform one segment of reconciling a system",
					Buckets: prometheus.DefBuckets,
				}, []string{"segment"},
			),
			ReconcileTime: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "controller_system_reconcile_seconds",
					Help:    "Time taken to reconcile a system",
					Buckets: prometheus.DefBuckets,
				}, []string{"result"},
			),
		},
	}

	err = systemReconcilerPodRestart.SetupWithManager(k8sManagerPodRestart, "styra-controller-pod-restart")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	managerCtxPodRestart, managerCancelPodRestart = context.WithCancel(context.Background())
	go func() {
		// Test setup function to systemReconcilerPodRestart that deploys a system with an ID and a Statefulset for a SLP
		// and restarts the SLP pods.
		defer ginkgo.GinkgoRecover()

		systemName := "test-pod-restart"
		systemNamespace := "default"

		systemToCreate := &styrav1beta1.System{
			ObjectMeta: metav1.ObjectMeta{
				Name:      systemName,
				Namespace: systemNamespace,
				Labels: map[string]string{
					"styra-controller/class": "styra-controller-pod-restart",
				},
			},
			Spec: styrav1beta1.SystemSpec{
				LocalPlane: &styrav1beta1.LocalPlane{
					Name: fmt.Sprintf("%v-slp", systemName),
				},
			},
		}

		gomega.Expect(k8sClient.Create(managerCtxPodRestart, systemToCreate)).To(gomega.Succeed())
		patch := client.MergeFrom(systemToCreate.DeepCopy())
		systemToCreate.Status.ID = "system_id"
		systemToCreate.Status.Ready = true
		gomega.Expect(k8sClient.Status().Patch(managerCtxPodRestart, systemToCreate, patch)).To(gomega.Succeed())

		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%v-slp", systemName),
				Namespace: systemNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": fmt.Sprintf("%v-slp", systemName)},
				},
				ServiceName: fmt.Sprintf("%v-slp", systemName),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": fmt.Sprintf("%v-slp", systemName)},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:    "busybox",
							Image:   "busybox",
							Command: []string{"sleep", "3600"},
						}},
					},
				},
			},
		}

		gomega.Expect(k8sClient.Create(managerCtxPodRestart, sts)).To(gomega.Succeed())

		err = k8sManagerPodRestart.Start(managerCtxPodRestart)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}()
})

var _ = ginkgo.AfterSuite(func() {
	if testing.Short() {
		return
	}
	ginkgo.By("tearing down the test environment")
	if managerCancel != nil {
		managerCancel()
	}
	if managerCancelPodRestart != nil {
		managerCancelPodRestart()
	}
	if testEnv != nil {
		err := testEnv.Stop()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
})
