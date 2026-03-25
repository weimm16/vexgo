package handler

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"vexgo/backend/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

var (
	S3Client     *minio.Client    // MinIO client instance
	S3Cfg        *config.S3Config // S3 configuration reference
	UseS3Storage bool             // Flag indicating whether S3 storage is active
)

// InitS3 initializes the MinIO client and verifies the connection.
// Returns an error if configuration is invalid or the bucket is unreachable.
func InitS3(cfg *config.S3Config) error {
	logrus.WithField("enabled", cfg.Enabled).Debug("Initializing S3 storage")

	if !cfg.IsEnabled() {
		UseS3Storage = false
		logrus.Info("S3 storage disabled, using local file storage")
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
		"region":   cfg.Region,
		"bucket":   cfg.Bucket,
	}).Debug("Validating S3 configuration")

	if err := cfg.Validate(); err != nil {
		logrus.WithError(err).Error("S3 configuration validation failed")
		return fmt.Errorf("S3 configuration error: %w", err)
	}

	logrus.Debug("S3 configuration validation passed")

	// Strip protocol prefix from endpoint, minio-go manages SSL separately
	endpoint := cfg.Endpoint
	useSSL := true
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		useSSL = false
		logrus.WithField("endpoint", endpoint).Debug("Using HTTP for S3 connection")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
		useSSL = true
		logrus.WithField("endpoint", endpoint).Debug("Using HTTPS for S3 connection")
	}
	endpoint = strings.TrimSuffix(endpoint, "/")

	logrus.WithField("endpoint", endpoint).Debug("Creating MinIO client")

	// Create MinIO client with static credentials
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": endpoint,
			"region":   cfg.Region,
		}).WithError(err).Error("Failed to create MinIO client")
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"region":   cfg.Region,
	}).Debug("MinIO client created successfully")

	// Verify connectivity by checking if the target bucket exists
	logrus.WithField("bucket", cfg.Bucket).Debug("Checking if S3 bucket exists")
	exists, err := client.BucketExists(context.TODO(), cfg.Bucket)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"bucket": cfg.Bucket,
		}).WithError(err).Error("Failed to check S3 bucket existence")
		return fmt.Errorf("failed to connect to S3: %w", err)
	}
	if !exists {
		logrus.WithField("bucket", cfg.Bucket).Error("S3 bucket does not exist")
		return fmt.Errorf("bucket %s does not exist", cfg.Bucket)
	}
	logrus.WithField("bucket", cfg.Bucket).Info("S3 bucket exists and is accessible")

	S3Client = client
	S3Cfg = cfg
	UseS3Storage = true

	logrus.WithField("bucket", cfg.Bucket).Info("S3 storage initialized successfully")
	return nil
}

// UploadFileToS3 uploads a file to the configured S3 bucket.
// Passing size as -1 lets minio-go handle multipart upload automatically.
// Returns the public URL of the uploaded file.
func UploadFileToS3(reader io.Reader, filename string, contentType string) (string, error) {
	if S3Client == nil {
		return "", fmt.Errorf("S3 storage not initialized")
	}

	// Fall back to extension-based detection if content type is not provided
	if contentType == "" {
		contentType = detectContentType(filename)
	}

	_, err := S3Client.PutObject(context.TODO(), S3Cfg.Bucket, filename, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	url := S3Cfg.GetURL(filename)
	fmt.Printf("Uploaded file, generated URL: %s\n", url)
	return url, nil
}

// DeleteFileFromS3 removes an object from the configured S3 bucket by key.
func DeleteFileFromS3(key string) error {
	if S3Client == nil {
		return fmt.Errorf("S3 storage not initialized")
	}

	err := S3Client.RemoveObject(context.TODO(), S3Cfg.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetFileURL constructs the public URL for a stored object.
// Priority: CustomDomain > path-style endpoint > virtual-hosted style.
func GetFileURL(key string) string {
	if S3Cfg == nil {
		return ""
	}

	// Use custom domain if configured (e.g. CDN domain)
	if S3Cfg.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", S3Cfg.CustomDomain, key)
	}

	if S3Cfg.ForcePath {
		// Path-style URL: https://endpoint/bucket/key
		endpoint := S3Cfg.Endpoint
		if !strings.HasPrefix(endpoint, "http") {
			endpoint = "https://" + endpoint
		}
		endpoint = strings.TrimSuffix(endpoint, "/")
		return fmt.Sprintf("%s/%s/%s", endpoint, S3Cfg.Bucket, key)
	}

	endpoint := S3Cfg.Endpoint
	if endpoint == "" {
		// Standard AWS virtual-hosted style URL
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3Cfg.Bucket, S3Cfg.Region, key)
	}

	// Virtual-hosted style for custom endpoints: https://bucket.endpoint/key
	endpoint = strings.Split(endpoint, ":")[0] // strip port if present
	return fmt.Sprintf("https://%s.%s/%s", S3Cfg.Bucket, endpoint, key)
}

// detectContentType returns the MIME type based on the file extension.
// Defaults to application/octet-stream for unknown types.
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
