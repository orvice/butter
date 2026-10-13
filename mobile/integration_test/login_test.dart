import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/client_storage.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  const allowHttp = bool.fromEnvironment('BUTTER_ALLOW_HTTP');
  testWidgets(
    'native password login, OS vault restoration and local logout',
    (tester) async {
      final storage = DeviceStorage();
      await storage.clearSession();
      final client = ClientContext(
        storage: storage,
        allowHttp: kDebugMode && allowHttp,
      );
      addTearDown(client.dispose);
      await tester.pumpWidget(ButterApp(context: client));

      Future<void> settled() async {
        for (var i = 0; i < 300 && client.busy; i++) {
          await tester.pump(const Duration(milliseconds: 100));
        }
        expect(client.busy, isFalse);
        expect(client.error, isNull);
        await tester.pumpAndSettle();
      }

      await tester.enterText(find.byKey(const Key('server-address')), server);
      await tester.tap(find.text('Verify and save'));
      await settled();
      await tester.enterText(find.byKey(const Key('username')), 'mobile-test');
      await tester.enterText(
        find.byKey(const Key('password')),
        'contract-test-password',
      );
      await tester.tap(find.byKey(const Key('sign-in')));
      await settled();
      expect(find.text('Choose Workspace'), findsOneWidget);
      await tester.tap(find.text('alpha'));
      await settled();
      expect(find.text('alpha-agent'), findsOneWidget);

      final restored = ClientContext(
        storage: DeviceStorage(),
        allowHttp: kDebugMode && allowHttp,
      );
      addTearDown(restored.dispose);
      await restored.start();
      expect(restored.user!.username, 'mobile-test');
      expect(restored.workspace, isNull);
      await client.logout();
      expect(await DeviceStorage().readSession(), isNull);
      expect(await DeviceStorage().readServer(), server);
    },
    skip: server.isEmpty,
  );
}
