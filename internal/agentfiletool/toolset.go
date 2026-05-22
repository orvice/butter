package agentfiletool

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"go.orx.me/apps/butter/internal/repo/agentfile"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type Toolset struct {
	repo       agentfile.Repository
	mounts     []*agentsv1.AgentFileMount
	maxBytes   int64
	toolsCache []tool.Tool
}

func NewToolset(repo agentfile.Repository, mounts []*agentsv1.AgentFileMount, maxBytes int64) (tool.Toolset, error) {
	if repo == nil || len(mounts) == 0 {
		return nil, nil
	}
	if maxBytes <= 0 {
		maxBytes = 256 * 1024
	}
	ts := &Toolset{repo: repo, mounts: normalizeMounts(mounts), maxBytes: maxBytes}
	if len(ts.mounts) == 0 {
		return nil, nil
	}
	tools, err := ts.buildTools()
	if err != nil {
		return nil, err
	}
	ts.toolsCache = tools
	return ts, nil
}

func (t *Toolset) Name() string { return "agent_files" }

func (t *Toolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return t.toolsCache, nil
}

func normalizeMounts(mounts []*agentsv1.AgentFileMount) []*agentsv1.AgentFileMount {
	out := make([]*agentsv1.AgentFileMount, 0, len(mounts))
	seen := map[string]bool{}
	for _, mount := range mounts {
		if mount == nil || strings.TrimSpace(mount.GetSpaceId()) == "" {
			continue
		}
		mountPath := normalizeMountPath(mount.GetMountPath())
		if seen[mountPath] {
			continue
		}
		seen[mountPath] = true
		perm := mount.GetPermission()
		if perm == agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED {
			perm = agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ
		}
		out = append(out, &agentsv1.AgentFileMount{
			SpaceId:    strings.TrimSpace(mount.GetSpaceId()),
			MountPath:  mountPath,
			Permission: perm,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return len(out[i].GetMountPath()) > len(out[j].GetMountPath())
	})
	return out
}

func normalizeMountPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}

func (t *Toolset) buildTools() ([]tool.Tool, error) {
	defs := []struct {
		name        string
		description string
		makeTool    func(functiontool.Config) (tool.Tool, error)
	}{
		{
			name:        "agent_files_list_files",
			description: "List files available in the mounted agent file spaces. Paths are virtual mount paths such as /docs/notes.md.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.listFiles)
			},
		},
		{
			name:        "agent_files_read_file",
			description: "Read a text file from a mounted agent file space.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.readFile)
			},
		},
		{
			name:        "agent_files_write_file",
			description: "Create or replace a text file in a writable mounted agent file space.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.writeFile)
			},
		},
		{
			name:        "agent_files_append_file",
			description: "Append text to a file in a writable mounted agent file space.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.appendFile)
			},
		},
		{
			name:        "agent_files_delete_file",
			description: "Delete a file from a mounted agent file space when the mount allows deletion.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.deleteFile)
			},
		},
		{
			name:        "agent_files_search_files",
			description: "Search mounted text files by simple substring and return matching file paths with snippets.",
			makeTool: func(cfg functiontool.Config) (tool.Tool, error) {
				return functiontool.New(cfg, t.searchFiles)
			},
		},
	}
	out := make([]tool.Tool, 0, len(defs))
	for _, def := range defs {
		tool, err := def.makeTool(functiontool.Config{Name: def.name, Description: def.description})
		if err != nil {
			return nil, fmt.Errorf("create %s tool: %w", def.name, err)
		}
		out = append(out, tool)
	}
	return out, nil
}

type resolvedPath struct {
	mount     *agentsv1.AgentFileMount
	spacePath string
}

func (t *Toolset) resolve(virtualPath string, need agentsv1.AgentFileMountPermission) (resolvedPath, error) {
	clean, err := agentfile.NormalizePath(virtualPath)
	if err != nil {
		return resolvedPath{}, err
	}
	for _, mount := range t.mounts {
		mountPath := mount.GetMountPath()
		if clean != mountPath && mountPath != "/" && !strings.HasPrefix(clean, strings.TrimRight(mountPath, "/")+"/") {
			continue
		}
		if !hasPermission(mount.GetPermission(), need) {
			return resolvedPath{}, fmt.Errorf("mount %s does not allow %s", mountPath, permissionName(need))
		}
		rel := strings.TrimPrefix(clean, mountPath)
		if mountPath == "/" {
			rel = strings.TrimPrefix(clean, "/")
		}
		spacePath := "/" + strings.TrimLeft(rel, "/")
		if spacePath == "/" {
			return resolvedPath{}, agentfile.ErrInvalidPath
		}
		return resolvedPath{mount: mount, spacePath: spacePath}, nil
	}
	return resolvedPath{}, fmt.Errorf("path %q is not under any mounted agent file space", clean)
}

func (t *Toolset) resolvePrefix(virtualPrefix string) (resolvedPath, error) {
	clean := normalizeMountPath(virtualPrefix)
	for _, mount := range t.mounts {
		mountPath := mount.GetMountPath()
		if clean != mountPath && mountPath != "/" && !strings.HasPrefix(clean, strings.TrimRight(mountPath, "/")+"/") {
			continue
		}
		if !hasPermission(mount.GetPermission(), agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ) {
			return resolvedPath{}, fmt.Errorf("mount %s does not allow read", mountPath)
		}
		rel := strings.TrimPrefix(clean, mountPath)
		if mountPath == "/" {
			rel = strings.TrimPrefix(clean, "/")
		}
		rel = strings.TrimLeft(rel, "/")
		if rel == "" {
			return resolvedPath{mount: mount}, nil
		}
		return resolvedPath{mount: mount, spacePath: "/" + rel}, nil
	}
	return resolvedPath{}, fmt.Errorf("path %q is not under any mounted agent file space", clean)
}

func hasPermission(got, need agentsv1.AgentFileMountPermission) bool {
	return got >= need
}

func permissionName(p agentsv1.AgentFileMountPermission) string {
	switch p {
	case agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ:
		return "read"
	case agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE:
		return "write"
	case agentsv1.AgentFileMountPermission_AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE:
		return "delete"
	default:
		return "access"
	}
}

func (t *Toolset) runtime(ctx context.Context) (RuntimeContext, error) {
	info := RuntimeContextFrom(ctx)
	if info.WorkspaceID == "" || info.AgentName == "" {
		return RuntimeContext{}, fmt.Errorf("agent file runtime context is missing")
	}
	return info, nil
}

func virtualPath(mount *agentsv1.AgentFileMount, spacePath string) string {
	mountPath := strings.TrimRight(mount.GetMountPath(), "/")
	if mountPath == "" {
		mountPath = "/"
	}
	spacePath = strings.TrimLeft(spacePath, "/")
	if mountPath == "/" {
		return "/" + spacePath
	}
	return mountPath + "/" + spacePath
}
