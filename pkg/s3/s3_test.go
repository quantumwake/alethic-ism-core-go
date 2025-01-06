package s3_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	store "github.com/quantumwake/alethic-ism-core-go/pkg/s3"
	"os"
	"testing"
)

const (
	bucketName = "test-bucket"
	region     = "us-west-2"
	endpoint = "https://ism.dev.sfo3.digitaloceanspaces.com"
)

func TestS3Client_UploadFile(t *testing.T) {

	data, err := os.ReadFile("testdata/transcript1.txt")
	if err != nil {
		t.Fatal(err)
	}


	client, err := store.NewS3Client("test-bucket", "us-west-2")
	if err != nil {
		t.Fatal(err)
	}

	err = client.UploadFile(context.Background(), "transcript1.txt", data)


	type fields struct {
		Client     *s3.Client
		BucketName string
	}
	type args struct {
		ctx      context.Context
		key      string
		filePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}

}