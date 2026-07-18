package skill

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const skillMDContentType = "text/markdown; charset=utf-8"

// S3Client is the subset of the AWS S3 API the content store needs.
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

// NewS3ContentStore stores SKILL.md bodies in S3 under
// <key_prefix>/<content key> (issue #153 / ADR 0004).
func NewS3ContentStore(bucket string, client S3Client, keyPrefix string) ContentStore {
	return &s3ContentStore{
		bucket:    bucket,
		keyPrefix: strings.Trim(keyPrefix, "/"),
		client:    client,
	}
}

func (s *s3ContentStore) fullKey(key string) string {
	key = strings.TrimLeft(key, "/")
	if s.keyPrefix == "" {
		return key
	}
	return s.keyPrefix + "/" + key
}

func (s *s3ContentStore) Put(ctx context.Context, key, content string) error {
	fullKey := s.fullKey(key)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fullKey),
		Body:        bytes.NewReader([]byte(content)),
		ContentType: aws.String(skillMDContentType),
	})
	if err != nil {
		return fmt.Errorf("put skill content %q: %w", fullKey, err)
	}
	return nil
}

func (s *s3ContentStore) Get(ctx context.Context, key string) (string, error) {
	fullKey := s.fullKey(key)
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		// Map missing objects to the repository sentinel so both
		// ContentStore implementations share one error contract.
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return "", fmt.Errorf("skill content %q: %w", fullKey, ErrNotFound)
		}
		return "", fmt.Errorf("get skill content %q: %w", fullKey, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read skill content %q: %w", fullKey, err)
	}
	return string(data), nil
}

func (s *s3ContentStore) Delete(ctx context.Context, keys []string) error {
	// DeleteObjects accepts at most 1000 keys per call; skills delete one
	// SKILL.md today, so a single chunk suffices until resources land.
	for start := 0; start < len(keys); start += 1000 {
		end := min(start+1000, len(keys))
		objects := make([]types.ObjectIdentifier, 0, end-start)
		for _, key := range keys[start:end] {
			objects = append(objects, types.ObjectIdentifier{Key: aws.String(s.fullKey(key))})
		}
		_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return fmt.Errorf("delete skill contents: %w", err)
		}
	}
	return nil
}
