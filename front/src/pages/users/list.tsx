import { useState } from "react";
import { Navigate } from "react-router-dom";
import { toast } from "sonner";
import { UserPlus, LockKeyhole, Ban, CheckCircle2 } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { DataTable, type Column } from "@/components/data-table";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useCreateUser,
  useSetUserDisabled,
  useUpdateUserPassword,
  useUsers,
  type AuthUser,
} from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";

function displayName(user: AuthUser) {
  return user.display_name || user.displayName || user.username;
}

function createdAt(user: AuthUser) {
  const value = user.created_at || user.createdAt;
  return value ? new Date(value).toLocaleString() : "-";
}

export default function UserListPage() {
  const { user: currentUser, isAdmin, isLoading: isAuthLoading } = useAuth();
  const { data, isLoading } = useUsers({ enabled: isAdmin });
  const create = useCreateUser();
  const updatePassword = useUpdateUserPassword();
  const setDisabled = useSetUserDisabled();

  const [createOpen, setCreateOpen] = useState(false);
  const [passwordOpenFor, setPasswordOpenFor] = useState<AuthUser | null>(null);
  const [username, setUsername] = useState("");
  const [name, setName] = useState("");
  const [role, setRole] = useState("user");
  const [password, setPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");

  const users = data?.users ?? [];

  if (isAuthLoading || !currentUser) {
    return (
      <div className="space-y-6">
        <PageHeader title="Users" />
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">Loading…</CardContent>
        </Card>
      </div>
    );
  }

  if (!isAdmin) return <Navigate to="/profile" replace />;

  function resetCreateForm() {
    setUsername("");
    setName("");
    setRole("user");
    setPassword("");
  }

  function handleCreate() {
    if (!username.trim()) {
      toast.error("Username is required");
      return;
    }
    if (!password) {
      toast.error("Password is required");
      return;
    }
    create.mutate(
      {
        username: username.trim(),
        password,
        display_name: name.trim(),
        role,
      },
      {
        onSuccess: () => {
          toast.success("User created");
          setCreateOpen(false);
          resetCreateForm();
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function handleUpdatePassword() {
    if (!passwordOpenFor) return;
    if (!newPassword) {
      toast.error("Password is required");
      return;
    }
    updatePassword.mutate(
      { id: passwordOpenFor.id, password: newPassword },
      {
        onSuccess: () => {
          toast.success("Password updated");
          setPasswordOpenFor(null);
          setNewPassword("");
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function toggleDisabled(user: AuthUser) {
    const disabled = !user.disabled;
    setDisabled.mutate(
      { id: user.id, disabled },
      {
        onSuccess: () => toast.success(disabled ? "User disabled" : "User enabled"),
        onError: (e) => toast.error(e.message),
      },
    );
  }

  const columns: Column<AuthUser>[] = [
    {
      header: "User",
      cell: (u) => (
        <div>
          <div className="font-medium">{displayName(u)}</div>
          <div className="text-xs text-muted-foreground">@{u.username}</div>
        </div>
      ),
    },
    {
      header: "Role",
      cell: (u) => <Badge variant={u.role === "admin" ? "default" : "secondary"}>{u.role || "user"}</Badge>,
    },
    {
      header: "Status",
      cell: (u) => (u.disabled ? <Badge variant="destructive">Disabled</Badge> : <Badge>Active</Badge>),
    },
    {
      header: "Created",
      cell: (u) => <span className="text-xs text-muted-foreground">{createdAt(u)}</span>,
    },
    {
      header: "Actions",
      cell: (u) => (
        <div className="flex flex-wrap gap-2">
          <Button variant="ghost" size="sm" onClick={() => setPasswordOpenFor(u)}>
            <LockKeyhole className="mr-1 h-3 w-3" /> Password
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => toggleDisabled(u)}
            disabled={currentUser?.id === u.id && !u.disabled}
          >
            {u.disabled ? <CheckCircle2 className="mr-1 h-3 w-3" /> : <Ban className="mr-1 h-3 w-3" />}
            {u.disabled ? "Enable" : "Disable"}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader title="Users" />

      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>Dashboard Users</CardTitle>
            <CardDescription>Create and manage MongoDB-backed dashboard accounts.</CardDescription>
          </div>
          <Button className="w-full sm:w-auto" onClick={() => setCreateOpen(true)}>
            <UserPlus className="mr-2 h-4 w-4" /> New User
          </Button>
        </CardHeader>
        <CardContent>
          <DataTable columns={columns} data={users} isLoading={isLoading} emptyMessage="No users found." />
        </CardContent>
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create user</DialogTitle>
            <DialogDescription>Add a new account that can sign in to the dashboard.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input id="username" value={username} onChange={(e) => setUsername(e.target.value)} autoFocus />
            </div>
            <div className="space-y-2">
              <Label htmlFor="display-name">Display name</Label>
              <Input id="display-name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Role</Label>
              <Select value={role} onValueChange={(value) => setRole(value ?? "user")}>
                <SelectTrigger>
                  <SelectValue placeholder="Role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={create.isPending}>{create.isPending ? "Creating…" : "Create"}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!passwordOpenFor} onOpenChange={(open) => !open && setPasswordOpenFor(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update password</DialogTitle>
            <DialogDescription>Set a new password for {passwordOpenFor?.username}.</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="new-password">New password</Label>
            <Input id="new-password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} autoFocus />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPasswordOpenFor(null)}>Cancel</Button>
            <Button onClick={handleUpdatePassword} disabled={updatePassword.isPending}>
              {updatePassword.isPending ? "Updating…" : "Update"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
