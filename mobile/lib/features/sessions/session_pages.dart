import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import 'session_controller.dart';
import 'session_events.dart';
import 'session_key.dart';

Future<void> openSessions(BuildContext context, ClientContext client) async {
  final ctl = client.sessions;
  if (ctl == null) return;
  ctl.draftFilters = ctl.appliedFilters;
  unawaited(ctl.loadPage());
  await openManagementPage(context, client, SessionsPage(client: client));
}

class SessionsPage extends StatefulWidget {
  const SessionsPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<SessionsPage> createState() => _SessionsPageState();
}

class _SessionsPageState extends State<SessionsPage> {
  SessionController? get _ctl => widget.client.sessions;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final ctl = _ctl;
      if (ctl == null) return;
      ctl.draftFilters = ctl.appliedFilters;
      unawaited(ctl.loadPage());
    });
  }

  @override
  Widget build(BuildContext context) {
    final ctl = _ctl;
    if (ctl == null) return const SizedBox.shrink();
    return ManagementPage(
      client: widget.client,
      title: 'Sessions',
      child: ListenableBuilder(
        listenable: ctl,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              'Session explorer',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'Browse ADK sessions you may access in this Workspace. '
              'Replies use ReplySession and do not open Chat threads for '
              'cron, Linear or other entry points.',
            ),
            const SizedBox(height: 16),
            _FilterCard(ctl: ctl),
            const SizedBox(height: 16),
            if (ctl.listState == SessionLoad.loading && ctl.sessions.isEmpty)
              const Center(child: CircularProgressIndicator()),
            if (ctl.listError != null)
              Semantics(
                liveRegion: true,
                child: Text(
                  ctl.listError!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (ctl.listState == SessionLoad.ready && ctl.sessions.isEmpty)
              const Text('No sessions match the current filters.'),
            for (final info in ctl.sessions)
              _SessionRow(info: info, onTap: () => _openDetail(info)),
            const SizedBox(height: 12),
            _PaginationBar(ctl: ctl),
          ],
        ),
      ),
    );
  }

  Future<void> _openDetail(SessionInfo info) async {
    final ctl = _ctl;
    if (ctl == null) return;
    await Navigator.push<void>(
      context,
      MaterialPageRoute(
        builder: (_) => SessionDetailPage(
          client: widget.client,
          sessionKey: SessionKey.fromInfo(info),
        ),
      ),
    );
    if (mounted) unawaited(ctl.loadPage());
  }
}

class _FilterCard extends StatefulWidget {
  const _FilterCard({required this.ctl});
  final SessionController ctl;

  @override
  State<_FilterCard> createState() => _FilterCardState();
}

class _FilterCardState extends State<_FilterCard> {
  late final _app = TextEditingController(
    text: widget.ctl.draftFilters.appName,
  );
  late final _user = TextEditingController(
    text: widget.ctl.draftFilters.userId,
  );

  @override
  void dispose() {
    _app.dispose();
    _user.dispose();
    super.dispose();
  }

  Future<void> _pickStart() async {
    final date = await showDatePicker(
      context: context,
      firstDate: DateTime(2020),
      lastDate: DateTime.now().add(const Duration(days: 1)),
      initialDate: widget.ctl.draftFilters.startTime ?? DateTime.now(),
    );
    if (date == null || !mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(
        widget.ctl.draftFilters.startTime ?? DateTime.now(),
      ),
    );
    if (time == null || !mounted) return;
    setState(() {
      widget.ctl.draftFilters = widget.ctl.draftFilters.copyWith(
        startTime: DateTime(
          date.year,
          date.month,
          date.day,
          time.hour,
          time.minute,
        ),
      );
    });
  }

