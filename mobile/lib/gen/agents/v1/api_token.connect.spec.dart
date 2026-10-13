//
//  Generated code. Do not modify.
//  source: agents/v1/api_token.proto
//

import "package:connectrpc/connect.dart" as connect;
import "api_token.pb.dart" as agentsv1api_token;

abstract final class APITokenService {
  /// Fully-qualified name of the APITokenService service.
  static const name = 'agents.v1.APITokenService';

  static const listAPITokens = connect.Spec(
    '/$name/ListAPITokens',
    connect.StreamType.unary,
    agentsv1api_token.ListAPITokensRequest.new,
    agentsv1api_token.ListAPITokensResponse.new,
  );

  static const createAPIToken = connect.Spec(
    '/$name/CreateAPIToken',
    connect.StreamType.unary,
    agentsv1api_token.CreateAPITokenRequest.new,
    agentsv1api_token.CreateAPITokenResponse.new,
  );

  static const revokeAPIToken = connect.Spec(
    '/$name/RevokeAPIToken',
    connect.StreamType.unary,
    agentsv1api_token.RevokeAPITokenRequest.new,
    agentsv1api_token.RevokeAPITokenResponse.new,
  );
}
