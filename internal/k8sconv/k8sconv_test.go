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

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/pkg/ocp"
	"github.com/bankdata/styra-controller/pkg/styra"
	"gopkg.in/yaml.v2"
)

var _ = ginkgo.Describe("OpaConfToK8sOPAConfigMap", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		slpURL            string
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMap(test.opaconf, test.slpURL, test.opaDefaultConfig, nil)

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

var _ = ginkgo.Describe("OpaConfToK8sSLPConfigMap", func() {

	type test struct {
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OpaConfToK8sSLPConfigMap(test.opaconf)

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

var _ = ginkgo.Describe("OpaConfToK8sOPAConfigMapNoSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMapNoSLP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapNoSLP(test.opaconf, test.opaDefaultConfig, nil)

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

var _ = ginkgo.Describe("OpaCustomConfToK8sWithSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		customConfig      map[string]interface{}
		slpURL            string
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaCustomConfToK8sWithSLP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMap(test.opaconf, test.slpURL, test.opaDefaultConfig, test.customConfig)

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

var _ = ginkgo.Describe("OpaCustomConfToK8sNoSLP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaCustomConfToK8sNoSLP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapNoSLP(test.opaconf, test.opaDefaultConfig, test.customConfig)

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
var _ = ginkgo.Describe("OpaConfToK8sOPAConfigMapforOCP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMapforOCP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapforOCP(test.opaconf, test.opaDefaultConfig, test.customConfig)

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
				ServiceURL:     "https://minio/ocp",
				ServiceName:    "s3",
				BundleService:  "s3",
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
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s3
    persist: true
    test: 123
persistence_directory: /opa-bundles
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

// Test without PersistBundle config
// And custom config overriding opaconf
var _ = ginkgo.Describe("OpaConfToK8sOPAConfigMapforOCP", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           ocp.OPAConfig
		customConfig      map[string]interface{}
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMapforOCP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapforOCP(test.opaconf, test.opaDefaultConfig, test.customConfig)

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
				BundleResource: "bundles/system/bundle.tar.gz",
				ServiceURL:     "https://minio/ocp",
				ServiceName:    "s3",
				BundleService:  "s3",
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
bundles:
  authz:
    resource: bundles/system/bundle.tar.gz
    service: s4
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
