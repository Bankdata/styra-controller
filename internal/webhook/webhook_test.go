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

package webhook

import (
	"bytes"
	"context"
	"errors"
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
		systemDatasourceChangedOCP:  systemURL,
		libraryDatasourceChangedOCP: libraryURL,
	}
}

var _ = ginkgo.Describe("New client creation", func() {
	ginkgo.It("should create a client with system and library webhook URLs", func() {
		systemURL := "http://example.com/system"
		libraryURL := "http://example.com/library"

		c := New(systemURL, libraryURL)

		gomega.Expect(c).NotTo(gomega.BeNil())
		gomega.Expect(c).Should(gomega.BeAssignableToTypeOf(&client{}))
	})

	ginkgo.It("should create a client with empty URLs", func() {
		c := New("", "")

		gomega.Expect(c).NotTo(gomega.BeNil())
		gomega.Expect(c).Should(gomega.BeAssignableToTypeOf(&client{}))
	})
})

var _ = ginkgo.Describe("Succeed to update system datasource", func() {

	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\"}"

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

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, datasourceID)

		gomega.Expect(err).To(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("Fail to update system datasource", func() {

	datasourceID := "systems/id_system/test_datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to system webhook
	expectedBody := "{\"datasourceId\":\"systems/id_system/test_datasource\"}"

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

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, datasourceID)

		gomega.Expect(err.Error()).To(
			gomega.BeEquivalentTo(
				"Failed to create request to webhook: response status code is 403, response body is forbidden",
			))
	})
})

var _ = ginkgo.Describe("Succeed to update library datasource", func() {

	datasourceID := "libraries/libraryID/datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to library webhook
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

		err := c.LibraryDatasourceChangedOCP(context.Background(), testlogger, datasourceID)

		gomega.Expect(err).To(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("Fail to update library datasource", func() {

	datasourceID := "libraries/libraryID/datasource"

	testlogger := testr.New(&testing.T{})

	//expected body of call to library webhook
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

		err := c.LibraryDatasourceChangedOCP(context.Background(), testlogger, datasourceID)

		gomega.Expect(err.Error()).To(
			gomega.BeEquivalentTo(
				"Failed to create request to webhook: response status code is 403, response body is forbidden",
			))
	})
})

var _ = ginkgo.Describe("Skip webhook call when not configured", func() {
	testlogger := testr.New(&testing.T{})

	ginkgo.It("should return nil when system webhook URL is empty", func() {
		roundTripFunc := func(_ *http.Request) *http.Response {
			ginkgo.Fail("Should not call webhook when URL is empty")
			return nil
		}

		c := NewTestClient(roundTripFunc, "", "")

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("should return nil when library webhook URL is empty", func() {
		roundTripFunc := func(_ *http.Request) *http.Response {
			ginkgo.Fail("Should not call webhook when URL is empty")
			return nil
		}

		c := NewTestClient(roundTripFunc, "", "")

		err := c.LibraryDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).To(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("Handle HTTP request errors", func() {
	testlogger := testr.New(&testing.T{})

	ginkgo.It("should handle request creation error for system datasource", func() {
		c := &client{
			hc:                         http.Client{},
			systemDatasourceChangedOCP: "://invalid-url",
		}

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("Failed to create request to webhook"))
	})

	ginkgo.It("should handle request creation error for library datasource", func() {
		c := &client{
			hc:                          http.Client{},
			libraryDatasourceChangedOCP: "://invalid-url",
		}

		err := c.LibraryDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("Failed to create request to webhook"))
	})

	ginkgo.It("should handle HTTP client error when response body read fails", func() {
		errorRoundTrip := func(_ *http.Request) *http.Response {
			return &http.Response{
				Header:     make(http.Header),
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`error message`)),
			}
		}

		c := NewTestClient(errorRoundTrip, "http://localhost:8080/webhook", "")

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("response status code is 500"))
	})

	ginkgo.It("should handle HTTP client Do error", func() {
		errorTransport := func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		}

		c := &client{
			hc: http.Client{
				Transport: testTransport(errorTransport),
			},
			systemDatasourceChangedOCP: "http://localhost:8080/webhook",
		}

		err := c.SystemDatasourceChangedOCP(context.Background(), testlogger, "test-datasource")

		gomega.Expect(err).NotTo(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("Failed in call to webhook"))
	})
})

type testTransport func(req *http.Request) (*http.Response, error)

func (t testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}
