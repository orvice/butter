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
import '../a2ui/form_logic.dart';
import '../a2ui/protocol.dart';
import '../a2ui/surface_store.dart';
import 'backoff.dart';
import 'chat_model.dart';
import 'history.dart';
import 'last_run_input.dart';
import 'run_effects.dart';
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
  bool _polling = false;
  bool _skipAttach = false;
  ChatMessage? _sentTurn;
  RestorableTurn? _restorable;
  Map<String, dynamic>? sharedState;
  String? stopError;
  bool _stopping = false;
  final _effects = _ChatRunEffects();
  final A2uiSurfaceStore surfaces = A2uiSurfaceStore();

  ChatPhase phase = ChatPhase.idle;
  ChatNotice? notice;
  String? sendError;
  String? openError;
  bool opening = false;
  ChatMessage? liveReply;
  List<InterruptInfo> interrupts = const [];

  bool get isPolling => _polling;

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
      !_polling &&
      notice?.kind != NoticeKind.running;

  bool get canStop =>
      _threadId != null &&
      agent != null &&
      (phase == ChatPhase.streaming || _polling) &&
      !_stopping;

  /// Input that may be put back in the composer after a failed or stopped run.
  RestorableTurn? get restorableInput => _restorable;

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
    _polling = false;
    _skipAttach = false;
    _sentTurn = null;
    _restorable = null;
    sharedState = null;
    stopError = null;
    _stopping = false;
    _effects.bind();
    surfaces
      ..setSubmitter(null)
      ..clear();
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
      final historyBody = await _agui.history(
        token: token,
        workspace: workspaceId,
        agentId: agent,
        threadId: id,
      );
      if (conv != _conv || _disposed) return;
      final history = parseThreadHistory(historyBody);
      UiSnapshot? snapshot;
      try {
        final uiBody = await _agui.uiSnapshot(
          token: token,
          workspace: workspaceId,
          agentId: agent,
          threadId: id,
        );
        if (conv == _conv && !_disposed) {
          snapshot = parseUiSnapshot(uiBody);
        }
      } catch (_) {
        // History alone still loads; surfaces may catch up on refresh.
      }
      if (conv != _conv || _disposed) return;
      _applyHistory(conv, history, snapshot: snapshot);
      openError = null;
    } catch (failure) {
      if (conv != _conv || _disposed) return;
      if (_expired(failure)) return;
      if (_polling && transientAguiRead(failure)) return;
      openError = describeChatFailure(failure);
    }
    opening = false;
    _emit();
  }

  void _applyHistory(int conv, ThreadHistory history, {UiSnapshot? snapshot}) {
    _messages
      ..clear()
      ..addAll(history.messages);
    interrupts = history.interrupts;
    _restoreSurfaces(history, snapshot);
    _restorable = _restorableFor(history);
    final running = history.running;
    if (running != null && conv == _conv && phase == ChatPhase.idle) {
      notice = null;
      unawaited(_followExpectedRun(conv, running.runId));
      return;
    }
    notice = _noticeFor(history);
    liveReply = null;
  }

  void _restoreSurfaces(ThreadHistory history, UiSnapshot? snapshot) {
    surfaces.setSubmitter(_canSubmitForms ? _submitForm : null);
    if (snapshot == null) return;
    final placed = history.surfaces
        .where((p) => p.messageId != null && p.messageId!.isNotEmpty)
        .map((p) => p.surfaceId);
    surfaces.restore(snapshot, placed);
  }

  bool get _canSubmitForms =>
      _threadId != null &&
      agent != null &&
      phase == ChatPhase.idle &&
      !opening &&
      !_polling &&
      notice?.kind != NoticeKind.running;

  Future<void> _submitForm(
    SurfaceEntry entry,
    Map<String, String> values,
  ) async {
    final form = entry.form;
    final thread = _threadId, agentId = _agentId;
    if (form == null || thread == null || agentId == null) {
      throw FormSubmitError('Submitting is not available right now.');
    }
    final runId = _newId();
    final body = buildFormSubmitRequest(
      threadId: thread,
      runId: runId,
      resume: RunResume(
        interruptId: form.interruptId,
        payload: submissionPayload(entry.id, form, values),
      ),
    );
    final conv = ++_conv;
    sendError = null;
    notice = null;
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
      if (conv != _conv || _disposed) {
        throw FormSubmitError('The submission failed.');
      }
      phase = ChatPhase.idle;
      if (failure is AguiHttpError) {
        throw _formErrorFromHttp(failure);
      }
      if (failure is AguiNotSent) {
        throw FormSubmitError('Could not reach the server. Nothing was sent.');
      }
      throw FormSubmitError(describeChatFailure(failure));
    }
    if (conv != _conv || _disposed) {
      opened.abort();
      return;
    }
    _started = true;
    phase = ChatPhase.streaming;
    _runStream = opened;
    _bindRunEffects();
    final ended = await _streamRun(
      conv,
      opened,
      runId,
      RunFold(_effects),
      fromPost: true,
    );
    if (conv != _conv || _disposed) return;
    if (!ended) await _pollRunEnded(conv, partial: liveReply);
  }

  void _bindRunEffects() {
    surfaces.setSubmitter(_canSubmitForms ? _submitForm : null);
    _effects.bind(
      onState: (state) {
        sharedState = state;
        _emit();
      },
      onA2ui: (value) {
        final event = parseA2uiEventValue(value);
        if (event != null) surfaces.apply(event);
        _emit();
      },
    );
  }

  FormSubmitError _formErrorFromHttp(AguiHttpError failure) {
    final code = failure.code;
    final message =
        failure.message ??
        switch (code) {
          'form_answered' => 'This form was already submitted.',
          'form_stale' =>
            'This form changed on the server. Refresh and try again.',
          'form_invalid' => 'Fix the highlighted fields and submit again.',
          'form_unknown' => 'This form is no longer available.',
          _ => describeChatFailure(failure),
        };
    return FormSubmitError(
      message,
      code: code,
      fieldErrors: failure.fieldErrors,
    );
  }

  /// Submits one Human Input form. Errors stay on the form; drafts are kept.
  Future<void> submitSurfaceForm(String surfaceId) async {
    await surfaces.submitForm(surfaceId, surfaces.draftFor(surfaceId));
  }

  Iterable<SurfaceEntry> surfacesForMessage(String? messageId) {
    if (messageId == null) return const [];
    return surfaces.entries.where(
      (e) =>
          !e.deleted &&
          e.messageId == messageId &&
          (surfaces.isPlaced(e.id) || e.origin == SurfaceOrigin.stream),
    );
  }

  Iterable<SurfaceEntry> get orphanSurfaces =>
      surfaces.entries.where((e) => !e.deleted && !surfaces.isPlaced(e.id));

  RestorableTurn? _restorableFor(ThreadHistory history) {
    final last = history.lastRun;
    if (last != null) {
      return restorableFromLastRun(last, history.messages);
    }
    if (_sentTurn != null && notice?.kind == NoticeKind.unconfirmed) {
      return restorableFromUserTurn(_sentTurn!);
    }
    return null;
  }

  ChatNotice? _noticeFor(ThreadHistory history) {
    if (history.running != null) {
      return const ChatNotice(
        NoticeKind.running,
        'A reply is still being written on the server.',
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
    'The agent is waiting for your answer.',
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
    if (_threadId == null || !_started) return;
    if (phase == ChatPhase.streaming || _polling) {
      await _readHistory(_conv);
      return;
    }
    const uncertain = {
      NoticeKind.connectionLost,
      NoticeKind.unconfirmed,
      NoticeKind.fallback,
      NoticeKind.running,
    };
    if (notice != null && uncertain.contains(notice!.kind)) await refresh();
  }

  /// Puts [restorableInput] back in the composer when the caller applies it.
  void clearRestorable() {
    _restorable = null;
    _emit();
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
    _sentTurn = turn;
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
        unawaited(_reconcileAmbiguous(conv));
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
    _bindRunEffects();
    unawaited(
      _streamRun(conv, opened, runId, RunFold(_effects), fromPost: true).then((
        ended,
      ) async {
        if (!ended &&
            conv == _conv &&
            !_disposed &&
            notice?.kind == NoticeKind.fallback) {
          await _pollRunEnded(conv, partial: liveReply);
        }
      }),
    );
    _emit();
    return true;
  }

  Future<void> _reconcileAmbiguous(int conv) async {
    await _readHistory(conv);
    if (conv != _conv || _disposed) return;
    if (phase != ChatPhase.idle) return;
    if (notice?.kind == NoticeKind.unconfirmed) {
      _restorable = _sentTurn == null
          ? null
          : restorableFromUserTurn(_sentTurn!);
      _emit();
    }
  }

  Future<void> _followExpectedRun(int conv, String runId) async {
    if (_skipAttach) {
      await _pollRunEnded(conv);
      return;
    }
    phase = ChatPhase.streaming;
    liveReply = null;
    _bindRunEffects();
    _emit();
    AguiRun? run;
    try {
      run = await _agui.attach(
        token: token,
        workspace: workspaceId,
        agentId: _agentId!,
        threadId: _threadId!,
      );
    } catch (failure) {
      if (conv != _conv || _disposed) return;
      if (_expired(failure)) return;
      if (transientAguiRead(failure)) {
        await _pollRunEnded(conv);
        return;
      }
      openError = describeChatFailure(failure);
      phase = ChatPhase.idle;
      _emit();
      return;
    }
    if (conv != _conv || _disposed) {
      run?.abort();
      return;
    }
    if (run == null) {
      await _pollRunEnded(conv);
      return;
    }
    _runStream = run;
    final ended = await _streamRun(
      conv,
      run,
      runId,
      RunFold(_effects),
      fromPost: false,
    );
    if (conv != _conv || _disposed) return;
    if (!ended) await _pollRunEnded(conv);
  }

  /// Streams one run's events. Returns true when the run ended in-stream.
  Future<bool> _streamRun(
    int conv,
    AguiRun run,
    String expectedRunId,
    RunFold fold, {
    required bool fromPost,
  }) async {
    final completer = Completer<bool>();
    late final StreamSubscription<SseEvent> sub;
    var finished = false;

    void finish(
      ChatNotice? ending, {
      required bool commit,
      required bool ended,
    }) {
      if (finished || conv != _conv || _disposed) return;
      finished = true;
      unawaited(sub.cancel());
      run.abort();
      if (identical(_sub, sub)) _sub = null;
      if (identical(_runStream, run)) _runStream = null;
      if (commit) {
        final reply = fold.reply;
        if (reply.parts.isNotEmpty) {
          if (fromPost) {
            _messages.add(reply);
          } else {
            liveReply = null;
            _messages.add(reply);
          }
        }
      }
      liveReply = null;
      phase = ChatPhase.idle;
      if (ending != null) {
        notice = ending;
        _restorable = _restorableForOutcome(ending, fold);
      }
      _sentTurn = null;
      if (!completer.isCompleted) completer.complete(ended);
      _emit();
    }

    sub = run.events.listen(
      (event) {
        if (finished || conv != _conv || _disposed) return;
        try {
          fold.addData(event.data);
        } on FormatException {
          if (!fromPost) {
            finish(
              null,
              commit: liveReply != null && liveReply!.parts.isNotEmpty,
              ended: false,
            );
            return;
          }
          finish(
            const ChatNotice(
              NoticeKind.connectionLost,
              'The server sent something this app cannot read, so the live '
              'reply was dropped. The run may still be going; refresh to check.',
            ),
            commit: true,
            ended: false,
          );
          return;
        }
        if (fold.runId != null && fold.runId != expectedRunId) {
          finish(null, commit: false, ended: false);
          return;
        }
        final outcome = fold.outcome;
        if (outcome == null) {
          liveReply = fold.reply;
          _emit();
          return;
        }
        if (outcome.end == RunEnd.fallback &&
            outcome.fallbackReason == 'truncated') {
          _skipAttach = true;
        }
        interrupts = outcome.interrupts;
        final ending = _noticeForOutcome(outcome);
        finish(ending, commit: true, ended: outcome.end != RunEnd.fallback);
      },
      onError: (Object _) =>
          finish(fromPost ? _lost : null, commit: fromPost, ended: false),
      onDone: () {
        if (fold.outcome == null) {
          finish(fromPost ? _lost : null, commit: fromPost, ended: false);
        }
      },
      cancelOnError: true,
    );
    _sub = sub;
    return completer.future;
  }

  RestorableTurn? _restorableForOutcome(ChatNotice? ending, RunFold fold) {
    if (ending == null) return null;
    if (ending.kind != NoticeKind.failed && ending.kind != NoticeKind.stopped) {
      return null;
    }
    return _sentTurn == null ? null : restorableFromUserTurn(_sentTurn!);
  }

  Future<void> _pollRunEnded(int conv, {ChatMessage? partial}) async {
    if (conv != _conv || _disposed) return;
    _polling = true;
    liveReply = partial;
    phase = ChatPhase.idle;
    notice = const ChatNotice(
      NoticeKind.fallback,
      'Waiting for the server to finish this reply…',
    );
    _emit();
    final backoff = Backoff();
    while (conv == _conv && !_disposed) {
      await backoff.wait(live: () => conv == _conv && !_disposed);
      if (conv != _conv || _disposed) break;
      try {
        final body = await _agui.history(
          token: token,
          workspace: workspaceId,
          agentId: _agentId!,
          threadId: _threadId!,
        );
        if (conv != _conv || _disposed) break;
        final history = parseThreadHistory(body);
        if (history.running != null) continue;
        UiSnapshot? snapshot;
        try {
          final uiBody = await _agui.uiSnapshot(
            token: token,
            workspace: workspaceId,
            agentId: _agentId!,
            threadId: _threadId!,
          );
          snapshot = parseUiSnapshot(uiBody);
        } catch (_) {}
        _polling = false;
        liveReply = null;
        _applyHistory(conv, history, snapshot: snapshot);
        break;
      } catch (failure) {
        if (conv != _conv || _disposed) break;
        if (_expired(failure)) break;
        if (!transientAguiRead(failure)) {
          openError = describeChatFailure(failure);
          _polling = false;
          break;
        }
      }
    }
    _polling = false;
    _emit();
  }

  /// Stop asks the server to end the detached run. 202 accepts; 204 means
  /// nothing was running. Never shows "stopped" until the thread read says so.
  Future<void> stopRun() async {
    final id = _threadId, agent = _agentId;
    if (_disposed || id == null || agent == null || !canStop) return;
    stopError = null;
    _stopping = true;
    _emit();
    final conv = _conv;
    try {
      final stopped = await _agui.stop(
        token: token,
        workspace: workspaceId,
        agentId: agent,
        threadId: id,
      );
      if (conv != _conv || _disposed) return;
      _abortStreamOnly();
      if (stopped != null) {
        await _pollRunEnded(conv, partial: liveReply);
      } else {
        await _readHistory(conv);
      }
    } catch (failure) {
      if (conv != _conv || _disposed) return;
      if (_expired(failure)) return;
      stopError = describeChatFailure(failure);
    } finally {
      if (conv == _conv) _stopping = false;
      _emit();
    }
  }

  void _abortStreamOnly() {
    final sub = _sub;
    _sub = null;
    if (sub != null) unawaited(sub.cancel());
    _runStream?.abort();
    _runStream = null;
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

final class _ChatRunEffects extends RunEffects {
  void Function(Map<String, dynamic>? state)? _onState;
  void Function(Object? value)? _onA2ui;

  void bind({
    void Function(Map<String, dynamic>? state)? onState,
    void Function(Object? value)? onA2ui,
  }) {
    _onState = onState;
    _onA2ui = onA2ui;
  }

  @override
  void a2ui(Object? value) => _onA2ui?.call(value);

  @override
  void state(Map<String, dynamic>? next) => _onState?.call(next);
}
