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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = Describe("CreateInvitation", func() {

	type test struct {
		email                   bool
		name                    string
		responseCode            int
		responseBody            string
		createInvitationRequest *styra.CreateInvitationRequest
		expectStyraErr          bool
	}

	DescribeTable("CreateInvitation", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.createInvitationRequest)).To(Succeed())
			Expect(bs).To(Equal(b.Bytes()))
			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/invitations?email=" + strconv.FormatBool(test.email)))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateInvitation(context.Background(), test.email, test.name)
		if test.expectStyraErr {
			Expect(res).To(BeNil())
			target := &styra.HTTPError{}
			Expect(errors.As(err, &target)).To(BeTrue())
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(test.responseCode))
		}
	},

		Entry("something", test{
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

		Entry("styra http error", test{
			name: "name",
			createInvitationRequest: &styra.CreateInvitationRequest{
				UserID: "name",
			},
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
