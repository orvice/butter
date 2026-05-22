package agentfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ContentStore persists file version bodies outside the metadata repository.
type ContentStore interface {
	Put(ctx context.Context, key, content, contentType string) error
	Get(ctx context.Context, key string) (content string, contentType string, err error)
	Delete(ctx context.Context, keys []string) error
}

type S3Client interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type s3ContentStore struct {
	bucket    string
	keyPrefix string
	client    S3Client
}

func NewS3ContentStore(bucket string, client S3Client, keyPrefix string) ContentStore {
	return &s3ContentStore{
		bucket:    bucket,
		client:    client,
		keyPrefix: strings.Trim(keyPrefix, "/"),
	}
}

func (s *s3ContentStore) fullKey(key string) string {
	key = strings.TrimLeft(key, "/")
	if s.keyPrefix == "" {
		return key
	}
	return s.keyPrefix + "/" + key
}

func (s *s3ContentStore) Put(ctx context.Context, key, content, contentType string) error {
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	fullKey := s.fullKey(key)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fullKey),
		Body:        bytes.NewReader([]byte(content)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put agent file content %q: %w", fullKey, err)
	}
	return nil
}

func (s *s3ContentStore) Get(ctx context.Context, key string) (string, string, error) {
	fullKey := s.fullKey(key)
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return "", "", fmt.Errorf("get agent file content %q: %w", fullKey, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read agent file content %q: %w", fullKey, err)
	}
	return string(data), aws.ToString(resp.ContentType), nil
}

func (s *s3ContentStore) Delete(ctx context.Context, keys []string) error {
	for chunk := range chunks(keys, 1000) {
		if len(chunk) == 0 {
			continue
		}
		objects := make([]types.ObjectIdentifier, 0, len(chunk))
		for _, key := range chunk {
			objects = append(objects, types.ObjectIdentifier{Key: aws.String(s.fullKey(key))})
		}
		_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return fmt.Errorf("delete agent file contents: %w", err)
		}
	}
	return nil
}

func chunks[T any](items []T, size int) func(func([]T) bool) {
	return func(yield func([]T) bool) {
		if size <= 0 {
			size = len(items)
		}
		for len(items) > 0 {
			n := size
			if n > len(items) {
				n = len(items)
			}
			if !yield(items[:n]) {
				return
			}
			items = items[n:]
		}
	}
}
