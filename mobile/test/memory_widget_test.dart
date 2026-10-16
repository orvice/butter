import 'package:butter_mobile/features/memory/memory_pages.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';

void main() {
  testWidgets('memory settings shows connection form', (tester) async {
    final rig = await signedIn(
      tester,
      mcpPollInterval: const Duration(days: 1),
      daemonPollInterval: const Duration(days: 1),
    );
    await tester.pumpWidget(
      MaterialApp(home: MemorySettingsPage(client: rig.client)),
    );
    await tester.pump();
    expect(find.text('Memory'), findsWidgets);
    expect(find.byKey(const Key('memory-base-url')), findsOneWidget);
  });
}
