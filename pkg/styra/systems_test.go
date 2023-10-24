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

var _ = ginkgo.Describe("GetSystem", func() {
	type test struct {
		systemID       string
		responseCode   int
		responseBody   string
		expected200    *styra.SystemConfig
		expectStyraErr bool
	}

	ginkgo.DescribeTable("GetSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/systems/" + test.systemID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetSystem(context.Background(), test.systemID)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.SystemConfig).To(gomega.Equal(test.expected200))
		}
	},

		ginkgo.Entry("something", test{
			systemID:     "systemID",
			responseCode: http.StatusOK,
			responseBody: `{
				"result": {
					"decision_mappings": {
						"dm1": {
							"allowed": {
								"expected": true,
								"negated": false,
								"path": "path"
							},
							"columns": [
							{
								"key": "key",
								"path": "path",
								"type": "type"
							}
							],
							"reason": {
								"path": "path"
							}
						}
					}
				}
			}`,
			expected200: &styra.SystemConfig{
				DecisionMappings: map[string]styra.DecisionMapping{
					"dm1": {
						Allowed: &styra.DecisionMappingAllowed{
							Expected: true,
							Negated:  false,
							Path:     "path",
						},
						Columns: []styra.DecisionMappingColumn{
							{
								Key:  "key",
								Path: "path",
								Type: "type",
							},
						},
						Reason: &styra.DecisionMappingReason{
							Path: "path",
						},
					},
				},
			},
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("UpdateSystem", func() {

	type test struct {
		responseCode         int
		responseBody         string
		id                   string
		request              *styra.UpdateSystemRequest
		expectedSystemConfig *styra.SystemConfig
		expectStyraErr       bool
	}

	ginkgo.DescribeTable("UpdateSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.request)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))

			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPut))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/systems/" + test.id))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpdateSystem(context.Background(), test.id, test.request)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.SystemConfig).To(gomega.Equal(test.expectedSystemConfig))
		}
	},

		ginkgo.Entry("something", test{
			responseCode: http.StatusOK,
			responseBody: `{
					"result": {
						"name": "mysystem",
						"read_only": true,
						"type": "systemtype",
						"id": "systemid"
					}
				}`,
			id: "systemid",
			request: &styra.UpdateSystemRequest{
				SystemConfig: &styra.SystemConfig{
					Name:     "mysystem",
					Type:     "systemtype",
					ReadOnly: true,
				},
			},
			expectedSystemConfig: &styra.SystemConfig{
				Name:     "mysystem",
				ReadOnly: true,
				Type:     "systemtype",
				ID:       "systemid",
			},
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("CreateSystem", func() {

	type test struct {
		responseCode         int
		responseBody         string
		request              *styra.CreateSystemRequest
		expectedSystemConfig *styra.SystemConfig
		expectStyraErr       bool
	}

	ginkgo.DescribeTable("CreateSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.request)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))

			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/systems"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateSystem(context.Background(), test.request)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.SystemConfig).To(gomega.Equal(test.expectedSystemConfig))
		}
	},

		ginkgo.Entry("something", test{
			responseCode: http.StatusOK,
			responseBody: `{
					"result": {
						"name": "mysystem",
						"read_only": true,
						"type": "systemtype",
						"id": "systemid"
					}
				}`,
			request: &styra.CreateSystemRequest{
				SystemConfig: &styra.SystemConfig{
					Name:     "mysystem",
					Type:     "systemtype",
					ReadOnly: true,
				},
			},
			expectedSystemConfig: &styra.SystemConfig{
				Name:     "mysystem",
				ReadOnly: true,
				Type:     "systemtype",
				ID:       "systemid",
			},
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("DeleteSystem", func() {

	type test struct {
		systemID       string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	ginkgo.DescribeTable("DeleteSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodDelete))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/systems/" + test.systemID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeleteSystem(context.Background(), test.systemID)
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
			systemID:     "systemId",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("DecisionMappingsEquals", func() {
	ginkgo.It("returns true if both are nil", func() {
		gomega.Expect(styra.DecisionMappingsEquals(nil, nil)).To(gomega.BeTrue())
	})

	ginkgo.It("returns false if one is nil", func() {
		expected := styra.DecisionMappingsEquals(map[string]styra.DecisionMapping{}, nil)
		gomega.Expect(expected).To(gomega.BeFalse())
		expected = styra.DecisionMappingsEquals(nil, map[string]styra.DecisionMapping{})
		gomega.Expect(expected).To(gomega.BeFalse())
	})

	ginkgo.It("treats columns slice as a map sorted on their Key", func() {
		expected := styra.DecisionMappingsEquals(
			map[string]styra.DecisionMapping{
				"test": {Columns: []styra.DecisionMappingColumn{
					{Key: "test1"},
					{Key: "test2"},
				}},
			},
			map[string]styra.DecisionMapping{
				"test": {Columns: []styra.DecisionMappingColumn{
					{Key: "test2"},
					{Key: "test1"},
				}},
			},
		)
		gomega.Expect(expected).To(gomega.BeTrue())
	})
})
