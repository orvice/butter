import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import 'chat_controller.dart';
import 'chat_model.dart';
import 'image_chooser.dart';
import 'message_view.dart';
import 'run_request.dart';
import 'threads.dart';

/// Threads of the selected Workspace, with search, rename and a new chat.
class ThreadListPage extends StatefulWidget {
  const ThreadListPage({
    super.key,
    required this.client,
    this.chooser = const ImageChooser(),
  });
  final ClientContext client;
  final ImageChooser chooser;
  @override
  State<ThreadListPage> createState() => _ThreadListPageState();
}

class _ThreadListPageState extends State<ThreadListPage> {
  late final ChatController? _chat = widget.client.chat;
  final _search = TextEditingController();

  @override
  void initState() {
    super.initState();
    unawaited(_chat?.loadThreads());
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Future<void> _open(Widget page) async {
    await openManagementPage(context, widget.client, page);
    // Back on the list: titles and new threads are read again, once.
    if (mounted) unawaited(_chat?.loadThreads());
  }

  Future<void> _newChat() async {
    final chat = _chat!;
    final agent = await showModalBottomSheet<Agent>(
      context: context,
      builder: (context) => _AgentPicker(agents: chat.agents),
    );
    if (agent == null || !mounted) return;
    chat.startThread(agent);
    await _open(ChatPage(client: widget.client, chooser: widget.chooser));
  }

  Future<void> _openThread(ThreadRow row) async {
    final chat = _chat!;
    unawaited(chat.openThread(row));
    await _open(ChatPage(client: widget.client, chooser: widget.chooser));
  }

  Future<void> _rename(ThreadRow row) async {
    final chat = _chat!;
    chat.renameError = null;
    await showDialog<void>(
      context: context,
      builder: (context) => _RenameDialog(chat: chat, row: row),
    );
  }

  @override
  Widget build(BuildContext context) {
    final chat = _chat;
    if (chat == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Chat')),
        body: const Center(child: Text('Sign in and choose a Workspace.')),
      );
    }
    return AnimatedBuilder(
      animation: chat,
      builder: (context, _) {
        final rows = chat.visibleThreads;
        final loading = chat.threadsState == ThreadsState.loading;
        return Scaffold(
          appBar: AppBar(title: const Text('Chat')),
          floatingActionButton: FloatingActionButton.extended(
            key: const Key('new-chat'),
            onPressed: _newChat,
            icon: const Icon(Icons.add),
            label: const Text('New chat'),
          ),
          body: SafeArea(
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
                  child: TextField(
                    key: const Key('thread-search'),
                    controller: _search,
                    textInputAction: TextInputAction.search,
                    decoration: const InputDecoration(
                      prefixIcon: Icon(Icons.search),
                      labelText: 'Search threads',
                    ),
                    onChanged: (value) => unawaited(chat.search(value)),
                  ),
                ),
                if (loading)
                  const LinearProgressIndicator(semanticsLabel: 'Loading'),
                if (chat.threadsError != null)
                  _ErrorRow(
                    text: chat.threadsError!,
                    action: 'Retry',
                    onAction: chat.loadThreads,
                  ),
                Expanded(
                  child: RefreshIndicator(
                    onRefresh: chat.loadThreads,
                    child: ListView(
                      physics: const AlwaysScrollableScrollPhysics(),
                      padding: const EdgeInsets.only(bottom: 88),
                      children: [
                        if (!loading &&
                            chat.threadsState == ThreadsState.ready &&
                            rows.isEmpty)
                          Padding(
                            padding: const EdgeInsets.all(24),
                            child: Text(
                              chat.query.trim().isEmpty
                                  ? 'No threads yet. Start a new chat.'
                                  : 'No threads match.',
                            ),
                          ),
                        for (final row in rows)
                          ListTile(
                            title: Text(
                              row.title,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            subtitle: Text(
                              row.agent == null
                                  ? 'Agent unavailable'
                                  : agentLabel(row.agent!),
                            ),
                            trailing: PopupMenuButton<String>(
                              tooltip: 'Thread options',
                              onSelected: (_) => _rename(row),
                              itemBuilder: (_) => const [
                                PopupMenuItem(
                                  value: 'rename',
                                  child: Text('Rename'),
                                ),
                              ],
                            ),
                            onTap: () => _openThread(row),
                          ),
                        if (chat.hasMoreThreads && chat.query.trim().isEmpty)
                          Padding(
                            padding: const EdgeInsets.all(12),
                            child: OutlinedButton(
                              onPressed: loading ? null : chat.moreThreads,
                              child: const Text('Load more threads'),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _AgentPicker extends StatelessWidget {
  const _AgentPicker({required this.agents});
  final List<Agent> agents;
  @override
  Widget build(BuildContext context) => SafeArea(
    child: ListView(
      shrinkWrap: true,
      children: [
        const ListTile(title: Text('Chat with')),
        if (agents.isEmpty)
          const ListTile(
            title: Text('No agents to chat with in this Workspace.'),
          ),
        for (final agent in agents)
          ListTile(
            title: Text(agentLabel(agent)),
            subtitle: agent.description.isEmpty
                ? null
                : Text(agent.description, maxLines: 2),
            onTap: () => Navigator.pop(context, agent),
          ),
      ],
    ),
  );
}

class _RenameDialog extends StatefulWidget {
  const _RenameDialog({required this.chat, required this.row});
  final ChatController chat;
  final ThreadRow row;
  @override
  State<_RenameDialog> createState() => _RenameDialogState();
}

class _RenameDialogState extends State<_RenameDialog> {
  late final _title = TextEditingController(text: widget.row.session.title);
  bool _saving = false;

  @override
  void dispose() {
    _title.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    final ok = await widget.chat.renameThread(widget.row, _title.text);
    if (!mounted) return;
    if (ok) {
      Navigator.pop(context);
    } else {
      setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Rename thread'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        TextField(
          key: const Key('rename-field'),
          controller: _title,
          autofocus: true,
          enabled: !_saving,
          decoration: const InputDecoration(labelText: 'Title'),
          onSubmitted: (_) => _saving ? null : _save(),
        ),
        if (widget.chat.renameError != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Semantics(
              liveRegion: true,
              child: Text(
                widget.chat.renameError!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ),
          ),
      ],
    ),
    actions: [
      TextButton(
        onPressed: _saving ? null : () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        onPressed: _saving ? null : _save,
        child: const Text('Save'),
      ),
    ],
  );
}

class _ErrorRow extends StatelessWidget {
  const _ErrorRow({required this.text, required this.action, this.onAction});
  final String text;
  final String action;
  final Future<void> Function()? onAction;
  @override
  Widget build(BuildContext context) => Semantics(
    liveRegion: true,
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Row(
        children: [
          Expanded(
            child: Text(
              text,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
          if (onAction != null)
            TextButton(onPressed: onAction, child: Text(action)),
        ],
      ),
    ),
  );
}

/// The open conversation. The thread is opened (or drafted) on the controller
/// before this page is pushed; leaving the page only detaches from a run.
class ChatPage extends StatefulWidget {
  const ChatPage({
    super.key,
    required this.client,
    this.chooser = const ImageChooser(),
  });
  final ClientContext client;
  final ImageChooser chooser;
  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  late final ChatController? _chat = widget.client.chat;
  // The thread this page was opened on; only that one is left with the page.
  late final String? _thread = _chat?.threadId;
  final _text = TextEditingController();
  final _tray = ImageTray();
  String? _attachProblem;

  @override
  void initState() {
    super.initState();
    _text.addListener(() => setState(() {}));
  }

  @override
  void dispose() {
    _text.dispose();
    _leave();
    super.dispose();
  }

  /// Leaving only closes this device's view of a run. Never during the unmount
  /// itself (listeners rebuild), and only for the thread this page showed.
  void _leave() {
    final chat = _chat;
    if (chat == null || chat.threadId != _thread) return;
    scheduleMicrotask(() {
      if (chat.threadId == _thread) chat.leave();
    });
  }

  bool get _hasTurn => _text.text.trim().isNotEmpty || _tray.images.isNotEmpty;

  Future<void> _pick() async {
    final result = await widget.chooser.choose(
      maxImagesPerMessage - _tray.images.length,
    );
    if (!mounted) return;
    var problem = result.problem;
    setState(() {
      for (final image in result.images) {
        try {
          _tray.add(image);
        } on ChatInputError catch (error) {
          problem ??= error.message;
        }
      }
      _attachProblem = problem;
    });
  }

  Future<void> _send() async {
    final chat = _chat!;
    final text = _text.text;
    final images = List.of(_tray.images);
    setState(() {
      _text.clear();
      _tray.clear();
      _attachProblem = null;
    });
    final sent = await chat.send(text, images);
    if (sent || !mounted) return;
    // Nothing was sent: the turn goes back to the composer.
    setState(() {
      if (_text.text.isEmpty) _text.text = text;
      for (final image in images) {
        try {
          _tray.add(image);
        } on ChatInputError {
          // Another turn was started meanwhile; keep what fits.
        }
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final chat = _chat;
    if (chat == null || chat.disposed) {
      return Scaffold(
        appBar: AppBar(title: const Text('Chat')),
        body: const Center(child: Text('Sign in and choose a Workspace.')),
      );
    }
    return AnimatedBuilder(
      animation: chat,
      builder: (context, _) {
        final agent = chat.agent;
        final name = agent == null ? 'Agent' : agentLabel(agent);
        final live = chat.liveReply;
        final items = <ChatMessage>[?live, ...chat.messages.reversed];
        return PopScope(
          onPopInvokedWithResult: (didPop, _) {
            if (didPop && chat.threadId == _thread) chat.leave();
          },
          child: Scaffold(
            appBar: AppBar(title: Text(agent == null ? 'Chat' : name)),
            body: SafeArea(
              child: Column(
                children: [
                  if (chat.opening || chat.phase != ChatPhase.idle)
                    const LinearProgressIndicator(semanticsLabel: 'Working'),
                  if (chat.openError != null)
                    _ErrorRow(
                      text: chat.openError!,
                      action: 'Retry',
                      onAction: chat.refresh,
                    ),
                  Expanded(
                    child: items.isEmpty && !chat.opening
                        ? const Center(child: Text('Send a message to begin.'))
                        : SelectionArea(
                            child: ListView.builder(
                              reverse: true,
                              padding: const EdgeInsets.all(12),
                              itemCount: items.length,
                              itemBuilder: (context, index) => Padding(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 4,
                                ),
                                child: MessageView(
                                  message: items[index],
                                  agentName: name,
                                ),
                              ),
                            ),
                          ),
                  ),
                  if (chat.notice != null) _NoticeBanner(chat: chat),
                  if (agent == null && chat.threadId != null && !chat.opening)
                    const Padding(
                      padding: EdgeInsets.all(12),
                      child: Text(
                        'This thread\'s agent is not available, so it cannot be continued.',
                      ),
                    ),
                  _composer(context, chat),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _composer(BuildContext context, ChatController chat) {
    final working = chat.phase != ChatPhase.idle;
    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 4, 8, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (chat.sendError != null)
            _ErrorRow(text: chat.sendError!, action: ''),
          if (_attachProblem != null)
            _ErrorRow(text: _attachProblem!, action: ''),
          if (_tray.images.isNotEmpty)
            SizedBox(
              height: 72,
              child: ListView(
                scrollDirection: Axis.horizontal,
                children: [
                  for (final (i, image) in _tray.images.indexed)
                    Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: Stack(
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(8),
                            child: Image.memory(
                              image.bytes,
                              width: 64,
                              height: 64,
                              fit: BoxFit.cover,
                              semanticLabel: 'Image ${i + 1}: ${image.name}',
                            ),
                          ),
                          Positioned(
                            top: -8,
                            right: -8,
                            child: IconButton(
                              tooltip: 'Remove image ${i + 1}',
                              visualDensity: VisualDensity.compact,
                              iconSize: 18,
                              icon: const Icon(Icons.cancel),
                              onPressed: () =>
                                  setState(() => _tray.removeAt(i)),
                            ),
                          ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              IconButton(
                tooltip: 'Attach image',
                icon: const Icon(Icons.image_outlined),
                onPressed: chat.canSend ? _pick : null,
              ),
              Expanded(
                child: TextField(
                  key: const Key('composer'),
                  controller: _text,
                  enabled: chat.agent != null,
                  minLines: 1,
                  maxLines: 6,
                  keyboardType: TextInputType.multiline,
                  textInputAction: TextInputAction.newline,
                  decoration: const InputDecoration(
                    hintText: 'Message',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
              ),
              if (working)
                const IconButton(
                  tooltip: 'Stop is not available in this app yet',
                  icon: Icon(Icons.stop_circle_outlined),
                  onPressed: null,
                )
              else
                IconButton(
                  key: const Key('send'),
                  tooltip: 'Send',
                  icon: const Icon(Icons.send),
                  onPressed: chat.canSend && _hasTurn ? _send : null,
                ),
            ],
          ),
        ],
      ),
    );
  }
}

class _NoticeBanner extends StatelessWidget {
  const _NoticeBanner({required this.chat});
  final ChatController chat;

  @override
  Widget build(BuildContext context) {
    final notice = chat.notice!;
    final scheme = Theme.of(context).colorScheme;
    final failure = notice.kind == NoticeKind.failed;
    final canRefresh = switch (notice.kind) {
      NoticeKind.connectionLost ||
      NoticeKind.unconfirmed ||
      NoticeKind.fallback ||
      NoticeKind.running => true,
      _ => false,
    };
    return Semantics(
      liveRegion: true,
      child: Material(
        color: failure ? scheme.errorContainer : scheme.secondaryContainer,
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(notice.text),
              for (final interrupt in chat.interrupts)
                if (interrupt.message.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text('“${interrupt.message}”'),
                  ),
              if (canRefresh) ...[
                const SizedBox(height: 4),
                Text(
                  'Reconnecting to a live run is not available in this app yet.',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: chat.opening ? null : chat.refresh,
                    child: const Text('Refresh'),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
