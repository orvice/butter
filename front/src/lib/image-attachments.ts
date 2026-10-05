// Client-side mirror of the backend multimodal input limits every entry
// point enforces (internal/userinput: the RPCs' InputParts and AG-UI content
// parts), so users get immediate feedback instead of a round-trip rejection.
// The backend remains the source of truth.
export const ALLOWED_IMAGE_MIME_TYPES = [
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
] as const;

export const MAX_IMAGE_BYTES = 10 * 1024 * 1024; // 10 MiB per image
export const MAX_TOTAL_IMAGE_BYTES = 20 * 1024 * 1024; // 20 MiB per request
export const MAX_IMAGE_COUNT = 10;

export const IMAGE_FILE_ACCEPT = ALLOWED_IMAGE_MIME_TYPES.join(",");

export interface AcceptImageFilesResult {
  accepted: File[];
  errors: string[];
}

function isAllowedImageMimeType(type: string): boolean {
  return (ALLOWED_IMAGE_MIME_TYPES as readonly string[]).includes(type);
}

function formatMiB(bytes: number): string {
  return `${Math.round(bytes / (1024 * 1024))} MiB`;
}

// acceptImageFiles validates `incoming` files against the backend limits,
// taking already-attached files into account for the count and total-size
// caps. Valid files are returned in `accepted`; each rejection produces a
// user-facing error string.
export function acceptImageFiles(existing: File[], incoming: File[]): AcceptImageFilesResult {
  const accepted: File[] = [];
  const errors: string[] = [];
  let count = existing.length;
  let totalBytes = existing.reduce((sum, f) => sum + f.size, 0);

  for (const file of incoming) {
    if (!isAllowedImageMimeType(file.type)) {
      errors.push(`${file.name}: unsupported type ${file.type || "unknown"}; accepted: JPEG, PNG, GIF, WebP`);
      continue;
    }
    if (file.size > MAX_IMAGE_BYTES) {
      errors.push(`${file.name}: exceeds the ${formatMiB(MAX_IMAGE_BYTES)} per-image limit`);
      continue;
    }
    if (count + 1 > MAX_IMAGE_COUNT) {
      errors.push(`${file.name}: at most ${MAX_IMAGE_COUNT} images per message`);
      continue;
    }
    if (totalBytes + file.size > MAX_TOTAL_IMAGE_BYTES) {
      errors.push(`${file.name}: total attachments exceed the ${formatMiB(MAX_TOTAL_IMAGE_BYTES)} limit`);
      continue;
    }
    count += 1;
    totalBytes += file.size;
    accepted.push(file);
  }

  return { accepted, errors };
}

export function validateImageFiles(files: File[]): string[] {
  return acceptImageFiles([], files).errors;
}
