import '../../gen/agents/v1/repobinding.pb.dart';

/// Deep copy of a binding for editing; unknown fields ride along.
class RepoBindingDraft {
  RepoBindingDraft._(this._binding);

  factory RepoBindingDraft.empty() =>
      RepoBindingDraft._(WorkspaceRepoBinding(contentSchemaVersion: 1));

  factory RepoBindingDraft.fromStored(WorkspaceRepoBinding stored) =>
      RepoBindingDraft._(stored.deepCopy());

  final WorkspaceRepoBinding _binding;

  String get gitHostId => _binding.gitHostId;
  set gitHostId(String v) => _binding.gitHostId = v;

  String get repository => _binding.repository;
  set repository(String v) => _binding.repository = v;

  String get branch => _binding.branch;
  set branch(String v) => _binding.branch = v;

  String get rootPath => _binding.rootPath;
  set rootPath(String v) => _binding.rootPath = v;

  RepoBindingWriteMode get writeMode => _binding.writeMode;
  set writeMode(RepoBindingWriteMode v) => _binding.writeMode = v;

  bool get credentialSet => _binding.credentialSet;

  WorkspaceRepoBinding toPutMessage() {
    final out = _binding.deepCopy()
      ..clearCredentialSet()
      ..clearCredentialUpdatedAt()
      ..clearStatus()
      ..clearCreatedAt()
      ..clearUpdatedAt()
      ..clearObservedCommitSha()
      ..clearActiveCommitSha()
      ..clearLastSyncedAt()
      ..clearLastSyncError()
      ..clearLastPublicationError()
      ..publicationErrors.clear()
      ..clearWebhookSecretSet()
      ..clearWorkspaceId();
    if (out.contentSchemaVersion == 0) {
      out.contentSchemaVersion = 1;
    }
    if (out.writeMode ==
        RepoBindingWriteMode.REPO_BINDING_WRITE_MODE_UNSPECIFIED) {
      out.writeMode =
          RepoBindingWriteMode.REPO_BINDING_WRITE_MODE_DIRECT_COMMIT;
    }
    return out;
  }

  String? validate() {
    if (gitHostId.trim().isEmpty) return 'Choose a Git host.';
    if (repository.trim().isEmpty) return 'Repository is required.';
    if (branch.trim().isEmpty) return 'Branch is required.';
    return null;
  }
}
