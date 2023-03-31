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

package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"

	v1 "github.com/bankdata/styra-controller/api/config/v1"
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
)

var _ = Describe("ProjectConfig", func() {

	Describe("ToV2Alpha1", func() {
		It("converts to v2alpha1", func() {
			v1cfg := &v1.ProjectConfig{
				ControllerManagerConfigurationSpec: cfg.ControllerManagerConfigurationSpec{
					CacheNamespace: "test",
				},
				StyraToken:               "token",
				StyraAddress:             "addr",
				StyraSystemUserRoles:     []string{"role1", "role2"},
				StyraSystemPrefix:        "prefix",
				StyraSystemSuffix:        "suffix",
				LogLevel:                 42,
				SentryDSN:                "https://my-sentry.com",
				SentryDebug:              true,
				Environment:              "test",
				SentryHTTPSProxy:         "https://my-proxy.com",
				ControllerClass:          "class",
				DatasourceWebhookAddress: "https://my-webhook.com",
				WebhooksDisabled:         true,
				MigrationEnabled:         true,
				GitCredentials: []*v1.GitCredential{
					{
						User:       "user",
						Password:   "password",
						RepoPrefix: "https://github.com/my-org",
					},
					{
						User:       "other-user",
						Password:   "other-password",
						RepoPrefix: "https://github.com/my-other-org",
					},
				},
			}

			expected := &v2alpha1.ProjectConfig{
				ControllerManagerConfigurationSpec: cfg.ControllerManagerConfigurationSpec{
					CacheNamespace: "test",
				},
				ControllerClass:    "class",
				DisableCRDWebhooks: true,
				EnableMigrations:   true,
				GitCredentials: []*v2alpha1.GitCredential{
					{
						User:       "user",
						Password:   "password",
						RepoPrefix: "https://github.com/my-org",
					},
					{
						User:       "other-user",
						Password:   "other-password",
						RepoPrefix: "https://github.com/my-other-org",
					},
				},
				LogLevel:        42,
				SystemPrefix:    "prefix",
				SystemSuffix:    "suffix",
				SystemUserRoles: []string{"role1", "role2"},
				Sentry: &v2alpha1.SentryConfig{
					DSN:         "https://my-sentry.com",
					Debug:       true,
					Environment: "test",
					HTTPSProxy:  "https://my-proxy.com",
				},
				NotificationWebhook: &v2alpha1.NotificationWebhookConfig{
					Address: "https://my-webhook.com",
				},
				Styra: v2alpha1.StyraConfig{
					Token:   "token",
					Address: "addr",
				},
			}

			Ω(v1cfg.ToV2Alpha1()).To(Equal(expected))
		})
	})
})
