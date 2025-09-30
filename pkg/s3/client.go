package s3

import (
	configv2alpha2 "github.com/bankdata/styra-controller/api/config/v2alpha2"
)

func NewS3Client(awsObjectStorage configv2alpha2.AWSObjectStorage) (S3Client, error) {
	config := Config{
		AccessKeyID:     awsObjectStorage.AdminCredentials.AccessKeyID,
		SecretAccessKey: awsObjectStorage.AdminCredentials.SecretAccessKey,
		Region:          awsObjectStorage.Region,
		UseSSL:          awsObjectStorage.URL == "", // Use SSL for AWS, not for custom endpoints
		PathStyle:       awsObjectStorage.URL != "", // Use path style for custom endpoints
	}

	if awsObjectStorage.URL != "" {
		//TODO fix route
		//Strip http:
		config.Endpoint = "localhost:9000"
	}

	return NewAWSS3Client(config)
}
