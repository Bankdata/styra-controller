package s3

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/madmin-go/v3"
	miniocred "github.com/minio/minio-go/v7/pkg/credentials"
)

type awsS3Client struct {
	client *s3.Client
	config Config
}

func NewAWSS3Client(cfg Config) (S3Client, error) {
	ctx := context.TODO()

	// Create AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"", // session token (empty)
		)),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom options
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			// Custom endpoint for MinIO or other S3-compatible storage
			scheme := "http"
			if cfg.UseSSL {
				scheme = "https"
			}
			o.BaseEndpoint = aws.String(fmt.Sprintf("%s://%s", scheme, cfg.Endpoint))
		}
		if cfg.PathStyle {
			o.UsePathStyle = true
		}
	})

	return &awsS3Client{
		client: client,
		config: cfg,
	}, nil
}

// TODO support AWS S3
// UserExists checks if a user exists in MinIO
func (c *awsS3Client) ExistsUser(ctx context.Context, accessKey string) (bool, error) {
	// Create MinIO admin client
	adminClient, err := madmin.NewWithOptions(c.config.Endpoint, &madmin.Options{
		Creds:  miniocred.NewStaticV4(c.config.AccessKeyID, c.config.SecretAccessKey, ""),
		Secure: c.config.UseSSL,
	})
	if err != nil {
		return false, fmt.Errorf("failed to create MinIO admin client: %w", err)
	}

	// Get user info
	_, err = adminClient.GetUserInfo(ctx, accessKey)
	if err != nil {
		if strings.Contains(err.Error(), "The specified user does not exist") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get user info for %s: %w", accessKey, err)
	}

	return true, nil
}

// TODO support AWS S3
// CreateSystemBundleUser creates a user in MinIO with read-only access to a specific bucketPath
func (c *awsS3Client) CreateSystemBundleUser(ctx context.Context, accessKey string, bucketName string, uniqueName string) (string, error) {
	// Create MinIO admin client
	adminClient, err := madmin.NewWithOptions(c.config.Endpoint, &madmin.Options{
		Creds:  miniocred.NewStaticV4(c.config.AccessKeyID, c.config.SecretAccessKey, ""),
		Secure: c.config.UseSSL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create MinIO admin client: %w", err)
	}

	// Create the user
	secretKey, err := generateBase64Secret(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret key: %w", err)
	}

	// Add the user
	err = adminClient.AddUser(ctx, accessKey, secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create user %s: %w", accessKey, err)
	}

	// Create a read-only policy for the specific bucket
	policyName := fmt.Sprintf("readonly-%s-%s", bucketName, uniqueName)
	policyDocument := fmt.Sprintf(`{
	    "Version": "2012-10-17",
	    "Statement": [
	        {
	            "Effect": "Allow",
	            "Action": [
	                "s3:GetObject"
	            ],
	            "Resource": [
	                "arn:aws:s3:::%s/bundles/%s/*"
	            ]
	        },
	        {
	            "Effect": "Allow",
	            "Action": [
	                "s3:ListBucket"
	            ],
	            "Resource": [
	                "arn:aws:s3:::%s"
	            ],
	            "Condition": {
	                "StringLike": {
	                    "s3:prefix": [
	                        "bundles/%s/*"
	                    ]
	                }
	            }
	        }
	    ]
	}`, bucketName, uniqueName, bucketName, uniqueName)

	// Add the policy
	err = adminClient.AddCannedPolicy(ctx, policyName, []byte(policyDocument))
	if err != nil {
		return "", fmt.Errorf("failed to create policy %s: %w", policyName, err)
	}

	// Attach the policy to the user
	_, err = adminClient.AttachPolicy(ctx, madmin.PolicyAssociationReq{
		Policies: []string{policyName},
		User:     accessKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach policy %s to user %s: %w", policyName, accessKey, err)
	}

	return secretKey, nil
}

func (c *awsS3Client) SetNewUserSecretKey(ctx context.Context, accessKey string) (string, error) {
	// Create MinIO admin client
	adminClient, err := madmin.NewWithOptions(c.config.Endpoint, &madmin.Options{
		Creds:  miniocred.NewStaticV4(c.config.AccessKeyID, c.config.SecretAccessKey, ""),
		Secure: c.config.UseSSL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create MinIO admin client: %w", err)
	}

	// Update existing user with new secret key
	newSecretKey, err := generateBase64Secret(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate new secret key: %w", err)
	}

	err = adminClient.SetUser(ctx, accessKey, newSecretKey, madmin.AccountEnabled)
	if err != nil {
		return "", fmt.Errorf("failed to update user %s secret key: %w", accessKey, err)
	}
	return newSecretKey, nil
}

// Helper function to generate base64 encoded secret
func generateBase64Secret(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
