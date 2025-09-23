package s3

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	madmin "github.com/minio/madmin-go/v3"
	miniocred "github.com/minio/minio-go/v7/pkg/credentials"
)

type minioClient struct {
	config      Config
	adminClient *madmin.AdminClient
}

func newMinioClient(cfg Config) (Client, error) {
	// Create MinIO admin client
	adminClient, err := madmin.NewWithOptions(cfg.Endpoint, &madmin.Options{
		Creds:  miniocred.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO admin client: %w", err)
	}

	return &minioClient{
		adminClient: adminClient,
		config:      cfg,
	}, nil
}

// UserExists checks if a user exists in MinIO
func (c *minioClient) UserExists(ctx context.Context, accessKey string) (bool, error) {
	_, err := c.adminClient.GetUserInfo(ctx, accessKey)
	if err != nil {
		if strings.Contains(err.Error(), "The specified user does not exist") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get user info for %s: %w", accessKey, err)
	}

	return true, nil
}

// CreateSystemBundleUser creates a user in MinIO with read-only access to a specific bucketPath
func (c *minioClient) CreateSystemBundleUser(
	ctx context.Context,
	accessKey string,
	bucketName string,
	uniqueName string) (string, error) {
	secretKey, err := generateBase64Secret(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret key: %w", err)
	}

	err = c.adminClient.AddUser(ctx, accessKey, secretKey)
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

	err = c.adminClient.AddCannedPolicy(ctx, policyName, []byte(policyDocument))
	if err != nil {
		return "", fmt.Errorf("failed to create policy %s: %w", policyName, err)
	}

	_, err = c.adminClient.AttachPolicy(ctx, madmin.PolicyAssociationReq{
		Policies: []string{policyName},
		User:     accessKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach policy %s to user %s: %w", policyName, accessKey, err)
	}

	return secretKey, nil
}

func (c *minioClient) SetNewUserSecretKey(ctx context.Context, accessKey string) (string, error) {
	// Update existing user with new secret key
	newSecretKey, err := generateBase64Secret(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate new secret key: %w", err)
	}

	err = c.adminClient.SetUser(ctx, accessKey, newSecretKey, madmin.AccountEnabled)
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
