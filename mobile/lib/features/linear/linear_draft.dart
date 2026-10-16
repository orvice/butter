import 'package:fixnum/fixnum.dart';

import '../../gen/agents/v1/linear.pb.dart';
import '../agents/leaf_draft.dart';

/// Write-only on create; never read back from the server.
enum LinearSecretEdit { unset, keep, replace, clear }

class LinearAppDraft {
  LinearAppDraft._(this._app);

  factory LinearAppDraft.create() => LinearAppDraft._(LinearApp());

  factory LinearAppDraft.fromStored(LinearApp stored) =>
      LinearAppDraft._(stored.deepCopy());

  final LinearApp _app;

  LinearApp get stored => _app;

  String get id => _app.id;
  set id(String v) => _app.id = v;

  String get displayName => _app.displayName;
  set displayName(String v) => _app.displayName = v;

  String get clientId => _app.clientId;
  set clientId(String v) => _app.clientId = v;

  String get agentId => _app.agentId;
  set agentId(String v) => _app.agentId = v;

  bool get inboundEnabled => _app.inboundEnabled;
  set inboundEnabled(bool v) => _app.inboundEnabled = v;

  Int64 get revision => _app.revision;
  set revision(Int64 v) => _app.revision = v;

  List<String> get allowedUserIds => List.unmodifiable(_app.allowedUserIds);
  set allowedUserIds(List<String> ids) {
    _app.allowedUserIds
      ..clear()
      ..addAll(ids);
  }

  RunLimitMode get runLimitMode => runLimitModeOf(
    present: _app.hasMaxRunSeconds(),
    seconds: _app.maxRunSeconds,
  );

  set runLimitMode(RunLimitMode mode) {
    applyRunLimit(
      mode,
      _app.maxRunSeconds,
      clear: _app.clearMaxRunSeconds,
      set: (v) => _app.maxRunSeconds = v,
    );
  }

  void setRunLimitSeconds(int seconds) {
    _app.maxRunSeconds = seconds;
  }

  String clientSecret = '';
  String webhookSecret = '';
  LinearSecretEdit clientSecretEdit = LinearSecretEdit.unset;
  LinearSecretEdit webhookSecretEdit = LinearSecretEdit.unset;

  LinearApp buildForUpdate() => _app.deepCopy();

  LinearApp buildForCreate() {
    final app = _app.deepCopy()
      ..clearId()
      ..clearRevision()
      ..clearCredentialState()
      ..clearClientSecretSet()
      ..clearWebhookSecretSet()
      ..clearCallbackUrl()
      ..clearWebhookUrl()
      ..clearCreatedAt()
      ..clearUpdatedAt()
      ..clearWorkspaceId();
    return app;
  }

  static List<String> parseAllowedUserIds(String text) {
    return text
        .split(RegExp(r'[\s,]+'))
        .map((s) => s.trim())
        .where((s) => s.isNotEmpty)
        .toList();
  }

  static String formatAllowedUserIds(Iterable<String> ids) => ids.join('\n');

  String? validate({required bool creating}) {
    if (displayName.trim().isEmpty) return 'Display name is required.';
    if (creating && clientId.trim().isEmpty) return 'Client ID is required.';
    if (agentId.isEmpty) return 'Choose the Agent this App routes to.';
    if (creating) {
      if (clientSecret.trim().isEmpty) {
        return 'Enter the OAuth client secret when registering the App.';
      }
      if (webhookSecret.trim().isEmpty) {
        return 'Enter the webhook signing secret when registering the App.';
      }
    }
    return null;
  }
}
