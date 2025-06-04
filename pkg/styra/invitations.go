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

package styra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

const (
	endpointV1Invitations = "/v1/invitations"
)

// CreateInvitationResponse is the response type for calls to the
// POST /v1/invitations endpoint in the Styra API.
type CreateInvitationResponse struct {
	StatusCode int
	Body       []byte
}

// CreateInvitationRequest is the request body for the
// POST /v1/invitations endpoint in the Styra API.
type CreateInvitationRequest struct {
	UserID string `json:"user_id"`
}

// CreateInvitation calls the POST /v1/invitations endpoint in the Styra API.
func (c *Client) CreateInvitation(ctx context.Context, email bool, name string) (*CreateInvitationResponse, error) {
	createInvitationData := CreateInvitationRequest{
		UserID: name,
	}

	res, err := c.request(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s?email=%s", endpointV1Invitations, strconv.FormatBool(email)),
		createInvitationData,
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := CreateInvitationResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}
