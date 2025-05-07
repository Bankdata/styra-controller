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
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	endpointV1Workspace = "/v1/workspace"
)

// UpdateWorkspaceRequest is the request type for calls to the PUT /v1/workspace endpoint
// in the Styra API.
type UpdateWorkspaceRequest struct {
	DecisionsExporter *ExporterConfig `json:"decisions_exporter,omitempty"`
	ActivityExporter  *ExporterConfig `json:"activity_exporter,omitempty"`
}

// UpdateWorkspaceResponse is the response type for calls to the PUT /v1/workspace endpoint
// in the Styra API.
type UpdateWorkspaceResponse struct {
	StatusCode int
	Body       []byte
}

// ExporterConfig is the configuration for the decision exporter in the Styra API.
type ExporterConfig struct {
	Interval string       `json:"interval,omitempty"`
	Kafka    *KafkaConfig `json:"kafka,omitempty"`
}

// KafkaConfig is the configuration for the Kafka exporter in the Styra API.
type KafkaConfig struct {
	Authentication string    `json:"authentication"`
	Brokers        []string  `json:"brokers"`
	RequredAcks    string    `json:"required_acks"`
	Topic          string    `json:"topic"`
	TLS            *KafkaTLS `json:"tls"`
}

// KafkaTLS is the TLS configuration for the Kafka exporter in the Styra API.
type KafkaTLS struct {
	ClientCert         string `json:"client_cert"`
	RootCA             string `json:"rootca"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

// UpdateWorkspace calls the PUT /v1/workspace endpoint in the Styra API.
func (c *Client) UpdateWorkspace(
	ctx context.Context,
	request *UpdateWorkspaceRequest,
) (*UpdateWorkspaceResponse, error) {
	return c.UpdateWorkspaceRaw(ctx, request)
}

// UpdateWorkspaceRaw calls the PUT /v1/workspace endpoint in the Styra API.
func (c *Client) UpdateWorkspaceRaw(
	ctx context.Context,
	request interface{},
) (*UpdateWorkspaceResponse, error) {
	res, err := c.request(ctx, http.MethodPatch, endpointV1Workspace, request)
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

	r := UpdateWorkspaceResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	return &r, nil
}
