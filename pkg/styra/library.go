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

	"github.com/pkg/errors"
)

const (
	endpointV1Libraries = "/v1/libraries"
)

type getLibraryJSONResponse struct {
	Result *LibraryEntityExpanded `json:"result"`
}

// GetLibraryResponse is the response type for calls to the
// GET /v1/libraries/{library} endpoint in the Styra API.
type GetLibraryResponse struct {
	Statuscode            int
	Body                  []byte
	LibraryEntityExpanded *LibraryEntityExpanded
}

// LibraryEntityExpanded is the type that defines of a Library
type LibraryEntityExpanded struct {
	DataSources   []LibraryDatasourceConfig   `json:"datasources"`
	Description   string                      `json:"description"`
	ID            string                      `json:"id"`
	ReadOnly      bool                        `json:"read_only"`
	SourceControl *LibrarySourceControlConfig `json:"source_control"`
}

// LibraryDatasourceConfig defines metadata of a datasource
type LibraryDatasourceConfig struct {
	Category string `json:"category"`
	ID       string `json:"id"`
}

// LibrarySourceControlConfig is a struct from styra where we only use a single field
// but kept for clarity when comparing to the API
type LibrarySourceControlConfig struct {
	LibraryOrigin *LibraryGitRepoConfig `json:"library_origin"`
}

// LibraryGitRepoConfig defines the Git configurations a library can be defined by
type LibraryGitRepoConfig struct {
	Commit      string `json:"commit"`
	Credentials string `json:"credentials"`
	Path        string `json:"path"`
	Reference   string `json:"reference"`
	URL         string `json:"url"`
}

// UpsertLibraryRequest is the request body for the
// PUT /v1/libraries/{library} endpoint in the Styra API.
type UpsertLibraryRequest struct {
	Description   string                      `json:"description"`
	ReadOnly      bool                        `json:"read_only"`
	SourceControl *LibrarySourceControlConfig `json:"source_control"`
}

// UpsertLibraryResponse is the response body for the
// PUT /v1/libraries/{library} endpoint in the Styra API.
type UpsertLibraryResponse struct {
	StatusCode int
	Body       []byte
}

// GetLibrary calls the GET /v1/libraries/{library} endpoint in the
// Styra API.
func (c *Client) GetLibrary(ctx context.Context, id string) (*GetLibraryResponse, error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", endpointV1Libraries, id), nil)
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

	var jsonRes getLibraryJSONResponse
	if err := json.Unmarshal(body, &jsonRes); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal body")
	}

	return &GetLibraryResponse{
		Statuscode:            res.StatusCode,
		Body:                  body,
		LibraryEntityExpanded: jsonRes.Result,
	}, nil
}

// UpsertLibrary calls the PUT /v1/libraries/{library} endpoint in the
// Styra API.
func (c *Client) UpsertLibrary(ctx context.Context, id string, reqBody *UpsertLibraryRequest,
) (*UpsertLibraryResponse, error) {
	res, err := c.request(ctx, http.MethodPut, fmt.Sprintf("%s/%s", endpointV1Libraries, id), reqBody)
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

	resp := UpsertLibraryResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &resp, nil
}
