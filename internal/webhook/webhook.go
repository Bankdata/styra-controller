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
	DatasourceChanged(context.Context, logr.Logger, string, string) error
}

type client struct {
	hc  http.Client
	url string
}

// New creates a new webhook notification Client.
func New(url string) Client {
	return &client{
		hc:  http.Client{},
		url: url,
	}
}

// DatasourceChanged notifies the webhook that a datasource has changed.
func (client *client) DatasourceChanged(ctx context.Context, log logr.Logger, systemID string, dsID string) error {

	body := map[string]string{"systemId": systemID, "datasourceId": dsID}
	jsonData, err := json.Marshal(body)

	if err != nil {
		log.Error(err, "Failed to marshal request body")
		return errors.Wrap(err, "Failed to marshal request body")
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, client.url, bytes.NewBuffer(jsonData))
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
		return errors.Errorf("response status code is %d, request body is %s", resp.StatusCode, bodyString)
	}

	log.Info("Called webhook successfully")
	return nil
}
