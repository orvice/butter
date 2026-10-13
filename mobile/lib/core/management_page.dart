import 'package:flutter/material.dart';

import '../app/client_context.dart';

Future<void> openManagementPage(
  BuildContext context,
  ClientContext client,
  Widget page,
) => Navigator.push<void>(
  context,
  MaterialPageRoute(
    builder: (_) => _ScopedRoute(client: client, page: page),
  ),
);

class _ScopedRoute extends StatefulWidget {
  const _ScopedRoute({required this.client, required this.page});
  final ClientContext client;
  final Widget page;
  @override
  State<_ScopedRoute> createState() => _ScopedRouteState();
}

class _ScopedRouteState extends State<_ScopedRoute> {
  late final _scope = widget.client.scope;
  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: widget.client,
    builder: (context, _) =>
        _scope == widget.client.scope && widget.client.user != null
        ? widget.page
        : Scaffold(
            appBar: AppBar(title: const Text('Context changed')),
            body: SafeArea(
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('Context changed. Return to the home screen.'),
                    TextButton(
                      onPressed: () =>
                          Navigator.popUntil(context, (route) => route.isFirst),
                      child: const Text('Return home'),
                    ),
                  ],
                ),
              ),
            ),
          ),
  );
}

/// Displays request progress and safe errors for an authorized route.
class ManagementPage extends StatelessWidget {
  const ManagementPage({
    super.key,
    required this.client,
    required this.title,
    required this.child,
  });
  final ClientContext client;
  final String title;
  final Widget child;
  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: client,
    builder: (context, _) => Scaffold(
      appBar: AppBar(title: Text(title)),
      body: SafeArea(
        child: Column(
          children: [
            if (client.busy)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (client.writeNotice != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(client.writeNotice!),
                ),
              ),
            if (client.error != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(
                    client.error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              ),
            Expanded(child: child),
          ],
        ),
      ),
    ),
  );
}

Future<bool> confirmChange(
  BuildContext context,
  String title,
  String description, {
  String action = 'Confirm',
}) async =>
    await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: Text(description),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(action),
          ),
        ],
      ),
    ) ??
    false;

/// Back navigation warns before discarding in-memory edits; writes cannot be
/// repeated by a navigation/lifecycle callback.
class DraftGuard extends StatefulWidget {
  const DraftGuard({
    super.key,
    required this.dirty,
    required this.busy,
    required this.child,
  });
  final bool dirty;
  final bool busy;
  final Widget child;
  @override
  State<DraftGuard> createState() => _DraftGuardState();
}

class _DraftGuardState extends State<DraftGuard> {
  bool _allowPop = false;
  bool _confirming = false;
  @override
  Widget build(BuildContext context) => PopScope(
    canPop: _allowPop || (!widget.dirty && !widget.busy),
    onPopInvokedWithResult: (didPop, result) async {
      if (didPop || widget.busy || _confirming) return;
      _confirming = true;
      final discard = await confirmChange(
        context,
        'Discard unsaved changes?',
        'Your draft is only kept on this screen.',
        action: 'Discard',
      );
      _confirming = false;
      if (!mounted || !discard) return;
      setState(() => _allowPop = true);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) Navigator.pop(context);
      });
    },
    child: widget.child,
  );
}
