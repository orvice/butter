# Repository binding PATs live outside the public model, encrypted per binding

A Workspace Repository Binding (issue #214) authenticates against its Git host with a
personal access token. The acceptance criteria require that plaintext PATs never appear
in public models, API responses, logs, or error details — a property that is easiest to
guarantee structurally rather than by remembering to redact.

The PAT is therefore **not a field of the `WorkspaceRepoBinding` proto at all**. The
proto only carries the derived, harmless facts `credential_set` and
`credential_updated_at`. Ciphertext moves through a dedicated seam on
`repobinding.Repository` (`SetCredential` / `GetCredential`), so a binding read can
never drag the credential along into a ConnectRPC response, a protojson-encoded Mongo
`spec` field, or a structured log of the message. The Mongo implementation keeps the
ciphertext in its own document field beside the spec; `Put` ignores caller-supplied
credential fields and re-derives them from stored state on every read.

Encryption is AES-GCM (`internal/secretbox`, extracted from the `mcpoauth` cipher
pattern) with a server-wide key from `git.encryption_key`. Each binding owns its own
independently encrypted PAT — credentials are not pooled or shared across workspaces.
The application service encrypts before the repository ever sees the value and decrypts
only inside `ValidateWorkspaceRepoBinding`, immediately before handing the token to the
provider adapter. Provider adapters (`internal/gitprovider`) drop response bodies when
mapping HTTP failures onto their sentinel error taxonomy, so provider errors are
credential-free by construction; the persisted validation status stores only those
sanitized details.

Consequences: replacing a credential is a write-only operation (`SetWorkspaceRepoBinding-
Credential`) that resets validation status to `UNVALIDATED`; there is no way to read a
PAT back, and rotating `git.encryption_key` invalidates stored credentials (validation
then fails with a "replace the credential" precondition error rather than decrypting
garbage). If later phases need per-binding key rotation or KMS-backed keys, the cipher
sits behind one constructor and can be swapped without touching the repository seam.
