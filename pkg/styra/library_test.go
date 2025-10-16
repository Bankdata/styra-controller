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

	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("GetLibrary", func() {
	type test struct {
		libraryID                     string
		responseCode                  int
		responseBody                  string
		expectedLibraryEntityExpanded *styra.LibraryEntityExpanded
		expectStyraErr                bool
	}

	ginkgo.DescribeTable("GetLibrary", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/libraries/" + test.libraryID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetLibrary(context.Background(), test.libraryID)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &httperror.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.Statuscode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.LibraryEntityExpanded).To(gomega.Equal(test.expectedLibraryEntityExpanded))
		}
	},

		ginkgo.Entry("happy path", test{
			libraryID:    "test1",
			responseCode: http.StatusOK,
			responseBody: `
			{
				"result": {
					"id": "test1",
					"description": "test2",
					"read_only": true,
					"source_control": {
						"use_workspace_settings": false,
						"origin": {
							"url": "https://github.com/test.git",
							"reference": "refs/heads/master",
							"commit": "",
							"credentials": "gitCreds",
							"path": ""
						},
						"library_origin": {
							"url": "https://github.com/test.git",
							"reference": "refs/heads/master",
							"commit": "",
							"credentials": "path/to/creds",
							"path": ""
						}
					},
					"policies": [],
					"datasources": []
				}
			}
	  	`,
			expectedLibraryEntityExpanded: &styra.LibraryEntityExpanded{
				DataSources: []styra.LibraryDatasourceConfig{},
				Description: "test2",
				ID:          "test1",
				ReadOnly:    true,
				SourceControl: &styra.LibrarySourceControlConfig{
					LibraryOrigin: &styra.LibraryGitRepoConfig{
						Commit:      "",
						Credentials: "path/to/creds",
						Path:        "",
						Reference:   "refs/heads/master",
						URL:         "https://github.com/test.git",
					},
				},
			},
		}),

		ginkgo.Entry("unexpected status code", test{
			libraryID:      "test",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("UpsertLibrary", func() {

	type test struct {
		libraryID            string
		upsertLibraryRequest *styra.UpsertLibraryRequest
		responseCode         int
		responseBody         string
		expectedBody         []byte
		expectStyraErr       bool
	}

	ginkgo.DescribeTable("UpsertLibrary", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			// First make sure the body is readable
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Then make sure the body is encoded correctly
			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.upsertLibraryRequest)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))

			// Then ensure the correct http request is made
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPut))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/libraries/" + test.libraryID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpsertLibrary(context.Background(), test.libraryID, test.upsertLibraryRequest)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &httperror.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.Body).To(gomega.Equal(test.expectedBody))
		}
	},

		ginkgo.Entry("happy", test{
			libraryID:    "test",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			upsertLibraryRequest: &styra.UpsertLibraryRequest{
				Description: "test2",
				ReadOnly:    true,
				SourceControl: &styra.LibrarySourceControlConfig{
					LibraryOrigin: &styra.LibraryGitRepoConfig{
						Commit:      "",
						Credentials: "path/to/creds",
						Path:        "",
						Reference:   "refs/heads/master",
						URL:         "https://github.com/test.git",
					},
				},
			},
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("sad", test{
			libraryID:      "test",
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
			target := &httperror.HTTPError{}
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
