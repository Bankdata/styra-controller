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
	"encoding/json"
	"errors"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = Describe("GetDatasource", func() {
	type test struct {
		datasourceID             string
		responseCode             int
		responseBody             string
		expectedDatasourceConfig *styra.DatasourceConfig
		expectStyraErr           bool
	}

	DescribeTable("GetDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			Expect(r.Method).To(Equal(http.MethodGet))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetDatasource(context.Background(), test.datasourceID)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.DatasourceConfig).To(Equal(test.expectedDatasourceConfig))
		}

	},

		Entry("happy path", test{
			datasourceID: "test",
			responseCode: http.StatusOK,
			responseBody: `
{
      "result": {
          "category": "git/rego",
          "commit": "",
          "credentials": "libraries/global/test/git",
          "description": "",
          "enabled": true,
          "id": "global/test",
          "path": "",
          "reference": "refs/heads/master",
          "type": "pull",
          "url": "https://test.com/test.git"
      }
}
	  	`,
			expectedDatasourceConfig: &styra.DatasourceConfig{
				Category:    "git/rego",
				Credentials: "libraries/global/test/git",
				Enabled:     true,
				ID:          "global/test",
				Reference:   "refs/heads/master",
				Type:        "pull",
				URL:         "https://test.com/test.git",
			},
		}),

		Entry("unexpected status code", test{
			datasourceID:   "test",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("UpsertDatasource", func() {

	type test struct {
		datasourceID            string
		responseCode            int
		responseBody            string
		upsertDatasourceRequest *styra.UpsertDatasourceRequest
		expectedBody            []byte
		expectStyraErr          bool
	}

	DescribeTable("UpsertDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.upsertDatasourceRequest)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))
			Expect(r.Method).To(Equal(http.MethodPut))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpsertDatasource(context.Background(), test.datasourceID, test.upsertDatasourceRequest)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.Body).To(Equal(test.expectedBody))
		}
	},

		Entry("something", test{
			datasourceID: "datasourceID",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			upsertDatasourceRequest: &styra.UpsertDatasourceRequest{
				Category: "datasourceCategory",
			},
			expectedBody: []byte(`expected response from styra api`)},
		),

		Entry("styra http error", test{
			datasourceID:   "datasourceID",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("DeleteDatasource", func() {

	type test struct {
		datasourceID   string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	DescribeTable("DeleteDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("")))
			Expect(r.Method).To(Equal(http.MethodDelete))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeleteDatasource(context.Background(), test.datasourceID)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.Body).To(Equal(test.expectedBody))
		}
	},

		Entry("something", test{
			datasourceID: "datasourceID",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		Entry("styra http error", test{
			datasourceID:   "datasourceID",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
