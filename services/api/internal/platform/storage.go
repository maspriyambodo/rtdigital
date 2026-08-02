package platform

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maspriyambodo/rtdigital/services/api/internal/config"
)

type Storage struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

type PresignedURL struct {
	URL     string
	Headers map[string]string
}

type ObjectMetadata struct {
	SizeBytes   int64
	ContentType string
}

func NewStorage(ctx context.Context, cfg config.R2Config) (*Storage, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, _ string, _ ...interface{}) (aws.Endpoint, error) {
		if service != s3.ServiceID {
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		}

		return aws.Endpoint{
			URL:               cfg.Endpoint,
			SigningRegion:     "auto",
			HostnameImmutable: !cfg.UsePathStyle,
		}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("load storage configuration: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.UsePathStyle
	})

	return &Storage{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
	}, nil
}

func (s *Storage) PresignUpload(
	ctx context.Context,
	key string,
	contentType string,
	lifetime time.Duration,
) (PresignedURL, error) {
	request, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(lifetime))
	if err != nil {
		return PresignedURL{}, fmt.Errorf("presign upload for key %q: %w", key, err)
	}

	return PresignedURL{
		URL: request.URL,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
	}, nil
}

func (s *Storage) HeadObject(ctx context.Context, key string) (ObjectMetadata, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectMetadata{}, fmt.Errorf("head object for key %q: %w", key, err)
	}

	return ObjectMetadata{
		SizeBytes:   aws.ToInt64(result.ContentLength),
		ContentType: aws.ToString(result.ContentType),
	}, nil
}

func (s *Storage) PresignDownload(
	ctx context.Context,
	key string,
	lifetime time.Duration,
) (string, error) {
	request, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(lifetime))
	if err != nil {
		return "", fmt.Errorf("presign download for key %q: %w", key, err)
	}

	return request.URL, nil
}

func (s *Storage) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object for key %q: %w", key, err)
	}
	return nil
}
