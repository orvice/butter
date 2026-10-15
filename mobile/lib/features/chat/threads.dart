import 'dart:convert';

import 'package:flutter/foundation.dart';

import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import 'run_request.dart';

// An AG-UI thread is the session `agui-{threadId}` under the ADK app `agui`.
// Which Workspace and agent it belongs to is recorded once, when the server
// creates the session, in the hidden A2UI binding (ADR-0014): the only place
// a client can read it from, and the same rule the web Chat applies.

const aguiAppName = 'agui';
const _sessionPrefix = 'agui-';
const _bindingKey = 'butter:a2ui:binding';
const maxThreadTitleCodePoints = 100;

String sessionIdOf(String threadId) => '$_sessionPrefix$threadId';

String? _threadIdOf(SessionInfo session) =>
    session.sessionId.startsWith(_sessionPrefix)
    ? session.sessionId.substring(_sessionPrefix.length)
    : null;

/// The agent a session holds [threadId] with in [workspaceId]: the one its
/// binding names, only when that binding is for this Workspace and this
/// thread. Anything else is null (another app, an unbound thread from before
/// bindings existed, another Workspace or thread, a broken binding).
({String agentId})? _binding(SessionInfo session, String workspaceId) {
  if (session.appName != aguiAppName) return null;
  final threadId = _threadIdOf(session);
  if (threadId == null || threadId.isEmpty) return null;
  final value = session.state.fields[_bindingKey];
  if (value == null) return null;
  Object? raw;
  if (value.hasStringValue()) {
    try {
      raw = jsonDecode(value.stringValue);
    } on FormatException {
      return null;
    }
  } else if (value.hasStructValue()) {
    raw = {
      for (final e in value.structValue.fields.entries)
        e.key: e.value.hasStringValue() ? e.value.stringValue : null,
    };
  }
  if (raw is! Map) return null;
  final workspace = raw['workspace_id'];
  final agent = raw['agent_id'];
  final thread = raw['thread_id'];
  if (workspace is! String ||
      agent is! String ||
      thread is! String ||
      workspace != workspaceId ||
      thread != threadId ||
      agent.isEmpty) {
    return null;
  }
  return (agentId: agent);
}

/// One thread as the list shows it.
@immutable
final class ThreadRow {
  const ThreadRow({
    required this.session,
    required this.threadId,
    required this.agentId,
    required this.title,
    this.agent,
  });
  final SessionInfo session;
  final String threadId;
  final String agentId;

  /// The agent as the Workspace's list has it; null when the list lacks it.
  final Agent? agent;
  final String title;

  /// A copy with the session replaced (a rename's answer) and the title it
  /// shows recomputed.
  ThreadRow withSession(SessionInfo next) => ThreadRow(
    session: next,
    threadId: threadId,
    agentId: agentId,
    agent: agent,
    title: threadTitle(next, agent),
  );
}

String agentLabel(Agent agent) =>
    agent.displayName.trim().isNotEmpty ? agent.displayName.trim() : agent.name;

/// The title a thread shows: its own, else its agent's name, else a
/// placeholder.
String threadTitle(SessionInfo session, Agent? agent) {
  final own = session.title.trim();
  if (own.isNotEmpty) return own;
  final name = agent == null ? '' : agentLabel(agent).trim();
  return name.isNotEmpty ? name : 'Untitled thread';
}

/// The caller's threads in one Workspace, in the order given (ListSessions
/// answers newest first), each with the agent its binding names.
List<ThreadRow> threadRows(
  Iterable<SessionInfo> sessions,
  String workspaceId,
  Iterable<Agent> agents,
) {
  final byId = {
    for (final agent in agents)
      if (agent.agentId.isNotEmpty) agent.agentId: agent,
  };
  return [
    for (final session in sessions)
      if (_binding(session, workspaceId) case final binding?)
        ThreadRow(
          session: session,
          threadId: _threadIdOf(session)!,
          agentId: binding.agentId,
          agent: byId[binding.agentId],
          title: threadTitle(session, byId[binding.agentId]),
        ),
  ];
}

/// The rows whose title contains [query], ignoring case and surrounding
/// space.
List<ThreadRow> searchThreads(List<ThreadRow> rows, String query) {
  final needle = query.trim().toLowerCase();
  if (needle.isEmpty) return rows;
  return [
    for (final row in rows)
      if (row.title.toLowerCase().contains(needle)) row,
  ];
}

/// A title as the server accepts it: trimmed, on one line, not blank, at most
/// 100 Unicode code points.
String normalizeThreadTitle(String input) {
  final title = input.trim().split(RegExp(r'\s+')).join(' ');
  if (title.isEmpty) throw const ChatInputError('Enter a title.');
  if (title.runes.length > maxThreadTitleCodePoints) {
    throw const ChatInputError('A title can be at most 100 characters.');
  }
  return title;
}
