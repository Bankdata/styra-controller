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

package ocp

import (
	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
)

// OPAConfig stores the information going into the ConfigMap for the OPA
type OPAConfig struct {
	BundleService        *OPAServiceConfig
	LogService           *OPAServiceConfig
	UniqueName           string
	Namespace            string
	BundleResource       string
	DecisionLogReporting configv2alpha2.DecisionLogReporting
}

// OPAServiceConfig defines a services added to the OPAs' config files.
type OPAServiceConfig struct {
	Name                         string              `json:"name" yaml:"name"`
	Credentials                  *ServiceCredentials `json:"credentials" yaml:"credentials"`
	ResponseHeaderTimeoutSeconds int                 `json:"response_header_timeout_seconds,omitempty" yaml:"response_header_timeout_seconds,omitempty"` //nolint:lll
	URL                          string              `json:"url" yaml:"url"`
}

// ServiceCredentials defines the structure for service credentials.
type ServiceCredentials struct {
	Bearer *Bearer    `json:"bearer,omitempty" yaml:"bearer,omitempty"`
	S3     *S3Signing `json:"s3_signing,omitempty" yaml:"s3_signing,omitempty"`
}

// S3Signing defines the structure for S3 signing configuration.
type S3Signing struct {
	S3EnvironmentCredentials map[string]EmptyStruct `json:"environment_credentials" yaml:"environment_credentials"`
}

// EmptyStruct is an empty struct used for mapping empty values in S3EnvironmentCredentials
type EmptyStruct struct{}

// Bearer defines the structure for bearer token credentials.
type Bearer struct {
	TokenPath string `json:"token_path" yaml:"token_path"`
}
