package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	region     string
	endpoint   string
}

func NewS3Storage(accessKey, secretKey, endpoint, region, bucketName string) (*S3Storage, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}

	if endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(resolver))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &S3Storage{
		client:     client,
		presigner:  presigner,
		bucketName: bucketName,
		region:     region,
		endpoint:   endpoint,
	}, nil
}

func (s *S3Storage) GeneratePresignedPutURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	request, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String("image/jpeg"), // Restrict to image/jpeg
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return request.URL, nil
}

func (s *S3Storage) GeneratePresignedGetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return request.URL, nil
}

func (s *S3Storage) GetPublicURL(key string) string {
	if s.endpoint != "" {
		// e.g., Supabase or MinIO (http://endpoint/bucket/key)
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucketName, key)
	}
	// AWS Native
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
}

func (s *S3Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %v", err)
	}

	return nil
}
