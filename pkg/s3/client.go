package s3

import (
	"strings"

	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
)

func NewS3Client(awsObjectStorage configv2alpha2.AWSObjectStorage) (S3Client, error) {
	config := Config{
		AccessKeyID:     awsObjectStorage.AdminCredentials.AccessKeyID,
		SecretAccessKey: awsObjectStorage.AdminCredentials.SecretAccessKey,
		Region:          awsObjectStorage.Region,
		PathStyle:       awsObjectStorage.URL != "", // Use path style for custom endpoints
	}

	if awsObjectStorage.URL != "" && strings.HasPrefix(awsObjectStorage.URL, "https://") {
		config.Endpoint = strings.TrimPrefix(awsObjectStorage.URL, "https://")
		config.UseSSL = true
	} else {
		config.Endpoint = strings.TrimPrefix(awsObjectStorage.URL, "http://")
		config.UseSSL = false
	}

	return NewAWSS3Client(config)
}
