import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { ImageUp, Trash2, Upload } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useChangePassword, useUpdateProfile, type AuthUser } from "@/api/auth";
import { useUploadAvatar } from "@/api/uploads";
import { useAuth } from "@/hooks/use-auth";

function displayName(user: { display_name?: string; displayName?: string; username: string }) {
  return user.display_name || user.displayName || user.username;
}

function avatarUrl(user: { avatar_url?: string; avatarUrl?: string }) {
  return user.avatar_url || user.avatarUrl || "";
}

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();

  if (!user) return null;

  return <ProfileContent key={user.id} user={user} refreshUser={refreshUser} />;
}

function ProfileContent({ user, refreshUser }: { user: AuthUser; refreshUser: (user: AuthUser) => void }) {
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();
  const uploadAvatar = useUploadAvatar();

  const [name, setName] = useState(() => displayName(user));
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const storedAvatar = avatarUrl(user);

  const avatarPreview = useMemo(() => {
    if (!avatarFile) return "";
    return URL.createObjectURL(avatarFile);
  }, [avatarFile]);

  useEffect(() => {
    return () => {
      if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    };
  }, [avatarPreview]);

  function handleAvatarFileChange(file: File | undefined) {
    if (!file) {
      setAvatarFile(null);
      return;
    }
    if (!file.type.startsWith("image/")) {
      toast.error("Choose an image file");
      return;
    }
    setAvatarFile(file);
  }

  function handleUploadAvatar() {
    if (!avatarFile) {
      toast.error("Choose an avatar image");
      return;
    }
    uploadAvatar.mutate(
      { file: avatarFile },
      {
        onSuccess: (res) => {
          updateProfile.mutate(
            { display_name: name.trim() || user.username, avatar_url: res.url },
            {
              onSuccess: (resp) => {
                if (resp.user) refreshUser(resp.user);
                setAvatarFile(null);
                toast.success("Avatar updated");
              },
              onError: (e) => toast.error(e.message),
            },
          );
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function handleRemoveAvatar() {
    if (!storedAvatar) return;
    updateProfile.mutate(
      { display_name: name.trim() || user.username, avatar_url: "" },
      {
        onSuccess: (res) => {
          if (res.user) refreshUser(res.user);
          setAvatarFile(null);
          toast.success("Avatar removed");
        },
        onError: (e) => toast.error(e.message),
      },
    );
  }

  function handleUpdateProfile() {
    const trimmed = name.trim();
    if (!trimmed) {
      toast.error("Display name is required");
      return;
    }
    updateProfile.mutate(
      { display_name: trimmed, avatar_url: storedAvatar },
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
          <CardTitle>Avatar</CardTitle>
          <CardDescription>Upload an image to the configured static storage.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-md border bg-muted">
              {avatarPreview || storedAvatar ? (
                <img
                  src={avatarPreview || storedAvatar}
                  alt=""
                  className="h-full w-full object-cover"
                />
              ) : (
                <ImageUp className="h-8 w-8 text-muted-foreground" />
              )}
            </div>
            <div className="min-w-0 flex-1 space-y-2">
              <Label htmlFor="avatar-file">Image file</Label>
              <Input
                id="avatar-file"
                type="file"
                accept="image/png,image/jpeg,image/gif,image/webp"
                onChange={(e) => handleAvatarFileChange(e.target.files?.[0])}
              />
              <div className="text-xs text-muted-foreground">PNG, JPEG, GIF, and WebP images are supported.</div>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              onClick={handleUploadAvatar}
              disabled={!avatarFile || uploadAvatar.isPending || updateProfile.isPending}
            >
              <Upload className="mr-2 h-4 w-4" />
              {uploadAvatar.isPending || updateProfile.isPending ? "Uploading..." : "Upload avatar"}
            </Button>
            {storedAvatar ? (
              <Button
                type="button"
                variant="outline"
                onClick={handleRemoveAvatar}
                disabled={updateProfile.isPending}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Remove avatar
              </Button>
            ) : null}
          </div>
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
