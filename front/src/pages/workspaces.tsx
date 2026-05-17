import { useEffect, useMemo, useState } from "react";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { toast } from "sonner";
import { Building2, Trash2, UserPlus } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
import { DeleteDialog } from "@/components/delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useUsers, type AuthUser } from "@/api/auth";
import {
  useAddWorkspaceMember,
  useDeleteWorkspace,
  useRemoveWorkspaceMember,
  useUpdateWorkspace,
  useUpdateWorkspaceMember,
  useWorkspaceMembers,
} from "@/api/workspaces";
import { useWorkspace } from "@/hooks/use-workspace";
import type { WorkspaceMember } from "@/gen/agents/v1/workspace_pb";

type WorkspaceMemberRecord = WorkspaceMember & {
  workspace_id?: string;
  user_id?: string;
  created_at?: Timestamp;
};

const MEMBER_ROLES = ["owner", "admin", "member"];

function workspaceMemberUserId(member: WorkspaceMemberRecord) {
  return member.userId || member.user_id || "";
}

function workspaceMemberCreatedAt(member: WorkspaceMemberRecord) {
  return member.createdAt || member.created_at;
}

function userLabel(user: AuthUser | undefined) {
  if (!user) return "Unknown user";
  return user.display_name || user.displayName || user.username;
}

function formatDate(value?: Timestamp) {
  if (!value) return "-";
  const millis = Number(value.seconds ?? 0) * 1000 + Math.floor((value.nanos ?? 0) / 1_000_000);
  return new Date(millis).toLocaleString();
}

