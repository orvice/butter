import 'dart:io';

import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/server.dart';

void main() {
  test(
    'authenticated unary requests refuse redirects instead of forwarding a token',
    () async {
      final destination = TestServer();
      await destination.start();
      addTearDown(destination.close);
      final redirector = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      addTearDown(() => redirector.close(force: true));
      redirector.listen((request) async {
        await request.drain<void>();
        request.response.statusCode = 307;
        request.response.headers.set(
          'location',
          '${destination.url}/api/agents.v1.AuthService/Me',
        );
        await request.response.close();
      });
      final api = ButterApi(
        ServerAddress.parse(
          'http://127.0.0.1:${redirector.port}',
          allowHttp: true,
        ),
      );
      addTearDown(api.close);
      await expectLater(
        api.me('private-token'),
        throwsA(isA<ConnectException>()),
      );
      expect(destination.requests, isEmpty);
    },
  );
}
