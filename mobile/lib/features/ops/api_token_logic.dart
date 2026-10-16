import '../../gen/agents/v1/api_token.pb.dart';

/// Personal / automation API tokens (not daemon worker credentials).
bool isUserApiToken(APIToken token) =>
    token.kind == APITokenKind.API_TOKEN_KIND_USER;

String apiTokenExpiryLabel(APIToken token) {
  if (token.hasExpiresAt()) {
    final at = token.expiresAt.toDateTime().toLocal();
    return 'Expires ${at.toIso8601String()}';
  }
  return 'Does not expire';
}
