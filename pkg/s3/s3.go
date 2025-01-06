package s3

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	Client     *s3.Client
	BucketName string
}

func NewS3Client(bucketName string, region string, endpoint string, accessKey string, secretKey string) (*S3Client, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, errors.New("endpoint, access key, and secret key are required")
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint, // DigitalOcean Spaces endpoint
			SigningRegion: region,   // Region is required for signing
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Path-style addressing for DigitalOcean Spaces
	})

	return &S3Client{
		Client:     client,
		BucketName: bucketName,
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
