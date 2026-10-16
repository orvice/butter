import 'package:butter_mobile/core/repo_binding_errors.dart';
import 'package:butter_mobile/features/repo_binding/repo_binding_controller.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('agentContentGitOwned mirrors active_commit_sha', () {
    final c = RepoBindingController(
      newApi: () => throw UnimplementedError(),
      token: 't',
      workspaceId: 'ws',
      canManage: () => true,
      onUnauthenticated: () {},
    );
    addTearDown(c.dispose);
    c.binding = WorkspaceRepoBinding(activeCommitSha: 'deadbeef');
    expect(c.agentContentGitOwned, isTrue);
    expect(c.hasActiveRevision, isTrue);
    c.binding = WorkspaceRepoBinding();
    expect(c.agentContentGitOwned, isFalse);
  });

  test('permission denied maps to owner/admin message', () {
    final msg = repoBindingFailureMessage(
      ConnectException(Code.permissionDenied, 'denied'),
      RepoBindingOperation.save,
    );
    expect(msg, contains('owner or admin'));
  });
}
