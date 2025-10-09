package styra_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/bankdata/styra-controller/pkg/httperror"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("DeletePolicy", func() {

	type test struct {
		policyName     string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	ginkgo.DescribeTable("DeletePolicy", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(bs).To(gomega.Equal([]byte("")))
			gomega.Expect(r.Method).To(gomega.Equal(http.MethodDelete))
			gomega.Expect(r.URL.String()).To(gomega.Equal("http://test.com/v1/policies/" + test.policyName))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeletePolicy(context.Background(), test.policyName)
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
			policyName:   "policyname",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		ginkgo.Entry("styra http error", test{
			policyName:     "policyname",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
