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
	"reflect"
	"sort"

	"github.com/bankdata/styra-controller/api/styra/v1beta1"
	"github.com/pkg/errors"
)

const (
	endpointV1Systems = "/v1/systems"
)

// UpdateSystemRequest is the request body for the the PUT /v1/systems/{system}
// endpoint in the Styra API.
type UpdateSystemRequest struct {
	*SystemConfig
}

// UpdateSystemResponse is the response body for the PUT /v1/systems/{system}
// endpoint in the Styra API.
type UpdateSystemResponse struct {
	StatusCode   int
	Body         []byte
	SystemConfig *SystemConfig
}

// CreateSystemRequest is the request body for the POST /v1/systems
// endpoint in the Styra API.
type CreateSystemRequest struct {
	*SystemConfig
}

// CreateSystemResponse is the response body for the POST /v1/systems
// endpoint in the Styra API.
type CreateSystemResponse struct {
	StatusCode   int
	Body         []byte
	SystemConfig *SystemConfig
}

// DeleteSystemResponse is the response type for calls to
// the DELETE /v1/systems/{system} endpoint in the Styra API.
type DeleteSystemResponse struct {
	StatusCode int
	Body       []byte
}

// GetSystemResponse is the response body for the GET /v1/systems{system}
// endpoint in the Styra API.
type GetSystemResponse struct {
	StatusCode   int
	Body         []byte
	SystemConfig *SystemConfig
}

