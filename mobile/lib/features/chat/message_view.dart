import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_markdown_plus/flutter_markdown_plus.dart';
import 'package:markdown/markdown.dart' as md;

import 'chat_model.dart';

/// How much of a tool's arguments and result is drawn. The full text stays in
/// the message; a very long result must not stall scrolling.
const _argumentsShown = 2000;
const _resultShown = 4000;

String _clip(String text, int limit) =>
    text.length <= limit ? text : '${text.substring(0, limit)}…';

/// One message: a user turn (right) or the agent's reply (left, with the
/// agent's name).
class MessageView extends StatelessWidget {
  const MessageView({
    super.key,
    required this.message,
    required this.agentName,
  });
  final ChatMessage message;
  final String agentName;

  @override
  Widget build(BuildContext context) {
    final user = message.role == ChatRole.user;
    final scheme = Theme.of(context).colorScheme;
    final body = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (final part in message.parts)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 2),
            child: switch (part) {
              TextPart() =>
                user ? Text(part.text) : ChatMarkdown(data: part.text),
              ImagePart() => MessageImage(base64: part.base64),
              UnavailableImagePart() => const Chip(
                avatar: Icon(Icons.broken_image_outlined, size: 18),
                label: Text('Image unavailable'),
              ),
              ToolPart() => ToolCard(tool: part),
            },
          ),
      ],
    );
    return Align(
      alignment: user ? Alignment.centerRight : Alignment.centerLeft,
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: MediaQuery.sizeOf(context).width * (user ? 0.85 : 0.95),
        ),
        child: Card(
          color: user ? scheme.primaryContainer : scheme.surfaceContainerHigh,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (!user)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        ExcludeSemantics(
                          child: CircleAvatar(
                            radius: 10,
                            child: Text(
                              agentName.isEmpty
                                  ? '?'
                                  : agentName.characters.first.toUpperCase(),
                              style: const TextStyle(fontSize: 11),
                            ),
                          ),
                        ),
                        const SizedBox(width: 6),
                        Flexible(
                          child: Text(
                            agentName,
                            overflow: TextOverflow.ellipsis,
                            style: Theme.of(context).textTheme.labelMedium,
                          ),
                        ),
                      ],
                    ),
                  ),
                body,
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// A reply's Markdown. Remote images are never fetched (that would let a
/// model's text make the device call out), links are shown but not opened,
/// and code blocks carry a copy button.
class ChatMarkdown extends StatelessWidget {
  const ChatMarkdown({super.key, required this.data});
  final String data;

  @override
  Widget build(BuildContext context) => MarkdownBody(
    data: data,
    extensionSet: md.ExtensionSet.gitHubFlavored,
    builders: {'pre': CodeBlockBuilder()},
    imageBuilder: (uri, title, alt) => Text('[image: ${alt ?? uri}]'),
  );
}

class CodeBlockBuilder extends MarkdownElementBuilder {
  @override
  bool isBlockElement() => true;

  @override
  Widget? visitElementAfterWithContext(
    BuildContext context,
    md.Element element,
    TextStyle? preferredStyle,
    TextStyle? parentStyle,
  ) {
    final code = element.textContent.trimRight();
    String? language;
    final first = element.children?.firstOrNull;
    if (first is md.Element) {
      final css = first.attributes['class'];
      if (css != null && css.startsWith('language-')) {
        language = css.substring('language-'.length);
      }
    }
    return _CodeBlock(code: code, language: language);
  }
}

class _CodeBlock extends StatelessWidget {
  const _CodeBlock({required this.code, this.language});
  final String code;
  final String? language;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 4),
      decoration: BoxDecoration(
        color: scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.only(left: 12),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    language ?? '',
                    style: Theme.of(context).textTheme.labelSmall,
                  ),
                ),
                IconButton(
                  tooltip: 'Copy code',
                  visualDensity: VisualDensity.compact,
                  icon: const Icon(Icons.copy, size: 18),
                  onPressed: () async {
                    final messenger = ScaffoldMessenger.maybeOf(context);
                    await Clipboard.setData(ClipboardData(text: code));
                    messenger?.showSnackBar(
                      const SnackBar(content: Text('Code copied')),
                    );
                  },
                ),
              ],
            ),
          ),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
            child: Text(
              code,
              style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
            ),
          ),
        ],
      ),
    );
  }
}

/// A tool call: its name and status in the header, the arguments and result
/// when opened.
class ToolCard extends StatelessWidget {
  const ToolCard({super.key, required this.tool});
  final ToolPart tool;

  @override
  Widget build(BuildContext context) {
    final status = switch (tool.status) {
      ToolStatus.calling => 'Calling…',
      ToolStatus.running => 'Running…',
      ToolStatus.done => 'Done',
    };
    final label = Theme.of(context).textTheme.labelSmall;
    return Card(
      margin: EdgeInsets.zero,
      child: ExpansionTile(
        dense: true,
        leading: const Icon(Icons.build_outlined, size: 18),
        title: Text(tool.name),
        trailing: Text(status),
        expandedCrossAxisAlignment: CrossAxisAlignment.start,
        childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
        children: [
          Text('Arguments', style: label),
          SelectableText(
            _clip(
              tool.arguments.isEmpty ? '{}' : tool.arguments,
              _argumentsShown,
            ),
            style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
          ),
          if (tool.result != null) ...[
            const SizedBox(height: 8),
            Text('Result', style: label),
            SelectableText(
              _clip(tool.result!, _resultShown),
              style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
            ),
          ],
        ],
      ),
    );
  }
}

/// An inline image decoded once, not on every frame of a streaming reply.
class MessageImage extends StatefulWidget {
  const MessageImage({super.key, required this.base64});
  final String base64;
  @override
  State<MessageImage> createState() => _MessageImageState();
}

class _MessageImageState extends State<MessageImage> {
  Uint8List? _bytes;
  @override
  void initState() {
    super.initState();
    _bytes = _decode(widget.base64);
  }

  @override
  void didUpdateWidget(MessageImage old) {
    super.didUpdateWidget(old);
    if (old.base64 != widget.base64) _bytes = _decode(widget.base64);
  }

  static Uint8List? _decode(String data) {
    try {
      return base64Decode(data);
    } on FormatException {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    const unavailable = Chip(
      avatar: Icon(Icons.broken_image_outlined, size: 18),
      label: Text('Image unavailable'),
    );
    final bytes = _bytes;
    if (bytes == null) return unavailable;
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: Image.memory(
        bytes,
        width: 220,
        fit: BoxFit.cover,
        semanticLabel: 'Attached image',
        errorBuilder: (_, _, _) => unavailable,
      ),
    );
  }
}
