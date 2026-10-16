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

    test('switching away from OAuth drops its settings', () {
      final d = McpDraft.edit(oauthStored())..authKind = McpAuthKind.none;
      final s = d.build();
      expect(s.auth.type, MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE);
      expect(s.auth.hasOauth2(), isFalse);
    });
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

  group('headers under every sign-in', () {
    MCPServer withHeaders(MCPServerAuthType type) => _stored(type: type)
      ..headers['Authorization'] = 'Bearer old'
      ..headers['X-Team'] = 'alpha';

    for (final type in [
      MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE,
      MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
    ]) {
      test('a name-only edit keeps the headers of a $type server', () {
        final stored = withHeaders(type);
        if (type == MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2) {
          stored.auth.ensureOauth2().clientId = 'cid';
        }
        final d = McpDraft.edit(stored)..name = 'Renamed';
        expect(d.firstError(), isNull);
        final built = d.build();
        expect(built.name, 'Renamed');
        expect(built.headers, {
          'Authorization': 'Bearer old',
          'X-Team': 'alpha',
        });
        expect(built.auth.type, type);
      });
    }

    test('only an explicit removal drops a header, and a new one is added', () {
      final d = McpDraft.edit(
        withHeaders(MCPServerAuthType.MCP_SERVER_AUTH_TYPE_NONE),
      );
      final byKey = {for (final h in d.headers) h.key: h};
      byKey['X-Team']!.removed = true;
      d.headers.add(McpHeaderRow.added(key: 'X-New', replacement: 'n'));
      expect(d.firstError(), isNull);
      expect(d.build().headers, {'Authorization': 'Bearer old', 'X-New': 'n'});
    });

    test('they are validated even when the sign-in is not Headers', () {
      final d = McpDraft.edit(_stored())
        ..headers.add(McpHeaderRow.added(key: 'Bad Name', replacement: 'v'));
      expect(d.firstError(), contains('valid header name'));
      d.headers.last.key = 'X-Ok';
      d.headers.last.replacement = '';
      expect(d.firstError(), contains('value'));
    });

    test('switching sign-in no longer discards headers', () {
      final d = McpDraft.edit(
        withHeaders(MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS),
      )..authKind = McpAuthKind.none;
      expect(d.build().headers, hasLength(2));
      expect(d.discards(), isEmpty);
    });

    test('an unspecified type that behaves like another stays as stored', () {
      final stored = withHeaders(
        MCPServerAuthType.MCP_SERVER_AUTH_TYPE_UNSPECIFIED,
      );
      final d = McpDraft.edit(stored)..name = 'Renamed';
      expect(d.authKind, McpAuthKind.staticHeaders);
      expect(
        d.build().auth.type,
        MCPServerAuthType.MCP_SERVER_AUTH_TYPE_UNSPECIFIED,
      );
    });
  });

  group('installed presets keep what they hide', () {
    MCPServer hidden({bool oauth = true}) {
      final s = _stored(
        type: oauth
            ? MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2
            : MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
      );
      s.metadata[mcpPresetMetadataKey] = 'preset-x';
      s.headers['Authorization'] = '***';
      s.headers['X-Team'] = '***';
      if (oauth) s.ensureAuth().ensureOauth2().clientId = 'cid';
      return s;
    }

    test('are recognised by the marker and their credentials', () {
      expect(isInstalledPreset(hidden()), isTrue);
      expect(mcpHidesCredentials(hidden()), isTrue);
      expect(mcpHidesCredentials(_stored()), isFalse);
      final plain = _stored()..metadata[mcpPresetMetadataKey] = 'p';
      expect(mcpHidesCredentials(plain), isFalse);
      expect(
        mcpHidesCredentials(_stored()..headers['A'] = 'b'),
        isFalse,
        reason: 'an ordinary server returns what it stores',
      );
    });

    test('a name-only edit sends no mask and names every kept credential', () {
      final d = McpDraft.edit(hidden())..name = 'Renamed';
      expect(d.firstError(), isNull);
      final save = d.buildSave();
      expect(save.server.name, 'Renamed');
      expect(save.server.headers, isEmpty);
      expect(save.keepHeaderKeys, unorderedEquals(['Authorization', 'X-Team']));
      expect(save.keepClientSecret, isTrue);
      expect(save.server.auth.oauth2.clientSecret, '');
      expect(save.server.metadata[mcpPresetMetadataKey], 'preset-x');
      expect(save.server.writeToBuffer().toString(), isNot(contains('***')));
      expect(d.discards(), isEmpty);
    });

    test('needsKeep always agrees with the keep fields of the save', () {
      void agrees(McpDraft d) =>
          expect(d.needsKeep, d.buildSave().keepsCredentials);
      agrees(McpDraft.edit(hidden()));
      agrees(McpDraft.edit(hidden(oauth: false)));
      agrees(McpDraft.edit(hidden())..authKind = McpAuthKind.none);
      agrees(McpDraft.edit(hidden())..dynamicRegistration = true);
      final replaced = McpDraft.edit(hidden())
        ..secretEdit = SecretEdit.replace
        ..newSecret = 'x';
      for (final h in replaced.headers) {
        h.removed = true;
      }
      agrees(replaced);
      expect(replaced.needsKeep, isFalse);
      agrees(McpDraft.edit(_stored()..headers['A'] = 'b'));
      expect(McpDraft.create().needsKeep, isFalse);
    });

    test('a replaced header is sent as a value, a removed one is omitted', () {
      final d = McpDraft.edit(hidden());
      final byKey = {for (final h in d.headers) h.key: h};
      byKey['Authorization']!.replacement = 'Bearer fresh';
      byKey['X-Team']!.removed = true;
      final save = d.buildSave();
      expect(save.server.headers, {'Authorization': 'Bearer fresh'});
      expect(save.keepHeaderKeys, isEmpty);
      expect(d.discards(), [contains('"X-Team"')]);
    });

    test('a replaced or cleared client secret is explicit', () {
      final d = McpDraft.edit(hidden());
      expect(d.hasStoredSecret, isTrue, reason: 'one may exist, unseen');
      d.secretEdit = SecretEdit.replace;
      d.newSecret = 'fresh';
      var save = d.buildSave();
      expect(save.keepClientSecret, isFalse);
      expect(save.server.auth.oauth2.clientSecret, 'fresh');
      d.secretEdit = SecretEdit.clear;
      save = d.buildSave();
      expect(save.keepClientSecret, isFalse);
      expect(save.server.auth.oauth2.clientSecret, '');
      expect(d.discards(), [contains('client secret')]);
    });

    test('changing the client ID cannot keep the hidden secret', () {
      final d = McpDraft.edit(hidden())..clientId = 'other';
      expect(d.validateSecret(), contains('belongs to its client ID'));
      expect(d.firstError(), isNotNull);
      d.secretEdit = SecretEdit.replace;
      d.newSecret = 'new';
      expect(d.firstError(), isNull);
    });

    test('name, filter, timeout and endpoints edit with credentials kept', () {
      final d = McpDraft.edit(hidden(oauth: false))
        ..name = 'Renamed'
        ..url = 'https://other.example/mcp'
        ..toolFilter = 'echo'
        ..timeout = '30';
      final save = d.buildSave();
      expect(save.server.url, 'https://other.example/mcp');
      expect(save.server.toolFilter, ['echo']);
      expect(save.server.timeoutSeconds, 30);
      expect(save.keepHeaderKeys, hasLength(2));
      expect(save.keepClientSecret, isFalse);
    });

    test(
      'dropping OAuth or its registered client discards the hidden client',
      () {
        final off = McpDraft.edit(hidden())..authKind = McpAuthKind.none;
        expect(off.discards().single, contains('client ID'));
        final dynamicRegistration = McpDraft.edit(hidden())
          ..dynamicRegistration = true;
        expect(dynamicRegistration.discards().single, contains('client ID'));
        expect(dynamicRegistration.buildSave().keepClientSecret, isFalse);
      },
    );

    test('an ordinary server still sends and keeps its own values', () {
      final stored = _stored(
        type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
      )..headers['X'] = 'real';
      stored.auth.ensureOauth2()
        ..clientId = 'cid'
        ..clientSecret = 'csecret';
      final save = (McpDraft.edit(stored)..name = 'n2').buildSave();
      expect(save.keepHeaderKeys, isEmpty);
      expect(save.keepClientSecret, isFalse);
      expect(save.server.headers, {'X': 'real'});
      expect(save.server.auth.oauth2.clientSecret, 'csecret');
    });
  });

  group('credential-destructive changes name their target', () {
    test('only discarding edits ask; ordinary ones do not', () {
      final stored = _stored(
        type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
      )..headers['X-Key'] = 'real';
      stored.auth.ensureOauth2()
        ..clientId = 'cid'
        ..clientSecret = 'csecret';

      expect((McpDraft.edit(stored)..name = 'n').discards(), isEmpty);
      expect(
        (McpDraft.edit(stored)
              ..secretEdit = SecretEdit.replace
              ..newSecret = 'rotated')
            .discards(),
        isEmpty,
        reason: 'rotating a secret is not discarding it',
      );

      final removeHeader = McpDraft.edit(stored)..headers.single.removed = true;
      expect(removeHeader.discards().single, contains('"X-Key"'));

      final clear = McpDraft.edit(stored)..secretEdit = SecretEdit.clear;
      expect(clear.discards().single, contains('client secret'));

      final off = McpDraft.edit(stored)..authKind = McpAuthKind.staticHeaders;
      expect(off.discards().single, contains('client ID and client secret'));
    });

    test('a created server never asks', () {
      expect(McpDraft.create().discards(), isEmpty);
    });

    test('typed secrets are forgotten once submitted, replacements remain', () {
      final stored = _stored(
        type: MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
      )..headers['X-Key'] = 'real';
      stored.auth.ensureOauth2()
        ..clientId = 'cid'
        ..clientSecret = 'csecret';
      final d = McpDraft.edit(stored)
        ..secretEdit = SecretEdit.replace
        ..newSecret = 'typed-secret'
        ..headers.single.replacement = 'typed-header'
        ..headers.add(
          McpHeaderRow.added(key: 'X-New', replacement: 'typed-new'),
        );
      d.clearTypedSecrets();
      expect(d.newSecret, '');
      expect(d.headers.map((h) => h.replacement), ['', '']);
      expect(d.secretEdit, SecretEdit.replace);
      expect(d.firstError(), isNotNull, reason: 'it must be typed again');
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
