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
	endpointV1Users = "/v1/users"
)

// GetUserResponse is the response type for calls to the GET /v1/users/{userId} endpoint
// in the Styra API.
type GetUserResponse struct {
	StatusCode int
	Body       []byte
}

// GetUsersResponse is the response type for calls to the GET /v1/users endpoint
// in the Styra API.
type GetUsersResponse struct {
	Users []User
}

// Struct to unmarshal the JSON response from the GET /v1/users endpoint
type getUsersJSONResponse struct {
	Result []User `json:"result"`
}

// User is the struct for a user in the Styra API.
type User struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

// GetUsers calls the GET /v1/users endpoint in the Styra API.
func (c *Client) GetUsers(ctx context.Context) (*GetUsersResponse, bool, error) {
	const cacheKey = "allUsersResponse"

	// Check if the response is in the cache
	if cachedResponse, found := c.Cache.Get(cacheKey); found {
		return cachedResponse.(*GetUsersResponse), true, nil
	}

	res, err := c.GetUserEndpoint(ctx, endpointV1Users)
	if err != nil {
		return nil, false, err
	}

	var js getUsersJSONResponse
	if err := json.Unmarshal(res.Body, &js); err != nil {
		return nil, false, errors.Wrap(err, "could not unmarshal body: ")
	}

	r := GetUsersResponse{
		Users: js.Result,
	}

	// Cache the response
	c.Cache.Set(cacheKey, &r, 0)

	return &r, false, nil
}

// GetUser calls the GET /v1/users/{userId} endpoint in the Styra API.
func (c *Client) GetUser(ctx context.Context, name string) (*GetUserResponse, error) {
	return c.GetUserEndpoint(ctx, fmt.Sprintf("%s/%s", endpointV1Users, name))
}

// GetUserEndpoint is a helper function to call the Styra API.
func (c *Client) GetUserEndpoint(ctx context.Context, endpoint string) (*GetUserResponse, error) {
	res, err := c.request(ctx, http.MethodGet, endpoint, nil)
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

	r := GetUserResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}
