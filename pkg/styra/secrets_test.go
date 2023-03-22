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

var _ = Describe("CreateUpdateSecret", func() {
	type test struct {
		secretID                   string
		responseCode               int
		responseBody               string
		createUpdateSecretsRequest *styra.CreateUpdateSecretsRequest
		expectStyraErr             bool
	}

	DescribeTable("CreateUpdateSecret", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			var b bytes.Buffer
			Expect(json.NewEncoder(&b).Encode(test.createUpdateSecretsRequest)).To(Succeed())
			Expect(bs).To(Equal((b.Bytes())))
			Expect(r.Method).To(Equal(http.MethodPut))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/secrets/" + test.secretID))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.CreateUpdateSecret(context.Background(), test.secretID, test.createUpdateSecretsRequest)
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
			secretID:     "name",
			responseCode: http.StatusOK,
			responseBody: `{"test"}`,
			createUpdateSecretsRequest: &styra.CreateUpdateSecretsRequest{
				Description: "description",
				Name:        "name",
				Secret:      "secret",
			},
		}),

		Entry("styra http error", test{
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
