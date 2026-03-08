package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// FileEntry represents a single object in S3.
type FileEntry struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}

type Client struct {
	Client     *s3.Client
	BucketName string
}

type ClientConfig struct {
	BucketName string
	Region     string
	Endpoint   string
	AccessKey  string
	SecretKey  string
}

func NewClient(ctx context.Context, clientConfig ClientConfig) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(clientConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			clientConfig.AccessKey,
			clientConfig.SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if clientConfig.Endpoint != "" {
			o.BaseEndpoint = &clientConfig.Endpoint
		}
	})

	return &Client{
		Client:     client,
		BucketName: clientConfig.BucketName,
	}, nil
}

func (s *Client) UploadFile(ctx context.Context, key string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.BucketName,
		Key:    &key,
		Body:   file,
	})
	return err
}

func (s *Client) DownloadFile(ctx context.Context, key string, destPath string) error {
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	defer output.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, output.Body)
	return err
}

func (s *Client) UploadBytes(ctx context.Context, key string, data []byte) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.BucketName,
		Key:    &key,
		Body:   bytes.NewReader(data),
	})
	return err
}

func (s *Client) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(output.Body)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetObjectStream returns a streaming reader for the given S3 key.
// The caller is responsible for closing the returned ReadCloser.
func (s *Client) GetObjectStream(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}
	return output.Body, nil
}

// ListObjects lists all objects under the given prefix, with optional glob pattern filtering.
// Pattern uses filepath.Match syntax (e.g. "*.csv", "*.{csv,xlsx}").
// Pass empty pattern to list all objects under the prefix.
func (s *Client) ListObjects(ctx context.Context, prefix, pattern string) ([]FileEntry, error) {
	var entries []FileEntry

	input := &s3.ListObjectsV2Input{
		Bucket: &s.BucketName,
		Prefix: &prefix,
	}

	// Paginate through all results
	paginator := s3.NewListObjectsV2Paginator(s.Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
		}

		for _, obj := range page.Contents {
			key := ""
			if obj.Key != nil {
				key = *obj.Key
			}

			// Apply glob pattern filter if specified
			if pattern != "" {
				baseName := filepath.Base(key)
				matched, err := filepath.Match(pattern, baseName)
				if err != nil {
					return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
				}
				if !matched {
					continue
				}
			}

			entry := FileEntry{
				Key:  key,
				Size: *obj.Size,
			}
			if obj.LastModified != nil {
				entry.LastModified = *obj.LastModified
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
