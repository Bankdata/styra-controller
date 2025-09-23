package s3

import (
	"context"
)

type S3Client interface {
	ExistsUser(ctx context.Context, accessKey string) (bool, error)
	CreateSystemBundleUser(ctx context.Context, accessKey, bucketName string, uniqueName string) (string, error)
	SetNewUserSecretKey(ctx context.Context, accessKey string) (string, error)
}

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UseSSL          bool
	PathStyle       bool // Required for MinIO and some S3-compatible storage
}

type UserInfo struct {
	AccessKey  string `json:"accessKey"`
	Status     string `json:"status"`
	PolicyName string `json:"policyName"`
}

const (
	AWSSecretNameKeyID     = "AWS_ACCESS_KEY_ID"
	AWSSecretNameSecretKey = "AWS_SECRET_ACCESS_KEY"
	AWSSecretNameRegion    = "AWS_REGION"
)

type AWSCredentials struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
}
