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

package v2alpha1

import (
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//nolint:staticcheck // issue https://github.com/Bankdata/styra-controller/issues/82
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"

	"github.com/bankdata/styra-controller/api/config/v2alpha2"
)

//+kubebuilder:object:root=true

// ProjectConfig is the Schema for the projectconfigs API
type ProjectConfig struct {
	metav1.TypeMeta `json:",inline"`

	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// ControllerClass sets a controller class for this controller. This allows
	// the provided CRDs to target a specific controller. This is useful when
	// running multiple controllers in the same cluster.
	ControllerClass string `json:"controllerClass"`

	// DeletionProtectionDefault sets the default to use with regards to deletion
	// protection if it is not set on the resource.
	DeletionProtectionDefault bool `json:"deletionProtectionDefault"`

	// DisableCRDWebhooks disables the CRD webhooks on the controller. If running
	// multiple controllers in the same cluster, only one will need to have it's
	// webhooks enabled.
	DisableCRDWebhooks bool `json:"disableCRDWebhooks"`

	// EnableMigrations enables the system migration annotation. This should be
	// kept disabled unless migrations need to be done.
	EnableMigrations bool `json:"enableMigrations"`

	// GitCredentials holds a list of git credential configurations. The
	// RepoPrefix of the GitCredential will be matched angainst repository URL in
	// order to determine which credential to use. The GitCredential with the
	// longest matching RepoPrefix will be selected.
	GitCredentials []*GitCredential `json:"gitCredentials"`

	// LogLevel sets the logging level of the controller. A higher number gives
	// more verbosity. A number higher than 0 should only be used for debugging
	// purposes.
	LogLevel int `json:"logLevel"`

	NotificationWebhook *NotificationWebhookConfig `json:"notificationWebhook"`

	Sentry *SentryConfig `json:"sentry"`

	SSO *SSOConfig `json:"sso"`

	Styra StyraConfig `json:"styra"`

	// SystemPrefix is a prefix for all the systems that the controller creates
	// in Styra DAS. This is useful in order to be able to identify what
	// controller created a system in a shared Styra DAS instance.
	SystemPrefix string `json:"systemPrefix"`

	// SystemSuffix is a suffix for all the systems that the controller creates
	// in Styra DAS. This is useful in order to be able to identify what
	// controller created a system in a shared Styra DAS instance.
	SystemSuffix string `json:"systemSuffix"`

	// SystemUserRoles is a list of Styra DAS system level roles which the subjects of
	// a system will be granted.
	SystemUserRoles []string `json:"systemUserRoles"`
}

// StyraConfig contains configuration for connecting to the Styra DAS apis
type StyraConfig struct {
	// Address is the URL for the Styra DAS API server.
	Address string `json:"address"`

	// Token is a Styra DAS API token. These can be created in the Styra DAS GUI
	// or through the API. The token should have the `WorkspaceAdministrator` role.
	Token string `json:"token"`
}

// SentryConfig contains configuration for how errors should be reported to
// sentry.
type SentryConfig struct {
	// Debug enables Sentry client debugging.
	Debug bool `json:"debug"`

	// DSN is the Sentry project DSN.
	DSN string `json:"dsn"`

	// Environment sets the environment of the events sent to Sentry.
	Environment string `json:"environment"`

	// HTTPSProxy sets an HTTP proxy server for sentry to use.
	HTTPSProxy string `json:"httpsProxy"`
}

// NotificationWebhookConfig contains configuration for how to call the notification
// webhook.
type NotificationWebhookConfig struct {
	// Address is the URL to be called when the controller should do a webhook
	// notification. Currently the only supported notification is that a
	// datasource configuration has changed.
	Address string `json:"address"`
}

