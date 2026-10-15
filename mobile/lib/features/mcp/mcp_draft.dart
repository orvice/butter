import '../../core/url_validation.dart';
import '../../gen/agents/v1/agent.pb.dart';

enum McpAuthKind { none, staticHeaders, oauth2 }

/// The metadata key the server stamps on a server installed from a global
/// preset.
const mcpPresetMetadataKey = 'butter.global_mcp_preset_id';

bool isInstalledPreset(MCPServer server) =>
    (server.metadata[mcpPresetMetadataKey] ?? '').isNotEmpty;

/// Whether reads of [server] hide its credentials: a server installed from a
/// global preset is returned with header values masked and its client secret
/// empty, so what the App holds is not what is stored. Saving such a server
/// never sends the masks back; kept credentials are named to the server
/// instead (`keep_header_keys`, `keep_oauth_client_secret`).
bool mcpHidesCredentials(MCPServer server) =>
    isInstalledPreset(server) &&
    (server.headers.isNotEmpty ||
        (server.hasAuth() && server.auth.oauth2.clientId.isNotEmpty));

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
  McpHeaderRow.stored(this.key, String stored, {this.hidden = false})
    : original = stored;
  McpHeaderRow.added({this.key = '', this.replacement = ''})
    : original = null,
      hidden = false;

  String key;
  final String? original;
  bool removed = false;

  /// A new value for a stored header, or the value of an added one.
  String? replacement;

  bool get isStored => original != null;

  /// The stored value is not known to the App (a preset's masked value): it can
  /// only be kept by name, replaced or removed.
  final bool hidden;

  /// A stored header whose value is left as it is.
  bool get kept => isStored && !removed && replacement == null;

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

/// What to send to save an edited server: the whole message plus the stored
/// credentials it keeps without having them.
class McpSave {
  const McpSave(
    this.server, {
    this.keepHeaderKeys = const [],
    this.keepClientSecret = false,
  });
  final MCPServer server;
  final List<String> keepHeaderKeys;
  final bool keepClientSecret;

