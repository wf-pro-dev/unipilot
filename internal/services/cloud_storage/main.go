package cloudstorage

import (
	"os"
	"context"
	"fmt"
	"log"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

const (
	Region = "us-west-2"
	Bucket = "unipilot.bucket"
)

func S3Client() (*s3.Client, error) {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}


	// Load default config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))

	if err != nil {
		fmt.Println("Error loading default config:", err)
		return nil, err
	}

	// Create new S3 client
	svc := s3.NewFromConfig(cfg)

	return svc, nil
}
