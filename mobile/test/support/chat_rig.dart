import 'package:butter_mobile/app/client_context.dart';
import 'package:flutter_test/flutter_test.dart';

import 'chat_server.dart';
import 'memory_storage.dart';
import 'server.dart';

typedef Rig = ({ClientContext client, TestServer server, ChatServer chat});

/// A signed-in client with [workspace] selected, against the shared test
/// server with the AG-UI and session endpoints on top.
Future<Rig> rig({
  int pageSize = 50,
  String workspace = 'alpha',
  Duration idleTimeout = const Duration(seconds: 60),
}) async {
  final server = TestServer();
  await server.start();
  addTearDown(server.close);
  final chat = ChatServer(server);
  var next = 0;
  final client = ClientContext(
    storage: MemoryStorage(),
    allowHttp: true,
    chatPageSize: pageSize,
    chatIdleTimeout: idleTimeout,
    newChatId: () => 'id-${++next}',
  );
  addTearDown(client.dispose);
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace(workspace);
  return (client: client, server: server, chat: chat);
}