  /// Whether the save depends on the server keeping stored credentials the App
  /// cannot read, which needs a server that supports it.
  bool get keepsCredentials => keepHeaderKeys.isNotEmpty || keepClientSecret;
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
          McpHeaderRow.stored(
            e.key,
            e.value,
            hidden: isInstalledPreset(server),
          ),
      ],
      toolFilter = server.toolFilter.join('\n'),
      timeout = server.timeoutSeconds > 0 ? '${server.timeoutSeconds}' : '',
      dynamicRegistration =
          !(server.hasAuth() && server.auth.oauth2.clientId.isNotEmpty),
      secretEdit = SecretEdit.keep {
    if (server.hasAuth()) {
      final o = server.auth.oauth2;
      clientId = o.clientId;
      originalClientId = o.clientId;
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
  String originalClientId = '';
  SecretEdit secretEdit;
  String newSecret = '';
  String scopes = '';
  String authorizationUrl = '';
  String tokenUrl = '';
  String resourceMetadataUrl = '';
  String authorizationServerUrl = '';
  String resource = '';

  late final String _initial;

  /// Whether the server's credentials are hidden from this form.
  bool get hidesCredentials => base != null && mcpHidesCredentials(base!);

  /// Whether a stored client secret exists to keep, replace or clear. When the
  /// secret is hidden, a stored client ID means one may exist.
  bool get hasStoredSecret =>
      base != null &&
      base!.hasAuth() &&
      (base!.auth.oauth2.clientSecret.isNotEmpty ||
          (hidesCredentials && base!.auth.oauth2.clientId.isNotEmpty));

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
    if (hidesCredentials &&
        secretEdit == SecretEdit.keep &&
        originalClientId.isNotEmpty &&
        clientId.trim() != originalClientId) {
      return 'A client secret belongs to its client ID. Choose Replace and enter the new one, or Clear.';
    }
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
    if (authKind == McpAuthKind.staticHeaders &&
        headers.where((h) => !h.removed).isEmpty) {
      problems.add('Add at least one header, or choose another sign-in.');
    }
    // Headers are sent with every request whatever the sign-in, so they are
    // validated whatever the sign-in.
    for (final h in headers) {
      problems
        ..add(validateHeaderKey(h))
        ..add(validateHeaderValue(h));
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

  /// What leaving this form would discard from what is stored, each as a
  /// sentence that names its target. Empty for harmless edits.
  List<String> discards() {
    final b = base;
    if (b == null) return const [];
    final out = <String>[
      for (final h in headers)
        if (h.isStored && h.removed)
          'Remove the header "${h.key}" and its stored value.',
    ];
    final hadClient = b.hasAuth() && b.auth.oauth2.clientId.isNotEmpty;
    final hadSecret = hasStoredSecret;
    if (authKind == McpAuthKind.oauth2 &&
        !dynamicRegistration &&
        secretEdit == SecretEdit.clear &&
        hadSecret) {
      out.add('Clear the stored client secret.');
    }
    if (mcpAuthKindOf(b) == McpAuthKind.oauth2 && hadClient) {
      if (authKind != McpAuthKind.oauth2) {
        out.add(
          'Stop using OAuth 2.0: the stored client ID${hadSecret ? ' and client secret are' : ' is'} discarded.',
        );
      } else if (dynamicRegistration) {
        out.add(
          'Register the client automatically: the stored client ID${hadSecret ? ' and client secret are' : ' is'} discarded.',
        );
      }
    }
    return out;
  }

  /// Forgets every typed secret (new header values, a new client secret) once
  /// they have been submitted. Replacements stay chosen but empty, so sending
  /// another one needs them typed again.
  void clearTypedSecrets() {
    newSecret = '';
    for (final h in headers) {
      if (h.replacement != null) h.replacement = '';
    }
  }

  /// The complete message to store. Call only after [firstError] is null.
  /// For a server whose credentials are hidden, use [buildSave]: kept headers
  /// and a kept client secret are named there, not in this message.
  MCPServer build() => buildSave().server;

  /// Whether saving this draft would rely on the server keeping credentials
  /// the form cannot read. Always agrees with [buildSave]'s keep fields.
  bool get needsKeep =>
      !isCreate &&
      (headers.any((h) => !h.removed && h.hidden && h.kept) ||
          (authKind == McpAuthKind.oauth2 &&
              !dynamicRegistration &&
              secretEdit == SecretEdit.keep &&
              hidesCredentials &&
              originalClientId.isNotEmpty));

  McpSave buildSave() {
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

    // Headers are sent whatever the sign-in, so they are written whatever the
    // sign-in: only an explicit removal drops one.
    final keepHeaderKeys = <String>[];
    s.headers.clear();
    for (final h in headers) {
      if (h.removed) continue;
      if (h.hidden && h.kept) {
        keepHeaderKeys.add(h.key.trim());
        continue;
      }
      s.headers[h.key.trim()] = h.value!;
    }

    // The sign-in type is only rewritten when the person changed it, so an
    // unspecified type that behaves like another stays as stored.
    final kindChanged = base == null || mcpAuthKindOf(base!) != authKind;
    var keepClientSecret = false;
    switch (authKind) {
      case McpAuthKind.none:
        if (kindChanged) {
          s.ensureAuth()
            ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE
            ..clearOauth2();
        }
      case McpAuthKind.staticHeaders:
        if (kindChanged) {
          s.ensureAuth()
            ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS
            ..clearOauth2();
        }
      case McpAuthKind.oauth2:
        final auth = s.ensureAuth()
          ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2;
        final o = auth.ensureOauth2();
        if (dynamicRegistration) {
          o
            ..clearClientId()
            ..clearClientSecret();
        } else {
          o.clientId = clientId.trim();
          if (secretEdit == SecretEdit.keep && hidesCredentials) {
            // The secret is not known here: name it, never send a mask.
            o.clearClientSecret();
            keepClientSecret = originalClientId.isNotEmpty;
          } else {
            o.clientSecret = switch (secretEdit) {
              SecretEdit.keep => base?.auth.oauth2.clientSecret ?? '',
              SecretEdit.replace => newSecret,
              SecretEdit.clear => '',
            };
          }
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
    return McpSave(
      s,
      keepHeaderKeys: keepHeaderKeys,
      keepClientSecret: keepClientSecret,
    );
  }
}
