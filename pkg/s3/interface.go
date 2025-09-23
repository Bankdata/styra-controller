package s3

import (
	"context"
)

// Client is an interface for clients interacting with S3 compatible object storage.
type Client interface {
	UserExists(ctx context.Context, accessKey string) (bool, error)
	CreateSystemBundleUser(ctx context.Context, accessKey, bucketName string, uniqueName string) (string, error)
	SetNewUserSecretKey(ctx context.Context, accessKey string) (string, error)
}

// Config defines the configuration for a S3 client.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UseSSL          bool
	PathStyle       bool // Required for MinIO and some S3-compatible storage
}

// Credentials represents S3 credentials.
type Credentials struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

// Constants for keys in secret generated for OPAs
const (
	AWSSecretNameKeyID     = "AWS_ACCESS_KEY_ID"
	AWSSecretNameSecretKey = "AWS_SECRET_ACCESS_KEY"
	AWSSecretNameRegion    = "AWS_REGION"
)