// SSOConfig contains configuration for how to use SSO tokens for determining
// what groups a user belongs to. This can be used to grant members of a
// certain group access to systems.
type SSOConfig struct {
	// IdentityProvider is the ID of a configured Styra DAS identity provider.
	IdentityProvider string `json:"identityProvider"`

	// JWTGroupsClaim is the json path to a claim in issued JWTs which contain a
	// list of groups that the user belongs to.
	JWTGroupsClaim string `json:"jwtGroupsClaim"`
}

// GitCredential represents a set of credentials to be used for repositories
// that match the RepoPrefix.
type GitCredential struct {
	// User is a http basic auth username used for git.
	User string `json:"user"`

	// Password is a http basic auth password used for git.
	Password string `json:"password"`

	// RepoPrefix specifies a repo URL prefix. eg. if RepoPrefix is set to
	// `https://github.com/bankdata`, then this credentials would apply for any
	// repository under the bankdata github org.
	RepoPrefix string `json:"repoPrefix"`
}

// GetGitCredentialForRepo determines which default GitCredential to use for checking out the
// policy repository based on the URL to the policy repository.
func (c *ProjectConfig) GetGitCredentialForRepo(repo string) *GitCredential {
	sort.Slice(c.GitCredentials, func(i, j int) bool {
		return len(c.GitCredentials[i].RepoPrefix) > len(c.GitCredentials[j].RepoPrefix)
	})

	for _, gitCredential := range c.GitCredentials {
		if strings.HasPrefix(repo, gitCredential.RepoPrefix) {
			return gitCredential
		}
	}

	return nil
}

// ToV2Alpha2 returns this ProjectConfig converted to a v2alpha2.ProjectConfig
func (c *ProjectConfig) ToV2Alpha2() *v2alpha2.ProjectConfig {
	v2cfg := &v2alpha2.ProjectConfig{
		ControllerClass:           c.ControllerClass,
		DeletionProtectionDefault: c.DeletionProtectionDefault,
		DisableCRDWebhooks:        c.DisableCRDWebhooks,
		EnableMigrations:          c.EnableMigrations,
		LogLevel:                  c.LogLevel,
		Styra: v2alpha2.StyraConfig{
			Token:   c.Styra.Token,
			Address: c.Styra.Address,
		},
		SystemPrefix:    c.SystemPrefix,
		SystemSuffix:    c.SystemSuffix,
		SystemUserRoles: c.SystemUserRoles,
	}

	if c.SSO != nil {
		v2cfg.SSO = &v2alpha2.SSOConfig{
			IdentityProvider: c.SSO.IdentityProvider,
			JWTGroupsClaim:   c.SSO.JWTGroupsClaim,
		}
	}

	if c.GitCredentials != nil {
		v2cfg.GitCredentials = make([]*v2alpha2.GitCredential, len(c.GitCredentials))
		for i, c := range c.GitCredentials {
			v2cfg.GitCredentials[i] = &v2alpha2.GitCredential{
				User:       c.User,
				Password:   c.Password,
				RepoPrefix: c.RepoPrefix,
			}
		}
	}

	if c.LeaderElection != nil && c.LeaderElection.LeaderElect != nil && *c.LeaderElection.LeaderElect {
		v2cfg.LeaderElection = &v2alpha2.LeaderElectionConfig{
			LeaseDuration: c.LeaderElection.LeaseDuration,
			RenewDeadline: c.LeaderElection.RenewDeadline,
			RetryPeriod:   c.LeaderElection.RetryPeriod,
		}
	}

	if c.NotificationWebhook != nil {
		v2cfg.NotificationWebhook = &v2alpha2.NotificationWebhookConfig{
			Address: c.NotificationWebhook.Address,
		}
	}

	if c.Sentry != nil {
		v2cfg.Sentry = &v2alpha2.SentryConfig{
			DSN:         c.Sentry.DSN,
			Debug:       c.Sentry.Debug,
			Environment: c.Sentry.Environment,
			HTTPSProxy:  c.Sentry.HTTPSProxy,
		}
	}

	return v2cfg
}

func init() {
	SchemeBuilder.Register(&ProjectConfig{})
}
