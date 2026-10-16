import '../../core/url_validation.dart';
import '../../gen/agents/v1/agent.pb.dart';

/// The protocols this App can write.
enum RemoteKind { a2a, opencode, daemon }

/// The executors a Daemon agent can launch on its worker; the server refuses
/// any other.
const daemonExecutors = ['opencode', 'codex'];

bool isDaemonAgent(RemoteAgent agent) =>
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON;

/// Whether the App can write this agent: protocols it does not know are left
/// alone, so a save can never rewrite them as another protocol.
bool remoteAgentEditable(RemoteAgent agent) =>
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A ||
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP ||
    agent.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON;

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
      kind = switch (agent.protocol) {
        RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP =>
          RemoteKind.opencode,
        RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON => RemoteKind.daemon,
        _ => RemoteKind.a2a,
      },
      daemonRuntimeId = agent.daemonRuntimeId,
      acpRuntime = agent.acpRuntime,
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
  String daemonRuntimeId = '';
  String acpRuntime = '';
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
    daemonRuntimeId,
    acpRuntime,
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

  /// A Daemon agent has no URL: the worker is reached through its runtime.
  String? validateUrl() => kind == RemoteKind.daemon
      ? null
      : validateHttpUrl(url, required: true, what: 'a URL');

  String? validateRuntime() {
    if (kind != RemoteKind.daemon) return null;
    return daemonRuntimeId.trim().isEmpty
        ? 'Choose a daemon runtime of this Workspace.'
        : null;
  }

  String? validateAcpRuntime() {
    if (kind != RemoteKind.daemon) return null;
    final value = acpRuntime.trim();
    if (value.isEmpty) return 'Choose an executor.';
    if (!daemonExecutors.contains(value)) {
      return 'The server runs only opencode or codex. Choose one of them.';
    }
    return null;
  }

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
      validateRuntime(),
      validateAcpRuntime(),
      validatePassword(),
    ]) {
      if (p != null) return p;
    }
    return null;
  }

  /// What saving would discard from what is stored, as sentences that name
  /// their target. Empty for harmless edits.
  List<String> discards() {
    final b = base;
    if (b == null) return const [];
    final hadPassword = b.password.isNotEmpty;
    final hadCredentials = hadPassword || b.username.isNotEmpty;
    final wasOpenCode =
        b.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP;
    final wasDaemon =
        b.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON;
    final wasA2a = b.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A;
    final opencodeParts = [
      if (b.opencodeAgent.isNotEmpty) 'agent "${b.opencodeAgent}"',
      if (b.opencodeModel.isNotEmpty) 'model "${b.opencodeModel}"',
      if (hadCredentials) hadPassword ? 'username and password' : 'username',
    ];
    final binding =
        'the Daemon binding (runtime "${b.daemonRuntimeId}", executor "${b.acpRuntime}")';
    switch (kind) {
      case RemoteKind.daemon:
        return [
          if (wasOpenCode)
            'Switch to Daemon: the OpenCode URL "${b.url}"${opencodeParts.isEmpty ? '' : ', ${opencodeParts.join(', ')}'} ${opencodeParts.isEmpty ? 'is' : 'are'} discarded.',
          if (wasA2a && b.url.isNotEmpty)
            'Switch to Daemon: the agent URL "${b.url}" is discarded.',
        ];
      case RemoteKind.a2a:
        return [
          if (wasOpenCode && hadCredentials)
            'Switch to A2A: the OpenCode ${hadPassword ? 'username and password are' : 'username is'} discarded.',
          if (wasDaemon && b.daemonRuntimeId.isNotEmpty)
            'Switch to A2A: $binding is discarded.',
        ];
      case RemoteKind.opencode:
        return [
          if (wasOpenCode && passwordEdit == SecretEdit.clear && hadPassword)
            'Clear the stored password.',
          if (wasDaemon && b.daemonRuntimeId.isNotEmpty)
            'Switch to OpenCode HTTP: $binding is discarded.',
        ];
    }
  }

  /// Forgets the typed password once it has been submitted. A replacement stays
  /// chosen but empty, so sending another one needs it typed again.
  void clearTypedSecrets() => newPassword = '';

  /// The complete message to store. Call only after [firstError] is null.
  RemoteAgent build() {
    final a = base?.deepCopy() ?? RemoteAgent();
    if (isCreate) a.id = id.trim();
    a.name = name.trim();
    final was = base?.protocol;
    switch (kind) {
      case RemoteKind.opencode:
        a
          ..url = url.trim()
          ..protocol = RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP
          ..opencodeAgent = opencodeAgent.trim()
          ..opencodeModel = opencodeModel.trim()
          ..username = username.trim()
          ..password = effectivePassword
          ..clearDaemonRuntimeId()
          ..clearAcpRuntime();
      case RemoteKind.a2a:
        a
          ..url = url.trim()
          ..protocol = RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A
          ..clearOpencodeAgent()
          ..clearOpencodeModel()
          ..clearUsername()
          ..clearPassword()
          ..clearDaemonRuntimeId()
          ..clearAcpRuntime();
      case RemoteKind.daemon:
        a
          ..protocol = RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON
          ..daemonRuntimeId = daemonRuntimeId.trim()
          ..acpRuntime = acpRuntime.trim();
        // A Daemon agent that was already one keeps whatever else it holds;
        // coming from another protocol, that protocol's settings go (the
        // discards were confirmed).
        if (was != RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON) {
          a
            ..clearUrl()
            ..clearOpencodeAgent()
            ..clearOpencodeModel()
            ..clearUsername()
            ..clearPassword();
        }
    }
    return a;
  }
}
