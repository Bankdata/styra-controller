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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

const (
	endpointV1Datasources = "/v1/datasources"
)

// UpsertDatasourceRequest is the request body for the
// PUT /v1/datasources/{datasource} endpoint in the Styra API.
type UpsertDatasourceRequest struct {
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Commit      string `json:"commit,omitempty"`
	Credentials string `json:"credentials,omitempty"`
	Reference   string `json:"reference,omitempty"`
	URL         string `json:"url,omitempty"`
	Path        string `json:"path,omitempty"`
}

// UpsertDatasourceResponse is the response type for calls to the
// PUT /v1/datasources/{datasource} endpoint in the Styra API.
type UpsertDatasourceResponse struct {
	StatusCode int
	Body       []byte
}

// DeleteDatasourceResponse is the response type for calls to the
// DELETE /v1/datasources/{datasource} endpoint in the Styra API.
type DeleteDatasourceResponse struct {
	StatusCode int
	Body       []byte
}

// GetDatasourceResponse stores the response body for the
// GET /v1/datasources/{datasource} endpoint in the Styra API.
type GetDatasourceResponse struct {
	StatusCode       int
	Body             []byte
	DatasourceConfig *DatasourceConfig
}

type getDatasourceJSONResponse struct {
	Result *DatasourceConfig `json:"result,omitempty"`
}

// DatasourceConfig defines the structure of a datasource configuration.
type DatasourceConfig struct {
	Category    string `json:"category"`
	Type        string `json:"type,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
	Commit      string `json:"commit,omitempty"`
	Credentials string `json:"credentials,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	ID          string `json:"id,omitempty"`
	Path        string `json:"path,omitempty"`
	Reference   string `json:"reference,omitempty"`
	URL         string `json:"url,omitempty"`
}

// GetDatasource calls the GET /v1/datasources/{datasource} endpoint in the
// Styra API.
func (c *Client) GetDatasource(ctx context.Context, id string) (*GetDatasourceResponse, error) {
	res, err := c.request(ctx, http.MethodGet, path.Join(endpointV1Datasources, id), nil)
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

	var jsonRes getDatasourceJSONResponse
	if err := json.Unmarshal(body, &jsonRes); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal body")
	}

	return &GetDatasourceResponse{
		StatusCode:       res.StatusCode,
		Body:             body,
		DatasourceConfig: jsonRes.Result,
	}, nil
}

// UpsertDatasource calls the PUT /v1/datasources/{datasource} endpoint in the
// Styra API.
func (c *Client) UpsertDatasource(
	ctx context.Context,
	id string,
	request *UpsertDatasourceRequest,
) (*UpsertDatasourceResponse, error) {
	res, err := c.request(ctx, http.MethodPut, fmt.Sprintf("%s/%s", endpointV1Datasources, id), request)
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

	r := UpsertDatasourceResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}

// DeleteDatasource calls the DELETE /v1/datasources/{datasource} endpoint in
// the Styra API.
func (c *Client) DeleteDatasource(ctx context.Context, id string) (*DeleteDatasourceResponse, error) {
	res, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", endpointV1Datasources, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := DeleteDatasourceResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	return &r, nil
}
