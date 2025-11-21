package services

import (
	"admin-panel/internal/config"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type StorageService struct {
	s3Client  *s3.S3
	bucket    string
	accountID string
	publicURL string
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.Cloudflare.AccountID)

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("auto"),
		Endpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(
			cfg.Cloudflare.AccessKeyID,
			cfg.Cloudflare.SecretAccessKey,
			"",
		),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 session: %w", err)
	}

	return &StorageService{
		s3Client:  s3.New(sess),
		bucket:    cfg.Cloudflare.Bucket,
		accountID: cfg.Cloudflare.AccountID,
		publicURL: cfg.Cloudflare.PublicURL,
	}, nil
}

func (s *StorageService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s", uuid.New().String(), time.Now().Format("20060102150405"), ext)

	key := filename
	if folder != "" {
		key = fmt.Sprintf("%s/%s", folder, filename)
	}

	// Detect content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to R2
	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(fileBytes),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(fileBytes))),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return s.GetPublicURL(key), nil
}

func (s *StorageService) GetPublicURL(key string) string {
	// Jika ada custom public URL (domain sendiri)
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, key)
	}

	// Default R2 dev subdomain
	return fmt.Sprintf("https://pub-%s.r2.dev/%s", s.accountID, key)
}

func (s *StorageService) DeleteFile(key string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}

	return nil
}

func (s *StorageService) ExtractFilePathFromURL(url string) string {
	var prefix string
	if s.publicURL != "" {
		prefix = fmt.Sprintf("%s/", s.publicURL)
	} else {
		prefix = fmt.Sprintf("https://pub-%s.r2.dev/", s.accountID)
	}

	if len(url) > len(prefix) {
		return url[len(prefix):]
	}
	return ""
}
