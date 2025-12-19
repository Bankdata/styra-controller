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
	"github.com/bankdata/styra-controller/pkg/styra"
	"gopkg.in/yaml.v2"
)

var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMap", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		slpURL            string
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMap(test.opaconf, test.slpURL, test.opaDefaultConfig, nil)

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
				PersistBundle: true, // test that PersistBundle is ignored for non-OCP
			},
			opaconf: styra.OPAConfig{
				HostURL:    "styra-host-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-id-123",
				SystemType: "system-type-123",
			},
			slpURL: "slp-url-123",
			expectedCMContent: `services:
- name: styra
  url: slp-url-123
labels:
  system-id: system-id-123
  system-type: system-type-123
discovery:
  name: discovery
  service: styra
decision_logs:
  request_context:
    http:
      headers:
      - header1
      - header2
`,
		}),
	)
})

var _ = ginkgo.Describe("OPAConfToK8sSLPConfigMap", func() {

	type test struct {
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OPAConfToK8sSLPConfigMap(test.opaconf)

		gomega.Expect(err).To(gomega.BeNil())

		var actualMap, expectedMap map[string]interface{}
		actualYAML := cm.Data["slp.yaml"]
		expectedYAML := test.expectedCMContent

		err = yaml.Unmarshal([]byte(actualYAML), &actualMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal actual YAML")

		err = yaml.Unmarshal([]byte(expectedYAML), &expectedMap)
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Failed to unmarshal expected YAML")

		gomega.Expect(actualMap).To(gomega.Equal(expectedMap))
	},

		ginkgo.Entry("success", test{
			opaconf: styra.OPAConfig{
				HostURL:    "styra-host-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-id-123",
				SystemType: "system-type-123",
			},
			expectedCMContent: `services:
- name: styra
  url: styra-host-url-123
  credentials:
    bearer:
      token_path: /etc/slp/auth/token
labels:
  system-id: system-id-123
  system-type: system-type-123
discovery:
  name: discovery
  resource: /systems/system-id-123/discovery
  service: styra
`,
		}),
	)
})

var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMapNoSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMapNoSLP", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapNoSLP(test.opaconf, test.opaDefaultConfig, nil)

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
				PersistBundle: true, // test that PersistBundle is ignored for non-OCP
			},
			opaconf: styra.OPAConfig{
				HostURL:    "styra-host-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-id-123",
				SystemType: "system-type-123",
			},
			expectedCMContent: `services:
- name: styra
  url: styra-host-url-123
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
- name: styra-bundles
  url: styra-host-url-123/bundles
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
labels:
  system-id: system-id-123
  system-type: system-type-123
discovery:
  name: discovery
  prefix: /systems/system-id-123
  service: styra
decision_logs:
  request_context:
    http:
      headers:
      - header1
      - header2
`,
		}),
	)
})

var _ = ginkgo.Describe("OPACustomConfToK8sWithSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		customConfig      map[string]interface{}
		slpURL            string
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPACustomConfToK8sWithSLP", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMap(test.opaconf, test.slpURL, test.opaDefaultConfig, test.customConfig)

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
				PersistBundle: true, // test that PersistBundle is ignored for non-OCP
			},
			opaconf: styra.OPAConfig{
				HostURL:    "styra-host-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-id-123",
				SystemType: "system-type-123",
			},
			customConfig: map[string]interface{}{
				"distributed_tracing": map[string]interface{}{
					"type":    "grpc",
					"address": "localhost:1234",
				},
			},
			slpURL: "slp-url-123",
			expectedCMContent: `services:
- name: styra
  url: slp-url-123
labels:
  system-id: system-id-123
  system-type: system-type-123
discovery:
  name: discovery
  service: styra
decision_logs:
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

var _ = ginkgo.Describe("OPACustomConfToK8sNoSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPACustomConfToK8sNoSLP", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapNoSLP(test.opaconf, test.opaDefaultConfig, test.customConfig)

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
				PersistBundle: true, // test that PersistBundle is ignored for non-OCP
			},
			opaconf: styra.OPAConfig{
				HostURL:    "styra-host-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-id-123",
				SystemType: "system-type-123",
			},
			customConfig: map[string]interface{}{
				"distributed_tracing": map[string]interface{}{
					"type":    "grpc",
					"address": "localhost:1234",
				},
			},
			expectedCMContent: `services:
- name: styra
  url: styra-host-url-123
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
- name: styra-bundles
  url: styra-host-url-123/bundles
  credentials:
    bearer:
      token_path: /etc/opa/auth/token
labels:
  system-id: system-id-123
  system-type: system-type-123
discovery:
  name: discovery
  prefix: /systems/system-id-123
  service: styra
decision_logs:
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

// Test PersistBundle config
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMapforOCP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMapforOCP", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			test.customConfig,
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
var _ = ginkgo.Describe("OPAConfToK8sOPAConfigMapforOCP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OPAConfToK8sOPAConfigMapforOCP", func(test test) {
		cm, err := k8sconv.OPAConfToK8sOPAConfigMapforOCP(
			test.opaconf,
			test.opaDefaultConfig,
			test.customConfig,
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
