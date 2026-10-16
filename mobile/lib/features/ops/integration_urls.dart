import '../../core/server_address.dart';

/// Connect RPC base used by this App (`/api` on the configured host).
String connectApiBase(ServerAddress address) =>
    address.endpoint('api').toString();

/// OpenAI-compatible HTTP API base (`/api/v1`).
String openAiCompatibleBase(ServerAddress address) =>
    address.endpoint('api/v1').toString();
