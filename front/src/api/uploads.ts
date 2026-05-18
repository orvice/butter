import { useMutation } from "@tanstack/react-query";
import { TOKEN_KEY, WORKSPACE_KEY } from "@/lib/constants";
import { ApiError } from "./client";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export interface UploadResult {
  key: string;
  url: string;
  content_type?: string;
  contentType?: string;
  size: number;
}

export interface UploadAvatarInput {
  file: File;
  ownerKind?: string;
  ownerId?: string;
}

export interface UploadStaticInput {
  file: File;
  name?: string;
  contentType?: string;
}

async function upload<T>(path: string, formData: FormData): Promise<T> {
  const token = localStorage.getItem(TOKEN_KEY);
  const workspaceId = localStorage.getItem(WORKSPACE_KEY);
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(workspaceId ? { "X-Workspace-ID": workspaceId } : {}),
    },
    body: formData,
  });

  let data: unknown;
  try {
    data = await res.json();
  } catch {
    data = undefined;
  }

  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = "/login";
    throw new ApiError("unauthenticated", "Invalid or expired token");
  }

  if (!res.ok) {
    const error = data as { error?: string; msg?: string; code?: string } | undefined;
    throw new ApiError(error?.code || "upload_failed", error?.error || error?.msg || "Upload failed");
  }

  return data as T;
}

export function uploadAvatar(input: UploadAvatarInput) {
  const formData = new FormData();
  formData.append("file", input.file);
  formData.append("content_type", input.file.type);

  const path =
    input.ownerKind && input.ownerId
      ? `/api/uploads/avatar/${encodeURIComponent(input.ownerKind)}/${encodeURIComponent(input.ownerId)}`
      : "/api/uploads/avatar";

  return upload<UploadResult>(path, formData);
}

export function uploadStatic(input: UploadStaticInput) {
  const formData = new FormData();
  formData.append("file", input.file);
  formData.append("name", input.name || input.file.name);
  formData.append("content_type", input.contentType || input.file.type);

  return upload<UploadResult>("/api/uploads/static", formData);
}

export function useUploadAvatar() {
  return useMutation({ mutationFn: uploadAvatar });
}

export function useUploadStatic() {
  return useMutation({ mutationFn: uploadStatic });
}
