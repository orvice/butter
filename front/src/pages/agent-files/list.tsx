import { useMemo, useState } from "react";
import { toast } from "sonner";
import {
  FilePlus2,
  FileText,
  FolderOpen,
  Pencil,
  Plus,
  Save,
  Search,
  Trash2,
} from "lucide-react";
import {
  useAgentFile,
  useAgentFiles,
  useAgentFileSpaces,
  useCreateAgentFileSpace,
  useDeleteAgentFile,
  useDeleteAgentFileSpace,
  useSearchAgentFiles,
  useUpdateAgentFileSpace,
  useWriteAgentFile,
} from "@/api/agent-files";
import { DeleteDialog } from "@/components/delete-dialog";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import type { AgentFile, AgentFileSpace } from "@/types/api";

type SpaceFormState = {
  id: string;
  name: string;
  description: string;
};

type FileFormState = {
  path: string;
  content: string;
};

function formatBytes(value?: number | string) {
  const n = Number(value ?? 0);
  if (!Number.isFinite(n) || n <= 0) return "0 B";
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KiB`;
  return `${(n / 1024 / 1024).toFixed(1)} MiB`;
}

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString() : "-";
}

function normalizedPath(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return "";
  return trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
}

function fileSortKey(file: AgentFile) {
  return file.path || "";
}

export default function AgentFilesPage() {
  const { data: spacesData, isLoading: spacesLoading } = useAgentFileSpaces();
  const createSpace = useCreateAgentFileSpace();
  const updateSpace = useUpdateAgentFileSpace();
  const deleteSpace = useDeleteAgentFileSpace();
  const writeFile = useWriteAgentFile();
  const deleteFile = useDeleteAgentFile();

  const spaces = useMemo(() => spacesData?.spaces ?? [], [spacesData?.spaces]);
  const [explicitSpaceId, setExplicitSpaceId] = useState("");
  const selectedSpaceId = useMemo(() => {
    if (explicitSpaceId && spaces.some((space) => space.id === explicitSpaceId)) {
      return explicitSpaceId;
    }
    return spaces[0]?.id ?? "";
  }, [explicitSpaceId, spaces]);
  const selectedSpace = spaces.find((space) => space.id === selectedSpaceId);
  const [pathPrefix, setPathPrefix] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedPath, setSelectedPath] = useState("");
  const [spaceDialogOpen, setSpaceDialogOpen] = useState(false);
  const [editingSpace, setEditingSpace] = useState<AgentFileSpace | null>(null);
  const [spaceForm, setSpaceForm] = useState<SpaceFormState>({ id: "", name: "", description: "" });
  const [fileDialogOpen, setFileDialogOpen] = useState(false);
  const [fileForm, setFileForm] = useState<FileFormState>({ path: "", content: "" });
  const [fileFormDirty, setFileFormDirty] = useState(false);
  const [deleteSpaceTarget, setDeleteSpaceTarget] = useState<AgentFileSpace | null>(null);
  const [deleteFileTarget, setDeleteFileTarget] = useState<AgentFile | null>(null);

  const { data: filesData, isLoading: filesLoading } = useAgentFiles(selectedSpaceId, pathPrefix);
  const { data: activeFileData, isLoading: activeFileLoading } = useAgentFile(selectedSpaceId, selectedPath);
  const { data: searchData } = useSearchAgentFiles(selectedSpaceId, searchQuery);

  const files = useMemo(
    () => [...(filesData?.files ?? [])].sort((a, b) => fileSortKey(a).localeCompare(fileSortKey(b))),
    [filesData?.files],
  );
  const searchResults = searchData?.results ?? [];
  const activeFile = activeFileData?.file;
  const activeContent = activeFileData?.content ?? "";
  const displayedFileForm =
    selectedPath && activeFileData && !fileFormDirty
      ? { path: activeFileData.file?.path ?? selectedPath, content: activeContent }
      : fileForm;

  function openCreateSpaceDialog() {
    setEditingSpace(null);
    setSpaceForm({ id: "", name: "", description: "" });
    setSpaceDialogOpen(true);
  }

  function openEditSpaceDialog(space: AgentFileSpace) {
    setEditingSpace(space);
    setSpaceForm({ id: space.id ?? "", name: space.name ?? "", description: space.description ?? "" });
    setSpaceDialogOpen(true);
  }

  function saveSpace() {
    const name = spaceForm.name.trim();
    if (!name) {
      toast.error("File space name is required");
      return;
    }
    const payload: AgentFileSpace = {
      id: editingSpace?.id || spaceForm.id.trim() || undefined,
      name,
      description: spaceForm.description.trim(),
    };
    const mutation = editingSpace ? updateSpace : createSpace;
    mutation.mutate(payload, {
      onSuccess: (res) => {
        toast.success(editingSpace ? "File space updated" : "File space created");
        setSpaceDialogOpen(false);
        if (!selectedSpaceId && res.space?.id) {
          setExplicitSpaceId(res.space.id);
        }
      },
      onError: (error) => toast.error(error.message),
    });
  }

  function openCreateFileDialog() {
    if (!selectedSpaceId) {
      toast.error("Create a file space first");
      return;
    }
    setSelectedPath("");
    setFileForm({ path: "", content: "" });
    setFileFormDirty(false);
    setFileDialogOpen(true);
  }

  function openEditFile(file: AgentFile) {
    setSelectedPath(file.path);
    setFileForm({ path: file.path, content: "" });
    setFileFormDirty(false);
    setFileDialogOpen(true);
  }

  function saveFile() {
    if (!selectedSpaceId) return;
    const path = normalizedPath(displayedFileForm.path);
    if (!path) {
      toast.error("File path is required");
      return;
    }
    writeFile.mutate(
      {
        spaceId: selectedSpaceId,
        path,
        content: displayedFileForm.content,
        contentType: "text/plain; charset=utf-8",
      },
      {
        onSuccess: () => {
          toast.success("File saved");
          setSelectedPath(path);
          setFileDialogOpen(false);
        },
        onError: (error) => toast.error(error.message),
      },
    );
  }

  function confirmDeleteSpace() {
    if (!deleteSpaceTarget?.id) return;
    deleteSpace.mutate(deleteSpaceTarget.id, {
      onSuccess: () => {
        toast.success("File space deleted");
        if (selectedSpaceId === deleteSpaceTarget.id) {
          setExplicitSpaceId("");
          setSelectedPath("");
        }
        setDeleteSpaceTarget(null);
      },
      onError: (error) => toast.error(error.message),
    });
  }

  function confirmDeleteFile() {
    if (!selectedSpaceId || !deleteFileTarget?.path) return;
    deleteFile.mutate(
      { spaceId: selectedSpaceId, path: deleteFileTarget.path },
      {
        onSuccess: () => {
          toast.success("File deleted");
          if (selectedPath === deleteFileTarget.path) {
            setSelectedPath("");
            setFileDialogOpen(false);
          }
          setDeleteFileTarget(null);
        },
        onError: (error) => toast.error(error.message),
      },
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Files"
        description="Create workspace text file spaces and mount them into agents as built-in tools."
      />

      <div className="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
        <Card>
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>File Spaces</CardTitle>
              <CardDescription>Workspace-owned collections available to agent mounts.</CardDescription>
            </div>
            <Button size="sm" onClick={openCreateSpaceDialog}>
              <Plus className="mr-1 h-4 w-4" />
              New
            </Button>
          </CardHeader>
          <CardContent className="space-y-2">
            {spacesLoading ? (
              Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-16" />)
            ) : spaces.length === 0 ? (
              <div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">
                No file spaces yet.
              </div>
            ) : (
              spaces.map((space) => {
                const selected = space.id === selectedSpaceId;
                return (
                  <button
                    key={space.id}
                    type="button"
                    onClick={() => {
                      setExplicitSpaceId(space.id ?? "");
                      setSelectedPath("");
                    }}
                    className={`w-full rounded-md border p-3 text-left transition-colors ${
                      selected ? "border-primary bg-primary/10" : "hover:bg-muted"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <FolderOpen className="h-4 w-4 text-muted-foreground" />
                          <span className="truncate font-medium">{space.name}</span>
                        </div>
                        <div className="mt-1 truncate font-mono text-xs text-muted-foreground">{space.id}</div>
                        {space.description ? (
                          <p className="mt-2 line-clamp-2 text-xs text-muted-foreground">{space.description}</p>
                        ) : null}
                      </div>
                      <Badge variant={selected ? "default" : "outline"}>{selected ? "Open" : "Space"}</Badge>
                    </div>
                  </button>
                );
              })
            )}
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <CardTitle>{selectedSpace?.name ?? "Files"}</CardTitle>
                <CardDescription>
                  {selectedSpace ? selectedSpace.description || "Manage text files in this file space." : "Select or create a file space."}
                </CardDescription>
              </div>
              {selectedSpace ? (
                <div className="flex flex-wrap gap-2">
                  <Button variant="outline" size="sm" onClick={() => openEditSpaceDialog(selectedSpace)}>
                    <Pencil className="mr-1 h-4 w-4" />
                    Edit Space
                  </Button>
                  <Button variant="destructive" size="sm" onClick={() => setDeleteSpaceTarget(selectedSpace)}>
                    <Trash2 className="mr-1 h-4 w-4" />
                    Delete Space
                  </Button>
                  <Button size="sm" onClick={openCreateFileDialog}>
                    <FilePlus2 className="mr-1 h-4 w-4" />
                    New File
                  </Button>
                </div>
              ) : null}
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3 md:grid-cols-2">
                <div className="relative">
                  <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    className="pl-8"
                    value={pathPrefix}
                    onChange={(event) => setPathPrefix(event.target.value)}
                    placeholder="/notes"
                    disabled={!selectedSpaceId}
                  />
                </div>
                <div className="relative">
                  <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    className="pl-8"
                    value={searchQuery}
                    onChange={(event) => setSearchQuery(event.target.value)}
                    placeholder="Search file contents"
                    disabled={!selectedSpaceId}
                  />
                </div>
              </div>

              {searchQuery.trim() ? (
                <div className="rounded-md border bg-muted/30 p-3">
                  <div className="mb-2 text-xs font-medium uppercase tracking-[0.05em] text-muted-foreground">
                    Search Results
                  </div>
                  {searchResults.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No matches.</p>
                  ) : (
                    <div className="space-y-2">
                      {searchResults.map((result) => {
                        const file = result.file;
                        if (!file) return null;
                        return (
                          <button
                            key={file.path}
                            type="button"
                            onClick={() => openEditFile(file)}
                            className="w-full rounded-md border bg-card p-2 text-left hover:bg-muted"
                          >
                            <div className="font-mono text-xs">{file.path}</div>
                            {result.snippets?.[0] ? (
                              <div className="mt-1 line-clamp-2 text-xs text-muted-foreground">{result.snippets[0]}</div>
                            ) : null}
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              ) : null}

              {filesLoading ? (
                <div className="space-y-2">
                  {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-14" />)}
                </div>
              ) : files.length === 0 ? (
                <div className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
                  {selectedSpaceId ? "No files in this space yet." : "Select a file space to browse files."}
                </div>
              ) : (
                <div className="overflow-hidden rounded-md border">
                  {files.map((file) => (
                    <div key={file.path} className="grid gap-3 border-b p-3 last:border-b-0 md:grid-cols-[minmax(0,1fr)_120px_120px_130px_auto] md:items-center">
                      <button type="button" className="min-w-0 text-left" onClick={() => openEditFile(file)}>
                        <div className="flex min-w-0 items-center gap-2">
                          <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                          <span className="truncate font-mono text-sm">{file.path}</span>
                        </div>
                      </button>
                      <span className="text-xs text-muted-foreground">{formatBytes(file.size_bytes)}</span>
                      <span className="text-xs text-muted-foreground">v{String(file.version ?? 0)}</span>
                      <span className="text-xs text-muted-foreground">{formatDate(file.updated_at)}</span>
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon-sm" title="Edit file" onClick={() => openEditFile(file)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon-sm" title="Delete file" onClick={() => setDeleteFileTarget(file)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={spaceDialogOpen} onOpenChange={setSpaceDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingSpace ? "Edit File Space" : "New File Space"}</DialogTitle>
            <DialogDescription>File spaces are mounted into agents from the agent settings page.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {!editingSpace ? (
              <div className="space-y-2">
                <Label htmlFor="space-id">ID</Label>
                <Input
                  id="space-id"
                  value={spaceForm.id}
                  onChange={(event) => setSpaceForm((prev) => ({ ...prev, id: event.target.value.trim() }))}
                  placeholder="product-docs"
                />
              </div>
            ) : null}
            <div className="space-y-2">
              <Label htmlFor="space-name">Name</Label>
              <Input
                id="space-name"
                value={spaceForm.name}
                onChange={(event) => setSpaceForm((prev) => ({ ...prev, name: event.target.value }))}
                placeholder="Product Docs"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="space-description">Description</Label>
              <Textarea
                id="space-description"
                value={spaceForm.description}
                onChange={(event) => setSpaceForm((prev) => ({ ...prev, description: event.target.value }))}
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSpaceDialogOpen(false)}>Cancel</Button>
            <Button onClick={saveSpace} disabled={createSpace.isPending || updateSpace.isPending}>
              {createSpace.isPending || updateSpace.isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={fileDialogOpen} onOpenChange={setFileDialogOpen}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{selectedPath ? "Edit File" : "New File"}</DialogTitle>
            <DialogDescription>
              Files are UTF-8 text. Agents see them through their configured mount paths.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="file-path">Path</Label>
              <Input
                id="file-path"
                value={displayedFileForm.path}
                onChange={(event) => {
                  setFileFormDirty(true);
                  setFileForm((prev) => ({ ...prev, path: event.target.value }));
                }}
                placeholder="/notes/todo.md"
              />
            </div>
            {activeFileLoading ? <Skeleton className="h-60" /> : (
              <div className="space-y-2">
                <Label htmlFor="file-content">Content</Label>
                <Textarea
                  id="file-content"
                  className="min-h-80 font-mono text-sm"
                  value={displayedFileForm.content}
                  onChange={(event) => {
                    setFileFormDirty(true);
                    setFileForm((prev) => ({ ...prev, content: event.target.value }));
                  }}
                  placeholder="# Notes"
                />
              </div>
            )}
            {activeFile ? (
              <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                <Badge variant="outline">Version {String(activeFile.version ?? 0)}</Badge>
                <Badge variant="outline">{formatBytes(activeFile.size_bytes)}</Badge>
                <span>Updated {formatDate(activeFile.updated_at)}</span>
              </div>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setFileDialogOpen(false)}>Cancel</Button>
            <Button onClick={saveFile} disabled={writeFile.isPending}>
              <Save className="mr-1 h-4 w-4" />
              {writeFile.isPending ? "Saving..." : "Save File"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <DeleteDialog
        open={!!deleteSpaceTarget}
        onOpenChange={(open) => !open && setDeleteSpaceTarget(null)}
        title="Delete file space"
        description={`Delete ${deleteSpaceTarget?.name ?? "this file space"} and all files inside it? This cannot be undone.`}
        onConfirm={confirmDeleteSpace}
        loading={deleteSpace.isPending}
      />

      <DeleteDialog
        open={!!deleteFileTarget}
        onOpenChange={(open) => !open && setDeleteFileTarget(null)}
        title="Delete file"
        description={`Delete ${deleteFileTarget?.path ?? "this file"}? This cannot be undone.`}
        onConfirm={confirmDeleteFile}
        loading={deleteFile.isPending}
      />
    </div>
  );
}
