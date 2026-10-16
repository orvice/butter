import 'package:butter_mobile/features/repo_binding/repo_binding_draft.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('toPutMessage strips server-owned fields', () {
    final stored =
        WorkspaceRepoBinding(
            gitHostId: 'h1',
            repository: 'o/r',
            branch: 'main',
            credentialSet: true,
            activeCommitSha: 'abc',
            workspaceId: 'ws',
          )
          ..status = (RepoBindingStatus()
            ..state =
                RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_OK);
    final draft = RepoBindingDraft.fromStored(stored);
    final put = draft.toPutMessage();
    expect(put.credentialSet, isFalse);
    expect(put.activeCommitSha, isEmpty);
    expect(put.hasStatus(), isFalse);
    expect(put.workspaceId, isEmpty);
    expect(put.gitHostId, 'h1');
  });
}
