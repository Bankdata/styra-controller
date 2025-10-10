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

package styra

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/pkg/errors"
)

const (
	endpointV1Secrets = "/v1/secrets"
)

// DeleteSecretResponse is the response type for calls to the
// DELETE /v1/secrets/{secretId} endpoint in the Styra API.
type DeleteSecretResponse struct {
	StatusCode int
	Body       []byte
}

// CreateUpdateSecretResponse is the response type for calls to the
// PUT /v1/secrets/{secretId} endpoint in the Styra API.
type CreateUpdateSecretResponse struct {
	StatusCode int
	Body       []byte
}

// CreateUpdateSecretsRequest is the response body for the
// PUT /v1/secrets/{secretId} endpoint in the Styra API.
type CreateUpdateSecretsRequest struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Secret      string `json:"secret"`
}

// CreateUpdateSecret calls the PUT /v1/secrets/{secretId} endpoint in the
// Styra API.
func (c *Client) CreateUpdateSecret(
	ctx context.Context,
	secretID string,
	createUpdateSecretsRequest *CreateUpdateSecretsRequest,
) (*CreateUpdateSecretResponse, error) {
	res, err := c.request(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/%s", endpointV1Secrets, secretID),
		createUpdateSecretsRequest,
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
		err := httperror.NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := CreateUpdateSecretResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}

// DeleteSecret calls the DELETE /v1/secrets/{secretId} endpoint in the
// Styra API.
func (c *Client) DeleteSecret(
	ctx context.Context,
	secretID string,
) (*DeleteSecretResponse, error) {
	res, err := c.request(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/%s", endpointV1Secrets, secretID),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		err := httperror.NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := DeleteSecretResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}
