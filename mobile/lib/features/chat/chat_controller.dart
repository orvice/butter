import 'dart:async';
import 'dart:convert';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/agui_api.dart';
import '../../core/butter_api.dart';
import '../../core/chat_errors.dart';
import '../../core/sse.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import 'chat_model.dart';
import 'history.dart';
import 'run_fold.dart';
import 'run_request.dart';
import 'threads.dart';

enum ThreadsState { idle, loading, ready, failed }

/// Where the open thread's run stands on this device.
enum ChatPhase {
  idle,

  /// The request is being sent; no answer yet.
  sending,

  /// The server accepted the run and its events are arriving.
  streaming,
}

/// Why a notice is shown under the conversation.
enum NoticeKind {
  /// The run failed; the server said why.
  failed,

  /// A person (or the server) stopped the run. Not a failure.
  stopped,

  /// The live stream broke. The run may still be going on the server.
  connectionLost,

  /// The request went out but no answer arrived: the server may or may not
  /// have the message.
  unconfirmed,

  /// The server could not keep following the run for this stream.
  fallback,

  /// The run paused for input.
  waitingInput,

  /// Another run holds the thread right now.
  running,
}

@immutable
final class ChatNotice {
  const ChatNotice(this.kind, this.text);
  final NoticeKind kind;
  final String text;
}

/// Chat for one server, signed-in user and Workspace.
///
/// Its lifetime is that scope: the owner disposes it when the server, account
/// or Workspace changes, which closes its network use and drops everything it
/// holds, so no late response can reach another context. It is never reused.
class ChatController extends ChangeNotifier {
  ChatController({
    required ButterApi api,
    required this.token,
    required this.workspaceId,
    required this.userId,
    required AguiApi agui,
    required String Function() newId,
    required this.onUnauthenticated,
    this.pageSize = 50,
  }) : _api = api,
       _agui = agui,
       _newId = newId;

  final ButterApi _api;
  final AguiApi _agui;
  final String Function() _newId;
  final String token;
  final String workspaceId;
  final String userId;
  final int pageSize;

  /// Called when a request is refused as unauthenticated; the owner signs the
  /// person out. Nothing is retried.
  final void Function() onUnauthenticated;

  bool _disposed = false;
  bool get disposed => _disposed;

  // ---- Agents and the thread list ----

  List<Agent> _agents = const [];
  final _sessions = <SessionInfo>[];
  String _nextPage = '';
  int _epoch = 0;
  String _query = '';

  ThreadsState threadsState = ThreadsState.idle;
  String? threadsError;
  String? renameError;

  bool get hasMoreThreads => _nextPage.isNotEmpty;
  String get query => _query;

  /// The Workspace's agents a person can talk to. The server refuses a run on
  /// an agent that is not active (provisioning, deleting, deleted).
  List<Agent> get agents => List.unmodifiable([
    for (final a in _agents)
      if (a.lifecycleStatus ==
              AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE ||
          a.lifecycleStatus ==
              AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_UNSPECIFIED)
        a,
  ]);

  List<ThreadRow> get threads => threadRows(_sessions, workspaceId, _agents);

  List<ThreadRow> get visibleThreads => searchThreads(threads, _query);

  bool _live(int epoch) => !_disposed && epoch == _epoch;

  void _emit() {
    if (!_disposed) notifyListeners();
  }

  /// True when [failure] was an expired session, which the owner handles.
  bool _expired(Object failure) {
    final expired =
        (failure is ConnectException && failure.code == Code.unauthenticated) ||
        (failure is AguiHttpError && failure.status == 401);
    if (expired) onUnauthenticated();
    return expired;
  }

  /// Reads the agents and the first page of threads again. A newer load
  /// replaces an older one still in flight.
  Future<void> loadThreads() async {
    if (_disposed) return;
    final epoch = ++_epoch;
    threadsState = ThreadsState.loading;
    threadsError = null;
    _emit();
    try {
      final agents = <Agent>[];
      var page = '';
      final seen = <String>{};
      do {
        final reply = await _api.agents(token, workspaceId, pageToken: page);
        if (!_live(epoch)) return;
        agents.addAll(reply.agents);
        page = reply.nextPageToken;
        if (page.isNotEmpty && !seen.add(page)) {
          throw const FormatException('Agent paging did not advance.');
        }
      } while (page.isNotEmpty);
      _agents = List.unmodifiable(agents);
      _sessions.clear();
      _nextPage = '';
      await _readUntilRows(epoch);
      if (!_live(epoch)) return;
      threadsState = ThreadsState.ready;
    } catch (failure) {
      if (!_live(epoch)) return;
      if (_expired(failure)) return;
      threadsState = ThreadsState.failed;
      threadsError = describeChatFailure(failure);
    }
    _emit();
  }

