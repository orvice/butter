import 'dart:async';

import 'package:butter_mobile/core/native_oauth.dart';

/// The OS boundary of native sign-in: the system browser and incoming links.
class FakeOAuthPlatform implements NativeOAuthPlatform {
  final opened = <Uri>[];
  bool openResult = true;

  /// When set, the browser launcher answers only once this completes (a slow
  /// or blocked launcher); the URL counts as handed over at once.
  Completer<bool>? openGate;

  /// Makes the launcher throw instead of answering.
  bool openThrows = false;
  final _links = StreamController<Uri>.broadcast(sync: true);

  @override
  Future<bool> openBrowser(Uri url) async {
    opened.add(url);
    if (openThrows) throw StateError('launcher failed');
    final gate = openGate;
    if (gate != null) return gate.future;
    return openResult;
  }

  @override
  Stream<Uri> get links => _links.stream;

  /// The OS delivers the App's return link.
  void deliver(String link) => _links.add(Uri.parse(link));
}
