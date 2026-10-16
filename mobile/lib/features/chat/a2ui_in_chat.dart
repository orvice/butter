import 'package:flutter/material.dart';

import '../a2ui/surface_store.dart';
import '../a2ui/surface_widgets.dart';
import 'chat_controller.dart';
import 'chat_model.dart';
import 'message_view.dart';

/// One assistant or user message plus any A2UI surfaces tied to it.
class ChatMessageBlock extends StatelessWidget {
  const ChatMessageBlock({
    super.key,
    required this.chat,
    required this.message,
    required this.agentName,
    this.locked = false,
  });

  final ChatController chat;
  final ChatMessage message;
  final String agentName;
  final bool locked;

  @override
  Widget build(BuildContext context) {
    final surfaces = chat.surfacesForMessage(message.id).toList();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        MessageView(message: message, agentName: agentName),
        for (final entry in surfaces)
          Padding(
            padding: const EdgeInsets.only(top: 4, left: 4, right: 4),
            child: _SurfaceHost(chat: chat, entry: entry, locked: locked),
          ),
      ],
    );
  }
}

class OrphanSurfacesBlock extends StatelessWidget {
  const OrphanSurfacesBlock({
    super.key,
    required this.chat,
    this.locked = false,
  });
  final ChatController chat;
  final bool locked;

  @override
  Widget build(BuildContext context) {
    final orphans = chat.orphanSurfaces.toList();
    if (orphans.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'From earlier in this conversation',
            style: Theme.of(context).textTheme.labelMedium,
          ),
          for (final entry in orphans)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: _SurfaceHost(chat: chat, entry: entry, locked: locked),
            ),
        ],
      ),
    );
  }
}

class _SurfaceHost extends StatelessWidget {
  const _SurfaceHost({
    required this.chat,
    required this.entry,
    required this.locked,
  });

  final ChatController chat;
  final SurfaceEntry entry;
  final bool locked;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: chat.surfaces,
      builder: (context, _) {
        final current = chat.surfaces.entry(entry.id) ?? entry;
        final draft = chat.surfaces.draftFor(current.id);
        return A2uiSurfaceView(
          entry: current,
          draft: draft,
          locked: locked || !chat.canSend,
          onDraftChanged: (next) => chat.surfaces.setDraft(current.id, next),
          onSubmit: () => chat.submitSurfaceForm(current.id),
        );
      },
    );
  }
}
