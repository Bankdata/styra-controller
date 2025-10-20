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

package ocp

import (
	"context"
	"io"
	"net/http"
	"path"

	"github.com/bankdata/styra-controller/pkg/httperror"
	"github.com/pkg/errors"
)

const (
	endpointV1Bundles = "/v1/bundles"
)

// BundleConfig represents the configuration of a bundle in the OCP APIs.
type BundleConfig struct {
	Name          string            `json:"-" yaml:"-"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	ObjectStorage ObjectStorage     `json:"object_storage,omitempty" yaml:"object_storage,omitempty"`
	Requirements  []Requirement     `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	ExcludedFiles []string          `json:"excluded_files,omitempty" yaml:"excluded_files,omitempty"`
}

// ObjectStorage represents the object storage configuration for a bundle.
type ObjectStorage struct {
	AmazonS3 *AmazonS3 `json:"aws,omitempty" yaml:"aws,omitempty"`
	// GCPCloudStorage   *GCPCloudStorage   `json:"gcp,omitempty" yaml:"gcp,omitempty"`
	// AzureBlobStorage  *AzureBlobStorage  `json:"azure,omitempty" yaml:"azure,omitempty"`
	// FileSystemStorage *FileSystemStorage `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
}

// AmazonS3 defines the configuration for a bundle stored in an Amazon S3 bucket.
type AmazonS3 struct {
	Bucket      string `json:"bucket" yaml:"bucket"`
	Key         string `json:"key" yaml:"key"`
	Region      string `json:"region,omitempty" yaml:"region,omitempty"`
	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

// GCPCloudStorage defines the configuration for a Google Cloud Storage bucket.
type GCPCloudStorage struct {
	Project     string `json:"project" yaml:"project"`
	Bucket      string `json:"bucket" yaml:"bucket"`
	Object      string `json:"object" yaml:"object"`
	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// AzureBlobStorage defines the configuration for an Azure Blob Storage container.
type AzureBlobStorage struct {
	AccountURL  string `json:"account_url" yaml:"account_url"`
	Container   string `json:"container" yaml:"container"`
	Path        string `json:"path" yaml:"path"`
	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// FileSystemStorage defines the configuration for a local filesystem storage.
type FileSystemStorage struct {
	Path string `json:"path" yaml:"path"` // Path to the bundle on the local filesystem.
}

// PutBundleRequest is the request body for the
// PUT /v1/bundles/{name} endpoint in the OCP API.
type PutBundleRequest struct {
	Name          string            `json:"-" yaml:"-"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	ObjectStorage ObjectStorage     `json:"object_storage,omitempty" yaml:"object_storage,omitempty"`
	Requirements  []Requirement     `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	ExcludedFiles []string          `json:"excluded_files,omitempty" yaml:"excluded_files,omitempty"`
}

// PutBundleResponse is the response type for calls to the
// PUT /v1/bundles/{name} endpoint in the OCP API.
type PutBundleResponse struct {
	StatusCode string `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// PutBundle calls the PUT /v1/bundles/{name} endpoint in the OCP API.
func (c *Client) PutBundle(ctx context.Context, bundle *PutBundleRequest) (err error) {
	res, err := c.request(ctx, http.MethodPut, path.Join(endpointV1Bundles, bundle.Name), bundle, nil)
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
		return errors.Wrap(err, "PutBundle: could not read body")
	}

	if res.StatusCode != http.StatusOK {
		return httperror.NewHTTPError(res.StatusCode, string(body))
	}
	return nil
}

// DeleteBundle calls the DELETE /v1/bundles/{name} endpoint in the OCP API.
func (c *Client) DeleteBundle(ctx context.Context, name string) (err error) {
	res, err := c.request(ctx, http.MethodDelete, path.Join(endpointV1Bundles, name), nil, nil)
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
		return errors.Wrap(err, "DeleteBundle: could not read body")
	}

	if res.StatusCode != http.StatusOK {
		return httperror.NewHTTPError(res.StatusCode, string(body))
	}
	return nil
}
