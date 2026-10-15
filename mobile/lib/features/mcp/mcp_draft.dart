import '../../core/url_validation.dart';
import '../../gen/agents/v1/agent.pb.dart';

enum McpAuthKind { none, staticHeaders, oauth2 }

/// The metadata key the server stamps on a server installed from a global
/// preset.
const mcpPresetMetadataKey = 'butter.global_mcp_preset_id';

bool isInstalledPreset(MCPServer server) =>
    (server.metadata[mcpPresetMetadataKey] ?? '').isNotEmpty;

/// Why [server] cannot be edited in the App, or null.
///
/// A server installed from a global preset is shown redacted (header values
/// are masked and the client secret is empty), and an update stores exactly
/// what it is sent, so saving that view back would replace the real
/// credentials with the masks. A preset with no credentials has nothing to
/// corrupt.
String? mcpEditBlock(MCPServer server) {
  if (!isInstalledPreset(server)) return null;
  final hasHeaders = server.headers.isNotEmpty;
  final hasClient = server.hasAuth() && server.auth.oauth2.clientId.isNotEmpty;
  if (!hasHeaders && !hasClient) return null;
  return 'This server was installed from a global preset that holds credentials. Butter shows them hidden, and saving would overwrite them, so it cannot be edited here. Remove it and install the preset again, or edit it on the platform.';
}

McpAuthKind mcpAuthKindOf(MCPServer server) {
  final type = server.hasAuth()
      ? server.auth.type
      : MCPServerAuthType.MCP_SERVER_AUTH_TYPE_UNSPECIFIED;
  switch (type) {
    case MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2:
      return McpAuthKind.oauth2;
    case MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS:
      return McpAuthKind.staticHeaders;
    case MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE:
      return McpAuthKind.none;
    default:
      // An unspecified type keeps its old behavior: headers are sent when
      // present, otherwise nothing.
      return server.headers.isNotEmpty
          ? McpAuthKind.staticHeaders
          : McpAuthKind.none;
  }
}

/// One header of a static-header server. A header that came from the server
/// carries its stored value; the form never shows it, only offers to keep,
/// replace or remove it.
class McpHeaderRow {
  McpHeaderRow.stored(this.key, String stored) : original = stored;
  McpHeaderRow.added({this.key = '', this.replacement = ''}) : original = null;

  String key;
  final String? original;
  bool removed = false;

  /// A new value for a stored header, or the value of an added one.
  String? replacement;

  bool get isStored => original != null;

  /// The value that will be sent, or null for a removed header.
  String? get value => removed ? null : (replacement ?? original);
}

final _headerName = RegExp(r"^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$");

List<String> splitList(String text) {
  final out = <String>[];
  for (final part in text.split(RegExp(r'[\s,]+'))) {
    final t = part.trim();
    if (t.isNotEmpty && !out.contains(t)) out.add(t);
  }
  return out;
}

/// The editable form of an MCP server. It is built from the stored message and
/// builds the complete message back: every field the form does not edit
/// (`metadata`, `workspace_id`, fields this App does not know yet) is carried
/// over from the copy it started from.
class McpDraft {
  McpDraft.create()
    : base = null,
      transport = MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
      authKind = McpAuthKind.none,
      headers = [],
      dynamicRegistration = true,
      secretEdit = SecretEdit.keep {
    _initial = snapshot();
  }

  McpDraft.edit(MCPServer server)
    : base = server.deepCopy(),
      id = server.id,
      name = server.name,
      url = server.url,
      transport = server.transport,
      authKind = mcpAuthKindOf(server),
      headers = [
        for (final e in server.headers.entries)
          McpHeaderRow.stored(e.key, e.value),
      ],
      toolFilter = server.toolFilter.join('\n'),
      timeout = server.timeoutSeconds > 0 ? '${server.timeoutSeconds}' : '',
      dynamicRegistration =
          !(server.hasAuth() && server.auth.oauth2.clientId.isNotEmpty),
      secretEdit = SecretEdit.keep {
    if (server.hasAuth()) {
      final o = server.auth.oauth2;
      clientId = o.clientId;
      scopes = o.scopes.join(' ');
      authorizationUrl = o.authorizationUrl;
      tokenUrl = o.tokenUrl;
      resourceMetadataUrl = o.resourceMetadataUrl;
      authorizationServerUrl = o.authorizationServerUrl;
      resource = o.resource;
    }
    headers.sort((a, b) => a.key.compareTo(b.key));
    _initial = snapshot();
  }

  final MCPServer? base;
  bool get isCreate => base == null;

  String id = '';
  String name = '';
  String url = '';
  MCPServerTransport transport;
  McpAuthKind authKind;
  List<McpHeaderRow> headers;
  String toolFilter = '';
  String timeout = '';

  // OAuth 2.0
  bool dynamicRegistration;
  String clientId = '';
  SecretEdit secretEdit;
  String newSecret = '';
  String scopes = '';
  String authorizationUrl = '';
  String tokenUrl = '';
  String resourceMetadataUrl = '';
  String authorizationServerUrl = '';
  String resource = '';

  late final String _initial;

  /// Whether a stored client secret exists to keep, replace or clear.
  bool get hasStoredSecret =>
      base != null &&
      base!.hasAuth() &&
      base!.auth.oauth2.clientSecret.isNotEmpty;

  String snapshot() => [
    id,
    name,
    url,
    '$transport',
    '$authKind',
    for (final h in headers)
      '${h.key}=${h.removed ? '<removed>' : (h.replacement ?? '<kept>')}',
    toolFilter,
    timeout,
    '$dynamicRegistration',
    clientId,
    '$secretEdit',
    newSecret,
    scopes,
    authorizationUrl,
    tokenUrl,
    resourceMetadataUrl,
    authorizationServerUrl,
    resource,
  ].join('\u0000');

