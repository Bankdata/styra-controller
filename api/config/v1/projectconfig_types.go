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
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
