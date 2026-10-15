//
//  Generated code. Do not modify.
//  source: agents/v1/api_token.proto
//

import "package:connectrpc/connect.dart" as connect;
import "api_token.pb.dart" as agentsv1api_token;
import "api_token.connect.spec.dart" as specs;

extension type APITokenServiceClient(connect.Transport _transport) {
  Future<agentsv1api_token.ListAPITokensResponse> listAPITokens(
    agentsv1api_token.ListAPITokensRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.APITokenService.listAPITokens,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1api_token.CreateAPITokenResponse> createAPIToken(
    agentsv1api_token.CreateAPITokenRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.APITokenService.createAPIToken,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1api_token.RevokeAPITokenResponse> revokeAPIToken(
    agentsv1api_token.RevokeAPITokenRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.APITokenService.revokeAPIToken,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
