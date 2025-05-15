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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// ClientInterface defines the interface for the Styra client.
type ClientInterface interface {
	GetSystem(ctx context.Context, id string) (*GetSystemResponse, error)
	GetSystemByName(ctx context.Context, name string) (*GetSystemResponse, error)

	CreateUpdateSecret(
		ctx context.Context,
		secretID string,
		request *CreateUpdateSecretsRequest,
	) (*CreateUpdateSecretResponse, error)
	DeleteSecret(
		ctx context.Context,
		secretID string,
	) (*DeleteSecretResponse, error)

	GetUser(ctx context.Context, name string) (*GetUserResponse, error)
	GetUsers(ctx context.Context) (*GetUsersResponse, bool, error)
	InvalidateCache()

	CreateInvitation(ctx context.Context, email bool, name string) (*CreateInvitationResponse, error)

	ListRoleBindingsV2(ctx context.Context, params *ListRoleBindingsV2Params) (*ListRoleBindingsV2Response, error)

	CreateRoleBinding(ctx context.Context, request *CreateRoleBindingRequest) (*CreateRoleBindingResponse, error)

	UpdateRoleBindingSubjects(
		ctx context.Context,
		id string,
		request *UpdateRoleBindingSubjectsRequest,
	) (*UpdateRoleBindingSubjectsResponse, error)

	DeleteRoleBindingV2(ctx context.Context, id string) (*DeleteRoleBindingV2Response, error)

	GetDatasource(ctx context.Context, id string) (*GetDatasourceResponse, error)

	UpsertDatasource(
		ctx context.Context,
		id string,
		request *UpsertDatasourceRequest,
	) (*UpsertDatasourceResponse, error)

	DeleteDatasource(ctx context.Context, id string) (*DeleteDatasourceResponse, error)

	GetLibrary(ctx context.Context, id string) (*GetLibraryResponse, error)
	UpsertLibrary(ctx context.Context, id string, request *UpsertLibraryRequest) (*UpsertLibraryResponse, error)

	UpdateSystem(ctx context.Context, id string, request *UpdateSystemRequest) (*UpdateSystemResponse, error)

	DeleteSystem(ctx context.Context, id string) (*DeleteSystemResponse, error)

	CreateSystem(ctx context.Context, request *CreateSystemRequest) (*CreateSystemResponse, error)

	GetOPAConfig(ctx context.Context, systemID string) (OPAConfig, error)

	VerifyGitConfiguration(ctx context.Context, request *VerfiyGitConfigRequest) (*VerfiyGitConfigResponse, error)

	DeletePolicy(ctx context.Context, policyName string) (*DeletePolicyResponse, error)

	UpdateWorkspace(ctx context.Context, request *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error)
	UpdateWorkspaceRaw(ctx context.Context, request interface{}) (*UpdateWorkspaceResponse, error)
}

// Client is a client for the Styra APIs.
type Client struct {
	HTTPClient http.Client
	URL        string
	token      string
	Cache      *cache.Cache
}

// New creates a new Styra ClientInterface.
func New(url string, token string) ClientInterface {
	c := cache.New(1*time.Hour, 10*time.Minute)

	return &Client{
		URL:        url,
		HTTPClient: http.Client{},
		token:      token,
		Cache:      c,
	}
}

// InvalidateCache invalidates the entire cache
func (c *Client) InvalidateCache() {
	c.Cache.Flush()
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}, headers map[string]string) (*http.Request, error) {
	u := fmt.Sprintf("%s%s", c.URL, endpoint)

	var b bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			return nil, errors.Wrap(err, "could not encode body")
		}
	}

	r, err := http.NewRequestWithContext(ctx, method, u, &b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	r.Header.Set("Content-Type", "application/json")

	if headers != nil {
		for k, v := range headers {
			r.Header.Set(k, v)
		}
	}

	return r, nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, endpoint, body, headers)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not send request")
	}

	return res, nil
}
