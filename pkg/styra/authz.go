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
	"net/url"

	"github.com/pkg/errors"
)

const (
	endpointV2Rolebindings = "/v2/authz/rolebindings"
)

// Role represents a role in Styra.
type Role string

const (
	// RoleSystemViewer is the Styra SystemViewer role.
	RoleSystemViewer Role = "SystemViewer"

	// RoleSystemPolicyEditor is the Styra SystemPolicyEditor role.
	RoleSystemPolicyEditor Role = "SystemPolicyEditor"

	// RoleLibraryViewer is the Styra LibraryViewer role.
	RoleLibraryViewer Role = "LibraryViewer"
)

// RoleBindingKind is the kind of the role binding.
type RoleBindingKind string

const (
	// RoleBindingKindSystem is a RoleBindingKind used when the role is for a
	// System.
	RoleBindingKindSystem RoleBindingKind = "system"
	// RoleBindingKindLibrary is a RoleBindingKind used when the role is for a
	// Library.
	RoleBindingKindLibrary RoleBindingKind = "library"
)

// SubjectKind is the kind of a subject.
type SubjectKind string

const (
	// SubjectKindUser is a SubjectKind used when the subject is a user.
	SubjectKindUser SubjectKind = "user"

	// SubjectKindClaim is a SubjectKind used when the subject is a claim.
	SubjectKindClaim SubjectKind = "claim"
)

// ClaimConfig represents a claim configuration.
type ClaimConfig struct {
	IdentityProvider string `json:"identity_provider,omitempty"`
	Key              string `json:"key"`
	Value            string `json:"value"`
}

// ListRoleBindingsV2Params is the URL params for the
// GET /v2/authz/rolebindings endpoint in the Styra API.
type ListRoleBindingsV2Params struct {
	ResourceKind RoleBindingKind
	ResourceID   string
	RoleID       Role
	SubjectKind  SubjectKind
}

// CreateRoleBindingRequest is the request body for the
// POST /v2/authz/rolebindings endpoint in the Styra API.
type CreateRoleBindingRequest struct {
	ResourceFilter *ResourceFilter `json:"resource_filter"`
	RoleID         Role            `json:"role_id"`
	Subjects       []*Subject      `json:"subjects"`
}

// ResourceFilter is a resource filter. This is used to limit what resources
// are targeted in the Styra APIs.
type ResourceFilter struct {
	ID   string          `json:"id"`
	Kind RoleBindingKind `json:"kind"`
}

// UpdateRoleBindingSubjectsRequest is the request body for the
// POST /v2/authz/rolebindings/{id}/subjects endpoint in the Styra API.
type UpdateRoleBindingSubjectsRequest struct {
	Subjects []*Subject `json:"subjects"`
}

// ListRoleBindingsV2Response is the response body for the
// GET /v2/authz/rolebindings endpoint in the Styra API.
type ListRoleBindingsV2Response struct {
	Rolebindings []*RoleBindingConfig `json:"rolebindings"`
	StatusCode   int
	Body         []byte
}

// RoleBindingConfig defines the structure of a rolebinding configuration. This
// is used for binding a list of subjects to a specific role.
type RoleBindingConfig struct {
	ID       string     `json:"id"`
	Subjects []*Subject `json:"subjects"`
	RoleID   Role       `json:"role_id"`
}

// Subject specifies a subject.
type Subject struct {
	ID          string       `json:"id,omitempty"`
	Kind        SubjectKind  `json:"kind"`
	ClaimConfig *ClaimConfig `json:"claim_config,omitempty"`
}

// CreateRoleBindingResponse is the response body for the
// POST /v2/authz/rolebindings endpoint in the Styra API.
type CreateRoleBindingResponse struct {
	Rolebinding *RoleBindingConfig `json:"rolebinding"`
	StatusCode  int
	Body        []byte
}

