package agentfiletool

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/adk/tool"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type listFilesArgs struct {
	PathPrefix string `json:"path_prefix,omitempty" jsonschema_description:"Optional virtual path prefix to list, such as /docs."`
}

type fileInfo struct {
	Path        string `json:"path"`
	SizeBytes   int64  `json:"size_bytes"`
	Version     int64  `json:"version"`
	ContentType string `json:"content_type"`
}

type listFilesResult struct {
	Files []fileInfo `json:"files"`
}

type readFileArgs struct {
	Path    string `json:"path" jsonschema_description:"Virtual file path to read."`
	Version int64  `json:"version,omitempty" jsonschema_description:"Optional file version. Leave empty or 0 for latest."`
}

type readFileResult struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Version     int64  `json:"version"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type writeFileArgs struct {
	Path        string `json:"path" jsonschema_description:"Virtual file path to create or replace."`
	Content     string `json:"content" jsonschema_description:"UTF-8 text content to write."`
	ContentType string `json:"content_type,omitempty" jsonschema_description:"Optional content type. Defaults to text/plain."`
}

type writeFileResult struct {
	Path      string `json:"path"`
	Version   int64  `json:"version"`
	SizeBytes int64  `json:"size_bytes"`
}

type appendFileArgs struct {
	Path    string `json:"path" jsonschema_description:"Virtual file path to append to."`
	Content string `json:"content" jsonschema_description:"UTF-8 text content to append."`
}

type deleteFileArgs struct {
	Path string `json:"path" jsonschema_description:"Virtual file path to delete."`
}

type deleteFileResult struct {
	Deleted bool `json:"deleted"`
}

type searchFilesArgs struct {
	Query string `json:"query" jsonschema_description:"Substring to search for."`
	Limit int32  `json:"limit,omitempty" jsonschema_description:"Maximum number of results. Defaults to 20."`
}

type searchResult struct {
	Path     string   `json:"path"`
	Snippets []string `json:"snippets"`
	Version  int64    `json:"version"`
}

type searchFilesResult struct {
	Results []searchResult `json:"results"`
}

func (t *Toolset) listFiles(ctx tool.Context, args listFilesArgs) (listFilesResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return listFilesResult{}, err
	}
	var out []fileInfo
	for _, mount := range t.mounts {
		prefix := ""
		if strings.TrimSpace(args.PathPrefix) != "" {
			rp, err := t.resolvePrefix(args.PathPrefix)
			if err != nil {
				return listFilesResult{}, err
			}
			if rp.mount.GetSpaceId() != mount.GetSpaceId() {
				continue
			}
			prefix = rp.spacePath
		}
		files, err := t.repo.ListFiles(ctx, info.WorkspaceID, mount.GetSpaceId(), prefix)
		if err != nil {
			return listFilesResult{}, err
		}
		for _, file := range files {
			out = append(out, toFileInfo(virtualPath(mount, file.GetPath()), file))
		}
	}
	return listFilesResult{Files: out}, nil
}

func (t *Toolset) readFile(ctx tool.Context, args readFileArgs) (readFileResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return readFileResult{}, err
	}
	rp, err := t.resolve(args.Path, agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ)
	if err != nil {
		return readFileResult{}, err
	}
	file, content, err := t.repo.ReadFile(ctx, info.WorkspaceID, rp.mount.GetSpaceId(), rp.spacePath, args.Version)
	if err != nil {
		return readFileResult{}, err
	}
	return readFileResult{
		Path:        virtualPath(rp.mount, file.GetPath()),
		Content:     content,
		Version:     file.GetVersion(),
		ContentType: file.GetContentType(),
		SizeBytes:   file.GetSizeBytes(),
	}, nil
}

func (t *Toolset) writeFile(ctx tool.Context, args writeFileArgs) (writeFileResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return writeFileResult{}, err
	}
	if int64(len([]byte(args.Content))) > t.maxBytes {
		return writeFileResult{}, fmt.Errorf("content exceeds max file size of %d bytes", t.maxBytes)
	}
	rp, err := t.resolve(args.Path, agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE)
	if err != nil {
		return writeFileResult{}, err
	}
	file, err := t.repo.WriteFile(ctx, info.WorkspaceID, rp.mount.GetSpaceId(), rp.spacePath, args.Content, args.ContentType, nil)
	if err != nil {
		return writeFileResult{}, err
	}
	return writeFileResult{Path: virtualPath(rp.mount, file.GetPath()), Version: file.GetVersion(), SizeBytes: file.GetSizeBytes()}, nil
}

func (t *Toolset) appendFile(ctx tool.Context, args appendFileArgs) (writeFileResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return writeFileResult{}, err
	}
	rp, err := t.resolve(args.Path, agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE)
	if err != nil {
		return writeFileResult{}, err
	}
	file, current, err := t.repo.ReadFile(ctx, info.WorkspaceID, rp.mount.GetSpaceId(), rp.spacePath, 0)
	if err != nil {
		if !errors.Is(err, agentfile.ErrNotFound) {
			return writeFileResult{}, err
		}
		file = nil
	}
	next := current + args.Content
	if int64(len([]byte(next))) > t.maxBytes {
		return writeFileResult{}, fmt.Errorf("content exceeds max file size of %d bytes", t.maxBytes)
	}
	contentType := ""
	if file != nil {
		contentType = file.GetContentType()
	}
	written, err := t.repo.WriteFile(ctx, info.WorkspaceID, rp.mount.GetSpaceId(), rp.spacePath, next, contentType, nil)
	if err != nil {
		return writeFileResult{}, err
	}
	return writeFileResult{Path: virtualPath(rp.mount, written.GetPath()), Version: written.GetVersion(), SizeBytes: written.GetSizeBytes()}, nil
}

func (t *Toolset) deleteFile(ctx tool.Context, args deleteFileArgs) (deleteFileResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return deleteFileResult{}, err
	}
	rp, err := t.resolve(args.Path, agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE)
	if err != nil {
		return deleteFileResult{}, err
	}
	if err := t.repo.DeleteFile(ctx, info.WorkspaceID, rp.mount.GetSpaceId(), rp.spacePath); err != nil {
		return deleteFileResult{}, err
	}
	return deleteFileResult{Deleted: true}, nil
}

func (t *Toolset) searchFiles(ctx tool.Context, args searchFilesArgs) (searchFilesResult, error) {
	info, err := t.runtime(ctx)
	if err != nil {
		return searchFilesResult{}, err
	}
	var out []searchResult
	for _, mount := range t.mounts {
		results, err := t.repo.SearchFiles(ctx, info.WorkspaceID, mount.GetSpaceId(), args.Query, int(args.Limit))
		if err != nil {
			return searchFilesResult{}, err
		}
		for _, result := range results {
			file := result.GetFile()
			out = append(out, searchResult{
				Path:     virtualPath(mount, file.GetPath()),
				Snippets: result.GetSnippets(),
				Version:  file.GetVersion(),
			})
		}
	}
	return searchFilesResult{Results: out}, nil
}

func toFileInfo(p string, file *agentsv1.AgentFile) fileInfo {
	return fileInfo{
		Path:        p,
		SizeBytes:   file.GetSizeBytes(),
		Version:     file.GetVersion(),
		ContentType: file.GetContentType(),
	}
}
