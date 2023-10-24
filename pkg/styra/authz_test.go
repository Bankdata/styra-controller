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
	"fmt"
	"io"
	"net/http"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("ListRoleBindingsV2", func() {

	type test struct {
		responseCode             int
		responseBody             string
		listRoleBindingsV2Params *styra.ListRoleBindingsV2Params
		expectedResponse         *styra.ListRoleBindingsV2Response
		expectStyraErr           bool
	}

	ginkgo.DescribeTable("ListRoleBindingsV2", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.String()).To(gomega.Equal(fmt.Sprintf(
				"http://test.com/v2/authz/rolebindings?resource_id=%s&resource_kind=%s&role_id=%s",
				test.listRoleBindingsV2Params.ResourceID,
				test.listRoleBindingsV2Params.ResourceKind,
				test.listRoleBindingsV2Params.RoleID,
			)))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.ListRoleBindingsV2(context.Background(), test.listRoleBindingsV2Params)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.Rolebindings).To(gomega.Equal(test.expectedResponse.Rolebindings))
		}
	},

		ginkgo.Entry("something", test{
			responseCode: http.StatusOK,
			responseBody: `{
						"request_id": "string",
						"rolebindings": [
						  {
							"id": "rolbindingId",
							"subjects": [
							  {
								"id": "subjectId",
								"kind": "subjectKind"
							  }
							]
						  }
						]
			  		}`,
			listRoleBindingsV2Params: &styra.ListRoleBindingsV2Params{
				ResourceKind: "ResourceKind",
				ResourceID:   "ResourceID",
				RoleID:       "RoleID",
			},
			expectedResponse: &styra.ListRoleBindingsV2Response{
				Rolebindings: []*styra.RoleBindingConfig{
					{
						ID: "rolbindingId",
						Subjects: []*styra.Subject{{
							ID:   "subjectId",
							Kind: "subjectKind",
						}},
					}},
				StatusCode: http.StatusOK,
			},
		}),

		ginkgo.Entry("styra http error", test{
			listRoleBindingsV2Params: &styra.ListRoleBindingsV2Params{
				ResourceKind: "ResourceKind",
				ResourceID:   "ResourceID",
				RoleID:       "RoleID",
			},
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("CreateRolebinding", func() {

	type test struct {
		responseCode             int
		responseBody             string
		createRoleBindingRequest *styra.CreateRoleBindingRequest
		expectedResponse         *styra.CreateRoleBindingResponse
		expectStyraErr           bool
	}

	ginkgo.DescribeTable("CreateRolebinding", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.createRoleBindingRequest)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))

			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v2/authz/rolebindings"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateRoleBinding(context.Background(), test.createRoleBindingRequest)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
			gomega.Expect(res.Rolebinding).To(gomega.Equal(test.expectedResponse.Rolebinding))
		}
	},

		ginkgo.Entry("something", test{
			responseCode: http.StatusOK,
			responseBody: `{
				"rolebinding": {
					"id":"rolebindingId",
					"resource_filter": {
						"id": "resourceFilterId",
						"kind": "resourceFilterKind"
					},
					"role_id": "roleId",
					"subjects": [
						{
							"id": "subjectId",
							"kind": "subjectKind"
						}
					]
				}
			}`,
			createRoleBindingRequest: &styra.CreateRoleBindingRequest{
				ResourceFilter: &styra.ResourceFilter{
					ID:   "resourceFilterId",
					Kind: "resourceFilterKind",
				},
				RoleID: "roleId",
				Subjects: []*styra.Subject{
					{
						ID:   "subjectId",
						Kind: styra.SubjectKind("subjectKind"),
					},
				},
			},
			expectedResponse: &styra.CreateRoleBindingResponse{
				Rolebinding: &styra.RoleBindingConfig{
					ID:     "rolebindingId",
					RoleID: "roleId",
					Subjects: []*styra.Subject{
						{
							ID:   "subjectId",
							Kind: "subjectKind",
						},
					},
				},
				StatusCode: http.StatusOK,
				//What to do with body?
			},
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("UpdateRoleBindingSubjects", func() {

	type test struct {
		id                               string
		responseCode                     int
		responseBody                     string
		updateRoleBindingSubjectsRequest *styra.UpdateRoleBindingSubjectsRequest
		expectedBody                     []byte
		expectStyraErr                   bool
	}

	ginkgo.DescribeTable("UpdateRoleBindingSubjects", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.updateRoleBindingSubjectsRequest)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v2/authz/rolebindings/" + test.id + "/subjects"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpdateRoleBindingSubjects(context.Background(), test.id, test.updateRoleBindingSubjectsRequest)
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
			id:           "rolebindingId",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			updateRoleBindingSubjectsRequest: &styra.UpdateRoleBindingSubjectsRequest{
				Subjects: []*styra.Subject{
					{
						ID:   "subjectId",
						Kind: "subjectKind",
					},
				},
			},
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusNotFound,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("DeleteRoleBindingV2", func() {

	type test struct {
		id             string
		responseCode   int
		responseBody   string
		expectStyraErr bool
	}

	ginkgo.DescribeTable("DeleteRoleBindingV2",
		func(test test) {
			c := newTestClient(func(r *http.Request) *http.Response {
				gomega.Expect(r.Method).To(gomega.Equal(http.MethodDelete))
				gomega.Expect(r.URL.String()).To(gomega.Equal(fmt.Sprintf("http://test.com/v2/authz/rolebindings/%s", test.id)))

				return &http.Response{
					Header:     make(http.Header),
					StatusCode: test.responseCode,
					Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
				}
			})
			res, err := c.DeleteRoleBindingV2(context.Background(), test.id)
			if test.expectStyraErr {
				gomega.Expect(res).To(gomega.BeNil())
				target := &styra.HTTPError{}
				gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
			} else {
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
				gomega.Expect(res.Body).To(gomega.Equal([]byte(test.responseBody)))
			}
		},

		ginkgo.Entry("success", test{
			id:           "rolebindingId",
			responseCode: http.StatusOK,
			responseBody: `{"request_id":"test"}`,
		}),

		ginkgo.Entry("not found", test{
			id:           "rolebindingId",
			responseCode: http.StatusNotFound,
			responseBody: `{"request_id":"test"}`,
		}),

		ginkgo.Entry("styra http error", test{
			id:             "rolebindingId",
			responseCode:   http.StatusInternalServerError,
			responseBody:   `{"request_id":"test"}`,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.DescribeTable("SubjectsAreEqual",
	func(as []*styra.Subject, bs []*styra.Subject, expected bool) {
		gomega.Ω(styra.SubjectsAreEqual(as, bs)).To(gomega.Equal(expected))
	},

	ginkgo.Entry("returns false if not same length",
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		nil,
		false),

	ginkgo.Entry("returns true if subjects are equal only user",
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		true),

	ginkgo.Entry("returns true if subjects are equal",
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		true),

	ginkgo.Entry("returns false if ClaimConfig value changes",
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName1"},
			},
		},
		false),

	ginkgo.Entry("returns false if ClaimConfig key changes",
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups1", Value: "ADGName"},
			},
		},
		false),

	ginkgo.Entry("returns false if ClaimConfig IdentityProvider changes",
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest1", Key: "groups", Value: "ADGName"},
			},
		},
		false),

	ginkgo.Entry("returns true despite order of subjects",
		[]*styra.Subject{
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
			{ID: "test@test.dk", Kind: "user"},
		},
		[]*styra.Subject{
			{ID: "test@test.dk", Kind: "user"},
			{
				Kind:        "claim",
				ClaimConfig: &styra.ClaimConfig{IdentityProvider: "ADtest", Key: "groups", Value: "ADGName"},
			},
		},
		true),
)
