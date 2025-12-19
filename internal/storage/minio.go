package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	bucketName = "product-images"
)

// MinIOStorage handles object storage operations
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(endpoint, accessKey, secretKey string, useSSL bool) (*MinIOStorage, error) {
	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client: client,
		bucket: bucketName,
	}

	// Initialize bucket
	if err := storage.initBucket(); err != nil {
		return nil, fmt.Errorf("failed to initialize bucket: %w", err)
	}

	log.Printf("MinIO storage initialized successfully (bucket: %s)", bucketName)
	return storage, nil
}

// initBucket creates the bucket if it doesn't exist
func (s *MinIOStorage) initBucket() error {
	ctx := context.Background()

	// Check if bucket exists
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %w", err)
	}

	if !exists {
		// Create bucket
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("error creating bucket: %w", err)
		}
		log.Printf("Created bucket: %s", s.bucket)

		// Set bucket policy to public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, s.bucket)

		err = s.client.SetBucketPolicy(ctx, s.bucket, policy)
		if err != nil {
			log.Printf("Warning: Could not set bucket policy: %v", err)
		}
	}

	return nil
}

// UploadImage uploads an image to MinIO
func (s *MinIOStorage) UploadImage(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// Set cache control headers for optimal caching
	opts := minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=31536000, immutable", // 1 year cache
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	log.Printf("Uploaded image: %s (size: %d bytes)", objectName, size)
	return nil
}

// GetImageURL returns the URL for an image
func (s *MinIOStorage) GetImageURL(objectName string) string {
	// For local development, we'll use the direct URL
	// In production, you might want to use presigned URLs or a CDN
	return fmt.Sprintf("/images/%s", objectName)
}

// GetPresignedURL generates a presigned URL for temporary access
func (s *MinIOStorage) GetPresignedURL(objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GetObject retrieves an object from MinIO
func (s *MinIOStorage) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

// StatObject gets object metadata
func (s *MinIOStorage) StatObject(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to stat object: %w", err)
	}
	return info, nil
}

// DeleteImage deletes an image from MinIO
func (s *MinIOStorage) DeleteImage(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	log.Printf("Deleted image: %s", objectName)
	return nil
}

// ListImages lists all images in the bucket
func (s *MinIOStorage) ListImages(ctx context.Context, prefix string) ([]string, error) {
	var images []string

	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		images = append(images, object.Key)
	}

	return images, nil
}
