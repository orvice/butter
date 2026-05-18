import { useEffect, useState } from "react";
import { toast } from "sonner";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useChangePassword, useUpdateProfile } from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";

function displayName(user: { display_name?: string; displayName?: string; username: string }) {
  return user.display_name || user.displayName || user.username;
}

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();

  const [name, setName] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  useEffect(() => {
    if (user) setName(displayName(user));
  }, [user]);

  if (!user) return null;

  function handleUpdateProfile() {
    const trimmed = name.trim();
    if (!trimmed) {
      toast.error("Display name is required");
      return;
    }
    updateProfile.mutate(
      { display_name: trimmed },
      {
        onSuccess: (res) => {
          toast.success("Profile updated");
          if (res.user) refreshUser(res.user);
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function handleChangePassword() {
    if (!currentPassword) {
      toast.error("Current password is required");
      return;
    }
    if (!newPassword) {
      toast.error("New password is required");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("New passwords do not match");
      return;
    }
    changePassword.mutate(
      { current_password: currentPassword, new_password: newPassword },
      {
        onSuccess: () => {
          toast.success("Password changed");
          setCurrentPassword("");
          setNewPassword("");
          setConfirmPassword("");
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Profile" />

      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Your dashboard identity and sign-in details.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span>@{user.username}</span>
            <Badge variant={user.role === "admin" ? "default" : "secondary"}>{user.role || "user"}</Badge>
          </div>
          <div className="space-y-2">
            <Label htmlFor="display-name">Display name</Label>
            <Input id="display-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <Button onClick={handleUpdateProfile} disabled={updateProfile.isPending}>
            {updateProfile.isPending ? "Saving…" : "Save profile"}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Password</CardTitle>
          <CardDescription>Change the password you use to sign in.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="current-password">Current password</Label>
            <Input
              id="current-password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-password">New password</Label>
            <Input
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirm-password">Confirm new password</Label>
            <Input
              id="confirm-password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <Button onClick={handleChangePassword} disabled={changePassword.isPending}>
            {changePassword.isPending ? "Updating…" : "Change password"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
