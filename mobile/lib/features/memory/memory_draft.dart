import '../../gen/agents/v1/workspace_memory.pb.dart';

enum MemoryApiKeyAction { keep, replace, clear }

/// Edits to the workspace mem0 connection. Secrets are never persisted locally.
final class MemoryConfigDraft {
  MemoryConfigDraft({WorkspaceMemoryConfig? base})
    : baseUrl = base?.baseUrl ?? '',
      enabled = base?.enabled ?? true,
      credentialSet = base?.credentialSet ?? false;

  String baseUrl;
  bool enabled;
  final bool credentialSet;

  MemoryApiKeyAction keyAction = MemoryApiKeyAction.keep;
  String replacementKey = '';

  bool get dirty => true;

  PutWorkspaceMemoryConfigRequest buildPut() {
    final req = PutWorkspaceMemoryConfigRequest(
      baseUrl: baseUrl,
      enabled: enabled,
    );
    switch (keyAction) {
      case MemoryApiKeyAction.keep:
        break;
      case MemoryApiKeyAction.replace:
        req.apiKey = replacementKey;
      case MemoryApiKeyAction.clear:
        req.apiKey = '';
    }
    return req;
  }

  void clearTypedSecret() {
    replacementKey = '';
    if (keyAction == MemoryApiKeyAction.replace) {
      keyAction = MemoryApiKeyAction.keep;
    }
  }
}

bool isHttpUrl(String value) {
  final uri = Uri.tryParse(value.trim());
  return uri != null &&
      (uri.scheme == 'http' || uri.scheme == 'https') &&
      uri.host.isNotEmpty;
}

String? validateMemoryBaseUrl(String value) {
  if (value.trim().isEmpty) return 'Base URL is required.';
  if (!isHttpUrl(value)) return 'Enter an absolute http(s) URL.';
  return null;
}
