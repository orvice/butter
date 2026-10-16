//
//  Generated code. Do not modify.
//  source: agents/v1/githost.proto
//

import "package:connectrpc/connect.dart" as connect;
import "githost.pb.dart" as agentsv1githost;
import "githost.connect.spec.dart" as specs;

/// GitHostService manages the platform-level allowlist of Git hosts.
/// List/Get are available to any authenticated user (workspace owners need
/// them to configure a binding); Create/Update/Delete require global admin.
extension type GitHostServiceClient (connect.Transport _transport) {
  Future<agentsv1githost.ListGitHostsResponse> listGitHosts(
    agentsv1githost.ListGitHostsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GitHostService.listGitHosts,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1githost.GetGitHostResponse> getGitHost(
    agentsv1githost.GetGitHostRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GitHostService.getGitHost,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1githost.CreateGitHostResponse> createGitHost(
    agentsv1githost.CreateGitHostRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GitHostService.createGitHost,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1githost.UpdateGitHostResponse> updateGitHost(
    agentsv1githost.UpdateGitHostRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GitHostService.updateGitHost,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1githost.DeleteGitHostResponse> deleteGitHost(
    agentsv1githost.DeleteGitHostRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.GitHostService.deleteGitHost,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
