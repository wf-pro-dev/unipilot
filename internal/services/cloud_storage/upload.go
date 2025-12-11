package cloudstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func UploadFile(filePath, fileName, key string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file info: %w", err)
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
		Key:    aws.String(key), // format : user_id/assignment_id/file_name
		Body:   bytes.NewReader(buf.Bytes()),
		Metadata: map[string]string{
			"original-name": fileName,
			"upload-time":   time.Now().Format(time.RFC3339),
			"file-size":     fmt.Sprintf("%d", fileInfo.Size()),
		},
	})

	if err != nil {
		return fmt.Errorf("r2 operation failed: %w", err)
	}

	return nil
}

func UploadProfilePicture(filePath, fileName, key string) (string, error) {
	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return "", err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to get file info: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		return "", err
	}

	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
		Metadata: map[string]string{
			"original-name": fileName,
			"upload-time":   time.Now().Format(time.RFC3339),
			"file-size":     fmt.Sprintf("%d", fileInfo.Size()),
		},
	})
	if err != nil {
		return "", err
	}

	log.Printf("File uploaded to S3: %s", key)

	// Construct and return the public URL
	publicURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", Region, Bucket, key)
	log.Printf("Public URL: %s", publicURL)
	return publicURL, nil
}

func DeleteFile(key string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return err
	}

	// Upload file
	_, err = svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &noKey) {
			log.Printf("Object %s does not exist in %s.\n", key, Bucket)
			err = noKey
		} else if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				log.Printf("Access denied: cannot delete object %s from %s.\n", key, Bucket)
				err = nil
			}
		}
	}

	return nil
}

func CopyFile(oldKey, newKey string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return err
	}

	_, err = svc.CopyObject(context.TODO(), &s3.CopyObjectInput{
		CopySource: aws.String(fmt.Sprintf("%s/%s", Bucket, oldKey)),
		Bucket:     aws.String(Bucket),
		Key:        aws.String(newKey),
	})
	if err != nil {
		return err
	}

	return nil

}

func DownloadFile(key string) (io.Reader, error) {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return nil, err
	}

	// Download object
	result, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil

}
