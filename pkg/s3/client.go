// Package s3 contains a client for interacting with S3 compatible object storage.
package s3

import (
	"strings"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
)

// NewClient creates a new S3Client for MinIO
func NewClient(s3Handler configv2alpha2.S3Handler) (Client, error) {
	config := Config{
		AccessKeyID:     s3Handler.AccessKeyID,
		SecretAccessKey: s3Handler.SecretAccessKey,
		Region:          s3Handler.Region,
		PathStyle:       s3Handler.URL != "", // Use path style for custom endpoints
	}

	if s3Handler.URL != "" && strings.HasPrefix(s3Handler.URL, "https://") {
		config.Endpoint = strings.TrimPrefix(s3Handler.URL, "https://")
		config.UseSSL = true
	} else {
		config.Endpoint = strings.TrimPrefix(s3Handler.URL, "http://")
		config.UseSSL = false
	}

	return newMinioClient(config)
}