// SystemConfig represents the configuration of a system in the Styra APIs.
type SystemConfig struct {
	DecisionMappings     map[string]DecisionMapping `json:"decision_mappings,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Name                 string                     `json:"name"`
	ReadOnly             bool                       `json:"read_only"`
	SourceControl        *SourceControlConfig       `json:"source_control,omitempty"`
	Type                 string                     `json:"type"`
	ID                   string                     `json:"id"`
	Datasources          []*DatasourceConfig        `json:"datasources,omitempty"`
	DeploymentParameters *DeploymentParameters      `json:"deployment_parameters,omitempty"`
}

// DeploymentParameters are additional OPA deployment parameters for the
// system.
type DeploymentParameters struct {
	Discovery *v1beta1.DiscoveryOverrides `json:"discovery,omitempty"`
}

// SourceControlConfig defines the structure of a source control configuration.
type SourceControlConfig struct {
	Origin GitRepoConfig `json:"origin"`
}

// GitRepoConfig defines the structure of a git repo configuration.
type GitRepoConfig struct {
	Commit      string `json:"commit"`
	Credentials string `json:"credentials"`
	Path        string `json:"path"`
	Reference   string `json:"reference"`
	URL         string `json:"url"`
}

// DecisionMapping defines the structure of a decision mapping.
type DecisionMapping struct {
	Allowed *DecisionMappingAllowed `json:"allowed,omitempty"`
	Columns []DecisionMappingColumn `json:"columns,omitempty"`
	Reason  *DecisionMappingReason  `json:"reason,omitempty"`
}

// DecisionMappingAllowed defines the structure of the allow element in a
// decision mapping.
type DecisionMappingAllowed struct {
	Expected interface{} `json:"expected,omitempty"`
	Negated  bool        `json:"negated,omitempty"`
	Path     string      `json:"path"`
}

// DecisionMappingColumn defines the structure of the column element in a
// decision mapping.
type DecisionMappingColumn struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Type string `json:"type,omitempty"`
}

// DecisionMappingReason defines the structure of the reason element in a
// decision mapping.
type DecisionMappingReason struct {
	Path string `json:"path"`
}

type getSystemJSONResponse struct {
	Result *SystemConfig `json:"result"`
}

// VerfiyGitConfigRequest is the request body for the POST
// /v1/systems/source-control/verify-config endpoint in the Styra API.
type VerfiyGitConfigRequest struct {
	Commit      string `json:"commit"`
	ID          string `json:"id"`
	Credentials string `json:"credentials"`
	Path        string `json:"path"`
	Reference   string `json:"reference"`
	URL         string `json:"url"`
}

// VerfiyGitConfigResponse is the response type for calls to the POST
// /v1/systems/source-control/verify-config endpoint in the Styra API.
type VerfiyGitConfigResponse struct {
	StatusCode int
	Body       []byte
}

// GetSystem calls the GET /v1/systems{system} endpoint in the Styra API.
func (c *Client) GetSystem(ctx context.Context, id string) (*GetSystemResponse, error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", endpointV1Systems, id), nil)
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

	r := GetSystemResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	if r.StatusCode == http.StatusOK {
		var js getSystemJSONResponse
		if err := json.Unmarshal(r.Body, &js); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal body")
		}
		r.SystemConfig = js.Result
	}

	return &r, nil
}

// GetSystemByName calls the GET /v1/systems?name=<name> endpoint in the Styra API. If a system exists with this
// name it will be returned in the response. Otherwise, r.SystemConfig will be nil.
func (c *Client) GetSystemByName(ctx context.Context, name string) (*GetSystemResponse, error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s?name=%s", endpointV1Systems, name), nil)
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

	r := GetSystemResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	var js struct {
		Result []SystemConfig
	}
	if err := json.Unmarshal(r.Body, &js); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal body")
	}

	if len(js.Result) > 0 {
		r.SystemConfig = &js.Result[0]
	}

	return &r, nil
}

// CreateSystem calls the POST /v1/systems endpoint in the Styra API.
func (c *Client) CreateSystem(ctx context.Context, request *CreateSystemRequest) (*CreateSystemResponse, error) {
	res, err := c.request(ctx, http.MethodPost, endpointV1Systems, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call post system")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := CreateSystemResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	if r.StatusCode == http.StatusOK {
		var js getSystemJSONResponse
		if err := json.Unmarshal(r.Body, &js); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not unmarshal body, %v", string(r.Body)))
		}
		r.SystemConfig = js.Result
	}

	return &r, nil
}

// UpdateSystem calls the PUT /v1/systems/{system} endpoint in the Styra API.
func (c *Client) UpdateSystem(
	ctx context.Context,
	id string,
	request *UpdateSystemRequest,
) (*UpdateSystemResponse, error) {
	res, err := c.request(ctx, http.MethodPut, fmt.Sprintf("%s/%s", endpointV1Systems, id), request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call update system")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := UpdateSystemResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	if r.StatusCode == http.StatusOK {
		var js getSystemJSONResponse
		if err := json.Unmarshal(r.Body, &js); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal body")
		}
		r.SystemConfig = js.Result
	}

	return &r, nil
}

// VerifyGitConfiguration calls the POST
// /v1/systems/source-control/verify-config endpoint in the Styra API.
func (c *Client) VerifyGitConfiguration(
	ctx context.Context,
	request *VerfiyGitConfigRequest,
) (*VerfiyGitConfigResponse, error) {
	res, err := c.request(ctx, http.MethodPost, fmt.Sprintf("%s/source-control/verify-config", endpointV1Systems), request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call validate git config ")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := VerfiyGitConfigResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}

// DeleteSystem calls the DELETE /v1/systems/{system} endpoint in the Styra
// API.
func (c *Client) DeleteSystem(ctx context.Context, id string) (*DeleteSystemResponse, error) {
	res, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", endpointV1Systems, id), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call delete system")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusOK {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	r := DeleteSystemResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}

// DecisionMappingsEquals asserts if two decisionmappings are equal.
func DecisionMappingsEquals(dms1, dms2 map[string]DecisionMapping) bool {
	if dms1 == nil && dms2 == nil {
		return true
	}
	if dms1 == nil || dms2 == nil {
		return false
	}

	columnSort := func(cols []DecisionMappingColumn) func(i, j int) bool {
		return func(i, j int) bool {
			return cols[i].Key < cols[j].Key
		}
	}
	for _, v := range dms1 {
		sort.Slice(v.Columns, columnSort(v.Columns))
	}
	for _, v := range dms2 {
		sort.Slice(v.Columns, columnSort(v.Columns))
	}
	return reflect.DeepEqual(dms1, dms2)
}
