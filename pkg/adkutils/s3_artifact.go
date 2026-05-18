// Package adkutils contains helpers for integrating Butter with Google ADK.
package adkutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/adk/artifact"
	"google.golang.org/genai"
)

const defaultTextContentType = "text/plain"

// S3ArtifactClient is the subset of the AWS S3 client used by the artifact service.
type S3ArtifactClient interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type s3ArtifactService struct {
	bucket string
	client S3ArtifactClient
}

// NewS3ArtifactService creates an ADK artifact.Service backed by S3.
func NewS3ArtifactService(bucket string, client S3ArtifactClient) artifact.Service {
	return &s3ArtifactService{
		bucket: bucket,
		client: client,
	}
}

func artifactHasUserNamespace(filename string) bool {
	return strings.HasPrefix(filename, "user:")
}

func buildArtifactKey(appName, userID, sessionID, fileName string, version int64) string {
	if artifactHasUserNamespace(fileName) {
		return fmt.Sprintf("%s/%s/user/%s/%d", appName, userID, fileName, version)
	}
	return fmt.Sprintf("%s/%s/%s/%s/%d", appName, userID, sessionID, fileName, version)
}

func buildArtifactPrefix(appName, userID, sessionID, fileName string) string {
	if artifactHasUserNamespace(fileName) {
		return fmt.Sprintf("%s/%s/user/%s/", appName, userID, fileName)
	}
	return fmt.Sprintf("%s/%s/%s/%s/", appName, userID, sessionID, fileName)
}

func buildSessionPrefix(appName, userID, sessionID string) string {
	return fmt.Sprintf("%s/%s/%s/", appName, userID, sessionID)
}

func buildUserPrefix(appName, userID string) string {
	return fmt.Sprintf("%s/%s/user/", appName, userID)
}

// Save implements artifact.Service.
func (s *s3ArtifactService) Save(ctx context.Context, req *artifact.SaveRequest) (*artifact.SaveResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	version := req.Version
	if version == 0 {
		resp, err := s.versions(ctx, &artifact.VersionsRequest{
			AppName: req.AppName, UserID: req.UserID, SessionID: req.SessionID, FileName: req.FileName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list artifact versions: %w", err)
		}
		version = 1
		if len(resp.Versions) > 0 {
			version = slices.Max(resp.Versions) + 1
		}
	}

	body, contentType := partBody(req.Part)
	key := buildArtifactKey(req.AppName, req.UserID, req.SessionID, req.FileName, version)
	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	}); err != nil {
		return nil, fmt.Errorf("failed to put artifact %q: %w", key, err)
	}

	return &artifact.SaveResponse{Version: version}, nil
}

func partBody(part *genai.Part) ([]byte, string) {
	if part.InlineData != nil {
		return part.InlineData.Data, part.InlineData.MIMEType
	}
	return []byte(part.Text), defaultTextContentType
}

// Load implements artifact.Service.
func (s *s3ArtifactService) Load(ctx context.Context, req *artifact.LoadRequest) (_ *artifact.LoadResponse, err error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	version, err := s.resolveVersion(ctx, req.AppName, req.UserID, req.SessionID, req.FileName, req.Version)
	if err != nil {
		return nil, err
	}

	key := buildArtifactKey(req.AppName, req.UserID, req.SessionID, req.FileName, version)
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, fmt.Errorf("artifact %q not found: %w", key, fs.ErrNotExist)
		}
		return nil, fmt.Errorf("failed to get artifact %q: %w", key, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close artifact body: %w", closeErr)
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact %q: %w", key, err)
	}

	return &artifact.LoadResponse{Part: genai.NewPartFromBytes(data, aws.ToString(resp.ContentType))}, nil
}

// Delete implements artifact.Service.
func (s *s3ArtifactService) Delete(ctx context.Context, req *artifact.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	if req.Version != 0 {
		key := buildArtifactKey(req.AppName, req.UserID, req.SessionID, req.FileName, req.Version)
		if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}); err != nil {
			return fmt.Errorf("failed to delete artifact %q: %w", key, err)
		}
		return nil
	}

	resp, err := s.versions(ctx, &artifact.VersionsRequest{
		AppName: req.AppName, UserID: req.UserID, SessionID: req.SessionID, FileName: req.FileName,
	})
	if err != nil {
		return fmt.Errorf("failed to list artifact versions for delete: %w", err)
	}
	if len(resp.Versions) == 0 {
		return nil
	}

	keys := make([]string, 0, len(resp.Versions))
	for _, version := range resp.Versions {
		keys = append(keys, buildArtifactKey(req.AppName, req.UserID, req.SessionID, req.FileName, version))
	}

	g, gctx := errgroup.WithContext(ctx)
	for chunk := range slices.Chunk(keys, 1000) {
		chunk := slices.Clone(chunk)
		g.Go(func() error {
			objects := make([]types.ObjectIdentifier, 0, len(chunk))
			for _, key := range chunk {
				objects = append(objects, types.ObjectIdentifier{Key: aws.String(key)})
			}
			_, err := s.client.DeleteObjects(gctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s.bucket),
				Delete: &types.Delete{
					Objects: objects,
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete artifact versions: %w", err)
			}
			return nil
		})
	}

	return g.Wait()
}

