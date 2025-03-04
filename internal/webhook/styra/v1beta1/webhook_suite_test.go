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

package v1beta1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	//+kubebuilder:scaffold:imports
	"github.com/bankdata/styra-controller/api/styra/v1alpha1"
	"github.com/bankdata/styra-controller/api/styra/v1beta1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecs(t, "test/integration/webhook")
}

var _ = ginkgo.BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(ginkgo.GinkgoWriter), zap.UseDevMode(true)))

	if !ginkgo.Label("integration").MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		return
	}

	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = v1beta1.AddToScheme(scheme)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	//+kubebuilder:scaffold:scheme

	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "..", "config", "webhook")},
		},
		CRDInstallOptions: envtest.CRDInstallOptions{
			Scheme: scheme,
		},
	}

	cfg, err := testEnv.Start()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(cfg).NotTo(gomega.BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(k8sClient).NotTo(gomega.BeNil())

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookInstallOptions.LocalServingHost,
			Port:    webhookInstallOptions.LocalServingPort,
			CertDir: webhookInstallOptions.LocalServingCertDir,
		}),
		LeaderElection: false,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	err = SetupSystemWebhookWithManager(mgr)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	//+kubebuilder:scaffold:webhook

	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		defer ginkgo.GinkgoRecover()
		err = mgr.Start(ctx)
		if err != nil {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	gomega.Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		err = conn.Close()
		if err != nil {
			return err
		}
		return nil
	}).Should(gomega.Succeed())

})

var _ = ginkgo.AfterSuite(func() {
	if testing.Short() {
		return
	}

	if cancel != nil {
		cancel()
	}

	ginkgo.By("tearing down the test environment")
	if testEnv != nil {
		err := testEnv.Stop()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
})
