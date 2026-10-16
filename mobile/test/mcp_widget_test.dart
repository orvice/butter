import 'package:butter_mobile/features/mcp/mcp_controller.dart';
import 'package:butter_mobile/features/mcp/mcp_draft.dart';
import 'package:butter_mobile/features/mcp/mcp_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';

MCPServer server(String id, {String name = ''}) => MCPServer(
  id: id,
  name: name.isEmpty ? 'Server $id' : name,
  url: 'https://mcp.example.com/mcp',
  transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
);

MCPServer headerServer(String id) {
  final s = server(id)..headers['Authorization'] = 'Bearer very-secret-token';
  s.ensureAuth().type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS;
  return s;
}

MCPServer oauthServer(String id) {
  final s = server(id);
  s.ensureAuth()
    ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2
    ..ensureOauth2();
  return s;
}

Future<void> openList(WidgetTester tester, ConfigRig h) async {
  await openPage(tester, McpListPage(client: h.client));
  await tester.runAsync(() => h.client.mcp!.load());
  await tester.pumpAndSettle();
}

Future<void> openDetail(WidgetTester tester, ConfigRig h, String id) async {
  await tester.runAsync(
    () => tester.pumpWidget(
      MaterialApp(
        home: McpDetailPage(client: h.client, id: id),
      ),
    ),
  );
  await waitFor(tester, () => h.client.mcp!.item?.id == id);
}