// List implements artifact.Service.
func (s *s3ArtifactService) List(ctx context.Context, req *artifact.ListRequest) (*artifact.ListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	filenames := map[string]bool{}
	if err := s.fetchFilenames(ctx, buildSessionPrefix(req.AppName, req.UserID, req.SessionID), filenames); err != nil {
		return nil, fmt.Errorf("failed to list session artifacts: %w", err)
	}
	if err := s.fetchFilenames(ctx, buildUserPrefix(req.AppName, req.UserID), filenames); err != nil {
		return nil, fmt.Errorf("failed to list user artifacts: %w", err)
	}

	result := slices.Collect(maps.Keys(filenames))
	sort.Strings(result)
	return &artifact.ListResponse{FileNames: result}, nil
}

func (s *s3ArtifactService) fetchFilenames(ctx context.Context, prefix string, filenames map[string]bool) error {
	return s.listKeys(ctx, prefix, func(key string) error {
		segments := strings.Split(key, "/")
		if len(segments) < 2 {
			return fmt.Errorf("invalid artifact key %q", key)
		}
		filenames[segments[len(segments)-2]] = true
		return nil
	})
}

// Versions implements artifact.Service.
func (s *s3ArtifactService) Versions(ctx context.Context, req *artifact.VersionsRequest) (*artifact.VersionsResponse, error) {
	resp, err := s.versions(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Versions) == 0 {
		return nil, fmt.Errorf("artifact not found: %w", fs.ErrNotExist)
	}
	return resp, nil
}

func (s *s3ArtifactService) versions(ctx context.Context, req *artifact.VersionsRequest) (*artifact.VersionsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	prefix := buildArtifactPrefix(req.AppName, req.UserID, req.SessionID, req.FileName)
	versions := []int64{}
	if err := s.listKeys(ctx, prefix, func(key string) error {
		segments := strings.Split(key, "/")
		if len(segments) == 0 {
			return fmt.Errorf("invalid artifact key %q", key)
		}
		version, err := strconv.ParseInt(segments[len(segments)-1], 10, 64)
		if err != nil {
			return nil
		}
		versions = append(versions, version)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return &artifact.VersionsResponse{Versions: versions}, nil
}

func (s *s3ArtifactService) listKeys(ctx context.Context, prefix string, visit func(string) error) error {
	var continuationToken *string
	for {
		resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list objects with prefix %q: %w", prefix, err)
		}
		for _, obj := range resp.Contents {
			if obj.Key == nil {
				continue
			}
			if err := visit(aws.ToString(obj.Key)); err != nil {
				return err
			}
		}
		if !aws.ToBool(resp.IsTruncated) {
			return nil
		}
		continuationToken = resp.NextContinuationToken
	}
}

func (s *s3ArtifactService) resolveVersion(ctx context.Context, appName, userID, sessionID, fileName string, version int64) (int64, error) {
	if version != 0 {
		return version, nil
	}
	resp, err := s.versions(ctx, &artifact.VersionsRequest{
		AppName: appName, UserID: userID, SessionID: sessionID, FileName: fileName,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list artifact versions: %w", err)
	}
	if len(resp.Versions) == 0 {
		return 0, fmt.Errorf("artifact not found: %w", fs.ErrNotExist)
	}
	return slices.Max(resp.Versions), nil
}

// GetArtifactVersion implements artifact.Service.
func (s *s3ArtifactService) GetArtifactVersion(ctx context.Context, req *artifact.GetArtifactVersionRequest) (*artifact.GetArtifactVersionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	version, err := s.resolveVersion(ctx, req.AppName, req.UserID, req.SessionID, req.FileName, req.Version)
	if err != nil {
		return nil, err
	}

	key := buildArtifactKey(req.AppName, req.UserID, req.SessionID, req.FileName, version)
	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, fmt.Errorf("artifact %q not found: %w", key, fs.ErrNotExist)
		}
		return nil, fmt.Errorf("failed to head artifact %q: %w", key, err)
	}

	customMeta := make(map[string]any, len(resp.Metadata))
	for k, v := range resp.Metadata {
		customMeta[k] = v
	}

	createTime := float64(0)
	if resp.LastModified != nil {
		createTime = float64(resp.LastModified.Unix())
	}

	return &artifact.GetArtifactVersionResponse{
		ArtifactVersion: &artifact.ArtifactVersion{
			Version:        version,
			CanonicalURI:   fmt.Sprintf("s3://%s/%s", s.bucket, key),
			CustomMetadata: customMeta,
			CreateTime:     createTime,
			MimeType:       aws.ToString(resp.ContentType),
		},
	}, nil
}

func isS3NotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound")
}

var _ artifact.Service = (*s3ArtifactService)(nil)
