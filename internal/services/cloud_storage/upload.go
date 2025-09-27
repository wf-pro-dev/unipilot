package cloudstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFile(file *os.File, key string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return err
	}

	// Read the contents of the file into a buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		return err
	}

	// Upload file
	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key), // format : user_id/assignment_id/document_id/file_name
		Body:   bytes.NewReader(buf.Bytes()),
	})

	if err != nil {
		return err
	}

	return nil
}
