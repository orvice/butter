import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';
import 'workflow_draft.dart';

String workflowKindLabel(WorkflowNode node) {
  if (unknownKind(node) != null) return 'Unknown kind (${unknownKind(node)})';
  return switch (node.kind) {
    WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT => 'Agent',
    WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT => 'Human input',
    WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER => 'Router',
    WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN => 'Join',
    _ => 'Unsupported kind',
  };
}

IconData workflowKindIcon(WorkflowNode node) => switch (node.kind) {
  WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT => Icons.smart_toy_outlined,
  WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT => Icons.person_outline,
  WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER => Icons.call_split,
  WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN => Icons.call_merge,
  _ => Icons.help_outline,
};

/// One plain-text line per edge, shared by the graph's screen-reader label and
/// the Agent detail page.
String describeWorkflowEdge(WorkflowEdge edge) {
  final suffix = edge.isDefault
      ? ' (default)'
      : edge.route.isEmpty
      ? ''
      : ' (route "${edge.route}")';
  return '${edge.from.isEmpty ? '(unnamed)' : edge.from} → '
      '${edge.to.isEmpty ? '(unnamed)' : edge.to}$suffix';
}

const _nodeWidth = 136.0;
const _nodeHeight = 56.0;
const _columnGap = 56.0;
const _rowGap = 20.0;
const _pad = 12.0;

/// The graph, auto-laid-out in columns by distance from START. It reads the
/// same [WorkflowConfig] the editor changes; there is no separate layout
/// model to keep in step. Nodes are real buttons so the graph is operable
/// by touch and by assistive technology; the painted edges are mirrored as
/// text in [describeWorkflowEdge] lines for the same reason.
class WorkflowGraphView extends StatelessWidget {
  const WorkflowGraphView({
    super.key,
    required this.config,
    this.problemNodes = const {},
    this.onTapNode,
  });
  final WorkflowConfig config;

  /// Indexes of nodes to flag as having problems.
  final Set<int> problemNodes;
  final void Function(int index)? onTapNode;