  Future<void> _pickEnd() async {
    final date = await showDatePicker(
      context: context,
      firstDate: DateTime(2020),
      lastDate: DateTime.now().add(const Duration(days: 1)),
      initialDate: widget.ctl.draftFilters.endTime ?? DateTime.now(),
    );
    if (date == null || !mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(
        widget.ctl.draftFilters.endTime ?? DateTime.now(),
      ),
    );
    if (time == null || !mounted) return;
    setState(() {
      widget.ctl.draftFilters = widget.ctl.draftFilters.copyWith(
        endTime: DateTime(
          date.year,
          date.month,
          date.day,
          time.hour,
          time.minute,
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          TextField(
            controller: _app,
            decoration: const InputDecoration(
              labelText: 'Channel (app name)',
              hintText: 'Empty = all channels',
            ),
            onChanged: (v) => widget.ctl.draftFilters = widget.ctl.draftFilters
                .copyWith(appName: v.trim()),
          ),
          TextField(
            controller: _user,
            decoration: const InputDecoration(
              labelText: 'User ID',
              helperText:
                  'Workspace owners and admins may filter another user.',
            ),
            onChanged: (v) => widget.ctl.draftFilters = widget.ctl.draftFilters
                .copyWith(userId: v.trim()),
          ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: const Text('From'),
            subtitle: Text(
              widget.ctl.draftFilters.startTime
                      ?.toLocal()
                      .toString()
                      .split('.')
                      .first ??
                  'Any time',
            ),
            trailing: IconButton(
              tooltip: 'Clear start time',
              onPressed: widget.ctl.draftFilters.startTime == null
                  ? null
                  : () => setState(
                      () => widget.ctl.draftFilters = widget.ctl.draftFilters
                          .copyWith(clearStart: true),
                    ),
              icon: const Icon(Icons.clear),
            ),
            onTap: _pickStart,
          ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: const Text('To'),
            subtitle: Text(
              widget.ctl.draftFilters.endTime
                      ?.toLocal()
                      .toString()
                      .split('.')
                      .first ??
                  'Any time',
            ),
            trailing: IconButton(
              tooltip: 'Clear end time',
              onPressed: widget.ctl.draftFilters.endTime == null
                  ? null
                  : () => setState(
                      () => widget.ctl.draftFilters = widget.ctl.draftFilters
                          .copyWith(clearEnd: true),
                    ),
              icon: const Icon(Icons.clear),
            ),
            onTap: _pickEnd,
          ),
          FilledButton(
            onPressed: widget.ctl.listState == SessionLoad.loading
                ? null
                : () => widget.ctl.applyFilters(),
            child: const Text('Apply filters'),
          ),
        ],
      ),
    ),
  );
}

class _PaginationBar extends StatelessWidget {
  const _PaginationBar({required this.ctl});
  final SessionController ctl;

  @override
  Widget build(BuildContext context) => Row(
    children: [
      Expanded(
        child: Text(
          'Showing ${ctl.sessions.length} of ${ctl.total} sessions · page ${ctl.pageIndex + 1}',
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ),
      IconButton(
        tooltip: 'Previous page',
        onPressed: ctl.canPrevPage && ctl.listState != SessionLoad.loading
            ? ctl.prevPage
            : null,
        icon: const Icon(Icons.chevron_left),
      ),
      IconButton(
        tooltip: 'Next page',
        onPressed: ctl.canNextPage && ctl.listState != SessionLoad.loading
            ? ctl.nextPage
            : null,
        icon: const Icon(Icons.chevron_right),
      ),
    ],
  );
}

class _SessionRow extends StatelessWidget {
  const _SessionRow({required this.info, required this.onTap});
  final SessionInfo info;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final title = info.title.isNotEmpty ? info.title : info.sessionId;
    final updated = info.hasLastUpdateTime()
        ? info.lastUpdateTime.toDateTime().toLocal().toString().split('.').first
        : '—';
    return ListTile(
      title: Text(title, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: Text(
        '${info.appName} · ${info.userId}\n$updated · ${info.turnCount} turns',
      ),
      isThreeLine: true,
      trailing: const Icon(Icons.chevron_right),
      onTap: onTap,
    );
  }
}

class SessionDetailPage extends StatefulWidget {
  const SessionDetailPage({
    super.key,
    required this.client,
    required this.sessionKey,
  });
  final ClientContext client;
  final SessionKey sessionKey;