  /// Reads the next page, and carries on while pages hold no thread of this
  /// Workspace (the list also carries sessions that are not chat threads),
  /// within a bound, so the list is never empty with more behind it.
  Future<void> _readUntilRows(int epoch) async {
    var pages = 0;
    final before = threads.length;
    while (pages < 10) {
      pages++;
      await _readPage(epoch);
      if (!_live(epoch)) return;
      if (threads.length > before || !hasMoreThreads) return;
    }
  }

  Future<void> _readPage(int epoch) async {
    final reply = await _api.threadSessions(
      token,
      workspaceId,
      userId: userId,
      pageSize: pageSize,
      pageToken: _nextPage,
    );
    if (!_live(epoch)) return;
    if (reply.nextPageToken.isNotEmpty && reply.nextPageToken == _nextPage) {
      throw const FormatException('Thread paging did not advance.');
    }
    _sessions.addAll(reply.sessions);
    _nextPage = reply.nextPageToken;
    _emit();
  }

  Future<void> moreThreads() async {
    if (_disposed || !hasMoreThreads || threadsState == ThreadsState.loading) {
      return;
    }
    final epoch = ++_epoch;
    threadsState = ThreadsState.loading;
    threadsError = null;
    _emit();
    try {
      await _readUntilRows(epoch);
      if (!_live(epoch)) return;
    } catch (failure) {
      if (!_live(epoch)) return;
      if (_expired(failure)) return;
      threadsError = describeChatFailure(failure);
    }
    threadsState = ThreadsState.ready;
    _emit();
  }

  /// Filters the loaded threads by title. A search must not miss a thread
  /// that is simply on a page not read yet, so it reads the rest first.
  Future<void> search(String query) async {
    if (_disposed) return;
    _query = query;
    _emit();
    if (query.trim().isEmpty || !hasMoreThreads) return;
    final epoch = ++_epoch;
    threadsState = ThreadsState.loading;
    threadsError = null;
    _emit();
    try {
      var pages = 0;
      while (hasMoreThreads && pages < 200) {
        pages++;
        await _readPage(epoch);
        if (!_live(epoch)) return;
      }
    } catch (failure) {
      if (!_live(epoch)) return;
      if (_expired(failure)) return;
      threadsError = describeChatFailure(failure);
    }
    threadsState = ThreadsState.ready;
    _emit();
  }

  /// Renames a thread. The title is checked first; a refused rename leaves the
  /// old title and is not retried.
  Future<bool> renameThread(ThreadRow row, String input) async {
    if (_disposed) return false;
    renameError = null;
    final String title;
    try {
      title = normalizeThreadTitle(input);
    } on ChatInputError catch (error) {
      renameError = error.message;
      _emit();
      return false;
    }
    try {
      final updated = await _api.renameSession(
        token,
        workspaceId,
        appName: row.session.appName,
        userId: row.session.userId,
        sessionId: row.session.sessionId,
        title: title,
      );
      if (_disposed) return false;
      final index = _sessions.indexWhere(
        (s) => s.sessionId == row.session.sessionId,
      );
      if (index >= 0) {
        // Only the title changes; the rest, including the hidden binding that
        // makes this a thread, stays as the list read it.
        _sessions[index] = _sessions[index].deepCopy()
          ..title = updated.title.isEmpty ? title : updated.title;
      }
      _emit();
      return true;
    } catch (failure) {
      if (_disposed) return false;
      if (_expired(failure)) return false;
      renameError = describeChatFailure(failure);
      _emit();
      return false;
    }
  }

  // ---- The open conversation ----

  String? _agentId;
  String? _threadId;
  final _messages = <ChatMessage>[];
  StreamSubscription<SseEvent>? _sub;
  AguiRun? _runStream;
  int _conv = 0;
  // Whether the server has (or may have) this thread: false for a draft.
  bool _started = false;

  ChatPhase phase = ChatPhase.idle;
  ChatNotice? notice;
  String? sendError;
  String? openError;
  bool opening = false;
  ChatMessage? liveReply;
  List<InterruptInfo> interrupts = const [];

  String? get threadId => _threadId;
  String? get agentId => _agentId;
  bool get isOpen => _threadId != null;

  Agent? get agent {
    final id = _agentId;
    if (id == null) return null;
    for (final a in agents) {
      if (a.agentId == id) return a;
    }
    return null;
  }

  List<ChatMessage> get messages => List.unmodifiable(_messages);

