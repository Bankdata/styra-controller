package ocp

import (
	"context"
	"io"
	"net/http"
	"path"

	"github.com/bankdata/styra-controller/pkg/http_error"
	"github.com/pkg/errors"
)

const (
	endpointV1Bundles = "/v1/bundles"
)

// BundleConfig represents the configuration of a bundle in the Styra APIs.
type BundleConfig struct {
	Name          string            `json:"-" yaml:"-"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	ObjectStorage ObjectStorage     `json:"object_storage,omitempty" yaml:"object_storage,omitempty"`
	Requirements  []Requirement     `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	ExcludedFiles []string          `json:"excluded_files,omitempty" yaml:"excluded_files,omitempty"`
}

type ObjectStorage struct {
	AmazonS3 *AmazonS3 `json:"aws,omitempty" yaml:"aws,omitempty"`
	// GCPCloudStorage   *GCPCloudStorage   `json:"gcp,omitempty" yaml:"gcp,omitempty"`
	// AzureBlobStorage  *AzureBlobStorage  `json:"azure,omitempty" yaml:"azure,omitempty"`
	// FileSystemStorage *FileSystemStorage `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
}

type AmazonS3 struct {
	Bucket      string `json:"bucket" yaml:"bucket"`
	Key         string `json:"key" yaml:"key"`
	Region      string `json:"region,omitempty" yaml:"region,omitempty"`
	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
}

// // GCPCloudStorage defines the configuration for a Google Cloud Storage bucket.
// type GCPCloudStorage struct {
// 	Project     string `json:"project" yaml:"project"`
// 	Bucket      string `json:"bucket" yaml:"bucket"`
// 	Object      string `json:"object" yaml:"object"`
// 	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
// }

// // AzureBlobStorage defines the configuration for an Azure Blob Storage container.
// type AzureBlobStorage struct {
// 	AccountURL  string `json:"account_url" yaml:"account_url"`
// 	Container   string `json:"container" yaml:"container"`
// 	Path        string `json:"path" yaml:"path"`
// 	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
// }

// // FileSystemStorage defines the configuration for a local filesystem storage.
// type FileSystemStorage struct {
// 	Path string `json:"path" yaml:"path"` // Path to the bundle on the local filesystem.
// }

// PutBundleRequest is the request body for the
// PUT /v1/bundles/{name} endpoint in the Styra API.
type PutBundleRequest struct {
	Name          string            `json:"-" yaml:"-"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	ObjectStorage ObjectStorage     `json:"object_storage,omitempty" yaml:"object_storage,omitempty"`
	Requirements  []Requirement     `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	ExcludedFiles []string          `json:"excluded_files,omitempty" yaml:"excluded_files,omitempty"`
}

// PutBundleResponse is the response type for calls to the
// PUT /v1/bundles/{name} endpoint in the Styra API.
type PutBundleResponse struct {
	StatusCode string `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (r *Client) PutBundle(ctx context.Context, bundle *PutBundleRequest) error {
	res, err := r.request(ctx, http.MethodPut, path.Join(endpointV1Bundles, bundle.Name), bundle, nil)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "PutBundle: could not read body")
	}

	if res.StatusCode != http.StatusOK {
		return http_error.NewHTTPError(res.StatusCode, string(body))
	}
	return nil
}
