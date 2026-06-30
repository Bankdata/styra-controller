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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:object:root=true

// ProjectConfig is the Schema for the projectconfigs API
type ProjectConfig struct {
	metav1.TypeMeta `json:",inline"`

	// ControllerClass sets a controller class for this controller. This allows
	// the provided CRDs to target a specific controller. This is useful when
	// running multiple controllers in the same cluster.
	ControllerClass string `json:"controllerClass"`

	// NamespaceSelector defines criteria for only accepting namespaces matching the selector.
	// If a resource is in a namespace that matches the criteria defined in NamespaceSelector,
	// it will be included in reconciliation.
	// If not set, the controller will reconcile resources in all namespaces.
	NamespaceSelector *NamespaceSelector `json:"namespaceSelector,omitempty"`

	// DeletionProtectionDefault sets the default to use with regards to deletion
	// protection if it is not set on the resource.
	DeletionProtectionDefault bool `json:"deletionProtectionDefault"`

	// DisableCRDWebhooks disables the CRD webhooks on the controller. If running
	// multiple controllers in the same cluster, only one will need to have it's
	// webhooks enabled.
	DisableCRDWebhooks bool `json:"disableCRDWebhooks"`

	// LogLevel sets the logging level of the controller. A higher number gives
	// more verbosity. A number higher than 0 should only be used for debugging
	// purposes.
	LogLevel int `json:"logLevel"`

	LeaderElection *LeaderElectionConfig `json:"leaderElection"`

	OPA OPAConfig `json:"opa,omitempty"`

	// SystemPrefix is a prefix for all the systems that the controller creates.
	SystemPrefix string `json:"systemPrefix"`

	// SystemSuffix is a suffix for all the systems that the controller creates.
	SystemSuffix string `json:"systemSuffix"`

	// OPAControlPlaneConfig contains configuration for connecting to the
	// OPA Control Plane APIs. If this is not set, the controller will not
	// attempt to connect to the OPA Control Plane APIs.
	OPAControlPlaneConfig *OPAControlPlaneConfig `json:"opaControlPlaneConfig,omitempty"`
}

// LeaderElectionConfig contains configuration for leader election
type LeaderElectionConfig struct {
	LeaseDuration metav1.Duration `json:"leaseDuration"`
	RenewDeadline metav1.Duration `json:"renewDeadline"`
	RetryPeriod   metav1.Duration `json:"retryPeriod"`
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

// NamespaceSelector defines criteria for only accepting MatchPatterns namespaces for reconciliation.
type NamespaceSelector struct {
	// MatchPatterns is a list of glob patterns to match namespace names against.
	MatchPatterns []string `json:"matchPatterns,omitempty"`
}

// BundleObjectStorage defines the structure for object storage configuration used by bundles
type BundleObjectStorage struct {
	S3 *S3ObjectStorage `json:"s3,omitempty" yaml:"s3,omitempty"`
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

// OPAConfig contains default configuration for generated OPA config.
type OPAConfig struct {
	DecisionLogs           DecisionLog        `json:"decisionLogs,omitempty" yaml:"decisionLogs,omitempty"`
	Metrics                MetricsConfig      `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	PersistBundle          bool               `json:"persist_bundle,omitempty" yaml:"persist_bundle,omitempty"`
	PersistBundleDirectory string             `json:"persist_bundle_directory,omitempty" yaml:"persist_bundle_directory,omitempty"` //nolint:lll
	BundleServer           *OPABundleServer   `json:"bundleServer,omitempty" yaml:"bundleServer,omitempty"`
	DecisionAPIConfig      *DecisionAPIConfig `json:"decisionAPIConfig,omitempty" yaml:"decisionAPIConfig,omitempty"`
}

// OPABundleServer contains configuration for the OPA bundle server
type OPABundleServer struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	URL       string `json:"url,omitempty" yaml:"url,omitempty"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	TokenPath string `json:"tokenPath,omitempty" yaml:"tokenPath,omitempty"`
}

// MetricsConfig contains configuration for OPA metrics
type MetricsConfig struct {
	Prometheus PrometheusMetricsConfig `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}

// PrometheusMetricsConfig contains configuration for Prometheus metrics
type PrometheusMetricsConfig struct {
	HTTP HTTPMetricsConfig `json:"http,omitempty" yaml:"http,omitempty"`
}

// HTTPMetricsConfig contains configuration for HTTP metrics
type HTTPMetricsConfig struct {
	Buckets []float64 `json:"buckets,omitempty" yaml:"buckets,omitempty"`
}

// DecisionLog contains configuration for the decision logs
type DecisionLog struct {
	RequestContext RequestContext `json:"requestContext,omitempty"`
}

// DecisionAPIConfig contains configuration for decision log dispatch
type DecisionAPIConfig struct {
	Name       string               `json:"name,omitempty"`
	ServiceURL string               `json:"serviceUrl,omitempty"`
	TokenPath  string               `json:"tokenPath,omitempty"`
	Reporting  DecisionLogReporting `json:"reporting,omitempty"`
}

// DecisionLogReporting contains configuration for decision log reporting
type DecisionLogReporting struct {
	MaxDelaySeconds      int `json:"maxDelaySeconds,omitempty" yaml:"maxDelaySeconds,omitempty"`
	MinDelaySeconds      int `json:"minDelaySeconds,omitempty" yaml:"minDelaySeconds,omitempty"`
	UploadSizeLimitBytes int `json:"uploadSizeLimitBytes,omitempty" yaml:"uploadSizeLimitBytes,omitempty"`
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

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ProjectConfig{})
		return nil
	})
}