  @override
  State<SessionDetailPage> createState() => _SessionDetailPageState();
}

class _SessionDetailPageState extends State<SessionDetailPage> {
  SessionController? get _ctl => widget.client.sessions;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      unawaited(_ctl?.openSession(widget.sessionKey));
    });
  }

  @override
  Widget build(BuildContext context) {
    final ctl = _ctl;
    if (ctl == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: ctl,
      builder: (context, _) {
        final detail = ctl.detail;
        final info = detail?.session;
        return Scaffold(
          appBar: AppBar(
            title: Text(
              info?.title.isNotEmpty == true
                  ? info!.title
                  : widget.sessionKey.sessionId,
            ),
            actions: [
              IconButton(
                tooltip: 'Rename title',
                onPressed: ctl.renameBlocked ? null : () => _rename(ctl, info),
                icon: const Icon(Icons.edit),
              ),
              IconButton(
                tooltip: 'Delete session',
                onPressed: ctl.deleteBlocked ? null : () => _delete(ctl),
                icon: const Icon(Icons.delete_outline),
              ),
            ],
          ),
          body: SafeArea(child: _buildBody(ctl, detail, info)),
          floatingActionButton: info == null
              ? null
              : FloatingActionButton.extended(
                  onPressed: ctl.replyBlocked ? null : () => _reply(ctl),
                  icon: const Icon(Icons.reply),
                  label: const Text('Reply'),
                ),
        );
      },
    );
  }

  Widget _buildBody(
    SessionController ctl,
    SessionDetail? detail,
    SessionInfo? info,
  ) {
    if (ctl.detailState == SessionLoad.loading && detail == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (ctl.detailError != null && detail == null) {
      return Center(child: Text(ctl.detailError!));
    }
    if (detail == null || info == null) {
      return const Center(child: Text('Session not found.'));
    }
    final events = parseSessionEvents(detail.events);
    final partial =
        !ctl.showAllEvents &&
        events.length >= SessionController.recentEventsLimit;
    final duration = formatSessionDetailDuration(
      detail.hasDuration() ? detail.duration : null,
    );
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        if (ctl.renameOutcomeNotice != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(ctl.renameOutcomeNotice!),
          ),
        if (ctl.deleteError != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              ctl.deleteError!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
        if (ctl.replyError != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              ctl.replyError!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
        _InfoGrid(info: info, duration: duration, partial: partial),
        const SizedBox(height: 12),
        OutlinedButton(
          onPressed: ctl.detailState == SessionLoad.loading
              ? null
              : ctl.toggleFullHistory,
          child: Text(
            ctl.showAllEvents
                ? 'Show recent ${SessionController.recentEventsLimit}'
                : 'Show full history',
          ),
        ),
        const SizedBox(height: 16),
        Text('Memory context', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        if (info.state.fields.isEmpty)
          const Text('No session state recorded.')
        else
          SelectableText(
            const JsonEncoder.withIndent('  ').convert(
              info.state.fields.map((k, v) => MapEntry(k, v.toString())),
            ),
          ),
        const SizedBox(height: 16),
        Text(
          'Event log (${events.length}${partial ? '+' : ''})',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        if (partial)
          const Padding(
            padding: EdgeInsets.only(top: 4, bottom: 8),
            child: Text(
              'Showing the most recent ${SessionController.recentEventsLimit} events. '
              'Duration and turn count reflect this range.',
              style: TextStyle(fontSize: 12),
            ),
          ),
        for (final evt in events) _EventCard(event: evt),
        const SizedBox(height: 80),
      ],
    );
  }

  Future<void> _rename(SessionController ctl, SessionInfo? info) async {
    final controller = TextEditingController(text: info?.title ?? '');
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Rename session'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(labelText: 'Title'),
          maxLength: 100,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    await ctl.rename(widget.sessionKey, controller.text);
  }

  Future<void> _delete(SessionController ctl) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete session?'),
        content: Text(
          'Delete session "${widget.sessionKey.sessionId}" and its events? '
          'AG-UI threads are stopped on the server before deletion.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    final deleted = await ctl.deleteSession(widget.sessionKey);
    if (deleted && mounted) Navigator.pop(context);
  }

  Future<void> _reply(SessionController ctl) async {
    final agents = widget.client.agents;
    if (agents.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Load Agents in this Workspace first.')),
      );
      return;
    }
    var agentId = agents.first.agentId;
    final message = TextEditingController();
    final result = await showDialog<bool>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setLocal) => AlertDialog(
          title: const Text('Reply to session'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Text(
                  'Sends one message through ReplySession with the Workspace header. '
                  'This does not open or migrate the session into Chat.',
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: agentId,
                  decoration: const InputDecoration(labelText: 'Agent'),
                  items: [
                    for (final a in agents)
                      DropdownMenuItem(
                        value: a.agentId,
                        child: Text(
                          a.displayName.isEmpty ? a.name : a.displayName,
                        ),
                      ),
                  ],
                  onChanged: (v) => setLocal(() => agentId = v ?? agentId),
                ),
                TextField(
                  controller: message,
                  decoration: const InputDecoration(labelText: 'Message'),
                  minLines: 2,
                  maxLines: 6,
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Send'),
            ),
          ],
        ),
      ),
    );
    if (result != true || !mounted) return;
    final reply = await ctl.reply(
      key: widget.sessionKey,
      agentId: agentId,
      message: message.text,
    );
    if (!mounted) return;
    if (reply.response != null) {
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Agent reply'),
          content: SingleChildScrollView(
            child: SelectableText(reply.response!),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Close'),
            ),
          ],
        ),
      );
    } else if (reply.unconfirmed) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Reply not confirmed; it may still have run. Reload before sending again.',
          ),
        ),
      );
    } else if (reply.error != null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(reply.error!)));
    }
  }
}

