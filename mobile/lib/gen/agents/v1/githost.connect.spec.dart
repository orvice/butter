//
//  Generated code. Do not modify.
//  source: agents/v1/githost.proto
//

import "package:connectrpc/connect.dart" as connect;
import "githost.pb.dart" as agentsv1githost;

/// GitHostService manages the platform-level allowlist of Git hosts.
/// List/Get are available to any authenticated user (workspace owners need
/// them to configure a binding); Create/Update/Delete require global admin.
abstract final class GitHostService {
  /// Fully-qualified name of the GitHostService service.
  static const name = 'agents.v1.GitHostService';

  static const listGitHosts = connect.Spec(
    '/$name/ListGitHosts',
    connect.StreamType.unary,
    agentsv1githost.ListGitHostsRequest.new,
    agentsv1githost.ListGitHostsResponse.new,
  );

  static const getGitHost = connect.Spec(
    '/$name/GetGitHost',
    connect.StreamType.unary,
    agentsv1githost.GetGitHostRequest.new,
    agentsv1githost.GetGitHostResponse.new,
  );

  static const createGitHost = connect.Spec(
    '/$name/CreateGitHost',
    connect.StreamType.unary,
    agentsv1githost.CreateGitHostRequest.new,
    agentsv1githost.CreateGitHostResponse.new,
  );

  static const updateGitHost = connect.Spec(
    '/$name/UpdateGitHost',
    connect.StreamType.unary,
    agentsv1githost.UpdateGitHostRequest.new,
    agentsv1githost.UpdateGitHostResponse.new,
  );

  static const deleteGitHost = connect.Spec(
    '/$name/DeleteGitHost',
    connect.StreamType.unary,
    agentsv1githost.DeleteGitHostRequest.new,
    agentsv1githost.DeleteGitHostResponse.new,
  );
}
