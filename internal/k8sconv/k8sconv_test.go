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

package k8sconv_test

import (
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("OpaConfToK8sOPAConfigMap", func() {

	type test struct {
		opaDefaultConfig  configv2alpha2.OPAConfig
		opaconf           styra.OPAConfig
		slpURL            string
		expectedCMContent string
	}

	ginkgo.DescribeTable("OpaConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMap(test.opaconf, test.slpURL, test.opaDefaultConfig)

		gomega.Expect(err).To(gomega.BeNil())

		gomega.Expect(cm.Data["opa-conf.yaml"]).To(gomega.Equal(test.expectedCMContent))
	},

		ginkgo.Entry("success", test{
			opaDefaultConfig: configv2alpha2.OPAConfig{
				DecisionLogs: configv2alpha2.DecisionLog{
					RequestContextHTTPHeaders: strings.Split("header1,header2", ","),
				},
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

		gomega.Expect(cm.Data["slp.yaml"]).To(gomega.Equal(test.expectedCMContent))
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
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapNoSLP(test.opaconf, test.opaDefaultConfig)

		gomega.Expect(err).To(gomega.BeNil())

		gomega.Expect(cm.Data["opa-conf.yaml"]).To(gomega.Equal(test.expectedCMContent))
	},

		ginkgo.Entry("success", test{
			opaDefaultConfig: configv2alpha2.OPAConfig{
				DecisionLogs: configv2alpha2.DecisionLog{
					RequestContextHTTPHeaders: strings.Split("header1,header2", ","),
				},
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
