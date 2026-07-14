package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type EvidenceStore struct {
	s3Client    *s3.Client
	bucket      string
	evidenceDir string
}

type EvidenceObject struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}

func NewEvidenceStore(cfg Config) *EvidenceStore {
	client := s3.New(s3.Options{
		Region:       cfg.S3Region,
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretKey, "")),
		UsePathStyle: cfg.S3ForcePath,
	})
	return &EvidenceStore{
		s3Client:    client,
		bucket:      cfg.S3Bucket,
		evidenceDir: cfg.EvidenceDir,
	}
}

func (s *EvidenceStore) Open(ctx context.Context, evidence Evidence) (*EvidenceObject, error) {
	parsed, err := url.Parse(evidence.URI)
	if err != nil {
		return nil, fmt.Errorf("parse evidence URI: %w", err)
	}

	switch parsed.Scheme {
	case "s3":
		return s.openS3(ctx, evidence, parsed)
	case "file":
		return s.openFile(evidence, parsed)
	default:
		return nil, fmt.Errorf("evidence URI scheme %q is not downloadable", parsed.Scheme)
	}
}

func (s *EvidenceStore) openS3(ctx context.Context, evidence Evidence, parsed *url.URL) (*EvidenceObject, error) {
	if parsed.Host != s.bucket {
		return nil, fmt.Errorf("evidence bucket does not match configured bucket")
	}
	key := strings.TrimPrefix(parsed.Path, "/")
	if key == "" {
		return nil, fmt.Errorf("evidence object key is empty")
	}

	output, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get S3 evidence object: %w", err)
	}

	contentType := metadataString(evidence.Metadata, "content_type")
	if contentType == "" && output.ContentType != nil {
		contentType = *output.ContentType
	}
	contentLength := int64(-1)
	if output.ContentLength != nil {
		contentLength = *output.ContentLength
	}
	return &EvidenceObject{
		Body:          output.Body,
		ContentType:   defaultContentType(contentType),
		ContentLength: metadataInt64(evidence.Metadata, "size_bytes", contentLength),
		Filename:      evidenceFilename(evidence, key),
	}, nil
}

func (s *EvidenceStore) openFile(evidence Evidence, parsed *url.URL) (*EvidenceObject, error) {
	if parsed.Path == "" {
		return nil, fmt.Errorf("evidence file path is empty")
	}
	evidenceRoot, err := filepath.Abs(s.evidenceDir)
	if err != nil {
		return nil, fmt.Errorf("resolve evidence directory: %w", err)
	}
	objectPath, err := filepath.Abs(parsed.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve evidence file path: %w", err)
	}
	rel, err := filepath.Rel(evidenceRoot, objectPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("evidence file is outside configured evidence directory")
	}

	file, err := os.Open(objectPath)
	if err != nil {
		return nil, fmt.Errorf("open evidence file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat evidence file: %w", err)
	}

	return &EvidenceObject{
		Body:          file,
		ContentType:   defaultContentType(metadataString(evidence.Metadata, "content_type")),
		ContentLength: metadataInt64(evidence.Metadata, "size_bytes", info.Size()),
		Filename:      evidenceFilename(evidence, objectPath),
	}, nil
}

func evidenceFilename(evidence Evidence, fallbackPath string) string {
	if filename := sanitizeFilename(metadataString(evidence.Metadata, "filename")); filename != "" {
		return filename
	}
	if filename := sanitizeFilename(path.Base(strings.ReplaceAll(fallbackPath, "\\", "/"))); filename != "" {
		return filename
	}
	return evidence.ID
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = path.Base(value)
	if value == "." || value == "/" {
		return ""
	}
	return value
}

func defaultContentType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "application/octet-stream"
	}
	return value
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func metadataInt64(metadata map[string]any, key string, fallback int64) int64 {
	if metadata == nil {
		return fallback
	}
	switch value := metadata[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
