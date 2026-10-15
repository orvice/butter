import 'package:butter_mobile/features/mcp/mcp_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/core/url_validation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:fixnum/fixnum.dart';

MCPServer _stored({
  MCPServerAuthType type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE,
}) {
  final s = MCPServer(
    id: 'srv',
    name: 'Docs',
    url: 'https://mcp.example.com/mcp',
    transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
    workspaceId: 'w1',
  );
  s.metadata['owner'] = 'platform';
  s.auth = MCPServerAuth(type: type);
  return s;
}

void main() {
  group('create', () {
    test('a minimal server needs an ID, a name and an absolute URL', () {
      final d = McpDraft.create();
      expect(d.firstError(), isNotNull);
      d.id = 'docs';
      d.name = 'Docs';
      expect(d.firstError(), contains('URL'));
      d.url = 'mcp.example.com';
      expect(d.firstError(), contains('absolute'));
      d.url = 'https://mcp.example.com/mcp';
      expect(d.firstError(), isNull);
      final s = d.build();
      expect(s.id, 'docs');
      expect(
        s.transport,
        MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
      );
      expect(s.auth.type, MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE);
      expect(s.headers, isEmpty);
    });

    test('an ID cannot hold spaces and is trimmed', () {
      final d = McpDraft.create()
        ..id = 'a b'
        ..name = 'n'
        ..url = 'https://x.example';
      expect(d.validateId(), isNotNull);
      d.id = '  ok  ';
      expect(d.validateId(), isNull);
      expect(d.build().id, 'ok');
    });

    test('is dirty only once something is entered', () {
      final d = McpDraft.create();
      expect(d.dirty, isFalse);
      d.name = 'x';
      expect(d.dirty, isTrue);
    });

    test('timeout is whole seconds, empty meaning the default', () {
      final d = McpDraft.create()
        ..id = 'a'
        ..name = 'n'
        ..url = 'https://x.example';
      d.timeout = '-1';
      expect(d.validateTimeout(), isNotNull);
      d.timeout = '1.5';
      expect(d.validateTimeout(), isNotNull);
      d.timeout = '45';
      expect(d.build().timeoutSeconds, 45);
      d.timeout = '';
      expect(d.build().timeoutSeconds, 0);
    });

    test(
      'tool filter splits on lines, commas and spaces and drops repeats',
      () {
        final d = McpDraft.create()
          ..id = 'a'
          ..name = 'n'
          ..url = 'https://x.example'
          ..toolFilter = 'echo, search\nfetch  echo';
        expect(d.build().toolFilter, ['echo', 'search', 'fetch']);
      },
    );
  });

  group('static headers', () {
    McpDraft draft() => McpDraft.create()
      ..id = 'a'
      ..name = 'n'
      ..url = 'https://x.example'
      ..authKind = McpAuthKind.staticHeaders;

    test('need at least one valid, unique header', () {
      final d = draft();
      expect(d.firstError(), contains('header'));
      d.headers.add(McpHeaderRow.added(key: 'Authorization', replacement: ''));
      expect(d.firstError(), contains('value'));
      d.headers.single.replacement = 'Bearer abc';
      expect(d.firstError(), isNull);
      d.headers.add(McpHeaderRow.added(key: 'authorization', replacement: 'x'));
      expect(d.firstError(), contains('twice'));
      d.headers.last.key = 'Bad Name';
      expect(d.firstError(), contains('valid header name'));
      d.headers.last.key = 'X-Api';
      d.headers.last.replacement = 'a\r\nInjected: 1';
      expect(d.firstError(), contains('line breaks'));
    });

    test('are sent as the map, trimmed of key padding', () {
      final d = draft();
      d.headers.add(McpHeaderRow.added(key: ' X-Key ', replacement: 'v'));
      final s = d.build();
      expect(
        s.auth.type,
        MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
      );
      expect(s.headers, {'X-Key': 'v'});
    });

    test('keep, replace and remove a stored header', () {
      final stored =
          _stored(type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS)
            ..headers['Authorization'] = 'Bearer old'
            ..headers['X-Keep'] = 'kept-value'
            ..headers['X-Drop'] = 'gone';
      final d = McpDraft.edit(stored);
      expect(d.authKind, McpAuthKind.staticHeaders);
      final byKey = {for (final h in d.headers) h.key: h};
      byKey['Authorization']!.replacement = 'Bearer new';
      byKey['X-Drop']!.removed = true;
      expect(d.firstError(), isNull);
      expect(d.build().headers, {
        'Authorization': 'Bearer new',
        'X-Keep': 'kept-value',
      });
    });

    test('an unchanged edit keeps every stored value and is not dirty', () {
      final stored = _stored(
        type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
      )..headers['Authorization'] = 'Bearer old';
      final d = McpDraft.edit(stored);
      expect(d.dirty, isFalse);
      expect(d.build(), stored);
    });

    test('cleartext transport with credentials is flagged', () {
      final d = draft()..url = 'http://x.example';
      expect(d.sendsCredentialsInClear, isTrue);
      d.url = 'https://x.example';
      expect(d.sendsCredentialsInClear, isFalse);
    });
  });

  group('OAuth 2.0', () {
    MCPServer oauthStored() {
      final s = _stored(type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2);
      s.auth.ensureOauth2()
        ..clientId = 'cid'
        ..clientSecret = 'csecret'
        ..scopes.addAll(['read', 'write'])
        ..tokenUrl = 'https://as.example/token';
      return s;
    }

    test('dynamic registration sends no client', () {
      final d = McpDraft.edit(oauthStored())..dynamicRegistration = true;
      final o = d.build().auth.oauth2;
      expect(o.clientId, '');
      expect(o.clientSecret, '');
      expect(o.scopes, ['read', 'write']);
    });

    test(
      'the client secret is kept, replaced or cleared, never blanked by accident',
      () {
        final d = McpDraft.edit(oauthStored());
        expect(d.dynamicRegistration, isFalse);
        expect(d.hasStoredSecret, isTrue);
        expect(d.build().auth.oauth2.clientSecret, 'csecret');
        d.secretEdit = SecretEdit.replace;
        expect(d.validateSecret(), isNotNull);
        d.newSecret = 'fresh';
        expect(d.validateSecret(), isNull);
        expect(d.build().auth.oauth2.clientSecret, 'fresh');
        d.secretEdit = SecretEdit.clear;
        expect(d.build().auth.oauth2.clientSecret, '');
        expect(d.build().auth.oauth2.clientId, 'cid');
      },
    );

    test('a preconfigured client needs its ID', () {
      final d = McpDraft.create()
        ..id = 'a'
        ..name = 'n'
        ..url = 'https://x.example'
        ..authKind = McpAuthKind.oauth2
        ..dynamicRegistration = false;
      expect(d.firstError(), contains('client ID'));
      d.clientId = 'cid';
      expect(d.firstError(), isNull);
    });

    test('the endpoint overrides must be absolute http(s) URLs', () {
      final d = McpDraft.edit(oauthStored())..tokenUrl = 'ftp://as';
      expect(d.firstError(), contains('http'));
      d.tokenUrl = '';
      expect(d.firstError(), isNull);
      expect(d.build().auth.oauth2.tokenUrl, '');
    });

    test(
      'switching to another sign-in drops the OAuth settings and headers',
      () {
        final d = McpDraft.edit(oauthStored())..authKind = McpAuthKind.none;
        final s = d.build();
        expect(s.auth.type, MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE);
        expect(s.auth.hasOauth2(), isFalse);
        expect(s.headers, isEmpty);
      },
    );
  });

  group('preservation', () {
    test('unknown fields, metadata and the workspace survive an edit', () {
      final stored = _stored();
      stored.unknownFields.mergeVarintField(99, Int64(7));
      stored.auth.unknownFields.mergeVarintField(98, Int64(8));
      final d = McpDraft.edit(stored)..name = 'Renamed';
      final s = d.build();
      expect(s.name, 'Renamed');
      expect(s.metadata['owner'], 'platform');
      expect(s.workspaceId, 'w1');
      expect(s.id, 'srv');
      expect(s.unknownFields.getField(99)?.varints.single.toInt(), 7);
      expect(s.auth.unknownFields.getField(98)?.varints.single.toInt(), 8);
    });

    test('the ID of a stored server cannot be changed', () {
      final d = McpDraft.edit(_stored())..id = 'other';
      expect(d.build().id, 'srv');
    });
  });

  group('installed presets', () {
    MCPServer installed({bool headers = false, bool client = false}) {
      final s = _stored();
      s.metadata[mcpPresetMetadataKey] = 'preset-x';
      if (headers) s.headers['Authorization'] = '***';
      if (client) s.ensureAuth().ensureOauth2().clientId = 'cid';
      return s;
    }

    test('one holding redacted credentials cannot be edited', () {
      expect(isInstalledPreset(installed()), isTrue);
      expect(mcpEditBlock(installed(headers: true)), isNotNull);
      expect(mcpEditBlock(installed(client: true)), isNotNull);
    });

    test('one without credentials, and any ordinary server, can', () {
      expect(mcpEditBlock(installed()), isNull);
      expect(mcpEditBlock(_stored()..headers['A'] = 'b'), isNull);
    });
  });

  test('validateHttpUrl', () {
    expect(validateHttpUrl('', required: false), isNull);
    expect(validateHttpUrl('', required: true), isNotNull);
    expect(validateHttpUrl('ftp://a', required: true), isNotNull);
    expect(validateHttpUrl('https://', required: true), isNotNull);
    expect(validateHttpUrl('http://localhost:8080/x', required: true), isNull);
  });
}
