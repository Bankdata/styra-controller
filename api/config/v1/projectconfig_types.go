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

package v1

import (
	"github.com/bankdata/styra-controller/api/config/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//nolint:staticcheck // issue https://github.com/Bankdata/styra-controller/issues/82
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// ProjectConfig is the Schema for the projectconfigs API
type ProjectConfig struct {
	metav1.TypeMeta `json:",inline"`

	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	StyraToken               string           `json:"styraToken"`
	StyraAddress             string           `json:"styraAddress"`
	StyraSystemUserRoles     []string         `json:"styraSystemUserRoles"`
	StyraSystemPrefix        string           `json:"styraSystemPrefix"`
	StyraSystemSuffix        string           `json:"styraSystemSuffix"`
	LogLevel                 int              `json:"logLevel"`
	SentryDSN                string           `json:"sentryDSN"`
	SentryDebug              bool             `json:"sentryDebug"`
	Environment              string           `json:"environment"`
	SentryHTTPSProxy         string           `json:"sentryHTTPSProxy"`
	ControllerClass          string           `json:"controllerClass"`
	WebhooksDisabled         bool             `json:"webhooksDisabled"`
	DatasourceWebhookAddress string           `json:"datasourceWebhookAddress"`
	IdentityProvider         string           `json:"identityProvider"`
	JwtGroupClaim            string           `json:"jwtGroupClaim"`
	MigrationEnabled         bool             `json:"migrationEnabled"`
	GitCredentials           []*GitCredential `json:"gitCredentials"`
	// Only used by the now deprecated StyraSystem controller
	GitUser     string `json:"gitUser"`
	GitPassword string `json:"gitPassword"`
}

// GitCredential defines the structure of a git credential.
type GitCredential struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	RepoPrefix string `json:"repoPrefix"`
}

func init() {
	SchemeBuilder.Register(&ProjectConfig{})
}

// ToV2Alpha1 returns this ProjectConfig converted to a v2alpha1.ProjectConfig
func (c *ProjectConfig) ToV2Alpha1() *v2alpha1.ProjectConfig {
	v2cfg := &v2alpha1.ProjectConfig{
		ControllerManagerConfigurationSpec: c.ControllerManagerConfigurationSpec,
		ControllerClass:                    c.ControllerClass,
		DisableCRDWebhooks:                 c.WebhooksDisabled,
		EnableMigrations:                   c.MigrationEnabled,
		LogLevel:                           c.LogLevel,
		SystemPrefix:                       c.StyraSystemPrefix,
		SystemSuffix:                       c.StyraSystemSuffix,
		SystemUserRoles:                    c.StyraSystemUserRoles,
		Styra: v2alpha1.StyraConfig{
			Token:   c.StyraToken,
			Address: c.StyraAddress,
		},
	}

	if c.SentryDSN != "" {
		v2cfg.Sentry = &v2alpha1.SentryConfig{
			DSN:         c.SentryDSN,
			Debug:       c.SentryDebug,
			Environment: c.Environment,
			HTTPSProxy:  c.SentryHTTPSProxy,
		}
	}

	if c.DatasourceWebhookAddress != "" {
		v2cfg.NotificationWebhook = &v2alpha1.NotificationWebhookConfig{
			Address: c.DatasourceWebhookAddress,
		}
	}

	if c.GitCredentials != nil {
		v2cfg.GitCredentials = make([]*v2alpha1.GitCredential, len(c.GitCredentials))
		for i, c := range c.GitCredentials {
			v2cfg.GitCredentials[i] = &v2alpha1.GitCredential{
				User:       c.User,
				Password:   c.Password,
				RepoPrefix: c.RepoPrefix,
			}
		}
	}

	return v2cfg
}
