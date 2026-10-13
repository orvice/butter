//
//  Generated code. Do not modify.
//  source: agents/v1/repobinding.proto
//

import "package:connectrpc/connect.dart" as connect;
import "repobinding.pb.dart" as agentsv1repobinding;

/// WorkspaceRepoBindingService manages the current workspace's repository
/// binding. All RPCs derive the workspace from the X-Workspace-ID header.
/// Get is available to every workspace member; the mutating RPCs require the
/// caller to be a workspace owner or admin (global admins bypass).
abstract final class WorkspaceRepoBindingService {
  /// Fully-qualified name of the WorkspaceRepoBindingService service.
  static const name = 'agents.v1.WorkspaceRepoBindingService';

  /// GetWorkspaceRepoBinding returns the binding (unset when none exists)
  /// plus any overlapping bindings in other workspaces. Never returns
  /// credential material.
  static const getWorkspaceRepoBinding = connect.Spec(
    '/$name/GetWorkspaceRepoBinding',
    connect.StreamType.unary,
    agentsv1repobinding.GetWorkspaceRepoBindingRequest.new,
    agentsv1repobinding.GetWorkspaceRepoBindingResponse.new,
  );

  /// PutWorkspaceRepoBinding creates or replaces the workspace's binding
  /// configuration. Server-owned fields (credential, status, timestamps) are
  /// ignored on input. Every Put resets status to UNVALIDATED: each binding
  /// field (including write mode) changes what validation must prove.
  static const putWorkspaceRepoBinding = connect.Spec(
    '/$name/PutWorkspaceRepoBinding',
    connect.StreamType.unary,
    agentsv1repobinding.PutWorkspaceRepoBindingRequest.new,
    agentsv1repobinding.PutWorkspaceRepoBindingResponse.new,
  );

  /// DeleteWorkspaceRepoBinding safely detaches the workspace from Git-owned
  /// Agent Content. It first materializes the Active Revision snapshot back into
  /// database-managed Agent fields and reloads the runtime, then removes the
  /// binding and its encrypted PAT. It never modifies remote Git content, and
  /// refuses to proceed when there is no valid active snapshot unless the caller
  /// selects a supported recovery path. Requires owner/admin role.
  static const deleteWorkspaceRepoBinding = connect.Spec(
    '/$name/DeleteWorkspaceRepoBinding',
    connect.StreamType.unary,
    agentsv1repobinding.DeleteWorkspaceRepoBindingRequest.new,
    agentsv1repobinding.DeleteWorkspaceRepoBindingResponse.new,
  );

  /// OnboardWorkspaceRepository performs the one-time reconciliation of a freshly
  /// bound workspace's Agent Content with the repository (issue #219).
  /// EXPORT_CURRENT commits existing database-managed content to Git;
  /// IMPORT_REPOSITORY adopts repository content for matching Agent IDs. In both
  /// modes the workspace only switches to Git-owned content after the
  /// commit/import, read-back, validation, Effective Agent construction, and
  /// Active Revision publication all succeed. Requires owner/admin role.
  static const onboardWorkspaceRepository = connect.Spec(
    '/$name/OnboardWorkspaceRepository',
    connect.StreamType.unary,
    agentsv1repobinding.OnboardWorkspaceRepositoryRequest.new,
    agentsv1repobinding.OnboardWorkspaceRepositoryResponse.new,
  );

  /// SetWorkspaceRepoBindingCredential sets or replaces the binding's PAT.
  /// The PAT is write-only: it is encrypted before storage and never returned.
  static const setWorkspaceRepoBindingCredential = connect.Spec(
    '/$name/SetWorkspaceRepoBindingCredential',
    connect.StreamType.unary,
    agentsv1repobinding.SetWorkspaceRepoBindingCredentialRequest.new,
    agentsv1repobinding.SetWorkspaceRepoBindingCredentialResponse.new,
  );

  /// ValidateWorkspaceRepoBinding probes the repository with the stored
  /// credential (visibility/read, branch existence, write capability, PR/MR
  /// capability as required by the write mode), persists the outcome on the
  /// binding, and returns it.
  static const validateWorkspaceRepoBinding = connect.Spec(
    '/$name/ValidateWorkspaceRepoBinding',
    connect.StreamType.unary,
    agentsv1repobinding.ValidateWorkspaceRepoBindingRequest.new,
    agentsv1repobinding.ValidateWorkspaceRepoBindingResponse.new,
  );

  /// SyncWorkspaceRepository reads the bound repository's managed tree and
  /// Markdown blobs into the workspace-scoped DB cache. Requires owner or
  /// admin role. The sync is idempotent when the remote HEAD has not changed.
  static const syncWorkspaceRepository = connect.Spec(
    '/$name/SyncWorkspaceRepository',
    connect.StreamType.unary,
    agentsv1repobinding.SyncWorkspaceRepositoryRequest.new,
    agentsv1repobinding.SyncWorkspaceRepositoryResponse.new,
  );

