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

package k8sconv_test

import (
	"strings"

	"github.com/go-logr/logr"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"gopkg.in/yaml.v2"
)

// Test PersistBundle config
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMap", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			nil,
			test.customConfig,
			nil,
			logr.Discard())

		gomega.Expect(err).To(gomega.BeNil())

		var actualMap, expectedMap map[string]interface{}
		actualYAML := cm.Data["opa-conf.yaml"]
		expectedYAML := test.expectedCMContent

		err = yaml.Unmarshal([]byte(actualYAML), &actualMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal actual YAML")

		err = yaml.Unmarshal([]byte(expectedYAML), &expectedMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal expected YAML")

		gomega.Expect(actualMap).To(gomega.Equal(expectedMap))
	},
		ginkgo.Entry("success", test{
			opaDefaultConfig: configv2alpha2.OPAConfig{
				DecisionLogs: configv2alpha2.DecisionLog{
					RequestContext: configv2alpha2.RequestContext{
						HTTP: configv2alpha2.HTTP{
							Headers: strings.Split("header1,header2", ","),
						},
					},
				},
				PersistBundle:          true,
				PersistBundleDirectory: "/opa-bundles",
			},
			opaconf: ocp.OPAConfig{
				UniqueName:     "system",
				Namespace:      "system-ns",
				BundleResource: "bundles/system/bundle.tar.gz",
				BundleService: &ocp.OPAServiceConfig{
					Name: "s3",
					URL:  "https://minio/ocp",
					Credentials: &ocp.ServiceCredentials{
						S3: &ocp.S3Signing{
							S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
						},
					},
				},
				LogService: &ocp.OPAServiceConfig{
					Name: "logs",
					URL:  "https://log-service/ocp",
					Credentials: &ocp.ServiceCredentials{
						Bearer: &ocp.Bearer{
							TokenPath: "/etc/opa/auth/token",
						},
					},
				},
				DecisionLogReporting: configv2alpha2.DecisionLogReporting{
					UploadSizeLimitBytes: 1,
					MinDelaySeconds:      2,
					MaxDelaySeconds:      3,
				},
			},
			customConfig: map[string]interface{}{
				"distributed_tracing": map[string]interface{}{
					"type":    "grpc",
					"address": "localhost:1234",
				},
				"bundles": map[string]interface{}{
					"authz": map[string]interface{}{
						"test": 123,
					},
				},
			},
			expectedCMContent: `services:
- name: s3
  url: https://minio/ocp
  credentials:
    s3_signing:
      environment_credentials: {}
- name: logs
  url: https://log-service/ocp
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s3
    persist: true
    test: 123
persistence_directory: /opa-bundles
labels:
  namespace: system-ns
  unique-name: system
decision_logs:
  reporting:
    upload_size_limit_bytes: 1
    min_delay_seconds: 2
    max_delay_seconds: 3
  request_context:
    http:
      headers:
      - header1
      - header2
  service: logs
  resource_path: /logs
distributed_tracing:
  type: grpc
  address: localhost:1234
`,
		}),
	)
})