  bool get dirty => snapshot() != _initial;

  /// Whether the credentials this form sends would travel unencrypted.
  bool get sendsCredentialsInClear =>
      isCleartextUrl(url) &&
      (authKind == McpAuthKind.staticHeaders ||
          (authKind == McpAuthKind.oauth2 && !dynamicRegistration));

  // ---- Validation (the server validates again) ----

  String? validateId() {
    if (!isCreate) return null;
    final t = id.trim();
    if (t.isEmpty) return 'Enter an ID.';
    if (t.contains(RegExp(r'\s'))) return 'An ID cannot contain spaces.';
    return null;
  }

  String? validateName() => name.trim().isEmpty ? 'Enter a name.' : null;

  String? validateUrl() => validateHttpUrl(url, required: true, what: 'a URL');

  String? validateTimeout() {
    final t = timeout.trim();
    if (t.isEmpty) return null;
    final n = int.tryParse(t);
    if (n == null || n < 0 || n > 2147483647) {
      return 'Enter whole seconds, or leave empty for the default.';
    }
    return null;
  }

  String? validateHeaderKey(McpHeaderRow row) {
    if (row.removed) return null;
    final key = row.key.trim();
    if (key.isEmpty) return 'Enter a header name.';
    if (!_headerName.hasMatch(key)) return 'Not a valid header name.';
    final same = headers.where(
      (h) => !h.removed && h.key.trim().toLowerCase() == key.toLowerCase(),
    );
    if (same.length > 1) return 'This header is listed twice.';
    return null;
  }

  String? validateHeaderValue(McpHeaderRow row) {
    if (row.removed) return null;
    final v = row.value;
    if (v == null || v.isEmpty) return 'Enter a value, or remove the header.';
    if (v.contains(RegExp(r'[\r\n\u0000]'))) {
      return 'A header value cannot contain line breaks.';
    }
    return null;
  }

  String? validateClientId() {
    if (authKind != McpAuthKind.oauth2 || dynamicRegistration) return null;
    return clientId.trim().isEmpty ? 'Enter the client ID.' : null;
  }

  String? validateSecret() {
    if (authKind != McpAuthKind.oauth2 || dynamicRegistration) return null;
    if (secretEdit == SecretEdit.replace && newSecret.isEmpty) {
      return 'Enter the new client secret, or choose Keep or Clear.';
    }
    return null;
  }

  String? validateOAuthUrl(String value) => authKind != McpAuthKind.oauth2
      ? null
      : validateHttpUrl(value, required: false);

  /// The first problem, or null when the form can be saved.
  String? firstError() {
    final problems = <String?>[
      validateId(),
      validateName(),
      validateUrl(),
      validateTimeout(),
    ];
    if (authKind == McpAuthKind.staticHeaders) {
      if (headers.where((h) => !h.removed).isEmpty) {
        problems.add('Add at least one header, or choose another sign-in.');
      }
      for (final h in headers) {
        problems
          ..add(validateHeaderKey(h))
          ..add(validateHeaderValue(h));
      }
    }
    if (authKind == McpAuthKind.oauth2) {
      problems
        ..add(validateClientId())
        ..add(validateSecret())
        ..add(validateOAuthUrl(authorizationUrl))
        ..add(validateOAuthUrl(tokenUrl))
        ..add(validateOAuthUrl(resourceMetadataUrl))
        ..add(validateOAuthUrl(authorizationServerUrl))
        ..add(validateOAuthUrl(resource));
    }
    for (final p in problems) {
      if (p != null) return p;
    }
    return null;
  }

  /// The complete message to store. Call only after [firstError] is null.
  MCPServer build() {
    final s = base?.deepCopy() ?? MCPServer();
    if (isCreate) s.id = id.trim();
    s
      ..name = name.trim()
      ..url = url.trim()
      ..transport = transport;
    final seconds = int.tryParse(timeout.trim()) ?? 0;
    if (s.timeoutSeconds != seconds) s.timeoutSeconds = seconds;
    s.toolFilter
      ..clear()
      ..addAll(splitList(toolFilter));
    final auth = s.ensureAuth();
    switch (authKind) {
      case McpAuthKind.none:
        auth
          ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE
          ..clearOauth2();
        s.headers.clear();
      case McpAuthKind.staticHeaders:
        auth
          ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS
          ..clearOauth2();
        s.headers
          ..clear()
          ..addEntries([
            for (final h in headers)
              if (h.value != null) MapEntry(h.key.trim(), h.value!),
          ]);
      case McpAuthKind.oauth2:
        auth.type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2;
        s.headers.clear();
        final o = auth.ensureOauth2();
        if (dynamicRegistration) {
          o
            ..clearClientId()
            ..clearClientSecret();
        } else {
          o.clientId = clientId.trim();
          o.clientSecret = switch (secretEdit) {
            SecretEdit.keep => base?.auth.oauth2.clientSecret ?? '',
            SecretEdit.replace => newSecret,
            SecretEdit.clear => '',
          };
        }
        o.scopes
          ..clear()
          ..addAll(splitList(scopes));
        o
          ..authorizationUrl = authorizationUrl.trim()
          ..tokenUrl = tokenUrl.trim()
          ..resourceMetadataUrl = resourceMetadataUrl.trim()
          ..authorizationServerUrl = authorizationServerUrl.trim()
          ..resource = resource.trim();
    }
    return s;
  }
}
