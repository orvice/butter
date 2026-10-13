import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';

import 'app/butter_app.dart';
import 'app/client_context.dart';
import 'core/client_storage.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  final context = ClientContext(
    storage: DeviceStorage(),
    allowHttp: kDebugMode && const bool.fromEnvironment('BUTTER_ALLOW_HTTP'),
  );
  runApp(ButterApp(context: context));
  unawaited(context.start());
}
