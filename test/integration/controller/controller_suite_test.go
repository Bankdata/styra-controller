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
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/stretchr/testify/mock"
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
	"github.com/bankdata/styra-controller/pkg/styra"
	styraclientmock "github.com/bankdata/styra-controller/pkg/styra/mocks"
	//+kubebuilder:scaffold:imports
)

var (
	k8sClient       client.Client
	testEnv         *envtest.Environment
	managerCtx      context.Context
	managerCancel   context.CancelFunc
	styraClientMock *styraclientmock.ClientInterface
	webhookMock     *webhookmocks.Client
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
	webhookMock = &webhookmocks.Client{}

	systemReconciler := styractrls.SystemReconciler{
		Client:        k8sClient,
		Scheme:        k8sManager.GetScheme(),
		Styra:         styraClientMock,
		WebhookClient: webhookMock,
		Recorder:      k8sManager.GetEventRecorderFor("system-controller"),
		Config: &configv2alpha2.ProjectConfig{
			SystemUserRoles: []string{string(styra.RoleSystemViewer)},
			SSO: &configv2alpha2.SSOConfig{
				IdentityProvider: "AzureAD Bankdata",
				JWTGroupsClaim:   "groups",
			},
		},
	}

	err = systemReconciler.SetupWithManager(k8sManager)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	globalDatasourceReconciler := &styractrls.GlobalDatasourceReconciler{
		Config: &configv2alpha2.ProjectConfig{
			GitCredentials: []*configv2alpha2.GitCredential{
				{User: "test-user", Password: "test-secret"},
			},
		},
		Client: k8sClient,
		Styra:  styraClientMock,
	}

	err = globalDatasourceReconciler.SetupWithManager(k8sManager)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	managerCtx, managerCancel = context.WithCancel(context.Background())
	go func() {
		defer ginkgo.GinkgoRecover()
		err = k8sManager.Start(managerCtx)
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
	if testEnv != nil {
		err := testEnv.Stop()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
})
