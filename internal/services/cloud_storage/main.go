package cloudstorage

import (
	"context"

	"unipilot/internal/errors"
	"unipilot/internal/secrets"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	Region    = "us-west-2"
	Bucket    = "unipilot"
	BucketURL = "https://assets.wwwill.xyz"

	PrivateBucket = "unipilot-private"
)

func S3Client() (*s3.Client, error) {

	region, err := secrets.GetEnvVar("AWS_REGION")
	if err != nil {
		return nil, errors.Inherit(err, errors.ConfigEnvVarNotFound)
	}

	// Load default config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))

	if err != nil {
		return nil, errors.Wrap(err, errors.StorageClientFailed, "Failed to load AWS config")
	}

	// Create new S3 client
	svc := s3.NewFromConfig(cfg)

	return svc, nil
}
