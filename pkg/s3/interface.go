package s3

import (
	"context"
)

type S3Client interface {
	ExistsUser(ctx context.Context, accessKey string) (bool, error)
	CreateSystemBundleUser(ctx context.Context, accessKey, secretKey, bucketName string, uniqueName string) error
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
