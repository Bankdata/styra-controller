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

package config

import (
	"os"
	"path/filepath"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/bankdata/styra-controller/api/config/v2alpha2"
)

var _ = ginkgo.DescribeTable("deserialize",
	func(data []byte, expected *v2alpha2.ProjectConfig, shouldErr bool) {
		scheme := runtime.NewScheme()
		err := v2alpha2.AddToScheme(scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		actual, err := deserialize(data, scheme)
		if shouldErr {
			gomega.Ω(err).Should(gomega.HaveOccurred())
		} else {
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		}
		gomega.Ω(actual).Should(gomega.Equal(expected))
	},

	ginkgo.Entry("errors on unexpected api group",
		[]byte(`
apiVersion: myconfig.bankdata.dk/v1
kind: ProjectConfig
styra:
  token: my-token
`),
		nil,
		true,
	),

	ginkgo.Entry("can deserialize v2alpha2",
		[]byte(`
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: my-token
`),
		&v2alpha2.ProjectConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ProjectConfig",
				APIVersion: v2alpha2.GroupVersion.Identifier(),
			},
			OPAControlPlaneConfig: &v2alpha2.OPAControlPlaneConfig{
				Token: "my-token",
			},
		},
		false,
	),
)

var _ = ginkgo.Describe("Load", func() {
	var (
		scheme *runtime.Scheme
		tmpDir string
	)

	ginkgo.BeforeEach(func() {
		scheme = runtime.NewScheme()
		err := v2alpha2.AddToScheme(scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		tmpDir = ginkgo.GinkgoT().TempDir()
	})

	writeFile := func(name, content string) string {
		p := filepath.Join(tmpDir, name)
		err := os.WriteFile(p, []byte(content), 0o600)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		return p
	}

	ginkgo.It("loads a single config file", func() {
		f := writeFile("config.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: my-token
logLevel: 2
`)
		cfg, err := Load([]string{f}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("my-token"))
		gomega.Ω(cfg.LogLevel).Should(gomega.Equal(2))
	})

	ginkgo.It("merges multiple config files with later files overriding", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  address: https://styra.example.com
  token: base-token
logLevel: 1
systemPrefix: my-prefix
`)
		secrets := writeFile("secrets.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: secret-token
`)
		cfg, err := Load([]string{base, secrets}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		// Token should be overridden by secrets file
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("secret-token"))
		// Address should be preserved from base
		gomega.Ω(cfg.OPAControlPlaneConfig.Address).Should(gomega.Equal("https://styra.example.com"))
		// Other fields from base should be preserved
		gomega.Ω(cfg.LogLevel).Should(gomega.Equal(1))
		gomega.Ω(cfg.SystemPrefix).Should(gomega.Equal("my-prefix"))
	})

	ginkgo.It("merges three files in order", func() {
		f1 := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
logLevel: 0
systemPrefix: base
systemSuffix: base-suffix
`)
		f2 := writeFile("overlay.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
logLevel: 1
systemPrefix: overlay
`)
		f3 := writeFile("secrets.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: secret
`)
		cfg, err := Load([]string{f1, f2, f3}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.LogLevel).Should(gomega.Equal(1))
		gomega.Ω(cfg.SystemPrefix).Should(gomega.Equal("overlay"))
		gomega.Ω(cfg.SystemSuffix).Should(gomega.Equal("base-suffix"))
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("secret"))
	})

	ginkgo.It("returns an error when no files are specified", func() {
		_, err := Load([]string{}, scheme)
		gomega.Ω(err).Should(gomega.HaveOccurred())
	})

	ginkgo.It("returns an error when a file does not exist", func() {
		_, err := Load([]string{"/nonexistent/path.yaml"}, scheme)
		gomega.Ω(err).Should(gomega.HaveOccurred())
	})

	ginkgo.It("returns an error when second file does not exist", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
logLevel: 1
`)
		_, err := Load([]string{base, "/nonexistent/secrets.yaml"}, scheme)
		gomega.Ω(err).Should(gomega.HaveOccurred())
		gomega.Ω(err.Error()).Should(gomega.ContainSubstring("/nonexistent/secrets.yaml"))
	})

	ginkgo.It("returns an error when overlay file has invalid YAML", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
logLevel: 1
`)
		bad := writeFile("bad.yaml", `
not: valid: yaml: {{
`)
		_, err := Load([]string{base, bad}, scheme)
		gomega.Ω(err).Should(gomega.HaveOccurred())
	})

	ginkgo.It("preserves nested OPA config when overlay only sets secrets", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opa:
  bundleServer:
    url: https://minio-host
    path: /ocp
  metrics:
    prometheus:
      http:
        buckets:
        - 0.01
        - 0.1
        - 1
logLevel: 0
`)
		secrets := writeFile("secrets.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: secret-token
`)
		cfg, err := Load([]string{base, secrets}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.OPA.BundleServer).ShouldNot(gomega.BeNil())
		gomega.Ω(cfg.OPA.BundleServer.URL).Should(gomega.Equal("https://minio-host"))
		gomega.Ω(cfg.OPA.BundleServer.Path).Should(gomega.Equal("/ocp"))
		gomega.Ω(cfg.OPA.Metrics.Prometheus.HTTP.Buckets).Should(gomega.Equal([]float64{0.01, 0.1, 1}))
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("secret-token"))
	})

	ginkgo.It("overlay can override boolean fields", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
deletionProtectionDefault: false
disableCRDWebhooks: false
`)
		overlay := writeFile("overlay.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
deletionProtectionDefault: true
`)
		cfg, err := Load([]string{base, overlay}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.DeletionProtectionDefault).Should(gomega.BeTrue())
		// disableCRDWebhooks not in overlay, should remain false
		gomega.Ω(cfg.DisableCRDWebhooks).Should(gomega.BeFalse())
	})

	ginkgo.It("overlay can set OPA control plane config with nested S3", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  address: https://ocp-host/ocp
  bundleObjectStorage:
    s3:
      bucket: ocp
      region: us-east-1
      url: https://minio-host
      ocpConfigSecretName: minio
`)
		secrets := writeFile("secrets.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: ocp-secret-token
`)
		cfg, err := Load([]string{base, secrets}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.OPAControlPlaneConfig).ShouldNot(gomega.BeNil())
		gomega.Ω(cfg.OPAControlPlaneConfig.Address).Should(gomega.Equal("https://ocp-host/ocp"))
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("ocp-secret-token"))
		gomega.Ω(cfg.OPAControlPlaneConfig.BundleObjectStorage).ShouldNot(gomega.BeNil())
		gomega.Ω(cfg.OPAControlPlaneConfig.BundleObjectStorage.S3.Bucket).Should(gomega.Equal("ocp"))
	})

	ginkgo.It("overlay replaces default requirements entirely", func() {
		base := writeFile("base.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  defaultRequirements:
  - library-a
  - library-b
  - library-c
`)
		overlay := writeFile("overlay.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  defaultRequirements:
  - library-a
`)
		cfg, err := Load([]string{base, overlay}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
		gomega.Ω(cfg.OPAControlPlaneConfig.DefaultRequirements).Should(gomega.Equal([]string{"library-a"}))
	})

	ginkgo.It("simulates realistic ConfigMap + Secret split", func() {
		configMap := writeFile("config.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
controllerClass: production
deletionProtectionDefault: true
disableCRDWebhooks: false
logLevel: 0
systemPrefix: prod
systemSuffix: cluster-01
opaControlPlaneConfig:
  address: https://ocp.example.com
  bundleObjectStorage:
    s3:
      bucket: bundles
      region: eu-west-1
      url: https://s3.example.com
      ocpConfigSecretName: s3-config
  defaultRequirements:
    - base-library
opa:
  bundleServer:
    url: https://s3.example.com
    path: /bundles
`)
		secret := writeFile("config-secrets.yaml", `
apiVersion: config.bankdata.dk/v2alpha2
kind: ProjectConfig
opaControlPlaneConfig:
  token: real-ocp-token
`)
		cfg, err := Load([]string{configMap, secret}, scheme)
		gomega.Ω(err).ShouldNot(gomega.HaveOccurred())

		// Non-secret fields from ConfigMap
		gomega.Ω(cfg.ControllerClass).Should(gomega.Equal("production"))
		gomega.Ω(cfg.DeletionProtectionDefault).Should(gomega.BeTrue())
		gomega.Ω(cfg.LogLevel).Should(gomega.Equal(0))
		gomega.Ω(cfg.SystemPrefix).Should(gomega.Equal("prod"))
		gomega.Ω(cfg.SystemSuffix).Should(gomega.Equal("cluster-01"))
		gomega.Ω(cfg.OPAControlPlaneConfig.Address).Should(gomega.Equal("https://ocp.example.com"))
		gomega.Ω(cfg.OPAControlPlaneConfig.BundleObjectStorage.S3.Bucket).Should(gomega.Equal("bundles"))
		gomega.Ω(cfg.OPAControlPlaneConfig.DefaultRequirements).Should(gomega.Equal([]string{"base-library"}))
		gomega.Ω(cfg.OPA.BundleServer.URL).Should(gomega.Equal("https://s3.example.com"))

		// Secret fields from Secret overlay
		gomega.Ω(cfg.OPAControlPlaneConfig.Token).Should(gomega.Equal("real-ocp-token"))
	})
})
