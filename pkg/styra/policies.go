package styra

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// DeletePolicyResponse is the response type for calls to
// the DELETE /v1/policies/{policy} endpoint in the Styra API.
type DeletePolicyResponse struct {
	StatusCode int
	Body       []byte
}

// DeletePolicy calls the DELETE /v1/policies/{policy} endpoint in the Styra API.
func (c *Client) DeletePolicy(ctx context.Context, policyName string) (*DeletePolicyResponse, error) {
	res, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/v1/policies/%s", policyName), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not delete policy: %s", policyName))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := DeletePolicyResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}
