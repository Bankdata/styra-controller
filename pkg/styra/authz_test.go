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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = Describe("ListRoleBindingsV2", func() {

	type test struct {
		responseCode             int
		responseBody             string
		listRoleBindingsV2Params *styra.ListRoleBindingsV2Params
		expectedResponse         *styra.ListRoleBindingsV2Response
		expectStyraErr           bool
	}

	DescribeTable("ListRoleBindingsV2", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("")))
			Expect(r.Method).To(Equal(http.MethodGet))
			Expect(r.URL.String()).To(Equal(fmt.Sprintf(
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
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.Rolebindings).To(Equal(test.expectedResponse.Rolebindings))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
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

var _ = Describe("CreateRolebinding", func() {

	type test struct {
		responseCode             int
		responseBody             string
		createRoleBindingRequest *styra.CreateRoleBindingRequest
		expectedResponse         *styra.CreateRoleBindingResponse
		expectStyraErr           bool
	}

	DescribeTable("CreateRolebinding", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())

			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.createRoleBindingRequest)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))

			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.URL.String()).To(Equal("http://test.com/v2/authz/rolebindings"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateRoleBinding(context.Background(), test.createRoleBindingRequest)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
			Expect(res.Rolebinding).To(Equal(test.expectedResponse.Rolebinding))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("UpdateRoleBindingSubjects", func() {

	type test struct {
		id                               string
		responseCode                     int
		responseBody                     string
		updateRoleBindingSubjectsRequest *styra.UpdateRoleBindingSubjectsRequest
		expectedBody                     []byte
		expectStyraErr                   bool
	}

	DescribeTable("UpdateRoleBindingSubjects", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.updateRoleBindingSubjectsRequest)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))
			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.URL.String()).To(Equal("http://test.com/v2/authz/rolebindings/" + test.id + "/subjects"))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.UpdateRoleBindingSubjects(context.Background(), test.id, test.updateRoleBindingSubjectsRequest)
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

		Entry("styra http error", test{
			responseCode:   http.StatusNotFound,
			expectStyraErr: true,
		}),
	)
})

var _ = Describe("DeleteRoleBindingV2", func() {

	type test struct {
		id             string
		responseCode   int
		responseBody   string
		expectStyraErr bool
	}

	DescribeTable("DeleteRoleBindingV2",
		func(test test) {
			c := newTestClient(func(r *http.Request) *http.Response {
				Expect(r.Method).To(Equal(http.MethodDelete))
				Expect(r.URL.String()).To(Equal(fmt.Sprintf("http://test.com/v2/authz/rolebindings/%s", test.id)))

				return &http.Response{
					Header:     make(http.Header),
					StatusCode: test.responseCode,
					Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
				}
			})
			res, err := c.DeleteRoleBindingV2(context.Background(), test.id)
			if test.expectStyraErr {
				Expect(res).To(BeNil())
				target := &styra.HTTPError{}
				Expect(errors.As(err, &target)).To(BeTrue())
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(res.StatusCode).To(Equal(test.responseCode))
				Expect(res.Body).To(Equal([]byte(test.responseBody)))
			}
		},

		Entry("success", test{
			id:           "rolebindingId",
			responseCode: http.StatusOK,
			responseBody: `{"request_id":"test"}`,
		}),

		Entry("not found", test{
			id:           "rolebindingId",
			responseCode: http.StatusNotFound,
			responseBody: `{"request_id":"test"}`,
		}),

		Entry("styra http error", test{
			id:             "rolebindingId",
			responseCode:   http.StatusInternalServerError,
			responseBody:   `{"request_id":"test"}`,
			expectStyraErr: true,
		}),
	)
})

var _ = DescribeTable("SubjectsAreEqual",
	func(as []*styra.Subject, bs []*styra.Subject, expected bool) {
		Ω(styra.SubjectsAreEqual(as, bs)).To(Equal(expected))
	},

	Entry("returns false if not same length",
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		nil,
		false),

	Entry("returns true if subjects are equal only user",
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		[]*styra.Subject{{ID: "test@test.dk", Kind: "user"}},
		true),

	Entry("returns true if subjects are equal",
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

	Entry("returns false if ClaimConfig value changes",
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

	Entry("returns false if ClaimConfig key changes",
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

	Entry("returns false if ClaimConfig IdentityProvider changes",
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

	Entry("returns true despite order of subjects",
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