  /// GetRepositorySyncStatus returns the binding with its current sync state
  /// without triggering a new synchronization.
  static const getRepositorySyncStatus = connect.Spec(
    '/$name/GetRepositorySyncStatus',
    connect.StreamType.unary,
    agentsv1repobinding.GetRepositorySyncStatusRequest.new,
    agentsv1repobinding.GetRepositorySyncStatusResponse.new,
  );

  /// ListRepositoryEntries returns cached directory entries for a path within
  /// the managed tree. Reads only from the DB cache; never issues Git requests.
  static const listRepositoryEntries = connect.Spec(
    '/$name/ListRepositoryEntries',
    connect.StreamType.unary,
    agentsv1repobinding.ListRepositoryEntriesRequest.new,
    agentsv1repobinding.ListRepositoryEntriesResponse.new,
  );

  /// GetRepositoryFile returns the cached content of a single Markdown file.
  /// Reads only from the DB cache; never issues Git requests.
  static const getRepositoryFile = connect.Spec(
    '/$name/GetRepositoryFile',
    connect.StreamType.unary,
    agentsv1repobinding.GetRepositoryFileRequest.new,
    agentsv1repobinding.GetRepositoryFileResponse.new,
  );

  /// PublishWorkspaceRepository validates the current observed revision's
  /// agent content and, on success, atomically advances the Active Revision
  /// and reloads the runner. Fails gracefully when validation errors are
  /// present; the binding reports per-agent errors.
  static const publishWorkspaceRepository = connect.Spec(
    '/$name/PublishWorkspaceRepository',
    connect.StreamType.unary,
    agentsv1repobinding.PublishWorkspaceRepositoryRequest.new,
    agentsv1repobinding.PublishWorkspaceRepositoryResponse.new,
  );

  /// ConfigureWebhookSecret generates (or replaces) a webhook secret for the
  /// binding and returns the callback URL and the secret (shown only once).
  static const configureWebhookSecret = connect.Spec(
    '/$name/ConfigureWebhookSecret',
    connect.StreamType.unary,
    agentsv1repobinding.ConfigureWebhookSecretRequest.new,
    agentsv1repobinding.ConfigureWebhookSecretResponse.new,
  );

  /// AcceptRepositoryBaseline accepts a diverged (force-pushed) HEAD as the
  /// new baseline, re-syncs and publishes. Requires owner/admin role.
  static const acceptRepositoryBaseline = connect.Spec(
    '/$name/AcceptRepositoryBaseline',
    connect.StreamType.unary,
    agentsv1repobinding.AcceptRepositoryBaselineRequest.new,
    agentsv1repobinding.AcceptRepositoryBaselineResponse.new,
  );

  /// CommitAgentContent applies a changeset of PUT and DELETE operations to
  /// managed Agent Content paths and produces a single Git commit. In
  /// DIRECT_COMMIT mode the commit lands on the bound branch; in
  /// CHANGE_REQUEST mode a PR/MR is opened instead. Content is validated
  /// before the commit is created. Requires owner/admin role.
  static const commitAgentContent = connect.Spec(
    '/$name/CommitAgentContent',
    connect.StreamType.unary,
    agentsv1repobinding.CommitAgentContentRequest.new,
    agentsv1repobinding.CommitAgentContentResponse.new,
  );

  /// RollbackAgentContent restores managed Agent Content from a previously
  /// published revision in a new Git commit. The target revision's content
  /// is read from Git, validated, and committed to the bound branch (or a
  /// PR/MR in CHANGE_REQUEST mode). Requires owner/admin role.
  static const rollbackAgentContent = connect.Spec(
    '/$name/RollbackAgentContent',
    connect.StreamType.unary,
    agentsv1repobinding.RollbackAgentContentRequest.new,
    agentsv1repobinding.RollbackAgentContentResponse.new,
  );

  /// PurgeAgentContent permanently deletes an Agent's managed Content files from
  /// the repository in a new Git commit. It is a separate owner/admin action
  /// from DeleteAgent (which only tombstones the DB entity and retains content).
  /// It refuses when the effective path is still claimed by any non-deleted
  /// Agent in this workspace or in any workspace that shares the path via an
  /// overlapping binding. Requires owner/admin role.
  static const purgeAgentContent = connect.Spec(
    '/$name/PurgeAgentContent',
    connect.StreamType.unary,
    agentsv1repobinding.PurgeAgentContentRequest.new,
    agentsv1repobinding.PurgeAgentContentResponse.new,
  );
}
