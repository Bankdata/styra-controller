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
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/patrickmn/go-cache"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("GetUser", func() {

	type test struct {
		name           string
		responseCode   int
		responseBody   string
		expectStyraErr bool
	}

	ginkgo.DescribeTable("GetUser", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/users/" + test.name))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.GetUser(context.Background(), test.name)
		if test.expectStyraErr {
			gomega.Expect(res).To(gomega.BeNil())
			target := &styra.HTTPError{}
			gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
		} else {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(res.StatusCode).To(gomega.Equal(test.responseCode))
		}
	},

		ginkgo.Entry("something", test{
			name:         "name",
			responseCode: http.StatusOK,
			responseBody: `{
				"request_id": "id",
				"result": {
				  "enabled": false,
					"id": "name"
				}
			}`,
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})

var _ = ginkgo.Describe("GetUsers", func() {
	type test struct {
		responseCode   int
		responseBody   string
		expectStyraErr bool
	}

	ginkgo.DescribeTable("GetUsers",
		func(test test) {
			c := newTestClientWithCache(func(r *http.Request) *http.Response {
				bs, err := io.ReadAll(r.Body)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(bs).To(gomega.Equal([]byte("")))
				gomega.Expect(r.Method).To(gomega.Equal(http.MethodGet))
				gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/users"))

				return &http.Response{
					Header:     make(http.Header),
					StatusCode: test.responseCode,
					Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
				}
			}, cache.New(1*time.Hour, 10*time.Minute))

			// Call GetUsers
			res, _, err := c.GetUsers(context.Background())
			if test.expectStyraErr {
				gomega.Expect(res).To(gomega.BeNil())
				target := &styra.HTTPError{}
				gomega.Expect(errors.As(err, &target)).To(gomega.BeTrue())
			} else {
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(res.Users).ToNot(gomega.BeNil())
				gomega.Expect(res.Users[0].ID).To(gomega.Equal("user1"))
				gomega.Expect(res.Users[0].Enabled).To(gomega.BeTrue())
				gomega.Expect(res.Users[1].ID).To(gomega.Equal("user2"))
				gomega.Expect(res.Users[1].Enabled).To(gomega.BeFalse())
			}
		},

		ginkgo.Entry("successful response", test{
			responseCode: http.StatusOK,
			responseBody: `{
                "result": [
                    {"enabled": true, "id": "user1"},
                    {"enabled": false, "id": "user2"}
                ]
            }`,
		}),

		ginkgo.Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
