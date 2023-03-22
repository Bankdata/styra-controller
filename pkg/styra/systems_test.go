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

var _ = Describe("GetSystem", func() {
	type test struct {
		systemID       string
		responseCode   int
		responseBody   string
		expected200    *styra.SystemConfig
		expectStyraErr bool
	}

	DescribeTable("GetSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("")))
			Expect(r.Method).To(Equal(http.MethodGet))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/systems/" + test.systemID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetSystem(context.Background(), test.systemID)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.SystemConfig).To(Equal(test.expected200))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("UpdateSystem", func() {

	type test struct {
		responseCode         int
		responseBody         string
		id                   string
		request              *styra.UpdateSystemRequest
		expectedSystemConfig *styra.SystemConfig
		expectStyraErr       bool
	}

	DescribeTable("UpdateSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())

			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.request)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))

			Expect(r.Method).To(Equal(http.MethodPut))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/systems/" + test.id))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpdateSystem(context.Background(), test.id, test.request)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.SystemConfig).To(Equal(test.expectedSystemConfig))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("CreateSystem", func() {

	type test struct {
		responseCode         int
		responseBody         string
		request              *styra.CreateSystemRequest
		expectedSystemConfig *styra.SystemConfig
		expectStyraErr       bool
	}

	DescribeTable("CreateSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())

			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.request)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))

			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/systems"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateSystem(context.Background(), test.request)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.SystemConfig).To(Equal(test.expectedSystemConfig))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("DeleteSystem", func() {

	type test struct {
		systemID       string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	DescribeTable("DeleteSystem", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("")))
			Expect(r.Method).To(Equal(http.MethodDelete))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/systems/" + test.systemID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeleteSystem(context.Background(), test.systemID)
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
			systemID:     "systemId",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("DecisionMappingsEquals", func() {
	It("returns true if both are nil", func() {
		Expect(styra.DecisionMappingsEquals(nil, nil)).To(BeTrue())
	})

	It("returns false if one is nil", func() {
		expected := styra.DecisionMappingsEquals(map[string]styra.DecisionMapping{}, nil)
		Expect(expected).To(BeFalse())
		expected = styra.DecisionMappingsEquals(nil, map[string]styra.DecisionMapping{})
		Expect(expected).To(BeFalse())
	})

	It("treats columns slice as a map sorted on their Key", func() {
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
		Expect(expected).To(BeTrue())
	})
})