class _InfoGrid extends StatelessWidget {
  const _InfoGrid({
    required this.info,
    required this.duration,
    required this.partial,
  });
  final SessionInfo info;
  final String duration;
  final bool partial;

  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 16,
    runSpacing: 12,
    children: [
      _InfoChip(label: 'User', value: info.userId),
      _InfoChip(label: 'Channel', value: info.appName),
      _InfoChip(
        label: 'Last update',
        value: info.hasLastUpdateTime()
            ? info.lastUpdateTime
                  .toDateTime()
                  .toLocal()
                  .toString()
                  .split('.')
                  .first
            : '—',
      ),
      _InfoChip(
        label: partial ? 'Duration (recent)' : 'Duration',
        value: duration,
      ),
      _InfoChip(
        label: partial ? 'Turns (recent)' : 'Turns',
        value: '${info.turnCount}',
      ),
    ],
  );
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    mainAxisSize: MainAxisSize.min,
    children: [
      Text(label, style: Theme.of(context).textTheme.labelSmall),
      Text(value),
    ],
  );
}

class _EventCard extends StatelessWidget {
  const _EventCard({required this.event});
  final ParsedSessionEvent event;

  Future<void> _openTrace(BuildContext context) async {
    final url = event.traceUrl;
    if (url == null || url.isEmpty) return;
    final uri = Uri.tryParse(url);
    if (uri == null || uri.scheme != 'http' && uri.scheme != 'https') return;
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  }

  @override
  Widget build(BuildContext context) => Card(
    margin: const EdgeInsets.only(bottom: 8),
    child: Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Chip(label: Text(event.author)),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  event.timestamp?.toString().split('.').first ?? '',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ),
              if (event.traceUrl != null && event.traceUrl!.isNotEmpty)
                TextButton(
                  onPressed: () => _openTrace(context),
                  child: const Text('Trace'),
                )
              else if (event.traceId != null && event.traceId!.isNotEmpty)
                Text(
                  event.traceId!.length > 12
                      ? '${event.traceId!.substring(0, 12)}…'
                      : event.traceId!,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
            ],
          ),
          if (event.textParts.isEmpty &&
              event.functionCalls.isEmpty &&
              event.functionResponses.isEmpty)
            const Text('No content.')
          else ...[
            for (final part in event.textParts) SelectableText(part.text),
            for (final call in event.functionCalls)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Tool call: ${call.name}',
                      style: Theme.of(context).textTheme.labelMedium,
                    ),
                    SelectableText(
                      const JsonEncoder.withIndent('  ').convert(call.args),
                    ),
                  ],
                ),
              ),
            for (final resp in event.functionResponses)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Tool response: ${resp.name}',
                      style: Theme.of(context).textTheme.labelMedium,
                    ),
                    SelectableText(
                      const JsonEncoder.withIndent('  ').convert(resp.response),
                    ),
                  ],
                ),
              ),
          ],
        ],
      ),
    ),
  );
}
