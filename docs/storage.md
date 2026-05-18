# Object Storage & Static Assets

Butter persists user-uploaded assets (avatars, static files) in an
S3-compatible object store. Storage is delegated to the
[`butterfly.orx.me/core/store/s3`](https://butterfly.orz.ee/stores/s3.html)
helper bundled with the Butterfly core framework, so any service that the
core supports (AWS S3, MinIO, Cloudflare R2, Backblaze B2, etc.) works out
of the box.

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

## 3. Upload endpoints

All endpoints sit behind the standard auth middleware (cookie session or
API token + `X-Workspace-ID`).

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

Store the returned `url` on the user / agent record (or use the
[Auth/Agent services](api.md) once the proto fields land).

### `POST /api/uploads/avatar/:owner_kind/:owner_id`

Uploads an avatar for an arbitrary owner. `owner_kind` is a free-form
string (e.g. `user`, `agent`, `workspace`). Cross-owner uploads require the
admin role; self uploads (`owner_kind=user`, `owner_id=<caller>`) are
always allowed.

### `POST /api/uploads/static` (admin only)

General-purpose static asset upload. Multipart form:

| Field          | Required | Notes |
|----------------|----------|-------|
| `file`         | yes      | The asset bytes. |
| `name`         | no       | Override the object name. Defaults to the file's basename. |
| `content_type` | no       | Override the Content-Type. Defaults to the form part's type or `application/octet-stream`. |

Returns the same shape as `/avatar`. Asset is placed at
`<key_prefix>/static/<name>`.

## 4. Storing avatar URLs

Until `User.avatar_url` / `Agent.avatar_url` are added to the proto schema,
clients should persist the returned `url` field themselves (e.g. on a user
metadata document) and render it directly from the CDN. Stable keys make it
safe to overwrite an existing avatar: keys include a timestamp + random
suffix so each upload produces a new cacheable URL.

## 5. Troubleshooting

| Symptom | Likely cause |
|---------|--------------|
| `503 static storage is not configured` | `static.s3_bucket` is empty or doesn't match a `store.s3.<key>`. |
| `500 ... s3 client "xxx" is not configured` | The named client failed to initialize at startup — check core logs. |
| `413 payload exceeds max size` | Body larger than `static.max_upload_bytes`. Bump the limit or shrink the asset. |
| `415 unsupported content type` | Avatar endpoint only accepts PNG / JPEG / GIF / WebP. |
