# Object Storage, Static Assets & Artifacts

Butter can persist user-uploaded public assets (avatars, static files) and
private ADK artifacts in S3-compatible object stores. It can also persist
workspace-scoped Agent Files: text file spaces that agents mount through the
built-in `agent_files` toolset. Storage is delegated to the
[`butterfly.orx.me/core/store/s3`](https://butterfly.orz.ee/stores/s3.html)
helper bundled with the Butterfly core framework, so any service that the core
supports (AWS S3, MinIO, Cloudflare R2, Backblaze B2, etc.) works out of the
box.

## 1. Configure the S3 client

Add a `store.s3.<name>` block to `config/butter.yaml`. Each entry registers a
named client that the rest of the app can retrieve via
`s3.GetClient("<name>")` / `s3.GetBucket("<name>")`.

```yaml
store:
  s3:
    assets:
      endpoint: "s3.amazonaws.com"
      access_key_id: "AKIA..."
      secret_access_key: "..."
      region: "us-east-1"
      bucket: "butter-assets"
      use_ssl: true
      use_path_style: false

    # MinIO / self-hosted example
    local:
      endpoint: "localhost:9000"
      ak: "minioadmin"        # shorthand for access_key_id
      sk: "minioadmin"        # shorthand for secret_access_key
      region: "us-east-1"
      bucket: "butter-local"
      use_ssl: false
      use_path_style: true    # required for MinIO
```

The core framework initializes every client at startup. Boot logs show
`initialize s3 client name=... bucket=...` for each entry.

## 2. Enable static uploads

Once a client is registered, point Butter's upload service at it via the
top-level `static` block:

```yaml
static:
  s3_bucket: "assets"                  # must match a store.s3.<key>
  key_prefix: "butter"                 # all object keys prefixed with this
  cdn_base_url: "https://cdn.example.com"
  public_base_url: ""                  # fallback when CDN is unset
  max_upload_bytes: 5242880            # 5 MiB; 0 = default
```

| Field             | Description |
|-------------------|-------------|
| `s3_bucket`       | The `store.s3.<name>` entry to use. Empty disables uploads. |
| `key_prefix`      | Prepended to every object key (e.g. for sharing a bucket). |
| `cdn_base_url`    | Public CDN base URL. When set, returned URLs use this host. |
| `public_base_url` | Non-CDN fallback (e.g. the bucket's direct https URL). |
| `max_upload_bytes`| Per-request size limit. Default 5 MiB. |

Returned URLs are built as:

```
<cdn_base_url>/<key_prefix>/avatars/<owner_kind>/<owner_id>/<timestamp>-<random>.<ext>
```

If both `cdn_base_url` and `public_base_url` are empty, URLs fall back to
`s3://<bucket>/<key>` — useful for local development.

## 3. Enable ADK artifact persistence

Agents and tools may use ADK artifacts for per-app/user/session blobs such as
generated files or tool outputs. Configure a private S3 bucket via the top-level
`artifact` block:

```yaml
artifact:
  s3_bucket: "artifacts"      # must match a store.s3.<key>
  key_prefix: "artifacts"    # optional object key prefix
```

| Field | Description |
|-------|-------------|
| `s3_bucket` | The `store.s3.<name>` entry to use. Empty disables artifact persistence. |
| `key_prefix` | Prepended to every artifact object key. |

Artifact keys are internal and versioned by ADK identity:

```
<key_prefix>/<app_name>/<user_id>/<session_id>/<file_name>/<version>
<key_prefix>/<app_name>/<user_id>/user/<file_name>/<version>
```

Keep artifact buckets private. Unlike static uploads, artifacts are not returned
through a CDN URL by Butter and may contain user-specific or model-generated
content. If `artifact.s3_bucket` is empty or the named S3 client is not
registered, Butter runs without artifact persistence.

## 4. Enable Agent Files

Agent Files are workspace-owned text file spaces. Agents opt in by mounting
specific spaces in `Agent.config.file_mounts`; the runtime then exposes
`agent_files_*` tools for listing, reading, writing, appending, deleting, and
searching only those mounted paths.

```yaml
agent_files:
  s3_bucket: "agent-files"        # must match a store.s3.<key>
  key_prefix: "agent-files"       # optional object key prefix
  max_file_bytes: 262144          # 256 KiB default
```

| Field | Description |
|-------|-------------|
| `s3_bucket` | The `store.s3.<name>` entry to use for file contents. Empty falls back to in-memory content storage. |
| `key_prefix` | Prepended to every Agent Files object key. |
| `max_file_bytes` | Maximum UTF-8 text size accepted by write/append tools. Default 256 KiB. |

Metadata for file spaces and files is stored in MongoDB when
`storage_backend: mongo`; file contents are versioned in S3. In memory mode,
both metadata and contents are process-local.

Agent mount example:

```yaml
agents:
  - name: research-agent
    config:
      file_mounts:
        - space_id: product-docs
          mount_path: /docs
          permission: AGENT_FILE_MOUNT_PERMISSION_READ
        - space_id: research-notes
          mount_path: /notes
          permission: AGENT_FILE_MOUNT_PERMISSION_READ_WRITE
```

## 5. Upload endpoints

All endpoints sit behind the standard auth middleware (dashboard Bearer session,
root token, or workspace-bound API token; user/root-token callers should include
`X-Workspace-ID`).

### `POST /api/uploads/avatar`

Uploads an avatar for the **currently authenticated user**. Multipart form:

| Field          | Required | Notes |
|----------------|----------|-------|
| `file`         | yes      | The image bytes. |
| `content_type` | no       | Overrides the Content-Type detected from the part header. Must be one of `image/png`, `image/jpeg`, `image/gif`, `image/webp`. |

Response (200):

```json
{
  "key": "butter/avatars/user/u-123/20260518123045-9f3a1b2c.png",
  "url": "https://cdn.example.com/butter/avatars/user/u-123/20260518123045-9f3a1b2c.png",
  "content_type": "image/png",
  "size": 24580
}
```

Store the returned `url` on the user / agent record via the
[Auth/Agent services](api.md) — for user avatars call
`AuthService.UpdateProfile` with the `avatar_url` field; for agents put
the URL on `Agent.metadata.icon_url`.

### `POST /api/uploads/avatar/:owner_kind/:owner_id`

Uploads an avatar for an arbitrary owner. `owner_kind` is a free-form
string (e.g. `user`, `agent`, `workspace`). Admins may upload for any owner.
Non-admin users may upload their own `user` avatar and may upload `agent`
icons in their current workspace.

### `POST /api/uploads/static` (admin only)

General-purpose static asset upload. Multipart form:

| Field          | Required | Notes |
|----------------|----------|-------|
| `file`         | yes      | The asset bytes. |
| `name`         | no       | Override the object name. Defaults to the file's basename. |
| `content_type` | no       | Override the Content-Type. Defaults to the form part's type or `application/octet-stream`. |

Returns the same shape as `/avatar`. Asset is placed at
`<key_prefix>/static/<name>`.

## 6. Storing avatar URLs

- **User avatars** are persisted on `User.avatar_url`. After a successful
  `POST /api/uploads/avatar` the dashboard automatically calls
  `AuthService.UpdateProfile` with the returned `url` so the value
  survives logout/refresh; `Me` and `Login` responses both carry it.
  Pass an empty `avatar_url` to clear.
- **Agent icons** are persisted on `Agent.metadata.icon_url` (a free-form
  metadata map; no dedicated proto field yet). The dashboard reads this
  key when rendering agent cards.

Keys include a timestamp + random suffix, so each upload produces a new
cacheable URL — overwriting an old avatar is safe.

## 7. Troubleshooting

| Symptom | Likely cause |
|---------|--------------|
| `503 static storage is not configured` | `static.s3_bucket` is empty or doesn't match a `store.s3.<key>`. |
| `500 ... s3 client "xxx" is not configured` | The named client failed to initialize at startup — check core logs. |
| `artifact service disabled (artifact.s3_bucket not set)` | Artifact persistence is intentionally disabled. |
| `artifact service disabled: s3 client not registered` | `artifact.s3_bucket` does not match an initialized `store.s3.<key>`. |
| `agent files content store using memory` | `agent_files.s3_bucket` is empty; file contents will not survive process restart. |
| `agent files content store falling back to memory` | `agent_files.s3_bucket` does not match an initialized `store.s3.<key>`. |
| `413 payload exceeds max size` | Body larger than `static.max_upload_bytes`. Bump the limit or shrink the asset. |
| `415 unsupported content type` | Avatar endpoint only accepts PNG / JPEG / GIF / WebP. |