  /// A message may be sent when the thread is open on an agent that can run,
  /// nothing of this thread is running or being read, and no other run holds
  /// it.
  bool get canSend =>
      _threadId != null &&
      agent != null &&
      phase == ChatPhase.idle &&
      !opening &&
      notice?.kind != NoticeKind.running;

  void _closeConversation() {
    _conv++;
    final sub = _sub;
    _sub = null;
    if (sub != null) unawaited(sub.cancel());
    // Ends this device's connection now; the run goes on at the server.
    _runStream?.abort();
    _runStream = null;
    _agentId = null;
    _threadId = null;
    _started = false;
    _messages.clear();
    phase = ChatPhase.idle;
    notice = null;
    sendError = null;
    openError = null;
    opening = false;
    liveReply = null;
    interrupts = const [];
  }

  /// Opens a draft thread on [agent]. Nothing is sent, and no thread exists on
  /// the server, until the first message.
  void startThread(Agent agent) {
    if (_disposed) return;
    final wasSending = phase == ChatPhase.sending;
    _closeConversation();
    if (wasSending) _agui.abortAll();
    _agentId = agent.agentId;
    _threadId = _newId();
    _emit();
  }

  /// Leaves the open thread. A run in progress keeps going on the server:
  /// leaving only closes this device's view of it, never stops or repeats it.
  void leave() {
    if (_disposed) return;
    final wasSending = phase == ChatPhase.sending;
    _closeConversation();
    if (wasSending) _agui.abortAll();
    _emit();
  }

  /// Opens an existing thread and reads its history once.
  Future<void> openThread(ThreadRow row) async {
    if (_disposed) return;
    final wasSending = phase == ChatPhase.sending;
    _closeConversation();
    if (wasSending) _agui.abortAll();
    final conv = _conv;
    _agentId = row.agentId;
    _threadId = row.threadId;
    _started = true;
    opening = true;
    _emit();
    await _readHistory(conv);
  }

  Future<void> _readHistory(int conv) async {
    final id = _threadId, agent = _agentId;
    if (id == null || agent == null) return;
    try {
      final body = await _agui.history(
        token: token,
        workspace: workspaceId,
        agentId: agent,
        threadId: id,
      );
      if (conv != _conv || _disposed) return;
      final history = parseThreadHistory(body);
      _messages
        ..clear()
        ..addAll(history.messages);
      interrupts = history.interrupts;
      openError = null;
      notice = _noticeFor(history);
    } catch (failure) {
      if (conv != _conv || _disposed) return;
      if (_expired(failure)) return;
      openError = describeChatFailure(failure);
    }
    opening = false;
    _emit();
  }

  ChatNotice? _noticeFor(ThreadHistory history) {
    if (history.running != null) {
      return const ChatNotice(
        NoticeKind.running,
        'A reply is still being written on the server. Following it live is '
        'not available in this app yet; refresh to see how it stands.',
      );
    }
    final last = history.lastRun;
    if (last != null) {
      if (last.status == 'cancelled') {
        return const ChatNotice(
          NoticeKind.stopped,
          'The last run was stopped.',
        );
      }
      return ChatNotice(
        NoticeKind.failed,
        last.error.isEmpty ? 'The last run failed.' : _bound(last.error),
      );
    }
    if (history.interrupts.isNotEmpty) return _waitingNotice;
    return null;
  }

  static const _waitingNotice = ChatNotice(
    NoticeKind.waitingInput,
    'The agent is waiting for an answer. Forms and choices are not available '
    'in this app yet.',
  );

  static String _bound(String text) => boundText(text, 500);

  /// Reads the open thread again (read-only), for after a lost connection.
  Future<void> refresh() async {
    if (_disposed || _threadId == null || !_started) return;
    if (phase != ChatPhase.idle || opening) return;
    opening = true;
    _emit();
    await _readHistory(_conv);
  }

  /// Foreground recovery: reads again only when what is shown is uncertain.
  Future<void> recover() async {
    const uncertain = {
      NoticeKind.connectionLost,
      NoticeKind.unconfirmed,
      NoticeKind.fallback,
      NoticeKind.running,
    };
    if (notice != null && uncertain.contains(notice!.kind)) await refresh();
  }

  /// The app is going to the background: nothing keeps running here, so open
  /// connections end. A run in progress carries on at the server.
  void suspend() {
    if (_disposed) return;
    _agui.abortAll();
  }