export default function WorkspacePage() {
  const { workspaces, selectedWorkspace, selectedWorkspaceId, clearSelectedWorkspace } = useWorkspace();
  const { data: membersData, isLoading: membersLoading } = useWorkspaceMembers(selectedWorkspaceId);
  const { data: usersData, isLoading: usersLoading } = useUsers();
  const updateWorkspace = useUpdateWorkspace();
  const deleteWorkspace = useDeleteWorkspace();
  const addMember = useAddWorkspaceMember();
  const updateMember = useUpdateWorkspaceMember();
  const removeMember = useRemoveWorkspaceMember();

  const members = useMemo(() => membersData?.members ?? [], [membersData?.members]);
  const users = useMemo(() => usersData?.users ?? [], [usersData?.users]);

  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [memberUserId, setMemberUserId] = useState("");
  const [memberRole, setMemberRole] = useState("member");
  const [deleteOpen, setDeleteOpen] = useState(false);

  useEffect(() => {
    queueMicrotask(() => {
      setName(selectedWorkspace?.name ?? "");
      setSlug(selectedWorkspace?.slug ?? "");
      setDescription(selectedWorkspace?.description ?? "");
    });
  }, [selectedWorkspace?.description, selectedWorkspace?.id, selectedWorkspace?.name, selectedWorkspace?.slug]);

  const availableUsers = useMemo(
    () => users.filter((user) => !members.some((member) => workspaceMemberUserId(member) === user.id)),
    [members, users],
  );

  function handleWorkspaceUpdate(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedWorkspaceId) return;
    if (!name.trim()) {
      toast.error("Workspace name is required");
      return;
    }
    if (!slug.trim()) {
      toast.error("Workspace slug is required");
      return;
    }

    updateWorkspace.mutate(
      {
        id: selectedWorkspaceId,
        name: name.trim(),
        slug: slug.trim(),
        description: description.trim(),
      },
      {
        onSuccess: () => toast.success("Workspace updated"),
        onError: (error) => toast.error(error.message),
      },
    );
  }

  function handleAddMember() {
    if (!selectedWorkspaceId) return;
    if (!memberUserId) {
      toast.error("Select a user first");
      return;
    }

    addMember.mutate(
      {
        workspace_id: selectedWorkspaceId,
        user_id: memberUserId,
        role: memberRole,
      },
      {
        onSuccess: () => {
          toast.success("Member added");
          setMemberUserId("");
          setMemberRole("member");
        },
        onError: (error) => toast.error(error.message),
      },
    );
  }

  function handleMemberRoleChange(member: WorkspaceMemberRecord, role: string | null) {
    const userId = workspaceMemberUserId(member);
    if (!selectedWorkspaceId || !userId || !role || role === member.role) return;
    updateMember.mutate(
      {
        workspace_id: selectedWorkspaceId,
        user_id: userId,
        role,
      },
      {
        onSuccess: () => toast.success("Member role updated"),
        onError: (error) => toast.error(error.message),
      },
    );
  }

  function handleRemoveMember(member: WorkspaceMemberRecord) {
    const userId = workspaceMemberUserId(member);
    if (!selectedWorkspaceId || !userId) return;
    removeMember.mutate(
      {
        workspace_id: selectedWorkspaceId,
        user_id: userId,
      },
      {
        onSuccess: () => toast.success("Member removed"),
        onError: (error) => toast.error(error.message),
      },
    );
  }

  function handleDeleteWorkspace() {
    if (!selectedWorkspaceId) return;
    deleteWorkspace.mutate(selectedWorkspaceId, {
      onSuccess: () => {
        clearSelectedWorkspace();
        setDeleteOpen(false);
        toast.success("Workspace deleted");
      },
      onError: (error) => toast.error(error.message),
    });
  }

  const columns: Column<WorkspaceMemberRecord>[] = [
    {
      header: "Member",
      cell: (member) => {
        const userId = workspaceMemberUserId(member);
        const user = users.find((entry) => entry.id === userId);
        return (
          <div>
            <div className="font-medium">{userLabel(user)}</div>
            <div className="text-xs text-muted-foreground">{user?.username ? `@${user.username}` : userId}</div>
          </div>
        );
      },
    },
    {
      header: "Role",
      cell: (member) => (
        <Select value={member.role || "member"} onValueChange={(value) => handleMemberRoleChange(member, value)}>
          <SelectTrigger className="w-32">
            <SelectValue placeholder="Role" />
          </SelectTrigger>
          <SelectContent>
            {MEMBER_ROLES.map((role) => (
              <SelectItem key={role} value={role}>
                {role}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ),
    },
    {
      header: "Joined",
      cell: (member) => <span className="text-xs text-muted-foreground">{formatDate(workspaceMemberCreatedAt(member))}</span>,
    },
    {
      header: "Actions",
      cell: (member) => (
        <Button variant="ghost" size="sm" onClick={() => handleRemoveMember(member)} disabled={removeMember.isPending}>
          Remove
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader title="Workspace Settings" />

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Workspace Profile
            </CardTitle>
            <CardDescription>Update the active workspace metadata used across the dashboard.</CardDescription>
          </div>
          {workspaces.length > 0 ? <Badge variant="outline">{workspaces.length} available</Badge> : null}
        </CardHeader>
        <CardContent>
          {selectedWorkspace ? (
            <form className="space-y-4" onSubmit={handleWorkspaceUpdate}>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="workspace-name">Name</Label>
                  <Input
                    id="workspace-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    disabled={updateWorkspace.isPending}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="workspace-slug">Slug</Label>
                  <Input
                    id="workspace-slug"
                    value={slug}
                    onChange={(e) => setSlug(e.target.value.toLowerCase().trim().replace(/[^a-z0-9-]+/g, "-"))}
                    disabled={updateWorkspace.isPending}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="workspace-description">Description</Label>
                <Textarea
                  id="workspace-description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  disabled={updateWorkspace.isPending}
                  placeholder="Describe what this workspace is for"
                />
              </div>
              <div className="flex flex-wrap gap-3">
                <Button type="submit" disabled={updateWorkspace.isPending}>
                  {updateWorkspace.isPending ? "Saving..." : "Save changes"}
                </Button>
                <Button type="button" variant="destructive" onClick={() => setDeleteOpen(true)}>
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete workspace
                </Button>
              </div>
            </form>
          ) : (
            <div className="text-sm text-muted-foreground">Select a workspace from the header first.</div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle>Members</CardTitle>
            <CardDescription>Add people to the active workspace and adjust their roles.</CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            <Select
              value={memberUserId || undefined}
              onValueChange={(value) => setMemberUserId(value ?? "")}
              disabled={usersLoading || availableUsers.length === 0}
            >
              <SelectTrigger className="w-64">
                <SelectValue placeholder={usersLoading ? "Loading users" : "Select a user"} />
              </SelectTrigger>
              <SelectContent>
                {availableUsers.map((user) => (
                  <SelectItem key={user.id} value={user.id}>
                    {userLabel(user)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={memberRole} onValueChange={(value) => setMemberRole(value ?? "member")}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="Role" />
              </SelectTrigger>
              <SelectContent>
                {MEMBER_ROLES.map((role) => (
                  <SelectItem key={role} value={role}>
                    {role}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button onClick={handleAddMember} disabled={!selectedWorkspaceId || !memberUserId || addMember.isPending}>
              <UserPlus className="mr-2 h-4 w-4" />
              {addMember.isPending ? "Adding..." : "Add member"}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <DataTable columns={columns} data={members} isLoading={membersLoading} emptyMessage="No workspace members found." />
          {!usersLoading && users.length === 0 ? (
            <p className="text-sm text-muted-foreground">No dashboard users are available to add yet.</p>
          ) : null}
        </CardContent>
      </Card>

      <DeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete workspace"
        description={`Delete ${selectedWorkspace?.name || selectedWorkspace?.slug || "this workspace"}? This cannot be undone.`}
        onConfirm={handleDeleteWorkspace}
        loading={deleteWorkspace.isPending}
      />
    </div>
  );
}
