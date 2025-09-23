package ocp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	endpointV1Sources = "/v1/sources"
)

// PutSourceRequest is the request body for the
// PUT /v1/sources/{id} endpoint in the Styra API.
type PutSourceRequest struct {
	Source SourceConfig `json:"source"`
}

// PutSourceResponse is the response type for calls to the
// PUT /v1/sources/{id} endpoint in the Styra API.
type PutSourceResponse struct {
	ID     string       `json:"id"`
	Source SourceConfig `json:"source"`
}

// GetSourceResponse is the response type for calls to the
// GET /v1/sources/{id} endpoint in the Styra API.
type GetSourceResponse struct {
	Statuscode int
	Body       []byte
	Source     *SourceConfig
}

// BundleConfig represents the configuration of a bundle in the Styra APIs.
type SourceConfig struct {
	Name          string            `json:"name,omitempty" yaml:"name,omitempty"`
	Builtin       *string           `json:"builtin,omitempty" yaml:"builtin,omitempty"`
	Git           GitConfig         `json:"git,omitempty" yaml:"git,omitempty"`
	Datasources   []Datasource      `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	EmbeddedFiles map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	Directory     string            `json:"directory,omitempty" yaml:"directory,omitempty"` // Root directory for the source files, used to resolve file paths below.
	Paths         []string          `json:"paths,omitempty" yaml:"paths,omitempty"`
	Requirements  []Requirement     `json:"requirements,omitempty" yaml:"requirements,omitempty"`
}

// GitConfig represents the git source control configuration for a bundle.
type GitConfig struct {
	Repo          string `json:"repo"`
	Reference     string `json:"reference"`
	Path          string `json:"path"`
	IncludedFiles string `json:"included_files,omitempty"`
	ExcludedFiles string `json:"excluded_files,omitempty"`
	CredentialID  string `json:"credentials,omitempty"`
}

type Datasource struct {
	Name           string                 `json:"name" yaml:"name"`
	Path           string                 `json:"path" yaml:"path"`
	Type           string                 `json:"type" yaml:"type"`
	TransformQuery string                 `json:"transform_query,omitempty" yaml:"transform_query,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Credentials    *SecretRef             `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type SecretRef struct {
	Name  string `json:"-" yaml:"-"`
	value *Secret
}

type Secret struct {
	Name  string                 `json:"-" yaml:"-"`
	Value map[string]interface{} `json:"-" yaml:"-"`
}

type Requirement struct {
	Source *string        `json:"source,omitempty" yaml:"source,omitempty"`
	Git    GitRequirement `json:"git,omitempty" yaml:"git,omitempty"`
}

func NewRequirement(source string) Requirement {
	return Requirement{
		Source: &source,
	}
}

func ToRequirements(sources []string) []Requirement {
	requirements := make([]Requirement, len(sources))
	for i, source := range sources {
		requirements[i] = NewRequirement(source)
	}
	return requirements
}

type GitRequirement struct {
	Commit *string `json:"commit,omitempty" yaml:"commit,omitempty"`
}

// GetSource calls the GET /v1/sources/{id} endpoint in the OCP API.
func (c *Client) GetSource(ctx context.Context, path string) (*GetSourceResponse, error) {
	// TODO: verify that path does not contain a double /
	// TODO: maybe validate path does not contain 'data' and throw an error otherwise?
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", endpointV1Sources, path), nil, nil)
	if err != nil {
		fmt.Println("error in GetSource1:", err)
		fmt.Println("res:", res)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error in GetSource2:", err)
		fmt.Println("body:", body)
		return nil, errors.Wrap(err, "could not read body")
	}

	var sourceConfig SourceConfig
	if err := json.Unmarshal(body, &sourceConfig); err != nil {
		fmt.Println("error in GetSource3:", err)
		fmt.Println("res:", string(body))
		return nil, errors.Wrap(err, "could not unmarshal body")
	}
	return &GetSourceResponse{
		Statuscode: res.StatusCode,
		Body:       body,
		Source:     &sourceConfig,
	}, nil
}

// PutSource calls the PUT /v1/sources/{id} endpoint in the OCP API.
func (c *Client) PutSource(
	ctx context.Context,
	id string,
	request *PutSourceRequest,
) (*PutSourceResponse, error) {
	res, err := c.request(ctx, http.MethodPut, fmt.Sprintf("%s/%s", endpointV1Sources, id), request, nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	var putSourceResponse PutSourceResponse
	if err := json.Unmarshal(body, &putSourceResponse); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal body")
	}
	return &putSourceResponse, nil
}
