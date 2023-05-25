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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	//nolint:staticcheck // issue https://github.com/Bankdata/styra-controller/issues/82
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
)

var _ = Describe("ProjectConfig", func() {
	DescribeTable("GetGitCredentialForRepo",
		func(gitCredentials []*v2alpha1.GitCredential, repo string, expected *v2alpha1.GitCredential) {
			c := &v2alpha1.ProjectConfig{
				GitCredentials: gitCredentials,
			}
			Ω(c.GetGitCredentialForRepo(repo)).To(Equal(expected))
		},

		Entry("returns nil if the list of credentials is empty", nil, "test", nil),

		Entry("finds matching credential",
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
		Entry("returns longest matching credential",
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

	Describe("unmarshalling", func() {
		It("correctly unmarshals all fields", func() {
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
			Ω(v2alpha1.AddToScheme(scheme)).Should(Succeed())
			decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
			var c v2alpha1.ProjectConfig
			_, _, err := decoder.Decode(validConfig, nil, &c)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c.ControllerClass).Should(Equal("class"))
			Ω(c.DeletionProtectionDefault).Should(BeTrue())
			Ω(c.DisableCRDWebhooks).Should(BeTrue())
			Ω(c.EnableMigrations).Should(BeTrue())
			Ω(len(c.GitCredentials)).Should(Equal(1))
			Ω(c.GitCredentials[0].User).Should(Equal("my-git-user"))
			Ω(c.GitCredentials[0].Password).Should(Equal("my-git-password"))
			Ω(c.GitCredentials[0].RepoPrefix).Should(Equal("https://github.com/my-org"))
			Ω(c.LogLevel).Should(Equal(42))
			Ω(c.NotificationWebhook).ShouldNot(BeNil())
			Ω(c.NotificationWebhook.Address).Should(Equal("https://webhook.com"))
			Ω(c.Sentry).ShouldNot(BeNil())
			Ω(c.Sentry.Debug).Should(BeTrue())
			Ω(c.Sentry.DSN).Should(Equal("https://sentry.com"))
			Ω(c.Sentry.Environment).Should(Equal("test"))
			Ω(c.Sentry.HTTPSProxy).Should(Equal("https://proxy.com"))
			Ω(c.SSO).ShouldNot(BeNil())
			Ω(c.SSO.IdentityProvider).Should(Equal("my-provider"))
			Ω(c.SSO.JWTGroupsClaim).Should(Equal("groups"))
			Ω(c.Styra.Address).Should(Equal("https://styra.com"))
			Ω(c.Styra.Token).Should(Equal("token"))
			Ω(c.SystemPrefix).Should(Equal("prefix"))
			Ω(c.SystemSuffix).Should(Equal("suffix"))
			Ω(len(c.SystemUserRoles)).Should(Equal(1))
			Ω(c.SystemUserRoles[0]).Should(Equal("SystemViewer"))
		})
	})
})