// UpdateRoleBindingSubjectsResponse is the response type for calls to the
// POST /v2/authz/rolebindings/{id}/subjects endpoint in the Styra API.
type UpdateRoleBindingSubjectsResponse struct {
	StatusCode int
	Body       []byte
}

// DeleteRoleBindingV2Response is the response type for calls to the
// DELETE /v2/authz/rolebindings/{id}/subjects endpoint in the Styra API
type DeleteRoleBindingV2Response struct {
	StatusCode int
	Body       []byte
}

// ListRoleBindingsV2 calls the GET /v2/authz/rolebindings endpoint in the
// Styra API.
func (c *Client) ListRoleBindingsV2(
	ctx context.Context,
	params *ListRoleBindingsV2Params,
) (*ListRoleBindingsV2Response, error) {
	values := url.Values{}

	if params.ResourceKind != "" {
		values["resource_kind"] = []string{string(params.ResourceKind)}
	}
	if params.ResourceID != "" {
		values["resource_id"] = []string{params.ResourceID}
	}
	if params.RoleID != "" {
		values["role_id"] = []string{string(params.RoleID)}
	}
	if params.SubjectKind != "" {
		values["subject_kind"] = []string{string(params.SubjectKind)}
	}

	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("%s?%s", endpointV2Rolebindings, values.Encode()), nil)
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

	r := ListRoleBindingsV2Response{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	if r.StatusCode == http.StatusOK {
		if err := json.Unmarshal(r.Body, &r); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal body")
		}
	}

	return &r, nil
}

// CreateRoleBinding calls the POST /v2/authz/rolebindings endpoint in the
// Styra API.
func (c *Client) CreateRoleBinding(
	ctx context.Context,
	request *CreateRoleBindingRequest,
) (*CreateRoleBindingResponse, error) {
	res, err := c.request(ctx, http.MethodPost, endpointV2Rolebindings, request)
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

	r := CreateRoleBindingResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	if r.StatusCode == http.StatusOK {
		if err := json.Unmarshal(r.Body, &r); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal body")
		}
	}

	return &r, nil
}

// UpdateRoleBindingSubjects calls the POST /v2/authz/rolebindings/{id}/subjects
// endpoint in the Styra API.
func (c *Client) UpdateRoleBindingSubjects(
	ctx context.Context,
	id string,
	request *UpdateRoleBindingSubjectsRequest,
) (*UpdateRoleBindingSubjectsResponse, error) {
	fmt.Println("updating rolebinding")
	res, err := c.request(ctx, http.MethodPost, fmt.Sprintf("%s/%s/subjects", endpointV2Rolebindings, id), request)
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

	return &UpdateRoleBindingSubjectsResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}, nil
}

// DeleteRoleBindingV2 calls the DELETE /v2/authz/rolebindings/{id}/subjects
// endpoint in the Styra API.
func (c *Client) DeleteRoleBindingV2(ctx context.Context, id string) (*DeleteRoleBindingV2Response, error) {
	res, err := c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", endpointV2Rolebindings, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		err := NewHTTPError(res.StatusCode, string(body))
		return nil, err
	}

	return &DeleteRoleBindingV2Response{
		StatusCode: res.StatusCode,
		Body:       body,
	}, nil
}

// SubjectsAreEqual checks if two lists of Subjects are equal.
func SubjectsAreEqual(as []*Subject, bs []*Subject) bool {
	if len(as) != len(bs) {
		return false
	}

	for _, a := range as {
		found := false
		for _, b := range bs {
			if a.Kind == SubjectKindClaim {
				if a.Kind == b.Kind &&
					a.ClaimConfig.IdentityProvider == b.ClaimConfig.IdentityProvider &&
					a.ClaimConfig.Key == b.ClaimConfig.Key &&
					a.ClaimConfig.Value == b.ClaimConfig.Value {
					found = true
				}
			}
			if a.Kind == SubjectKindUser {
				if a.Kind == b.Kind && a.ID == b.ID {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}
