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

package v2alpha2

import (
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true

// ProjectConfig is the Schema for the projectconfigs API
type ProjectConfig struct {
	metav1.TypeMeta `json:",inline"`

	// ControllerClass sets a controller class for this controller. This allows
	// the provided CRDs to target a specific controller. This is useful when
	// running multiple controllers in the same cluster.
	ControllerClass string `json:"controllerClass"`

	// DeletionProtectionDefault sets the default to use with regards to deletion
	// protection if it is not set on the resource.
	DeletionProtectionDefault bool `json:"deletionProtectionDefault"`

	// ReadOnly sets the value of ReadOnly for systems
	//+kubebuilder:deprecatedversion:warning="ReadOnly field is deprecated, only used in Styra"
	// Deprecated: ReadOnly field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	ReadOnly bool `json:"readOnly"`

	// EnableDeltaBundlesDefault sets the default of whether systems have delta-bundles or not
	//+kubebuilder:deprecatedversion:warning="EnableDeltaBundlesDefault field is deprecated, only used in Styra"
	// Deprecated: EnableDeltaBundlesDefault field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	EnableDeltaBundlesDefault *bool `json:"enableDeltaBundlesDefault,omitempty"`

	// DisableCRDWebhooks disables the CRD webhooks on the controller. If running
	// multiple controllers in the same cluster, only one will need to have it's
	// webhooks enabled.
	DisableCRDWebhooks bool `json:"disableCRDWebhooks"`

	// EnableMigrations enables the system migration annotation. This should be
	// kept disabled unless migrations need to be done.
	EnableMigrations bool `json:"enableMigrations"`

	// DatasourceIgnorePatterns is a list of regex patterns, that allow datasources in styra
	// to be ignored based on their datasource id.
	//+kubebuilder:deprecatedversion:warning="DatasourceIgnorePatterns field is deprecated, only used in Styra"
	// Deprecated: DatasourceIgnorePatterns field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	DatasourceIgnorePatterns []string `json:"datasourceIgnorePatterns,omitempty"`

	// GitCredentials holds a list of git credential configurations. The
	// RepoPrefix of the GitCredential will be matched angainst repository URL in
	// order to determine which credential to use. The GitCredential with the
	// longest matching RepoPrefix will be selected.
	//+kubebuilder:deprecatedversion:warning="GitCredentials field is deprecated, only used in Styra"
	// Deprecated: GitCredentials field is deprecated, only used in Styra.
	//  This field will be removed in a future version.
	GitCredentials []*GitCredential `json:"gitCredentials"`

	// LogLevel sets the logging level of the controller. A higher number gives
	// more verbosity. A number higher than 0 should only be used for debugging
	// purposes.
	LogLevel int `json:"logLevel"`

	LeaderElection *LeaderElectionConfig `json:"leaderElection"`

	NotificationWebhooks *NotificationWebhooksConfig `json:"notificationWebhooks,omitempty"`

	Sentry *SentryConfig `json:"sentry"`

	//+kubebuilder:deprecatedversion:warning="SSO field is deprecated, only used in Styra"
	// Deprecated: SSO field is deprecated, only used in Styra. This field will be removed in a future version.
	SSO *SSOConfig `json:"sso"`

	//+kubebuilder:deprecatedversion:warning="Styra field is deprecated, use OPAControlPlaneConfig instead"
	// Deprecated: Use OPAControlPlaneConfig instead. This field will be removed in a future version.
	Styra StyraConfig `json:"styra"`

	OPA OPAConfig `json:"opa,omitempty"`

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
	//+kubebuilder:deprecatedversion:warning="SystemUserRoles field is deprecated, only used in Styra"
	// Deprecated: SystemUserRoles field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	SystemUserRoles []string `json:"systemUserRoles"`

	//+kubebuilder:deprecatedversion:warning="DecisionsExporter field is deprecated, only used in Styra"
	// Deprecated: DecisionsExporter field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	DecisionsExporter *ExporterConfig `json:"decisionsExporter,omitempty"`

	//+kubebuilder:deprecatedversion:warning="ActivityExporter field is deprecated, only used in Styra"
	// Deprecated: ActivityExporter field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	ActivityExporter *ExporterConfig `json:"activityExporter,omitempty"`

	PodRestart *PodRestartConfig `json:"podRestart,omitempty"`

	// OPAControlPlaneConfig contains configuration for connecting to the
	// OPA Control Plane APIs. If this is not set, the controller will not
	// attempt to connect to the OPA Control Plane APIs.
	OPAControlPlaneConfig *OPAControlPlaneConfig `json:"opaControlPlaneConfig,omitempty"`

	// UserCredentialHandler contains configuration for the controller to handle user credentials, e.g. in S3
	UserCredentialHandler *UserCredentialHandler `json:"userCredentialHandler,omitempty"`

	// EnableStyraReconciliation is a flag that sets whether the controller should use Styra
	// A Migration flag to enable/disable Styra DAS reconciliation for all systems and libraries.
	//+kubebuilder:deprecatedversion:warning="EnableStyraReconciliation field is deprecated.
	// Only used in migration versions Styra->OCP"
	// Deprecated: EnableStyraReconciliation field is deprecated, only used in migration versions Styra->OCP.
	// This field will be removed in a future version.
	EnableStyraReconciliation bool `json:"enableStyraReconciliation,omitempty"`

	// EnableOPAControlPlaneReconciliation is a flag that sets whether the controller should use OPAControlPlane
	// A Migration flag to enable/disable OPA Control Plane reconciliation for all systems and libraries.
	//+kubebuilder:deprecatedversion:warning="EnableOPAControlPlaneReconciliation field is deprecated,
	// only used in migration versions Styra->OCP"
	// Deprecated: EnableOPAControlPlaneReconciliation field is deprecated, only used in migration versions Styra->OCP.
	// This field will be removed in a future version.
	EnableOPAControlPlaneReconciliation bool `json:"enableOPAControlPlaneReconciliation,omitempty"`

	// EnableOPAControlPlaneReconciliationTestData is a flag that sets whether the controller should create
	// OPAControlPlane test data.
	// A Migration flag to allow adding test data to OPA Control Plane and not to modify k8s data
	//+kubebuilder:deprecatedversion:warning="EnableOPAControlPlaneReconciliationTestData field is deprecated.
	// Only used in migration versions Styra->OCP"
	// Deprecated: EnableOPAControlPlaneReconciliationTestData field is deprecated.
	// Only used in migration versions Styra->OCP. This field will be removed in a future version.
	EnableOPAControlPlaneReconciliationTestData bool `json:"enableOPAControlPlaneReconciliationTestData,omitempty"`
}

// PodRestartConfig contains configuration for restarting OPA and SLP pods
type PodRestartConfig struct {
	SLPRestart *SLPRestartConfig `json:"slpRestart,omitempty"`
	OPARestart *OPARestartConfig `json:"opaRestart,omitempty"`
}

// SLPRestartConfig contains configuration for restarting SLP pods
type SLPRestartConfig struct {
	Enabled        bool   `json:"enabled"`
	DeploymentType string `json:"deploymentType"` // DeploymentType only currently supports "StatefulSet""
}

// OPARestartConfig contains configuration for restarting OPA pods -- This is not yet implemented
type OPARestartConfig struct {
	Enabled        bool   `json:"enabled"`
	DeploymentType string `json:"deploymentType"`
}

// LeaderElectionConfig contains configuration for leader election
type LeaderElectionConfig struct {
	LeaseDuration metav1.Duration `json:"leaseDuration"`
	RenewDeadline metav1.Duration `json:"renewDeadline"`
	RetryPeriod   metav1.Duration `json:"retryPeriod"`
}

// StyraConfig contains configuration for connecting to the Styra DAS apis
type StyraConfig struct {
	// Address is the URL for the Styra DAS API server.
	Address string `json:"address"`

	// Token is a Styra DAS API token. These can be created in the Styra DAS GUI
	// or through the API. The token should have the `WorkspaceAdministrator` role.
	Token string `json:"token"`

	// Alternative to the "token" whice define the Styra DAS API token directly in the config file,
	// this "tokenSecretPath" will use a token from a secret (only if "token" is not set)
	TokenSecretPath string `json:"tokenSecretPath"`
}

// OPAControlPlaneConfig defines the config for the OPA Control Plane.
type OPAControlPlaneConfig struct {
	// Address is the URL for the OPA Control Plane API server.
	Address string `json:"address"`

	// Token is a OPA Control Plane API token.
	Token string `json:"token"`

	// GitCredentials is the name of a secret used by the OPA Control Plane Git integration.
	GitCredentials []*GitCredentials `json:"gitCredentials,omitempty"`

	// BundleObjectStorage is the object storage configuration to use for bundles.
	// Currently only supports aws
	BundleObjectStorage *BundleObjectStorage `json:"bundleObjectStorage,omitempty"`

	// DefaultRequirements is a list of requirements that will be added to all
	// systems/bundles created by the controller in the OCP, in addition to any requirements/datasources
	// specified on the System resource.
	DefaultRequirements []string `json:"defaultRequirements,omitempty"`

	// SystemDatasourceChanged is the URL to be called when a system datasource has changed.
	SystemDatasourceChanged string `json:"systemDatasourceChanged,omitempty"`
	// LibraryDatasourceChanged is the URL to be called when a library datasource has changed.
	LibraryDatasourceChanged string `json:"libraryDatasourceChanged,omitempty"`
}

// UserCredentialHandler defines the structure of possible user credential handlers
type UserCredentialHandler struct {
	S3 *S3Handler `json:"s3,omitempty"`
}

// S3Handler defines the structure for S3 handler configuration.
type S3Handler struct {
	Bucket          string `json:"bucket"`
	URL             string `json:"url"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
}

// BundleObjectStorage defines the structure for object storage configuration used by bundles
type BundleObjectStorage struct {
	S3 *S3ObjectStorage `json:"s3,omitempty"`
}

// S3ObjectStorage defines the structure for S3 object storage configuration.
type S3ObjectStorage struct {
	Bucket              string `json:"bucket"`
	URL                 string `json:"url,omitempty"`
	Region              string `json:"region"`
	OCPConfigSecretName string `json:"ocpConfigSecretName"`
}

// GitCredentials contains configuration for git credentials used by the OPA Control Plane.
type GitCredentials struct {
	ID string `json:"id"`

	// RepoPrefix specifies a repo URL prefix. eg. if RepoPrefix is set to
	// `https://github.com/bankdata`, then this credentials would apply for any
	// repository under the bankdata github org.
	RepoPrefix string `json:"repoPrefix"`
}

// OPAConfig contains default configuration for the opa config generated by the styra-controller
type OPAConfig struct {
	DecisionLogs DecisionLog `json:"decision_logs"`
}

// DecisionLog contains configuration for the decision logs
type DecisionLog struct {
	RequestContext RequestContext `json:"request_context"`
}

// RequestContext contains configuration for the RequestContext in the decision logs
type RequestContext struct {
	HTTP HTTP `json:"http"`
}

// HTTP contains configuration for the HTTP config in the RequestContext
type HTTP struct {
	// http headers that will be added to the decision logs
	Headers []string `json:"headers"`
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

// NotificationWebhooksConfig contains configuration for how to call the notification
// webhook.
type NotificationWebhooksConfig struct {
	// Address is the URL to be called when the controller should do a webhook
	// notification. Currently the only supported notification is that a
	// datasource configuration has changed.

	//+kubebuilder:deprecatedversion:warning="SystemDatasourceChanged field is deprecated, only used in Styra"
	// Deprecated: SystemDatasourceChanged field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	SystemDatasourceChanged string `json:"systemDatasourceChanged,omitempty"`
	//+kubebuilder:deprecatedversion:warning="LibraryDatasourceChanged field is deprecated, only used in Styra"
	// Deprecated: LibraryDatasourceChanged field is deprecated, only used in Styra.
	// This field will be removed in a future version.
	LibraryDatasourceChanged string `json:"libraryDatasourceChanged,omitempty"`
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

// ExporterConfigType defines the type of exporter config
type ExporterConfigType string

const (
	// ExporterConfigTypeDecisions is the type for decisions exporter config
	ExporterConfigTypeDecisions ExporterConfigType = "DecisionsExporter"
	// ExporterConfigTypeActivity is the type for activity exporter config
	ExporterConfigTypeActivity ExporterConfigType = "ActivityExporter"
)

// ExporterConfig contains configuration for exports
type ExporterConfig struct {
	Enabled  bool         `json:"enabled,omitempty"`
	Interval string       `json:"interval,omitempty"`
	Kafka    *KafkaConfig `json:"kafka,omitempty"`
}

// KafkaConfig contains configuration for exporting decisions to Kafka
type KafkaConfig struct {
	Brokers      []string   `json:"brokers"`
	Topic        string     `json:"topic"`
	RequiredAcks string     `json:"requiredAcks"`
	TLS          *TLSConfig `json:"tls,omitempty"`
}

// TLSConfig contains TLS configuration for Kafka decisions export.
type TLSConfig struct {
	ClientCertificateName string `json:"clientCertificateName"`
	ClientCertificate     string `json:"clientCertificate"`
	ClientKey             string `json:"clientKey"`
	RootCA                string `json:"rootCA"`
	InsecureSkipVerify    bool   `json:"insecureSkipVerify"`
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

// SLPRestartEnabled returns true if the OPA restart is enabled
func (c *ProjectConfig) SLPRestartEnabled() bool {
	if c.PodRestart == nil || c.PodRestart.SLPRestart == nil {
		return false
	}
	return c.PodRestart.SLPRestart.Enabled
}

// OPARestartEnabled returns true if the OPA restart is enabled
func (c *ProjectConfig) OPARestartEnabled() bool {
	if c.PodRestart == nil || c.PodRestart.OPARestart == nil {
		return false
	}
	return c.PodRestart.OPARestart.Enabled
}

func init() {
	SchemeBuilder.Register(&ProjectConfig{})
}
