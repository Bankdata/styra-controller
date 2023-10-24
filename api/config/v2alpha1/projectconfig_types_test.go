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

package v2alpha1_test

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	//nolint:staticcheck // issue https://github.com/Bankdata/styra-controller/issues/82
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
)

var _ = ginkgo.Describe("ProjectConfig", func() {
	ginkgo.DescribeTable("GetGitCredentialForRepo",
		func(gitCredentials []*v2alpha1.GitCredential, repo string, expected *v2alpha1.GitCredential) {
			c := &v2alpha1.ProjectConfig{
				GitCredentials: gitCredentials,
			}
			gomega.Ω(c.GetGitCredentialForRepo(repo)).To(gomega.Equal(expected))
		},

		ginkgo.Entry("returns nil if the list of credentials is empty", nil, "test", nil),

		ginkgo.Entry("finds matching credential",
			[]*v2alpha1.GitCredential{
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata",
				},
			},
			"https://github.com/bankdata/styra-controller.git",
			&v2alpha1.GitCredential{
				User:       "",
				Password:   "",
				RepoPrefix: "https://github.com/bankdata",
			},
		),
		ginkgo.Entry("returns longest matching credential",
			[]*v2alpha1.GitCredential{
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata",
				},
				{
					User:       "",
					Password:   "",
					RepoPrefix: "https://github.com/bankdata/styra-controller",
				},
			},
			"https://github.com/bankdata/styra-controller.git",
			&v2alpha1.GitCredential{
				User:       "",
				Password:   "",
				RepoPrefix: "https://github.com/bankdata/styra-controller",
			},
		),
	)

	ginkgo.Describe("unmarshalling", func() {
		ginkgo.It("correctly unmarshals all fields", func() {
			validConfig := []byte(`
apiVersion: config.bankdata.dk/v2alpha1
kind: ProjectConfig
controllerClass: "class"
deletionProtectionDefault: true
disableCRDWebhooks: true
enableMigrations: true
gitCredentials:
  - user: my-git-user
    password: my-git-password
    repoPrefix: https://github.com/my-org
logLevel: 42
notificationWebhook:
  address: "https://webhook.com"
sentry:
  debug: true
  dsn: "https://sentry.com"
  environment: "test"
  httpsProxy: "https://proxy.com"
sso:
  identityProvider: "my-provider"
  jwtGroupsClaim: "groups"
styra:
  address: "https://styra.com"
  token: "token"
systemPrefix: "prefix"
systemSuffix: "suffix"
systemUserRoles: 
  - SystemViewer
`)
			scheme := runtime.NewScheme()
			gomega.Ω(v2alpha1.AddToScheme(scheme)).Should(gomega.Succeed())
			decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
			var c v2alpha1.ProjectConfig
			_, _, err := decoder.Decode(validConfig, nil, &c)
			gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
			gomega.Ω(c.ControllerClass).Should(gomega.Equal("class"))
			gomega.Ω(c.DeletionProtectionDefault).Should(gomega.BeTrue())
			gomega.Ω(c.DisableCRDWebhooks).Should(gomega.BeTrue())
			gomega.Ω(c.EnableMigrations).Should(gomega.BeTrue())
			gomega.Ω(len(c.GitCredentials)).Should(gomega.Equal(1))
			gomega.Ω(c.GitCredentials[0].User).Should(gomega.Equal("my-git-user"))
			gomega.Ω(c.GitCredentials[0].Password).Should(gomega.Equal("my-git-password"))
			gomega.Ω(c.GitCredentials[0].RepoPrefix).Should(gomega.Equal("https://github.com/my-org"))
			gomega.Ω(c.LogLevel).Should(gomega.Equal(42))
			gomega.Ω(c.NotificationWebhook).ShouldNot(gomega.BeNil())
			gomega.Ω(c.NotificationWebhook.Address).Should(gomega.Equal("https://webhook.com"))
			gomega.Ω(c.Sentry).ShouldNot(gomega.BeNil())
			gomega.Ω(c.Sentry.Debug).Should(gomega.BeTrue())
			gomega.Ω(c.Sentry.DSN).Should(gomega.Equal("https://sentry.com"))
			gomega.Ω(c.Sentry.Environment).Should(gomega.Equal("test"))
			gomega.Ω(c.Sentry.HTTPSProxy).Should(gomega.Equal("https://proxy.com"))
			gomega.Ω(c.SSO).ShouldNot(gomega.BeNil())
			gomega.Ω(c.SSO.IdentityProvider).Should(gomega.Equal("my-provider"))
			gomega.Ω(c.SSO.JWTGroupsClaim).Should(gomega.Equal("groups"))
			gomega.Ω(c.Styra.Address).Should(gomega.Equal("https://styra.com"))
			gomega.Ω(c.Styra.Token).Should(gomega.Equal("token"))
			gomega.Ω(c.SystemPrefix).Should(gomega.Equal("prefix"))
			gomega.Ω(c.SystemSuffix).Should(gomega.Equal("suffix"))
			gomega.Ω(len(c.SystemUserRoles)).Should(gomega.Equal(1))
			gomega.Ω(c.SystemUserRoles[0]).Should(gomega.Equal("SystemViewer"))
		})
	})
})
