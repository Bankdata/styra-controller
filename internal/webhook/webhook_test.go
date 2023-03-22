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

package webhook

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/go-logr/logr/testr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(f roundTripFunc, url string) Client {
	return &client{
		hc: http.Client{
			Transport: roundTripFunc(f),
		},
		url: url,
	}
}

var _ = Describe("Succeed to update datasource", func() {

	systemID := "id_system"
	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to webhook
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\",\"systemId\":\"id_system\"}"

	It("should return nil as error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			Expect(r.URL.String()).To(Equal("http://localhost:8080/v1/datasources/webhook"))
			Expect(r.Method).To(Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			Expect(actualBody).To(BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`datasource updated`)),
			}
		}

		c := NewTestClient(roundTripFunc, "http://localhost:8080/v1/datasources/webhook")

		err := c.DatasourceChanged(context.Background(), testlogger, systemID, datasourceID)

		Expect(err).To(BeNil())
	})
})

var _ = Describe("Fail to update datasource", func() {

	systemID := "id_system"
	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to PDG
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\",\"systemId\":\"id_system\"}"

	It("should return an error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			Expect(r.URL.String()).To(Equal("http://localhost:8080/v1/datasources/webhook"))
			Expect(r.Method).To(Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			Expect(actualBody).To(BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewBufferString(`forbidden`)),
			}
		}

		c := NewTestClient(roundTripFunc, "http://localhost:8080/v1/datasources/webhook")

		err := c.DatasourceChanged(context.Background(), testlogger, systemID, datasourceID)

		Expect(err.Error()).To(BeEquivalentTo("response status code is 403, request body is forbidden"))
	})
})
