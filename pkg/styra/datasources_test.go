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

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("GetDatasource", func() {
	type test struct {
		datasourceID             string
		responseCode             int
		responseBody             string
		expectedDatasourceConfig *styra.DatasourceConfig
		expectStyraErr           bool
	}

	ginkgo.DescribeTable("GetDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetDatasource(context.Background(), test.datasourceID)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.DatasourceConfig).To(gomega.Equal(test.expectedDatasourceConfig))
		}

	},

		ginkgo.Entry("happy path", test{
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

		ginkgo.Entry("unexpected status code", test{
			datasourceID:   "test",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("UpsertDatasource", func() {

	type test struct {
		datasourceID            string
		responseCode            int
		responseBody            string
		upsertDatasourceRequest *styra.UpsertDatasourceRequest
		expectedBody            []byte
		expectStyraErr          bool
	}

	ginkgo.DescribeTable("UpsertDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.upsertDatasourceRequest)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPut))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpsertDatasource(context.Background(), test.datasourceID, test.upsertDatasourceRequest)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.Body).To(gomega.Equal(test.expectedBody))
		}
	},

		ginkgo.Entry("something", test{
			datasourceID: "datasourceID",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			upsertDatasourceRequest: &styra.UpsertDatasourceRequest{
				Category: "datasourceCategory",
			},
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("styra http error", test{
			datasourceID:   "datasourceID",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("DeleteDatasource", func() {

	type test struct {
		datasourceID   string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	ginkgo.DescribeTable("DeleteDatasource", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodDelete))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/datasources/" + test.datasourceID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeleteDatasource(context.Background(), test.datasourceID)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.Body).To(gomega.Equal(test.expectedBody))
		}
	},

		ginkgo.Entry("something", test{
			datasourceID: "datasourceID",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("styra http error", test{
			datasourceID:   "datasourceID",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
