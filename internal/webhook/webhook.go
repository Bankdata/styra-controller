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

// Package webhook contains helpers for the notifaction webhooks of the
// controller. These webhooks can be used to notify other systems when
// something happens in the controller.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Client defines the interface for the notification webhook client.
type Client interface {
	SystemDatasourceChanged(context.Context, logr.Logger, string, string) error
	LibraryDatasourceChanged(context.Context, logr.Logger, string) error
	SystemDatasourceChangedOCP(context.Context, logr.Logger, string) error
	LibraryDatasourceChangedOCP(context.Context, logr.Logger, string) error
}

type client struct {
	hc                          http.Client
	libraryDatasourceChanged    string
	systemDatasourceChanged     string
	systemDatasourceChangedOCP  string
	libraryDatasourceChangedOCP string
}

// New creates a new webhook notification Client.
func New(
	systemDatasourceChanged string,
	libraryDatasourceChanged string,
	systemDatasourceChangedOCP string,
	libraryDatasourceChangedOCP string) Client {
	return &client{
		hc:                          http.Client{},
		systemDatasourceChanged:     systemDatasourceChanged,
		libraryDatasourceChanged:    libraryDatasourceChanged,
		systemDatasourceChangedOCP:  systemDatasourceChangedOCP,
		libraryDatasourceChangedOCP: libraryDatasourceChangedOCP,
	}
}

// LibraryDatasourceChangedOCP notifies the webhook that a library datasource has changed in OCP.
func (client *client) LibraryDatasourceChangedOCP(ctx context.Context, log logr.Logger, dsID string) error {
	if client.libraryDatasourceChangedOCP == "" {
		log.Info("LibraryDatasourceChangedOCP webhook not configured")
		return nil
	}

	body := map[string]string{"datasourceID": dsID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Error(err, "Failed to marshal request body")
		return errors.Wrap(err, "Failed to marshal request body")
	}

	err = client.newRequest(ctx, log, client.libraryDatasourceChangedOCP, jsonData)
	if err != nil {
		log.Error(err, "Failed to create request to webhook")
		return errors.Wrap(err, "Failed to create request to webhook")
	}

	log.Info("Called library webhook successfully for OCP")
	return nil
}

// SystemDatasourceChangedOCP notifies the webhook that a system datasource has changed in OCP.
func (client *client) SystemDatasourceChangedOCP(ctx context.Context, log logr.Logger, dsID string) error {
	if client.systemDatasourceChangedOCP == "" {
		log.Info("SystemDatasourceChangedOCP webhook not configured")
		return nil
	}

	body := map[string]string{"datasourceId": dsID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Error(err, "Failed to marshal request body")
		return errors.Wrap(err, "Failed to marshal request body")
	}

	err = client.newRequest(ctx, log, client.systemDatasourceChangedOCP, jsonData)
	if err != nil {
		log.Error(err, "Failed to create request to webhook")
		return errors.Wrap(err, "Failed to create request to webhook")
	}

	log.Info("Called system webhook successfully for OCP")
	return nil
}

// LibraryDatasourceChanged notifies the webhook that a library datasource has changed.
func (client *client) LibraryDatasourceChanged(ctx context.Context, log logr.Logger, datasourceID string) error {
	if client.libraryDatasourceChanged == "" {
		log.Info("LibraryDatasourceChanged webhook not configured")
		return nil
	}

	body := map[string]string{"datasourceID": datasourceID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Error(err, "Failed to marshal request body")
		return errors.Wrap(err, "Failed to marshal request body")
	}

	err = client.newRequest(ctx, log, client.libraryDatasourceChanged, jsonData)
	if err != nil {
		log.Error(err, "Failed to create request to webhook")
		return errors.Wrap(err, "Failed to create request to webhook")
	}

	log.Info("Called library webhook successfully")
	return nil
}

// SystemDatasourceChanged notifies the webhook that a system datasource has changed.
func (client *client) SystemDatasourceChanged(
	ctx context.Context,
	log logr.Logger,
	systemID string,
	dsID string) error {
	if client.systemDatasourceChanged == "" {
		log.Info("systemDatasourceChanged webhook not configured")
		return nil
	}

	body := map[string]string{"systemId": systemID, "datasourceId": dsID}
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Error(err, "Failed to marshal request body")
		return errors.Wrap(err, "Failed to marshal request body")
	}

	err = client.newRequest(ctx, log, client.systemDatasourceChanged, jsonData)
	if err != nil {
		log.Error(err, "Failed to create request to webhook")
		return errors.Wrap(err, "Failed to create request to webhook")
	}

	log.Info("Called system webhook successfully")
	return nil
}

func (client *client) newRequest(ctx context.Context, log logr.Logger, clientType string, jsonData []byte) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, clientType, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(err, "Failed to create request to webhook")
		return errors.Wrap(err, "Failed to create request to webhook")
	}
	r.Header.Set("Content-Type", "application/json")

	resp, err := client.hc.Do(r)
	if err != nil {
		log.Error(err, "Failed in call to webhook")
		return errors.Wrap(err, "Failed in call to webhook")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Info("Response status code is not 2XX")
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Could not read response body")
			return errors.Errorf("Could not read response body")
		}
		bodyString := string(bodyBytes)
		return errors.Errorf("response status code is %d, response body is %s", resp.StatusCode, bodyString)
	}
	return nil
}