void main() {
  testWidgets('the list is the active Workspace only and shows no secret', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) {
        s.seed('alpha', headerServer('a1'));
        s.seed('beta', server('b1', name: 'Beta server'));
      },
    );
    await openList(tester, h);
    expect(find.text('Server a1'), findsOneWidget);
    expect(find.text('Beta server'), findsNothing);
    expect(
      find.textContaining('Static headers: Authorization'),
      findsOneWidget,
    );
    expect(find.textContaining('very-secret-token'), findsNothing);
    expect(h.server.mcpCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  testWidgets('adding a static-header server stores it and never shows the '
      'value', (tester) async {
    final h = await signedIn(tester);
    await openList(tester, h);
    expect(find.text('No MCP servers yet'), findsOneWidget);

    await tapScreen(tester, find.byKey(const Key('add-mcp')));
    await enter(tester, const Key('mcp-id'), 'docs');
    await enter(tester, const Key('mcp-name'), 'Docs');
    await enter(tester, const Key('mcp-url'), 'https://mcp.example.com/mcp');
    await tapScreen(
      tester,
      find.descendant(
        of: find.byKey(const Key('mcp-auth-kind')),
        matching: find.text('Headers'),
      ),
    );
    await tapScreen(tester, find.byKey(const Key('add-header')));
    await enter(tester, const Key('header-key-0'), 'Authorization');
    await enter(tester, const Key('header-value-0'), 'Bearer typed-secret');
    // The value is masked in its field.
    final field = tester.widget<TextField>(
      find.descendant(
        of: find.byKey(const Key('header-value-0')),
        matching: find.byType(TextField),
      ),
    );
    expect(field.obscureText, isTrue);
    await tapReal(tester, find.byKey(const Key('save-mcp')));
    await waitFor(tester, () => h.server.writes == 1);
    await tester.pumpAndSettle();

    final stored = h.server.servers['alpha']!['docs']!;
    expect(stored.headers, {'Authorization': 'Bearer typed-secret'});
    expect(
      stored.auth.type,
      MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
    );
    expect(find.byKey(const Key('mcp-docs')), findsOneWidget);
    expect(find.textContaining('typed-secret'), findsNothing);
  });

  testWidgets(
    'an invalid URL or missing header is refused before any request',
    (tester) async {
      final h = await signedIn(tester);
      await openList(tester, h);
      await tapScreen(tester, find.byKey(const Key('add-mcp')));
      await enter(tester, const Key('mcp-id'), 'docs');
      await enter(tester, const Key('mcp-name'), 'Docs');
      await enter(tester, const Key('mcp-url'), 'ftp://nope');
      await tapScreen(tester, find.byKey(const Key('save-mcp')));
      expect(find.textContaining('http or https'), findsWidgets);
      expect(h.server.writes, 0);

      await enter(tester, const Key('mcp-url'), 'https://mcp.example.com/mcp');
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('mcp-auth-kind')),
          matching: find.text('Headers'),
        ),
      );
      await tapScreen(tester, find.byKey(const Key('save-mcp')));
      expect(find.textContaining('at least one header'), findsOneWidget);
      expect(h.server.writes, 0);
    },
  );

  testWidgets('a duplicate ID is explained and nothing is repeated', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seed('alpha', server('a1')),
    );
    await openList(tester, h);
    await tapScreen(tester, find.byKey(const Key('add-mcp')));
    await enter(tester, const Key('mcp-id'), 'a1');
    await enter(tester, const Key('mcp-name'), 'Again');
    await enter(tester, const Key('mcp-url'), 'https://mcp.example.com/mcp');
    await tapReal(tester, find.byKey(const Key('save-mcp')));
    await waitFor(tester, () => h.server.writes == 1);
    expect(find.byKey(const Key('config-write-error')), findsOneWidget);
    expect(find.textContaining('already exists'), findsOneWidget);
  });

  testWidgets('editing keeps a stored header hidden and untouched', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seed('alpha', headerServer('a1')),
    );
    await openDetail(tester, h, 'a1');
    await tapScreen(tester, find.byKey(const Key('edit-mcp')));
    expect(find.text('Value kept (hidden)'), findsOneWidget);
    expect(find.textContaining('very-secret-token'), findsNothing);
    await enter(tester, const Key('mcp-name'), 'Renamed');
    await tapReal(tester, find.byKey(const Key('save-mcp')));
    await waitFor(tester, () => h.server.writes == 1);

    final stored = h.server.servers['alpha']!['a1']!;
    expect(stored.name, 'Renamed');
    expect(stored.headers['Authorization'], 'Bearer very-secret-token');
  });

  testWidgets('the connection test classifies the failure and hides the '
      'server text', (tester) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seed('alpha', server('a1')),
    );
    await openDetail(tester, h, 'a1');
    expect(find.text('Connection not tested'), findsOneWidget);

    h.server.statusState = MCPServerStatus_State.STATE_DISCONNECTED;
    h.server.statusDetail =
        'dial tcp 10.1.2.3:443: connection refused; Authorization: Bearer leak-9';
    await tapReal(tester, find.byKey(const Key('probe-mcp')));
    await waitFor(
      tester,
      () => h.client.mcp!.statuses['a1']?.kind == McpStatusKind.disconnected,
    );
    expect(
      find.textContaining('could not connect to the MCP server'),
      findsOneWidget,
    );
    expect(find.textContaining('leak-9'), findsNothing);
    expect(find.textContaining('10.1.2.3'), findsNothing);
    expect(h.server.writes, 0);
  });

  testWidgets('tools show which ones the filter allows', (tester) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seed('alpha', server('a1')),
    );
    h.server.toolList = [
      MCPTool(
        name: 'echo',
        serverId: 'a1',
        allowed: true,
        description: 'Echoes',
      ),
      MCPTool(name: 'danger', serverId: 'a1', allowed: false),
    ];
    await openDetail(tester, h, 'a1');
    await tapReal(tester, find.byKey(const Key('list-tools')));
    await waitFor(
      tester,
      () => h.client.mcp!.tools['a1']?.state.name == 'ready',
    );
    expect(find.textContaining('1 of 2 allowed'), findsOneWidget);
    expect(find.byKey(const Key('tool-echo')), findsOneWidget);
    expect(find.text('Blocked by the tool filter'), findsOneWidget);
  });

  testWidgets('delete asks first and a cancelled dialog sends nothing', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seed('alpha', server('a1')),
    );
    await openDetail(tester, h, 'a1');
    await tapReal(tester, find.byKey(const Key('delete-mcp')));
    await tester.pumpAndSettle();
    expect(find.text('Delete MCP server?'), findsOneWidget);
    await tapDialog(tester, find.text('Cancel'));
    expect(h.server.count('DeleteMCPServer'), 0);

    await tapReal(tester, find.byKey(const Key('delete-mcp')));
    await tester.pumpAndSettle();
    await tapDialog(tester, find.text('Delete'));
    await waitFor(tester, () => h.server.count('DeleteMCPServer') == 1);
    expect(h.server.servers['alpha'], isEmpty);
  });

  testWidgets('a lost reply leaves a notice and disables saving', (
    tester,
  ) async {
    final h = await signedIn(tester);
    await openList(tester, h);
    h.server.loseReply = true;
    await tapScreen(tester, find.byKey(const Key('add-mcp')));
    await enter(tester, const Key('mcp-id'), 'docs');
    await enter(tester, const Key('mcp-name'), 'Docs');
    await enter(tester, const Key('mcp-url'), 'https://mcp.example.com/mcp');
    await tapReal(tester, find.byKey(const Key('save-mcp')));
    await waitFor(tester, () => h.server.writes == 1);
    expect(find.byKey(const Key('config-write-notice')), findsOneWidget);
    final save = tester.widget<FilledButton>(find.byKey(const Key('save-mcp')));
    expect(save.onPressed, isNull);
    expect(h.server.count('CreateMCPServer'), 1);
  });

  group('global presets', () {
    MCPServer preset(String id) {
      final p = headerServer(id)
        ..headers['Authorization'] = 'Bearer preset-secret';
      return p;
    }

    testWidgets('are read-only, show no credential and install after '
        'confirmation', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.presets.add(preset('p1')),
      );
      await tester.runAsync(() => h.client.mcp!.loadPresets());
      await tester.runAsync(() => h.client.mcp!.load());
      await openPage(tester, McpPresetsPage(client: h.client));
      expect(find.text('Server p1'), findsOneWidget);
      expect(find.textContaining('Credentials hidden'), findsOneWidget);
      expect(find.textContaining('preset-secret'), findsNothing);
      // Presets are never written from the App.
      expect(find.byIcon(Icons.edit), findsNothing);

      await tapReal(tester, find.byKey(const Key('install-p1')));
      await tester.pumpAndSettle();
      expect(find.text('Install preset?'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(h.server.count('InstallGlobalMCPServer'), 0);

      await tapReal(tester, find.byKey(const Key('install-p1')));
      await tester.pumpAndSettle();
      await tapDialog(tester, find.text('Install'));
      await waitFor(
        tester,
        () => h.server.count('InstallGlobalMCPServer') == 1,
      );
      expect(find.text('Reinstall'), findsOneWidget);
      expect(
        find.textContaining('Installed in this Workspace'),
        findsOneWidget,
      );
    });

    MCPServer installedPreset() => headerServer('p1')
      ..headers['Authorization'] = 'Bearer preset-secret'
      ..metadata[mcpPresetMetadataKey] = 'p1';

    testWidgets('an installed preset holding credentials is edited with them '
        'kept, never with the masks sent back', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', installedPreset()),
      );
      await openDetail(tester, h, 'p1');
      expect(find.byKey(const Key('mcp-edit-block')), findsNothing);
      expect(
        tester
            .widget<FilledButton>(find.byKey(const Key('edit-mcp')))
            .onPressed,
        isNotNull,
      );
      // What is on screen is the hidden view: nothing secret, nothing masked.
      expect(find.textContaining('preset-secret'), findsNothing);

      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      expect(find.text('Value kept (hidden)'), findsOneWidget);
      await enter(tester, const Key('mcp-name'), 'Renamed');
      await enter(tester, const Key('mcp-timeout'), '45');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      expect(
        find.text('Discard stored credentials?'),
        findsNothing,
        reason: 'a harmless edit asks nothing',
      );
      await waitFor(
        tester,
        () => h.server.count('UpdateMCPServerKeepingCredentials') == 1,
      );
      expect(
        h.server.count('UpdateMCPServer'),
        0,
        reason: 'keeping credentials never goes through the ordinary request',
      );

      final call = h.server.mcpCalls.singleWhere(
        (c) => c.method == 'UpdateMCPServerKeepingCredentials',
      );
      final request = UpdateMCPServerKeepingCredentialsRequest.fromBuffer(
        call.body,
      );
      expect(request.keepHeaderKeys, ['Authorization']);
      expect(request.mcpServer.headers, isEmpty);
      expect(call.body, isNot(containsAllInOrder('***'.codeUnits)));
      final stored = h.server.servers['alpha']!['p1']!;
      expect(stored.name, 'Renamed');
      expect(stored.timeoutSeconds, 45);
      expect(stored.headers['Authorization'], 'Bearer preset-secret');
      expect(stored.metadata[mcpPresetMetadataKey], 'p1');
    });

    testWidgets('replacing a hidden header sends the value; removing one '
        'asks first and names it', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', installedPreset()),
      );
      await openDetail(tester, h, 'p1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      await tapScreen(tester, find.byKey(const Key('header-replace-0')));
      await enter(tester, const Key('header-value-0'), 'Bearer rotated');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('Discard stored credentials?'), findsNothing);
      await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);
      expect(
        h.server.servers['alpha']!['p1']!.headers['Authorization'],
        'Bearer rotated',
      );
    });
  });

  group('an older server that cannot keep hidden credentials', () {
    // Headers beside no sign-in, so removing them leaves a valid server.
    MCPServer installed() => server('p1')
      ..headers['Authorization'] = 'Bearer preset-secret'
      ..metadata[mcpPresetMetadataKey] = 'p1';

    testWidgets('blocks the save with an upgrade explanation and changes '
        'nothing; removing the hidden credential explicitly still works', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', installed());
          s.supportsKeeping = false;
        },
      );
      await openDetail(tester, h, 'p1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      await enter(tester, const Key('mcp-name'), 'Renamed');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await waitFor(
        tester,
        () => h.server.count('UpdateMCPServerKeepingCredentials') == 1,
      );
      expect(find.textContaining('upgrade Butter'), findsWidgets);
      expect(find.byKey(const Key('mcp-keep-unsupported')), findsOneWidget);
      expect(
        tester
            .widget<FilledButton>(find.byKey(const Key('save-mcp')))
            .onPressed,
        isNull,
        reason: 'the unsafe save is not offered again',
      );
      expect(h.server.count('UpdateMCPServer'), 0);
      expect(
        h.server.servers['alpha']!['p1']!.headers['Authorization'],
        'Bearer preset-secret',
      );
      expect(h.server.servers['alpha']!['p1']!.name, 'Server p1');

      // Removing the hidden header makes the save independent of the missing
      // method: it asks first, then uses the ordinary request.
      await tapScreen(tester, find.byKey(const Key('header-remove-0')));
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('mcp-auth-kind')),
          matching: find.text('None'),
        ),
      );
      expect(find.byKey(const Key('mcp-keep-unsupported')), findsNothing);
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pumpAndSettle();
      expect(find.text('Discard stored credentials?'), findsOneWidget);
      await tapDialogReal(tester, find.text('Save and discard'));
      await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);
      expect(h.server.servers['alpha']!['p1']!.headers, isEmpty);
      expect(h.server.count('UpdateMCPServerKeepingCredentials'), 1);
    });
  });

  group('credential-destructive saves ask first and name the target', () {
    testWidgets('removing a header: cancel sends nothing, confirm removes it', (
      tester,
    ) async {
      // Beside OAuth, since static headers need at least one header.
      final h = await signedIn(
        tester,
        seed: (s) =>
            s.seed('alpha', oauthServer('a1')..headers['X-Team'] = 'alpha'),
      );
      await openDetail(tester, h, 'a1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      await tapScreen(tester, find.byKey(const Key('header-remove-0')));
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pumpAndSettle();
      expect(find.text('Discard stored credentials?'), findsOneWidget);
      expect(find.textContaining('"X-Team"'), findsOneWidget);
      expect(find.textContaining('"Server a1"'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(h.server.count('UpdateMCPServer'), 0);
      expect(h.server.servers['alpha']!['a1']!.headers, {'X-Team': 'alpha'});

      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pumpAndSettle();
      await tapDialogReal(tester, find.text('Save and discard'));
      await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);
      expect(h.server.servers['alpha']!['a1']!.headers, isEmpty);
    });

    testWidgets('clearing the client secret, or dropping OAuth, asks first', (
      tester,
    ) async {
      final secured = oauthServer('a1');
      secured.auth.oauth2
        ..clientId = 'cid'
        ..clientSecret = 'csecret';
      final h = await signedIn(tester, seed: (s) => s.seed('alpha', secured));
      await openDetail(tester, h, 'a1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('secret-choice-Client secret')),
          matching: find.text('Clear'),
        ),
      );
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pumpAndSettle();
      expect(find.text('Discard stored credentials?'), findsOneWidget);
      expect(find.textContaining('client secret'), findsWidgets);
      await tapDialog(tester, find.text('Cancel'));
      expect(h.server.count('UpdateMCPServer'), 0);

      // Switching to another sign-in discards the registered client.
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('mcp-auth-kind')),
          matching: find.text('None'),
        ),
      );
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pumpAndSettle();
      expect(find.text('Discard stored credentials?'), findsOneWidget);
      expect(find.textContaining('Stop using OAuth 2.0'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(h.server.count('UpdateMCPServer'), 0);
    });

    testWidgets('an ordinary edit, a rotated secret and a new header ask '
        'nothing', (tester) async {
      final secured = oauthServer('a1');
      secured.auth.oauth2
        ..clientId = 'cid'
        ..clientSecret = 'csecret';
      final h = await signedIn(tester, seed: (s) => s.seed('alpha', secured));
      await openDetail(tester, h, 'a1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      await enter(tester, const Key('mcp-name'), 'Renamed');
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('secret-choice-Client secret')),
          matching: find.text('Replace'),
        ),
      );
      await enter(tester, const Key('secret-value-Client secret'), 'rotated');
      await tapScreen(tester, find.byKey(const Key('add-header')));
      await enter(tester, const Key('header-key-0'), 'X-Extra');
      await enter(tester, const Key('header-value-0'), 'extra-value');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('Discard stored credentials?'), findsNothing);
      await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);
      final stored = h.server.servers['alpha']!['a1']!;
      expect(stored.auth.oauth2.clientSecret, 'rotated');
      expect(stored.headers, {'X-Extra': 'extra-value'});
      expect(
        stored.auth.type,
        MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
        reason: 'headers are supplementary to OAuth',
      );
    });
  });

  group('typed secrets do not outlive their submission', () {
    String typed(WidgetTester tester, Key key) => tester
        .widget<TextField>(
          find.descendant(
            of: find.byKey(key),
            matching: find.byType(TextField),
          ),
        )
        .controller!
        .text;

    for (final outcome in ['denied', 'lost reply']) {
      testWidgets('a header value and a client secret, $outcome', (
        tester,
      ) async {
        final secured = oauthServer('a1')..headers['X-Old'] = 'old-value';
        secured.auth.oauth2
          ..clientId = 'cid'
          ..clientSecret = 'csecret';
        final h = await signedIn(tester, seed: (s) => s.seed('alpha', secured));
        await openDetail(tester, h, 'a1');
        await tapScreen(tester, find.byKey(const Key('edit-mcp')));
        await tapScreen(tester, find.byKey(const Key('header-replace-0')));
        await enter(tester, const Key('header-value-0'), 'typed-header');
        await tapScreen(
          tester,
          find.descendant(
            of: find.byKey(const Key('secret-choice-Client secret')),
            matching: find.text('Replace'),
          ),
        );
        await enter(
          tester,
          const Key('secret-value-Client secret'),
          'typed-secret',
        );
        await enter(tester, const Key('mcp-name'), 'Kept name');
        if (outcome == 'denied') {
          h.server.refusals['UpdateMCPServer'] = (
            403,
            'permission_denied',
            'no',
          );
        } else {
          h.server.loseReply = true;
        }
        await tapReal(tester, find.byKey(const Key('save-mcp')));
        await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);

        // Safe fields stay; every secret field is empty again.
        expect(typed(tester, const Key('mcp-name')), 'Kept name');
        expect(typed(tester, const Key('header-value-0')), isEmpty);
        expect(typed(tester, const Key('secret-value-Client secret')), isEmpty);
        expect(find.textContaining('typed-header'), findsNothing);
        expect(find.textContaining('typed-secret'), findsNothing);

        // Nothing is sent again by itself, and sending again needs the values
        // typed again.
        h.server.refusals.clear();
        h.server.loseReply = false;
        await tester.runAsync(
          () => Future<void>.delayed(const Duration(milliseconds: 100)),
        );
        expect(h.server.count('UpdateMCPServer'), 1);
        await tapReal(tester, find.byKey(const Key('save-mcp')));
        await tester.pumpAndSettle();
        expect(
          h.server.count('UpdateMCPServer'),
          1,
          reason: outcome == 'denied'
              ? 'an empty replacement is refused locally'
              : 'saving is off until the outcome is reconciled',
        );
      });
    }

    testWidgets('a header added to a new server is forgotten the same way', (
      tester,
    ) async {
      final h = await signedIn(tester);
      await openList(tester, h);
      h.server.refusals['CreateMCPServer'] = (403, 'permission_denied', 'no');
      await tapScreen(tester, find.byKey(const Key('add-mcp')));
      await enter(tester, const Key('mcp-id'), 'docs');
      await enter(tester, const Key('mcp-name'), 'Docs');
      await enter(tester, const Key('mcp-url'), 'https://mcp.example.com/mcp');
      await tapScreen(tester, find.byKey(const Key('add-header')));
      await enter(tester, const Key('header-key-0'), 'X-Key');
      await enter(tester, const Key('header-value-0'), 'typed-header');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await waitFor(tester, () => h.server.count('CreateMCPServer') == 1);
      expect(typed(tester, const Key('header-key-0')), 'X-Key');
      expect(typed(tester, const Key('header-value-0')), isEmpty);
    });
  });

  group('headers beside any sign-in', () {
    testWidgets('an OAuth server shows its headers and keeps them on a '
        'name-only edit', (tester) async {
      final s0 = oauthServer('a1')..headers['X-Team'] = 'alpha';
      final h = await signedIn(tester, seed: (s) => s.seed('alpha', s0));
      await openDetail(tester, h, 'a1');
      await tapScreen(tester, find.byKey(const Key('edit-mcp')));
      expect(find.text('Value kept (hidden)'), findsOneWidget);
      await enter(tester, const Key('mcp-name'), 'Renamed');
      await tapReal(tester, find.byKey(const Key('save-mcp')));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('Discard stored credentials?'), findsNothing);
      await waitFor(tester, () => h.server.count('UpdateMCPServer') == 1);
      final stored = h.server.servers['alpha']!['a1']!;
      expect(stored.headers, {'X-Team': 'alpha'});
      expect(stored.auth.type, MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2);
    });
  });

  group('OAuth', () {
    testWidgets('connect opens the system browser, then the state read shows '
        'the connection; disconnect asks first', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', oauthServer('a1')),
      );
      await openDetail(tester, h, 'a1');
      expect(find.text('Authorization not read yet'), findsOneWidget);

      await tapReal(tester, find.byKey(const Key('oauth-connect')));
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.waiting,
      );
      expect(h.platform.opened, hasLength(1));
      expect(find.textContaining('Finish signing in'), findsOneWidget);
      expect(find.byKey(const Key('oauth-cancel')), findsOneWidget);

      await tapReal(tester, find.byKey(const Key('oauth-check')));
      await waitFor(tester, () => !h.client.mcp!.oauthOf('a1').reading);
      expect(find.textContaining('Not connected yet'), findsOneWidget);

      h.server.connect(
        'alpha',
        'a1',
        at: DateTime.now().add(const Duration(seconds: 5)),
      );
      await tapReal(tester, find.byKey(const Key('oauth-check')));
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.connected,
      );
      expect(find.textContaining('Connected since'), findsOneWidget);

      await tapReal(tester, find.byKey(const Key('oauth-disconnect')));
      await tester.pumpAndSettle();
      expect(find.text('Disconnect "Server a1"?'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(h.server.count('DisconnectMCPServerOAuth'), 0);
      await tapReal(tester, find.byKey(const Key('oauth-disconnect')));
      await tester.pumpAndSettle();
      await tapDialog(tester, find.text('Disconnect'));
      await waitFor(
        tester,
        () => h.server.count('DisconnectMCPServerOAuth') == 1,
      );
      expect(find.text('Not connected'), findsOneWidget);
    });

    testWidgets('a waiting page shows the connection without Check being '
        'tapped', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', oauthServer('a1')),
        mcpPollInterval: const Duration(milliseconds: 50),
      );
      await openDetail(tester, h, 'a1');
      await tapReal(tester, find.byKey(const Key('oauth-connect')));
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.waiting,
      );
      final writes = h.server.writes;
      h.server.connect(
        'alpha',
        'a1',
        at: DateTime.now().add(const Duration(seconds: 5)),
      );
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.connected,
      );
      await tester.pump();
      expect(find.textContaining('Connected since'), findsOneWidget);
      expect(h.server.writes, writes, reason: 'watching only reads');
      expect(h.server.count('StartMCPServerOAuth'), 1);
    });

    testWidgets('stop waiting says the link stays valid and makes no call', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', oauthServer('a1')),
      );
      await openDetail(tester, h, 'a1');
      await tapReal(tester, find.byKey(const Key('oauth-connect')));
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.waiting,
      );
      await tapScreen(tester, find.byKey(const Key('oauth-cancel')));
      expect(find.textContaining('stays valid'), findsOneWidget);
      expect(h.server.count('DisconnectMCPServerOAuth'), 0);
    });

    testWidgets('a browser that will not open is reported', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', oauthServer('a1')),
      );
      h.platform.openResult = false;
      await openDetail(tester, h, 'a1');
      await tapReal(tester, find.byKey(const Key('oauth-connect')));
      await waitFor(
        tester,
        () => h.client.mcp!.oauthOf('a1').phase == McpOAuthPhase.failed,
      );
      expect(find.textContaining('Could not open the browser'), findsOneWidget);
    });
  });
}
