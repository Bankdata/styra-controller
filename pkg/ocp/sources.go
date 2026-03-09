/*
Copyright (C) 2026 Bankdata (bankdata@bankdata.dk)

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

package ocp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	configv2alpha3 "github.com/bankdata/styra-controller/api/config/v2alpha3"
	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/pkg/errors"
)

const (
	endpointV1Sources = "/v1/sources"
)

// PutSourceRequest is the request body for the
// PUT /v1/sources/{id} endpoint in the OCP API.
type PutSourceRequest struct {
	Name          string            `json:"name,omitempty" yaml:"name,omitempty"`
	Builtin       *string           `json:"builtin,omitempty" yaml:"builtin,omitempty"`
	Git           *GitConfig        `json:"git,omitempty" yaml:"git,omitempty"`
	Datasources   []Datasource      `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	EmbeddedFiles map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	// Root directory for the source files, used to resolve file paths below.
	Directory    string        `json:"directory,omitempty" yaml:"directory,omitempty"`
	Paths        []string      `json:"paths,omitempty" yaml:"paths,omitempty"`
	Requirements []Requirement `json:"requirements,omitempty" yaml:"requirements,omitempty"`
}

// PutSourceResponse is the response type for calls to the
// PUT /v1/sources/{id} endpoint in the OCP API.
type PutSourceResponse struct {
	StatusCode int
	Body       []byte
	Message    string
}

// GetSourceResponse is the response type for calls to the
// GET /v1/sources/{id} endpoint in the OCP API.
type GetSourceResponse struct {
	StatusCode int
	Body       []byte
	Source     *SourceConfig
	Message    string
}

// SourceConfig represents the configuration of a source in the OCP APIs.
type SourceConfig struct {
	Name          string            `json:"name,omitempty" yaml:"name,omitempty"`
	Builtin       *string           `json:"builtin,omitempty" yaml:"builtin,omitempty"`
	Git           GitConfig         `json:"git,omitempty" yaml:"git,omitempty"`
	Datasources   []Datasource      `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	EmbeddedFiles map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	// Root directory for the source files, used to resolve file paths below.
	Directory    string        `json:"directory,omitempty" yaml:"directory,omitempty"`
	Paths        []string      `json:"paths,omitempty" yaml:"paths,omitempty"`
	Requirements []Requirement `json:"requirements,omitempty" yaml:"requirements,omitempty"`
}

// GitConfig represents the git source control configuration for a source.
type GitConfig struct {
	Repo          string   `json:"repo"`
	Reference     string   `json:"reference,omitempty"`
	Commit        string   `json:"commit,omitempty"`
	Path          string   `json:"path,omitempty"`
	IncludedFiles []string `json:"included_files,omitempty"`
	ExcludedFiles []string `json:"excluded_files,omitempty"`
	CredentialID  string   `json:"credentials,omitempty"`
}

// Datasource represents a datasource for a source.
type Datasource struct {
	Name           string                 `json:"name" yaml:"name"`
	Path           string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Type           string                 `json:"type" yaml:"type"`
	TransformQuery string                 `json:"transform_query,omitempty" yaml:"transform_query,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Credentials    *SecretRef             `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// SecretRef represents a reference to a secret for a datasource in a source.
type SecretRef struct {
	Name string `json:"-" yaml:"-"`
}

// Secret represents a secret for a datasource in a source.
type Secret struct {
	Name  string                 `json:"-" yaml:"-"`
	Value map[string]interface{} `json:"-" yaml:"-"`
}

// Requirement represents a requirement for a bundle and a source.
type Requirement struct {
	Source          string
	RequirementType RequirementType
}

// RequirementType defines the different types of requirements
type RequirementType string

const (
	// RequirementTypeGit means that the requirement is a git source
	RequirementTypeGit RequirementType = "git"

	// RequirementTypeData means that the requirement is a data source
	RequirementTypeData RequirementType = "data"

	// RequirementTypeGitAndData means that the requirement is both a git and a data source
	RequirementTypeGitAndData RequirementType = "git_and_data"

	// RequirementTypeUnknown means that the requirement type is unknown
	RequirementTypeUnknown RequirementType = "unknown"
)

// NewRequirement creates a new Requirement for a bundle.
func NewRequirement(source string, sourceType RequirementType) Requirement {
	return Requirement{
		Source:          source,
		RequirementType: sourceType,
	}
}

// ToRequirements converts the default requirements to a list of bundle Requirements.
func ToRequirements(sources []configv2alpha3.DefaultRequirement) []Requirement {
	requirements := make([]Requirement, len(sources))
	for i, source := range sources {
		requirements[i] = NewRequirement(source.Name, RequirementType(source.RequirementType))
	}
	return requirements
}

// GetSource calls the GET /v1/sources/{id} endpoint in the OCP API.
func (c *Client) GetSource(ctx context.Context, path string) (resp *GetSourceResponse, err error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", endpointV1Sources, path), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not get source from OCP")
	}

	// Close body and overwrite returned error if it is not set already.
	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = errors.Wrap(closeErr, "error closing response body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read GetSource body")
	}

	if res.StatusCode != http.StatusOK {
		return nil, httperror.NewHTTPError(res.StatusCode, string(body))
	}

	var sourceConfig SourceConfig
	if err := json.Unmarshal(body, &sourceConfig); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal GetSource body")
	}

	return &GetSourceResponse{
		StatusCode: res.StatusCode,
		Body:       body,
		Message:    res.Status,
		Source:     &sourceConfig,
	}, nil
}

// PutSource calls the PUT /v1/sources/{id} endpoint in the OCP API.
func (c *Client) PutSource(
	ctx context.Context,
	id string,
	request *PutSourceRequest,
) (resp *PutSourceResponse, err error) {
	res, err := c.request(ctx, http.MethodPut, fmt.Sprintf("%s/%s", endpointV1Sources, id), request, nil)
	if err != nil {
		return nil, errors.Wrap(err, "PutSource: could not call OCP")
	}

	// Close body and overwrite returned error if it is not set already.
	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = errors.Wrap(closeErr, "error closing response body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "PutSource: could not read body")
	}

	if res.StatusCode != http.StatusOK {
		return nil, httperror.NewHTTPError(res.StatusCode, string(body))
	}

	return &PutSourceResponse{
		StatusCode: res.StatusCode,
		Body:       body,
		Message:    res.Status,
	}, nil
}

// DeleteSource calls the DELETE /v1/sources/{name} endpoint in the OCP API.
func (c *Client) DeleteSource(ctx context.Context, id string) (err error) {
	res, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", endpointV1Sources, id), nil, nil)
	if err != nil {
		return err
	}

	// Close body and overwrite returned error if it is not set already.
	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = errors.Wrap(closeErr, "error closing response body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "DeleteSource: could not read body")
	}

	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusOK {
		return httperror.NewHTTPError(res.StatusCode, string(body))
	}
	return nil
}
