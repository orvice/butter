import { useState } from "react";
import { FileText, Trash2 } from "lucide-react";
import { useAgentFileSpaces } from "@/api/agent-files";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { AgentFileMount, AgentFileMountPermission, AgentFileSpace } from "@/types/api";

const DEFAULT_PERMISSION: AgentFileMountPermission = "AGENT_FILE_MOUNT_PERMISSION_READ";

const PERMISSIONS: { value: AgentFileMountPermission; label: string }[] = [
  { value: "AGENT_FILE_MOUNT_PERMISSION_READ", label: "Read" },
  { value: "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE", label: "Read / write" },
  { value: "AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE", label: "Read / write / delete" },
];

interface AgentFileMountsFieldProps {
  value?: AgentFileMount[];
  onChange: (mounts: AgentFileMount[]) => void;
}

function mountPathFromSpace(space: AgentFileSpace) {
  const source = space.name || space.id || "files";
  const slug = source
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `/${slug || "files"}`;
}

function uniqueMountPath(base: string, mounts: AgentFileMount[]) {
  const used = new Set(mounts.map((mount) => mount.mount_path || "/"));
  if (!used.has(base)) return base;
  for (let i = 2; ; i += 1) {
    const candidate = `${base}-${i}`;
    if (!used.has(candidate)) return candidate;
  }
}

export function AgentFileMountsField({ value, onChange }: AgentFileMountsFieldProps) {
  const [addValue, setAddValue] = useState<string | undefined>();
  const { data, isLoading } = useAgentFileSpaces();
  const mounts = value ?? [];
  const spaces = data?.spaces ?? [];
  const spacesById = new Map(spaces.map((space) => [space.id ?? "", space]));
  const mountedSpaceIds = new Set(mounts.map((mount) => mount.space_id));
  const availableSpaces = spaces.filter((space) => space.id && !mountedSpaceIds.has(space.id));

  function addMount(spaceId: string) {
    const space = spacesById.get(spaceId);
    if (!space) return;
    onChange([
      ...mounts,
      {
        space_id: spaceId,
        mount_path: uniqueMountPath(mountPathFromSpace(space), mounts),
        permission: DEFAULT_PERMISSION,
      },
    ]);
    setAddValue(undefined);
  }

  function updateMount(index: number, patch: Partial<AgentFileMount>) {
    onChange(mounts.map((mount, i) => (i === index ? { ...mount, ...patch } : mount)));
  }

  function removeMount(index: number) {
    onChange(mounts.filter((_, i) => i !== index));
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <Select
          value={addValue}
          onValueChange={(spaceId) => {
            if (spaceId) addMount(spaceId);
          }}
          disabled={isLoading || availableSpaces.length === 0}
        >
          <SelectTrigger className="w-full sm:w-72">
            <SelectValue placeholder={isLoading ? "Loading file spaces..." : "Add file space"} />
          </SelectTrigger>
          <SelectContent>
            {availableSpaces.map((space) => (
              <SelectItem key={space.id} value={space.id ?? ""}>
                {space.name || space.id}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Badge variant="outline">{mounts.length} mounted</Badge>
      </div>

      {mounts.length === 0 ? (
        <div className="rounded-md border border-dashed px-3 py-4 text-sm text-muted-foreground">
          No file spaces mounted.
        </div>
      ) : (
        <div className="space-y-2">
          {mounts.map((mount, index) => {
            const space = spacesById.get(mount.space_id);
            const permission = mount.permission || DEFAULT_PERMISSION;
            return (
              <div key={`${mount.space_id}-${index}`} className="rounded-md border p-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <FileText className="h-4 w-4 text-muted-foreground" />
                      <span className="truncate font-medium">{space?.name ?? mount.space_id}</span>
                      <Badge variant="secondary">{mount.space_id}</Badge>
                    </div>
                    {space?.description ? (
                      <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{space.description}</p>
                    ) : null}
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    title="Remove file space"
                    onClick={() => removeMount(index)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>

                <div className="mt-3 grid gap-3 md:grid-cols-[minmax(0,1fr)_220px]">
                  <label className="space-y-1">
                    <span className="text-xs font-medium text-muted-foreground">Mount path</span>
                    <Input
                      value={mount.mount_path ?? ""}
                      onChange={(event) => updateMount(index, { mount_path: event.target.value })}
                      placeholder="/docs"
                    />
                  </label>
                  <label className="space-y-1">
                    <span className="text-xs font-medium text-muted-foreground">Permission</span>
                    <Select
                      value={permission}
                      onValueChange={(next) => updateMount(index, { permission: next as AgentFileMountPermission })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PERMISSIONS.map((item) => (
                          <SelectItem key={item.value} value={item.value}>
                            {item.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </label>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
