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

// Package ocp provides functionality for interacting with the OCP API.
package ocp

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

// ClientInterface defines the interface for the OCP client.
type ClientInterface interface {
	GetSource(ctx context.Context, id string) (*GetSourceResponse, error)
	PutSource(ctx context.Context, id string, request *PutSourceRequest) (*PutSourceResponse, error)
	PutBundle(ctx context.Context, bundle *PutBundleRequest) error
}

// Client is a client for the OCP APIs.
type Client struct {
	HTTPClient http.Client
	URL        string
	token      string
	Cache      *cache.Cache
}

// New creates a new OCP ClientInterface.
func New(url string, token string) ClientInterface {
	c := cache.New(6*time.Hour, 10*time.Minute)

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

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
	headers map[string]string,
) (*http.Request, error) {
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

	for k, v := range headers {
		r.Header.Set(k, v)
	}

	return r, nil
}

func (c *Client) request(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
	headers map[string]string,
) (*http.Response, error) {
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
