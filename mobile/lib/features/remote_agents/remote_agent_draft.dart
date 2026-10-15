import '../../core/url_validation.dart';
import '../../gen/agents/v1/agent.pb.dart';

/// The protocols this App can write. The Daemon protocol belongs to the Daemon
/// feature: those agents are shown, never edited here.
enum RemoteKind { a2a, opencode }

bool isDaemonAgent(RemoteAgent agent) =>
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON;

/// Whether the App can write this agent: the Daemon protocol and protocols it
/// does not know are left to the surfaces that own them.
bool remoteAgentEditable(RemoteAgent agent) =>
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A ||
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP;

String remoteProtocolLabel(RemoteAgentProtocol protocol) => switch (protocol) {
  RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A => 'A2A',
  RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP => 'OpenCode HTTP',
  RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON => 'Daemon',
  _ => 'Unspecified',
};

/// The editable form of a Remote Agent. It builds the complete message from the
/// stored copy it started from, so fields the form does not edit
/// (`workspace_id`, fields this App does not know yet) are carried over.
class RemoteAgentDraft {
  RemoteAgentDraft.create() : base = null, kind = RemoteKind.a2a {
    _initial = snapshot();
  }

  RemoteAgentDraft.edit(RemoteAgent agent)
    : base = agent.deepCopy(),
      id = agent.id,
      name = agent.name,
      url = agent.url,
      kind =
          agent.protocol ==
              RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP
          ? RemoteKind.opencode
          : RemoteKind.a2a,
      opencodeAgent = agent.opencodeAgent,
      opencodeModel = agent.opencodeModel,
      username = agent.username {
    _initial = snapshot();
  }

  final RemoteAgent? base;
  bool get isCreate => base == null;

  String id = '';
  String name = '';
  String url = '';
  RemoteKind kind;
  String opencodeAgent = '';
  String opencodeModel = '';
  String username = '';
  SecretEdit passwordEdit = SecretEdit.keep;
  String newPassword = '';

  late final String _initial;

  bool get hasStoredPassword => (base?.password ?? '').isNotEmpty;

  String snapshot() => [
    id,
    name,
    url,
    '$kind',
    opencodeAgent,
    opencodeModel,
    username,
    '$passwordEdit',
    newPassword,
  ].join('\u0000');

  bool get dirty => snapshot() != _initial;

  /// What the password will be when sent: kept, replaced or cleared.
  String get effectivePassword {
    if (kind != RemoteKind.opencode) return '';
    return switch (passwordEdit) {
      SecretEdit.keep => base?.password ?? '',
      SecretEdit.replace => newPassword,
      SecretEdit.clear => '',
    };
  }

  bool get sendsCredentialsInClear =>
      kind == RemoteKind.opencode &&
      effectivePassword.isNotEmpty &&
      isCleartextUrl(url);

  String? validateId() {
    if (!isCreate) return null;
    final t = id.trim();
    if (t.isEmpty) return 'Enter an ID.';
    if (t.contains(RegExp(r'\s'))) return 'An ID cannot contain spaces.';
    return null;
  }

  String? validateName() => name.trim().isEmpty ? 'Enter a name.' : null;

  String? validateUrl() => validateHttpUrl(url, required: true, what: 'a URL');

  String? validatePassword() {
    if (kind != RemoteKind.opencode) return null;
    if (passwordEdit == SecretEdit.replace && newPassword.isEmpty) {
      return 'Enter the new password, or choose Keep or Clear.';
    }
    if (username.trim().isNotEmpty && effectivePassword.isEmpty) {
      return 'A username needs a password. Enter one, or clear the username.';
    }
    return null;
  }

  String? firstError() {
    for (final p in [
      validateId(),
      validateName(),
      validateUrl(),
      validatePassword(),
    ]) {
      if (p != null) return p;
    }
    return null;
  }

  /// The complete message to store. Call only after [firstError] is null.
  RemoteAgent build() {
    final a = base?.deepCopy() ?? RemoteAgent();
    if (isCreate) a.id = id.trim();
    a
      ..name = name.trim()
      ..url = url.trim();
    if (kind == RemoteKind.opencode) {
      a
        ..protocol = RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP
        ..opencodeAgent = opencodeAgent.trim()
        ..opencodeModel = opencodeModel.trim()
        ..username = username.trim()
        ..password = effectivePassword;
    } else {
      a
        ..protocol = RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A
        ..clearOpencodeAgent()
        ..clearOpencodeModel()
        ..clearUsername()
        ..clearPassword();
    }
    // Daemon routing is only valid on the Daemon protocol; the server refuses
    // it elsewhere.
    a
      ..clearDaemonRuntimeId()
      ..clearAcpRuntime();
    return a;
  }
}
