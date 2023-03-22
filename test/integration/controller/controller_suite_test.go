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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	configv1 "github.com/bankdata/styra-controller/api/config/v1"
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
	RegisterFailHandler(Fail)

	RunSpecs(t, "test/integration/controller")
}

func resetMock(m *mock.Mock) {
	m.Calls = nil
	m.ExpectedCalls = nil
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	if !Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
		return
	}

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = styrav1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = styrav1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	styraClientMock = &styraclientmock.ClientInterface{}
	webhookMock = &webhookmocks.Client{}

	systemReconciler := styractrls.SystemReconciler{
		Client:        k8sClient,
		Scheme:        k8sManager.GetScheme(),
		Styra:         styraClientMock,
		WebhookClient: webhookMock,
		Recorder:      k8sManager.GetEventRecorderFor("system-controller"),
		Config: &configv1.ProjectConfig{
			StyraSystemUserRoles: []string{string(styra.RoleSystemViewer)},
			IdentityProvider:     "AzureAD Bankdata",
			JwtGroupClaim:        "groups",
		},
	}

	err = systemReconciler.SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	globalDatasourceReconciler := &styractrls.GlobalDatasourceReconciler{
		Config: &configv1.ProjectConfig{GitUser: "test-user", GitPassword: "test-secret"},
		Client: k8sClient,
		Styra:  styraClientMock,
	}

	err = globalDatasourceReconciler.SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	managerCtx, managerCancel = context.WithCancel(context.Background())
	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(managerCtx)
		Expect(err).NotTo(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	if testing.Short() {
		return
	}
	By("tearing down the test environment")
	if managerCancel != nil {
		managerCancel()
	}
	if testEnv != nil {
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	}
})
