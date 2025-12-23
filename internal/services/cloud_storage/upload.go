package cloudstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	Errors "errors"
	"time"

	"unipilot/internal/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func UploadFile(filePath, fileName, key string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return errors.Wrap(err, errors.StorageClientFailed, "Failed to get S3 client")
	}

	file, err := os.Open(filePath)
	if err != nil {
		if Errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, errors.StorageFileNotFound, "File not found")
		}
		return errors.Wrap(err, errors.FSOpenFailed, "Unable to open file")
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, errors.InternalError, "Unable to get file info")
	}

	// Read the contents of the file into a buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {

		return errors.Wrap(err, errors.InternalError, "Unable to read file")
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
		var apiErr *smithy.GenericAPIError
		if Errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return errors.Wrap(err, errors.AuthForbidden, "Access denied")
			}
			return errors.Wrap(err, errors.StorageApiFailed, "Failed to copy object")
		}

	}

	return nil
}

func UploadProfilePicture(filePath, fileName, key string) (string, error) {
	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return "", errors.Wrap(err, errors.StorageClientFailed, "Failed to get S3 client")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrap(err, errors.StorageFileNotFound, "Unable to open file")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", errors.Wrap(err, errors.InternalError, "Unable to get file info")
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return "", errors.Wrap(err, errors.InternalError, "Unable to read file")
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
		var apiErr *smithy.GenericAPIError
		if Errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return "", errors.Wrap(err, errors.AuthForbidden, "Access denied")
			}
			return "", errors.Wrap(err, errors.StorageApiFailed, "Failed to copy object")
		}

	}

	// Construct and return the public URL
	publicURL := fmt.Sprintf("https://assets.wwwill.xyz/%s", key)

	return publicURL, nil
}

func DeleteFile(key string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return errors.Wrap(err, errors.StorageClientFailed, "Failed to get S3 client")
	}

	// Upload file
	_, err = svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if Errors.As(err, &noKey) {
			return errors.Wrap(err, errors.StorageFileNotFound, "Object not found")
		} else if Errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return errors.Wrap(err, errors.AuthForbidden, "Access denied")
			}
			return errors.Wrap(err, errors.StorageApiFailed, "Failed to copy object")
		}

	}

	return nil
}

func CopyFile(oldKey, newKey string) error {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return errors.Wrap(err, errors.StorageClientFailed, "Failed to get S3 client")
	}

	_, err = svc.CopyObject(context.TODO(), &s3.CopyObjectInput{
		CopySource: aws.String(fmt.Sprintf("%s/%s", Bucket, oldKey)),
		Bucket:     aws.String(Bucket),
		Key:        aws.String(newKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if Errors.As(err, &noKey) {
			return errors.Wrap(err, errors.StorageFileNotFound, "Object not found")
		} else if Errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return errors.Wrap(err, errors.AuthForbidden, "Access denied")
			}
			return errors.Wrap(err, errors.StorageApiFailed, "Failed to copy object")
		}

	}

	return nil

}

func DownloadFile(key string) (io.Reader, error) {

	// Get S3 client
	svc, err := S3Client()
	if err != nil {
		return nil, errors.Wrap(err, errors.StorageClientFailed, "Failed to get S3 client")
	}

	// Download object
	result, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if Errors.As(err, &noKey) {
			return nil, errors.Wrap(err, errors.StorageFileNotFound, "Object not found")
		} else if Errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return nil, errors.Wrap(err, errors.AuthForbidden, "Access denied")
			}
			return nil, errors.Wrap(err, errors.StorageApiFailed, "Failed to copy object")
		}

	}

	return result.Body, nil

}
