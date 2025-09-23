package ocp

import (
	"context"
	"net/http"
	"path"
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
	AmazonS3          *AmazonS3          `json:"aws,omitempty" yaml:"aws,omitempty"`
	GCPCloudStorage   *GCPCloudStorage   `json:"gcp,omitempty" yaml:"gcp,omitempty"`
	AzureBlobStorage  *AzureBlobStorage  `json:"azure,omitempty" yaml:"azure,omitempty"`
	FileSystemStorage *FileSystemStorage `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
}

type AmazonS3 struct {
	Bucket      string     `json:"bucket" yaml:"bucket"`
	Key         string     `json:"key" yaml:"key"`
	Region      string     `json:"region,omitempty" yaml:"region,omitempty"`
	Credentials *SecretRef `json:"credentials,omitempty" yaml:"credentials,omitempty"` // If nil, use default credentials chain: environment variables,
	// shared credentials file, ECS or EC2 instance role. More details in s3.go.
	URL string `json:"url,omitempty" yaml:"url,omitempty"` // for test purposes
}

// GCPCloudStorage defines the configuration for a Google Cloud Storage bucket.
type GCPCloudStorage struct {
	Project     string     `json:"project" yaml:"project"`
	Bucket      string     `json:"bucket" yaml:"bucket"`
	Object      string     `json:"object" yaml:"object"`
	Credentials *SecretRef `json:"credentials,omitempty" yaml:"credentials,omitempty"` // If nil, use default credentials chain: environment variables,
	// file created by gcloud auth application-default login, GCE/GKE metadata server. More details in s3.go.
}

// AzureBlobStorage defines the configuration for an Azure Blob Storage container.
type AzureBlobStorage struct {
	AccountURL  string     `json:"account_url" yaml:"account_url"`
	Container   string     `json:"container" yaml:"container"`
	Path        string     `json:"path" yaml:"path"`
	Credentials *SecretRef `json:"credentials,omitempty" yaml:"credentials,omitempty"` // If nil, use default credentials chain: environment variables,
	// managed identity, Azure CLI login. More details in s3.go.
}

// FileSystemStorage defines the configuration for a local filesystem storage.
type FileSystemStorage struct {
	Path string `json:"path" yaml:"path"` // Path to the bundle on the local filesystem.
}

// PutBundleRequest is the request body for the
// PUT /v1/bundles/{name} endpoint in the Styra API.
type PutBundleRequest struct {
	Bundle BundleConfig `json:"bundle"`
}

// PutBundleResponse is the response type for calls to the
// PUT /v1/bundles/{name} endpoint in the Styra API.
type PutBundleResponse struct {
	ID string `json:"id"`
}

func (r *Client) PutBundle(ctx context.Context, bundle *BundleConfig) (*PutBundleResponse, error) {
	_, err := r.request(ctx, http.MethodPut, path.Join(endpointV1Bundles, bundle.Name), &PutBundleRequest{
		Bundle: *bundle,
	}, nil)
	if err != nil {
		return nil, err
	}
	return &PutBundleResponse{ID: "res.ID"}, nil
}