// Test without PersistBundle config
// And custom config overriding opaconf
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMap", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			nil,
			test.customConfig,
			nil,
			logr.Discard())

		gomega.Expect(err).To(gomega.BeNil())

		var actualMap, expectedMap map[string]interface{}
		actualYAML := cm.Data["opa-conf.yaml"]
		expectedYAML := test.expectedCMContent

		err = yaml.Unmarshal([]byte(actualYAML), &actualMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal actual YAML")

		err = yaml.Unmarshal([]byte(expectedYAML), &expectedMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal expected YAML")

		gomega.Expect(actualMap).To(gomega.Equal(expectedMap))
	},

		ginkgo.Entry("success", test{
			opaDefaultConfig: configv2alpha2.OPAConfig{
				DecisionLogs: configv2alpha2.DecisionLog{
					RequestContext: configv2alpha2.RequestContext{
						HTTP: configv2alpha2.HTTP{
							Headers: strings.Split("header1,header2", ","),
						},
					},
				},
				PersistBundle: false,
			},
			opaconf: ocp.OPAConfig{
				LogService: &ocp.OPAServiceConfig{
					Name: "logs",
					URL:  "https://log-service/ocp",
					Credentials: &ocp.ServiceCredentials{
						Bearer: &ocp.Bearer{
							TokenPath: "/etc/opa/auth/token",
						},
					},
				},
				BundleResource: "bundles/system/bundle.tar.gz",
				BundleService: &ocp.OPAServiceConfig{
					Name: "s3",
					URL:  "https://minio/ocp",
					Credentials: &ocp.ServiceCredentials{
						S3: &ocp.S3Signing{
							S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
						},
					},
				},
				DecisionLogReporting: configv2alpha2.DecisionLogReporting{
					UploadSizeLimitBytes: 1048576,
					MinDelaySeconds:      1,
					MaxDelaySeconds:      30,
				},
			},
			customConfig: map[string]interface{}{
				"distributed_tracing": map[string]interface{}{
					"type":    "grpc",
					"address": "localhost:1234",
				},
				"bundles": map[string]interface{}{
					"authz": map[string]interface{}{
						"service": "s4", //test that custom config overrides default opaconf
					},
				},
			},
			expectedCMContent: `services:
- name: s3
  url: https://minio/ocp
  credentials:
    s3_signing:
      environment_credentials: {}
- name: logs
  url: https://log-service/ocp
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s4
decision_logs:
  reporting:
    upload_size_limit_bytes: 1048576
    min_delay_seconds: 1
    max_delay_seconds: 30
  service: logs
  resource_path: /logs
  request_context:
    http:
      headers:
      - header1
      - header2
distributed_tracing:
  type: grpc
  address: localhost:1234
`,
		}),
	)
})

// Test that spec.opa.config (opaConfig) takes highest precedence over customConfig on conflict.
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMap opaConfig precedence", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		opaConfig         map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			nil,
			test.customConfig,
			test.opaConfig,
			logr.Discard())

		gomega.Expect(err).To(gomega.BeNil())

		var actualMap, expectedMap map[string]interface{}
		actualYAML := cm.Data["opa-conf.yaml"]
		expectedYAML := test.expectedCMContent

		err = yaml.Unmarshal([]byte(actualYAML), &actualMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal actual YAML")

		err = yaml.Unmarshal([]byte(expectedYAML), &expectedMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal expected YAML")

		gomega.Expect(actualMap).To(gomega.Equal(expectedMap))
	},
		ginkgo.Entry("opaConfig wins over customConfig on conflict", test{
			opaconf: ocp.OPAConfig{
				BundleResource: "bundles/system/bundle.tar.gz",
				BundleService: &ocp.OPAServiceConfig{
					Name: "s3",
					URL:  "https://minio/ocp",
					Credentials: &ocp.ServiceCredentials{
						S3: &ocp.S3Signing{
							S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
						},
					},
				},
				LogService: &ocp.OPAServiceConfig{
					Name: "logs",
					URL:  "https://log-service/ocp",
					Credentials: &ocp.ServiceCredentials{
						Bearer: &ocp.Bearer{
							TokenPath: "/etc/opa/auth/token",
						},
					},
				},
				DecisionLogReporting: configv2alpha2.DecisionLogReporting{
					UploadSizeLimitBytes: 1048576,
					MinDelaySeconds:      1,
					MaxDelaySeconds:      30,
				},
			},
			// customConfig sets bundles.authz.service to "s4"
			customConfig: map[string]interface{}{
				"bundles": map[string]interface{}{
					"authz": map[string]interface{}{
						"service": "s4",
					},
				},
			},
			// opaConfig overrides the same key back to "s5" — must win
			opaConfig: map[string]interface{}{
				"bundles": map[string]interface{}{
					"authz": map[string]interface{}{
						"service": "s5",
					},
				},
			},
			expectedCMContent: `services:
- name: s3
  url: https://minio/ocp
  credentials:
    s3_signing:
      environment_credentials: {}
- name: logs
  url: https://log-service/ocp
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s5
decision_logs:
  reporting:
    upload_size_limit_bytes: 1048576
    min_delay_seconds: 1
    max_delay_seconds: 30
  service: logs
  resource_path: /logs
`,
		}),
		ginkgo.Entry("opaConfig adds keys not in customConfig", test{
			opaconf: ocp.OPAConfig{
				BundleResource: "bundles/system/bundle.tar.gz",
				BundleService: &ocp.OPAServiceConfig{
					Name: "s3",
					URL:  "https://minio/ocp",
					Credentials: &ocp.ServiceCredentials{
						S3: &ocp.S3Signing{
							S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
						},
					},
				},
				LogService: &ocp.OPAServiceConfig{
					Name: "logs",
					URL:  "https://log-service/ocp",
					Credentials: &ocp.ServiceCredentials{
						Bearer: &ocp.Bearer{
							TokenPath: "/etc/opa/auth/token",
						},
					},
				},
				DecisionLogReporting: configv2alpha2.DecisionLogReporting{
					UploadSizeLimitBytes: 1048576,
					MinDelaySeconds:      1,
					MaxDelaySeconds:      30,
				},
			},
			customConfig: nil,
			opaConfig: map[string]interface{}{
				"caching": map[string]interface{}{
					"inter_query_builtin_cache": map[string]interface{}{
						"max_size_bytes": 10000000,
					},
				},
			},
			expectedCMContent: `services:
- name: s3
  url: https://minio/ocp
  credentials:
    s3_signing:
      environment_credentials: {}
- name: logs
  url: https://log-service/ocp
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s3
decision_logs:
  reporting:
    upload_size_limit_bytes: 1048576
    min_delay_seconds: 1
    max_delay_seconds: 30
  service: logs
  resource_path: /logs
caching:
  inter_query_builtin_cache:
    max_size_bytes: 10000000
`,
		}),
	)
})

