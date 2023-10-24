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

package styra_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("GetOPAConfig", func() {

	type test struct {
		responseBody    string
		responseCode    int
		expectedOPAConf styra.OPAConfig
		expectStyraErr  bool
	}

	ginkgo.DescribeTable("GetOPAConfig", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/systems/test_id/assets/opa-config"))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		opaconf, err := c.GetOPAConfig(context.Background(), "test_id")
		if test.expectStyraErr {
			gomega.Expect(opaconf).To(gomega.Equal(styra.OPAConfig{}))
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(opaconf).To(gomega.Equal(test.expectedOPAConf))
		}
	},

		ginkgo.Entry("success", test{
			responseBody: `
discovery:
  name: discovery-123
  prefix: prefix-123
  service: service-123
labels:
  system-id: system-123
  system-type: custom-123
services:
  - credentials:
      bearer:
        token: opa-token-123
    url: styra-url-123
  - credentials:
      bearer:
        token: opa-token-1234
    url: styra-url-1234`,
			expectedOPAConf: styra.OPAConfig{
				HostURL:    "styra-url-123",
				Token:      "opa-token-123",
				SystemID:   "system-123",
				SystemType: "custom-123",
			},
			responseCode: http.StatusOK,
		}),
		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