  @override
  Widget build(BuildContext context) {
    final layout = WorkflowLayout.of(config);
    final theme = Theme.of(context);
    // Column 0 is START; reached nodes follow, unreached nodes get a last
    // column of their own so nothing in the model is invisible.
    final columns = <List<int?>>[
      [null],
      for (final layer in layout.layers) layer,
      if (layout.unreached.isNotEmpty) layout.unreached,
    ];
    final positions = <int?, Offset>{};
    var rows = 1;
    for (var c = 0; c < columns.length; c++) {
      rows = math.max(rows, columns[c].length);
      for (var r = 0; r < columns[c].length; r++) {
        positions[columns[c][r]] = Offset(
          _pad + c * (_nodeWidth + _columnGap),
          _pad + r * (_nodeHeight + _rowGap),
        );
      }
    }
    final width =
        _pad * 2 +
        columns.length * _nodeWidth +
        (columns.length - 1) * _columnGap;
    final height = _pad * 2 + rows * _nodeHeight + (rows - 1) * _rowGap;
    final indexByName = <String, int>{};
    for (var i = 0; i < config.nodes.length; i++) {
      indexByName.putIfAbsent(config.nodes[i].name, () => i);
    }
    final lines = <_DrawnEdge>[];
    for (final edge in config.edges) {
      final from = edge.from == startNodeName ? null : indexByName[edge.from];
      final to = indexByName[edge.to];
      final fromPos = edge.from == startNodeName
          ? positions[null]
          : (from == null ? null : positions[from]);
      final toPos = to == null ? null : positions[to];
      if (fromPos == null || toPos == null) continue;
      final label = edge.isDefault
          ? 'default'
          : edge.route.isEmpty
          ? ''
          : edge.route;
      lines.add(_DrawnEdge(fromPos, toPos, label));
    }
    final summary = [
      'Workflow graph with ${config.nodes.length} nodes and ${config.edges.length} edges.',
      for (final edge in config.edges) describeWorkflowEdge(edge),
    ].join(' ');
    return Semantics(
      container: true,
      label: summary,
      child: SingleChildScrollView(
        key: const Key('workflow-graph'),
        scrollDirection: Axis.horizontal,
        child: SizedBox(
          width: width,
          height: height,
          child: Stack(
            children: [
              ExcludeSemantics(
                child: CustomPaint(
                  size: Size(width, height),
                  painter: _EdgePainter(
                    lines,
                    theme.colorScheme.outline,
                    theme.textTheme.labelSmall!.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                    theme.colorScheme.surface,
                  ),
                ),
              ),
              Positioned(
                left: positions[null]!.dx,
                top: positions[null]!.dy,
                width: _nodeWidth,
                height: _nodeHeight,
                child: Card(
                  margin: EdgeInsets.zero,
                  color: theme.colorScheme.secondaryContainer,
                  child: Center(
                    child: Text(
                      startNodeName,
                      style: theme.textTheme.titleSmall,
                    ),
                  ),
                ),
              ),
              for (var i = 0; i < config.nodes.length; i++)
                Positioned(
                  left: positions[i]!.dx,
                  top: positions[i]!.dy,
                  width: _nodeWidth,
                  height: _nodeHeight,
                  child: _GraphNode(
                    node: config.nodes[i],
                    problem: problemNodes.contains(i),
                    onTap: onTapNode == null ? null : () => onTapNode!(i),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _GraphNode extends StatelessWidget {
  const _GraphNode({required this.node, required this.problem, this.onTap});
  final WorkflowNode node;
  final bool problem;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final name = node.name.isEmpty ? '(unnamed)' : node.name;
    return Semantics(
      button: onTap != null,
      label:
          '$name, ${workflowKindLabel(node)}${problem ? ', has problems' : ''}',
      excludeSemantics: true,
      child: Card(
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: BorderSide(
            color: problem ? scheme.error : scheme.outlineVariant,
            width: problem ? 2 : 1,
          ),
        ),
        child: InkWell(
          borderRadius: BorderRadius.circular(12),
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: Row(
              children: [
                Icon(
                  problem ? Icons.error_outline : workflowKindIcon(node),
                  size: 20,
                  color: problem ? scheme.error : null,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(name, maxLines: 1, overflow: TextOverflow.ellipsis),
                      Text(
                        workflowKindLabel(node),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DrawnEdge {
  _DrawnEdge(this.from, this.to, this.label);
  final Offset from;
  final Offset to;
  final String label;
}

class _EdgePainter extends CustomPainter {
  _EdgePainter(this.edges, this.color, this.labelStyle, this.labelBackground);
  final List<_DrawnEdge> edges;
  final Color color;
  final TextStyle labelStyle;
  final Color labelBackground;

  @override
  void paint(Canvas canvas, Size size) {
    final stroke = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5;
    final fill = Paint()..color = color;
    for (final edge in edges) {
      final forward = edge.to.dx > edge.from.dx;
      final start = Offset(
        forward ? edge.from.dx + _nodeWidth : edge.from.dx + _nodeWidth / 2,
        edge.from.dy + (forward ? _nodeHeight / 2 : _nodeHeight),
      );
      final end = Offset(
        forward ? edge.to.dx : edge.to.dx + _nodeWidth / 2,
        edge.to.dy + (forward ? _nodeHeight / 2 : _nodeHeight),
      );
      final path = Path()..moveTo(start.dx, start.dy);
      Offset anchor;
      double headAngle;
      if (forward) {
        final mid = (end.dx - start.dx) / 2;
        path.cubicTo(
          start.dx + mid,
          start.dy,
          end.dx - mid,
          end.dy,
          end.dx,
          end.dy,
        );
        anchor = Offset((start.dx + end.dx) / 2, (start.dy + end.dy) / 2);
        headAngle = 0;
      } else {
        // Same column or a back edge: loop underneath the nodes.
        final drop = math.max(start.dy, end.dy) + 24;
        path.cubicTo(start.dx, drop, end.dx, drop, end.dx, end.dy);
        anchor = Offset((start.dx + end.dx) / 2, drop - 6);
        headAngle = -math.pi / 2;
      }
      canvas.drawPath(path, stroke);
      final head = Path()
        ..moveTo(end.dx, end.dy)
        ..lineTo(
          end.dx - 8 * math.cos(headAngle - 0.4),
          end.dy - 8 * math.sin(headAngle - 0.4),
        )
        ..lineTo(
          end.dx - 8 * math.cos(headAngle + 0.4),
          end.dy - 8 * math.sin(headAngle + 0.4),
        )
        ..close();
      canvas.drawPath(head, fill);
      if (edge.label.isNotEmpty) {
        final text = TextPainter(
          text: TextSpan(text: edge.label, style: labelStyle),
          maxLines: 1,
          ellipsis: '…',
          textDirection: TextDirection.ltr,
        )..layout(maxWidth: _columnGap + _nodeWidth / 2);
        final rect = Rect.fromCenter(
          center: anchor,
          width: text.width + 6,
          height: text.height + 2,
        );
        canvas.drawRect(rect, Paint()..color = labelBackground);
        text.paint(canvas, rect.topLeft + const Offset(3, 1));
      }
    }
  }

  @override
  bool shouldRepaint(_EdgePainter old) => true;
}
