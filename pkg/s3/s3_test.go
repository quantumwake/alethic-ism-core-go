package s3_test

import (
	"context"
	"github.com/quantumwake/alethic-ism-core-go/pkg/s3"
	"testing"
)

const (
	bucketName = "test-bucket"
	region     = "us-west-2"
	endpoint   = "https://ism.dev.sfo3.digitaloceanspaces.com"
)

func TestS3Client_UploadFile(t *testing.T) {
	ctx := context.Background()
	filePath := "testdata/transcript1.txt"
	//data, err := os.ReadFile("testdata/transcript1.txt")
	//if err != nil {
	//	t.Fatal(err)
	//}

	clientConfig := s3.ClientConfig{
		BucketName: bucketName,
		Region:     region,
	}

	client, err := s3.NewClient(ctx, clientConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = client.UploadFile(context.Background(), "transcript1.txt", filePath)

	type fields struct {
		Client     *s3.Client
		BucketName string
	}
	type args struct {
		ctx      context.Context
		key      string
		filePath string
	}
	//tests := []struct {
	//	name    string
	//	fields  fields
	//	args    args
	//	wantErr bool
	//}

}