  /// Sends [text] and [images] as one turn. True when the turn left the
  /// composer: it was sent, or may have been (see [NoticeKind.unconfirmed]).
  /// False when nothing was sent; the composer keeps its content and
  /// [sendError] says why.
  Future<bool> send(String text, [List<ChatImage> images = const []]) async {
    final thread = _threadId, agentId = _agentId;
    if (_disposed || thread == null || agentId == null || !canSend) {
      return false;
    }
    sendError = null;
    final Uint8List body;
    final runId = _newId();
    final messageId = _newId();
    try {
      body = buildRunRequest(
        threadId: thread,
        runId: runId,
        messageId: messageId,
        text: text,
        images: images,
      );
    } on ChatInputError catch (error) {
      sendError = error.message;
      _emit();
      return false;
    }
    final conv = ++_conv;
    notice = null;
    interrupts = const [];
    final turn = ChatMessage(
      id: messageId,
      role: ChatRole.user,
      parts: [
        if (text.trim().isNotEmpty) TextPart(text),
        for (final image in images)
          ImagePart(
            mimeType: image.mimeType,
            base64: base64Encode(image.bytes),
          ),
      ],
    );
    _messages.add(turn);
    phase = ChatPhase.sending;
    _emit();

    final AguiRun opened;
    try {
      opened = await _agui.run(
        token: token,
        workspace: workspaceId,
        agentId: agentId,
        body: body,
      );
    } catch (failure) {
      if (conv != _conv || _disposed) return true;
      phase = ChatPhase.idle;
      if (failure is AguiAmbiguous) {
        _started = true;
        notice = const ChatNotice(
          NoticeKind.unconfirmed,
          'The server did not confirm this message. It may or may not have '
          'arrived and the agent may be running. It was not sent again; '
          'refresh to check.',
        );
        _emit();
        return true;
      }
      // Refused or never sent: the run did not start, so the turn goes back
      // to the composer.
      _messages.remove(turn);
      if (!_expired(failure)) {
        sendError = failure is AguiNotSent
            ? 'Could not reach the server. Nothing was sent.'
            : describeChatFailure(failure);
      }
      _emit();
      return false;
    }
    if (conv != _conv || _disposed) {
      opened.abort();
      return true;
    }
    _started = true;
    phase = ChatPhase.streaming;
    _runStream = opened;
    _follow(conv, opened, RunFold());
    _emit();
    return true;
  }

  void _follow(int conv, AguiRun run, RunFold fold) {
    late final StreamSubscription<SseEvent> sub;
    var finished = false;
    void finish(ChatNotice? ending, {required bool commit}) {
      if (finished || conv != _conv || _disposed) return;
      finished = true;
      unawaited(sub.cancel());
      run.abort();
      if (identical(_sub, sub)) _sub = null;
      if (identical(_runStream, run)) _runStream = null;
      if (commit) {
        final reply = fold.reply;
        if (reply.parts.isNotEmpty) _messages.add(reply);
      }
      liveReply = null;
      phase = ChatPhase.idle;
      notice = ending;
      _emit();
    }

    sub = run.events.listen(
      (event) {
        if (finished || conv != _conv || _disposed) return;
        try {
          fold.addData(event.data);
        } on FormatException {
          finish(
            const ChatNotice(
              NoticeKind.connectionLost,
              'The server sent something this app cannot read, so the live '
              'reply was dropped. The run may still be going; refresh to check.',
            ),
            commit: true,
          );
          return;
        }
        final outcome = fold.outcome;
        if (outcome == null) {
          liveReply = fold.reply;
          _emit();
          return;
        }
        interrupts = outcome.interrupts;
        finish(_noticeForOutcome(outcome), commit: true);
      },
      onError: (Object _) => finish(_lost, commit: true),
      onDone: () => finish(fold.outcome == null ? _lost : null, commit: true),
      cancelOnError: true,
    );
    _sub = sub;
  }

  static const _lost = ChatNotice(
    NoticeKind.connectionLost,
    'The connection to the live reply was lost. The run may still be going on '
    'the server; refresh to see how it stands.',
  );

  ChatNotice? _noticeForOutcome(RunOutcome outcome) => switch (outcome.end) {
    RunEnd.success => null,
    RunEnd.interrupt => _waitingNotice,
    RunEnd.error => ChatNotice(
      NoticeKind.failed,
      outcome.message == null || outcome.message!.isEmpty
          ? 'The run failed.'
          : _bound(outcome.message!),
    ),
    RunEnd.stopped => const ChatNotice(
      NoticeKind.stopped,
      'The run was stopped.',
    ),
    RunEnd.fallback => const ChatNotice(
      NoticeKind.fallback,
      'The server could not keep this live stream going. The run may still be '
      'going; refresh to see how it stands.',
    ),
  };

  @override
  void dispose() {
    if (_disposed) return;
    _disposed = true;
    _epoch++;
    _closeConversation();
    _agui.close();
    _sessions.clear();
    _agents = const [];
    _nextPage = '';
    _api.close();
    super.dispose();
  }
}
