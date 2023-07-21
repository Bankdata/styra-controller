package styra_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bankdata/styra-controller/pkg/styra"
)

var _ = Describe("DeletePolicy", func() {

	type test struct {
		policyName     string
		responseCode   int
		responseBody   string
		expectedBody   []byte
		expectStyraErr bool
	}

	DescribeTable("DeletePolicy", func(test test) {
		c := newTestClient(func(r *http.Request) *http.Response {
			bs, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("")))
			Expect(r.Method).To(Equal(http.MethodDelete))
			Expect(r.URL.String()).To(Equal("http://test.com/v1/policies/" + test.policyName))

			return &http.Response{
				Header:     make(http.Header),
				StatusCode: test.responseCode,
				Body:       io.NopCloser(bytes.NewBufferString(test.responseBody)),
			}
		})

		res, err := c.DeletePolicy(context.Background(), test.policyName)
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
			policyName:   "policyname",
			responseCode: http.StatusOK,
			responseBody: `expected response from styra api`,
			expectedBody: []byte(`expected response from styra api`)},
		),

		Entry("styra http error", test{
			policyName:     "policyname",
			responseCode:   http.StatusInternalServerError,
			expectStyraErr: true,
		}),
	)
})
