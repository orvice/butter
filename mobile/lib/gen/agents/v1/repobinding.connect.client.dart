//
//  Generated code. Do not modify.
//  source: agents/v1/repobinding.proto
//

import "package:connectrpc/connect.dart" as connect;
import "repobinding.pb.dart" as agentsv1repobinding;
import "repobinding.connect.spec.dart" as specs;

/// WorkspaceRepoBindingService manages the current workspace's repository
/// binding. All RPCs derive the workspace from the X-Workspace-ID header.
/// Get is available to every workspace member; the mutating RPCs require the
/// caller to be a workspace owner or admin (global admins bypass).
extension type WorkspaceRepoBindingServiceClient (connect.Transport _transport) {
  /// GetWorkspaceRepoBinding returns the binding (unset when none exists)
  /// plus any overlapping bindings in other workspaces. Never returns
  /// credential material.
  Future<agentsv1repobinding.GetWorkspaceRepoBindingResponse> getWorkspaceRepoBinding(
    agentsv1repobinding.GetWorkspaceRepoBindingRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.getWorkspaceRepoBinding,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PutWorkspaceRepoBinding creates or replaces the workspace's binding
  /// configuration. Server-owned fields (credential, status, timestamps) are
  /// ignored on input. Every Put resets status to UNVALIDATED: each binding
  /// field (including write mode) changes what validation must prove.
  Future<agentsv1repobinding.PutWorkspaceRepoBindingResponse> putWorkspaceRepoBinding(
    agentsv1repobinding.PutWorkspaceRepoBindingRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.putWorkspaceRepoBinding,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteWorkspaceRepoBinding safely detaches the workspace from Git-owned
  /// Agent Content. It first materializes the Active Revision snapshot back into
  /// database-managed Agent fields and reloads the runtime, then removes the
  /// binding and its encrypted PAT. It never modifies remote Git content, and
  /// refuses to proceed when there is no valid active snapshot unless the caller
  /// selects a supported recovery path. Requires owner/admin role.
  Future<agentsv1repobinding.DeleteWorkspaceRepoBindingResponse> deleteWorkspaceRepoBinding(
    agentsv1repobinding.DeleteWorkspaceRepoBindingRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.deleteWorkspaceRepoBinding,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// OnboardWorkspaceRepository performs the one-time reconciliation of a freshly
  /// bound workspace's Agent Content with the repository (issue #219).
  /// EXPORT_CURRENT commits existing database-managed content to Git;
  /// IMPORT_REPOSITORY adopts repository content for matching Agent IDs. In both
  /// modes the workspace only switches to Git-owned content after the
  /// commit/import, read-back, validation, Effective Agent construction, and
  /// Active Revision publication all succeed. Requires owner/admin role.
  Future<agentsv1repobinding.OnboardWorkspaceRepositoryResponse> onboardWorkspaceRepository(
    agentsv1repobinding.OnboardWorkspaceRepositoryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.onboardWorkspaceRepository,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SetWorkspaceRepoBindingCredential sets or replaces the binding's PAT.
  /// The PAT is write-only: it is encrypted before storage and never returned.
  Future<agentsv1repobinding.SetWorkspaceRepoBindingCredentialResponse> setWorkspaceRepoBindingCredential(
    agentsv1repobinding.SetWorkspaceRepoBindingCredentialRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.setWorkspaceRepoBindingCredential,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ValidateWorkspaceRepoBinding probes the repository with the stored
  /// credential (visibility/read, branch existence, write capability, PR/MR
  /// capability as required by the write mode), persists the outcome on the
  /// binding, and returns it.
  Future<agentsv1repobinding.ValidateWorkspaceRepoBindingResponse> validateWorkspaceRepoBinding(
    agentsv1repobinding.ValidateWorkspaceRepoBindingRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.validateWorkspaceRepoBinding,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SyncWorkspaceRepository reads the bound repository's managed tree and
  /// Markdown blobs into the workspace-scoped DB cache. Requires owner or
  /// admin role. The sync is idempotent when the remote HEAD has not changed.
  Future<agentsv1repobinding.SyncWorkspaceRepositoryResponse> syncWorkspaceRepository(
    agentsv1repobinding.SyncWorkspaceRepositoryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.syncWorkspaceRepository,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetRepositorySyncStatus returns the binding with its current sync state
  /// without triggering a new synchronization.
  Future<agentsv1repobinding.GetRepositorySyncStatusResponse> getRepositorySyncStatus(
    agentsv1repobinding.GetRepositorySyncStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.getRepositorySyncStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListRepositoryEntries returns cached directory entries for a path within
  /// the managed tree. Reads only from the DB cache; never issues Git requests.
  Future<agentsv1repobinding.ListRepositoryEntriesResponse> listRepositoryEntries(
    agentsv1repobinding.ListRepositoryEntriesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.listRepositoryEntries,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetRepositoryFile returns the cached content of a single Markdown file.
  /// Reads only from the DB cache; never issues Git requests.
  Future<agentsv1repobinding.GetRepositoryFileResponse> getRepositoryFile(
    agentsv1repobinding.GetRepositoryFileRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.getRepositoryFile,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PublishWorkspaceRepository validates the current observed revision's
  /// agent content and, on success, atomically advances the Active Revision
  /// and reloads the runner. Fails gracefully when validation errors are
  /// present; the binding reports per-agent errors.
  Future<agentsv1repobinding.PublishWorkspaceRepositoryResponse> publishWorkspaceRepository(
    agentsv1repobinding.PublishWorkspaceRepositoryRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.publishWorkspaceRepository,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ConfigureWebhookSecret generates (or replaces) a webhook secret for the
  /// binding and returns the callback URL and the secret (shown only once).
  Future<agentsv1repobinding.ConfigureWebhookSecretResponse> configureWebhookSecret(
    agentsv1repobinding.ConfigureWebhookSecretRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.configureWebhookSecret,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// AcceptRepositoryBaseline accepts a diverged (force-pushed) HEAD as the
  /// new baseline, re-syncs and publishes. Requires owner/admin role.
  Future<agentsv1repobinding.AcceptRepositoryBaselineResponse> acceptRepositoryBaseline(
    agentsv1repobinding.AcceptRepositoryBaselineRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.acceptRepositoryBaseline,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CommitAgentContent applies a changeset of PUT and DELETE operations to
  /// managed Agent Content paths and produces a single Git commit. In
  /// DIRECT_COMMIT mode the commit lands on the bound branch; in
  /// CHANGE_REQUEST mode a PR/MR is opened instead. Content is validated
  /// before the commit is created. Requires owner/admin role.
  Future<agentsv1repobinding.CommitAgentContentResponse> commitAgentContent(
    agentsv1repobinding.CommitAgentContentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.commitAgentContent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// RollbackAgentContent restores managed Agent Content from a previously
  /// published revision in a new Git commit. The target revision's content
  /// is read from Git, validated, and committed to the bound branch (or a
  /// PR/MR in CHANGE_REQUEST mode). Requires owner/admin role.
  Future<agentsv1repobinding.RollbackAgentContentResponse> rollbackAgentContent(
    agentsv1repobinding.RollbackAgentContentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.rollbackAgentContent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// PurgeAgentContent permanently deletes an Agent's managed Content files from
  /// the repository in a new Git commit. It is a separate owner/admin action
  /// from DeleteAgent (which only tombstones the DB entity and retains content).
  /// It refuses when the effective path is still claimed by any non-deleted
  /// Agent in this workspace or in any workspace that shares the path via an
  /// overlapping binding. Requires owner/admin role.
  Future<agentsv1repobinding.PurgeAgentContentResponse> purgeAgentContent(
    agentsv1repobinding.PurgeAgentContentRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceRepoBindingService.purgeAgentContent,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
