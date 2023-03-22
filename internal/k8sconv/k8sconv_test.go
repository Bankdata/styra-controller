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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/internal/k8sconv"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = Describe("OpaConfToK8sOPAConfigMap", func() {

	type test struct {
		opaconf           styra.OPAConfig
		slpURL            string
		expectedCMContent string
	}

	DescribeTable("OpaConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMap(test.opaconf, test.slpURL)

		Expect(err).To(BeNil())

		Expect(cm.Data["opa-conf.yaml"]).To(Equal(test.expectedCMContent))
	},

		Entry("success", test{
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
`,
		}),
	)
})

var _ = Describe("OpaConfToK8sSLPConfigMap", func() {

	type test struct {
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	DescribeTable("OpaConfToK8sOPAConfigMap", func(test test) {
		cm, err := k8sconv.OpaConfToK8sSLPConfigMap(test.opaconf)

		Expect(err).To(BeNil())

		Expect(cm.Data["slp.yaml"]).To(Equal(test.expectedCMContent))
	},

		Entry("success", test{
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

var _ = Describe("OpaConfToK8sOPAConfigMapNoSLP", func() {

	type test struct {
		opaconf           styra.OPAConfig
		expectedCMContent string
	}

	DescribeTable("OpaConfToK8sOPAConfigMapNoSLP", func(test test) {
		cm, err := k8sconv.OpaConfToK8sOPAConfigMapNoSLP(test.opaconf)

		Expect(err).To(BeNil())

		Expect(cm.Data["opa-conf.yaml"]).To(Equal(test.expectedCMContent))
	},

		Entry("success", test{
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
`,
		}),
	)
})