// Test that controller-level OPAConfig overlays the generated config but can be
// overridden by legacy customConfig and new opaConfig fields.
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMap controller opaConfig precedence", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		projectConfig     map[string]interface{}
		customConfig      map[string]interface{}
		opaConfig         map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			test.projectConfig,
			test.customConfig,
			test.opaConfig,
			logr.Discard())

		gomega.Expect(err).To(gomega.BeNil())

		var actualMap, expectedMap map[string]interface{}
		actualYAML := cm.Data["opa-conf.yaml"]
		expectedYAML := test.expectedCMContent

		err = yaml.Unmarshal([]byte(actualYAML), &actualMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal actual YAML")

		err = yaml.Unmarshal([]byte(expectedYAML), &expectedMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal expected YAML")

		gomega.Expect(actualMap).To(gomega.Equal(expectedMap))
	},
		ginkgo.Entry("controller config overrides generated labels and customConfig can override it", test{
			opaDefaultConfig: configv2alpha2.OPAConfig{
				DecisionLogs: configv2alpha2.DecisionLog{
					RequestContext: configv2alpha2.RequestContext{
						HTTP: configv2alpha2.HTTP{
							Headers: strings.Split("header1,header2", ","),
						},
					},
				},
			},
			opaconf: ocp.OPAConfig{
				BundleResource: "bundles/system/bundle.tar.gz",
				BundleService: &ocp.OPAServiceConfig{
					Name: "s3",
					URL:  "https://minio/ocp",
					Credentials: &ocp.ServiceCredentials{
						S3: &ocp.S3Signing{
							S3EnvironmentCredentials: map[string]ocp.EmptyStruct{},
						},
					},
				},
				LogService: &ocp.OPAServiceConfig{
					Name: "logs",
					URL:  "https://log-service/ocp",
					Credentials: &ocp.ServiceCredentials{
						Bearer: &ocp.Bearer{
							TokenPath: "/etc/opa/auth/token",
						},
					},
				},
				DecisionLogReporting: configv2alpha2.DecisionLogReporting{
					UploadSizeLimitBytes: 1048576,
					MinDelaySeconds:      1,
					MaxDelaySeconds:      30,
				},
			},
			projectConfig: map[string]interface{}{
				"labels": map[string]interface{}{
					"namespace": "project-ns",
					"role":      "project",
				},
			},
			customConfig: map[string]interface{}{
				"labels": map[string]interface{}{
					"namespace": "custom-ns",
				},
			},
			expectedCMContent: `services:
- name: s3
  url: https://minio/ocp
  credentials:
    s3_signing:
      environment_credentials: {}
- name: logs
  url: https://log-service/ocp
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s3
decision_logs:
  request_context:
    http:
      headers:
      - header1
      - header2
  reporting:
    upload_size_limit_bytes: 1048576
    min_delay_seconds: 1
    max_delay_seconds: 30
  service: logs
  resource_path: /logs
labels:
  namespace: custom-ns
  role: project
`,
		}),
	)
})
