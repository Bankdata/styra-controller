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
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(f roundTripFunc, systemURL string, libraryURL string) Client {
	return &client{
		hc: http.Client{
			Transport: roundTripFunc(f),
		},
		systemDatasourceChanged:  systemURL,
		libraryDatasourceChanged: libraryURL,
	}
}

var _ = ginkgo.Describe("Succeed to update system datasource", func() {

	systemID := "id_system"
	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\",\"systemId\":\"id_system\"}"

	ginkgo.It("should return nil as error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://localhost:8080/v1/datasources/webhook"))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			gomega.Expect(actualBody).To(gomega.BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`datasource updated`)),
			}
		}

		c := NewTestClient(roundTripFunc, "http://localhost:8080/v1/datasources/webhook", "")

		err := c.SystemDatasourceChanged(context.Background(), testlogger, systemID, datasourceID)

		gomega.Expect(err).To(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("Fail to update system datasource", func() {

	systemID := "id_system"
	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\",\"systemId\":\"id_system\"}"

	ginkgo.It("should return an error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://localhost:8080/v1/datasources/webhook"))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			gomega.Expect(actualBody).To(gomega.BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewBufferString(`forbidden`)),
			}
		}

		c := NewTestClient(roundTripFunc, "http://localhost:8080/v1/datasources/webhook", "")

		err := c.SystemDatasourceChanged(context.Background(), testlogger, systemID, datasourceID)

		gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("response status code is 403, response body is forbidden"))
	})
})

var _ = ginkgo.Describe("Succeed to update library datasource", func() {

	datasourceID := "libraries/libraryID/datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceID\":\"libraries/libraryID/datasource\"}"
	ginkgo.It("should return nil as error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://localhost:8080/v1/libraries/webhook"))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			gomega.Expect(actualBody).To(gomega.BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`datasource updated`)),
			}
		}

		c := NewTestClient(roundTripFunc, "", "http://localhost:8080/v1/libraries/webhook")

		err := c.LibraryDatasourceChanged(context.Background(), testlogger, datasourceID)

		gomega.Expect(err).To(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("Fail to update library datasource", func() {

	datasourceID := "libraries/libraryID/datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceID\":\"libraries/libraryID/datasource\"}"

	ginkgo.It("should return an error", func() {

		roundTripFunc := func(r *http.Request) *http.Response {
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://localhost:8080/v1/libraries/webhook"))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))

			body, _ := r.GetBody()
			bodyBytes, _ := io.ReadAll(body)
			actualBody := string(bodyBytes)
			gomega.Expect(actualBody).To(gomega.BeEquivalentTo(expectedBody))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewBufferString(`forbidden`)),
			}
		}

		c := NewTestClient(roundTripFunc, "", "http://localhost:8080/v1/libraries/webhook")

		err := c.LibraryDatasourceChanged(context.Background(), testlogger, datasourceID)

		gomega.Expect(err.Error()).To(gomega.BeEquivalentTo("response status code is 403, response body is forbidden"))
	})
})
