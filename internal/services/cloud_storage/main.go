package cloudstorage

import (
	"context"
	"fmt"

	"unipilot/internal/secrets"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	Region = "us-west-2"
	Bucket = "unipilot.bucket"
)

func S3Client() (*s3.Client, error) {

	region, err := secrets.GetEnvVar("AWS_REGION")
	if err != nil {
		return nil, err
	}

	// Load default config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))

	if err != nil {
		fmt.Println("Error loading default config:", err)
		return nil, err
	}

	// Create new S3 client
	svc := s3.NewFromConfig(cfg)

	return svc, nil
}
