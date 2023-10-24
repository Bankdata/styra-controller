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
	"strconv"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = ginkgo.Describe("CreateInvitation", func() {

	type test struct {
		email                   bool
		name                    string
		responseCode            int
		responseBody            string
		createInvitationRequest *styra.CreateInvitationRequest
		expectStyraErr          bool
	}

	ginkgo.DescribeTable("CreateInvitation", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			var b bytes.Buffer
			gomega.Expect(json.NewEncoder(&b).Encode(test.createInvitationRequest)).To(gomega.Succeed())
			gomega.Expect(bs).To(gomega.Equal(b.Bytes()))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/invitations?email=" +
				strconv.FormatBool(test.email)))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateInvitation(context.Background(), test.email, test.name)
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
				  "url": "url"
				}
			}`,
			createInvitationRequest: &styra.CreateInvitationRequest{
				UserID: "name",
			},
		}),

		ginkgo.Entry("styra http error", test{
			name: "name",
			createInvitationRequest: &styra.CreateInvitationRequest{
				UserID: "name",
			},
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
