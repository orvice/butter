import 'package:butter_mobile/features/agents/agent_lifecycle_controller.dart';
import 'package:butter_mobile/features/agents/agent_lifecycle_pages.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('tombstone dialog is distinct from purge wording', (
    tester,
  ) async {
    final lifecycle = AgentLifecycleController(
      newApi: () => throw UnimplementedError(),
      token: 't',
      workspaceId: 'ws',
      canManage: () => true,
      onUnauthenticated: () {},
    );
    addTearDown(lifecycle.dispose);

    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => TextButton(
            onPressed: () =>
                confirmDeleteAgent(context, lifecycle, 'agent-1', 'Helper'),
            child: const Text('open'),
          ),
        ),
      ),
    );
    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    expect(find.text('Tombstone Agent?'), findsOneWidget);
    expect(find.textContaining('not the same as'), findsOneWidget);
    expect(find.text('Tombstone'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
  });

  testWidgets('purge dialog names git content', (tester) async {
    final lifecycle = AgentLifecycleController(
      newApi: () => throw UnimplementedError(),
      token: 't',
      workspaceId: 'ws',
      canManage: () => true,
      onUnauthenticated: () {},
    );
    addTearDown(lifecycle.dispose);

    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => TextButton(
            onPressed: () =>
                confirmPurgeContent(context, lifecycle, 'agent-1', 'Helper'),
            child: const Text('purge'),
          ),
        ),
      ),
    );
    await tester.tap(find.text('purge'));
    await tester.pumpAndSettle();
    expect(find.text('Purge Git content?'), findsOneWidget);
    expect(find.text('Purge content'), findsOneWidget);
  });
}
