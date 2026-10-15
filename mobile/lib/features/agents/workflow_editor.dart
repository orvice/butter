import 'dart:collection';

import 'package:flutter/material.dart';

import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import 'workflow_draft.dart';
import 'workflow_graph_view.dart';

const _int32Max = 2147483647;

/// Agents an AGENT node may run: independent resources other than this
/// Workflow, each with a single parent. A node's own current reference is
/// always kept selectable, so an existing reference is never silently lost.
List<Agent> selectableWorkflowAgents(List<Agent> choices, String selfId) {
  final claimed = {
    for (final agent in choices)
      if (agent.agentId != selfId) ...agent.childAgentIds,
  };
  return [
    for (final agent in choices)
      if (agent.agentId.isNotEmpty &&
          agent.agentId != selfId &&
          !claimed.contains(agent.agentId))
        agent,
  ];
}

/// The draft's problems, with the Agent references checked once the editor
/// choices are loaded (before that the server still decides).
List<WorkflowIssue> workflowIssues(
  WorkflowDraft workflow, {
  required List<Agent> choices,
  required bool choicesReady,
  required String selfId,
}) => workflow.validate(
  knownAgentIds: choicesReady
      ? {
          for (final agent in selectableWorkflowAgents(choices, selfId))
            agent.agentId,
        }
      : null,
);

final class _Identity {
  const _Identity(this.object);
  final Object object;
  @override
  bool operator ==(Object other) =>
      other is _Identity && identical(object, other.object);
  @override
  int get hashCode => identityHashCode(object);
}

/// Structured node, edge and form editing for a Workflow Agent. It edits the
/// shared [WorkflowDraft] in place and never shows or accepts the graph as
/// one JSON document. Problems are shown where the user fixes them: on the
/// node, edge, form field or option.
class WorkflowEditor extends StatefulWidget {
  const WorkflowEditor({
    super.key,
    required this.workflow,
    required this.selfId,
    required this.choices,
    required this.choicesReady,
    required this.enabled,
    required this.showSummary,
    required this.onChanged,
  });
  final WorkflowDraft workflow;
  final String selfId;
  final List<Agent> choices;
  final bool choicesReady;
  final bool enabled;

  /// Lists every problem at the top, e.g. after a refused Save.
  final bool showSummary;
  final VoidCallback onChanged;
  @override
  State<WorkflowEditor> createState() => _WorkflowEditorState();
}

class _WorkflowEditorState extends State<WorkflowEditor> {
  final _expanded = HashSet<WorkflowNode>.identity();
  final _cardKeys = HashMap<WorkflowNode, GlobalKey>.identity();
  late List<WorkflowIssue> _issues;

  WorkflowDraft get _wf => widget.workflow;

  void _changed(VoidCallback change) {
    setState(change);
    widget.onChanged();
  }

  GlobalKey _cardKey(WorkflowNode node) =>
      _cardKeys.putIfAbsent(node, () => GlobalKey());

  void _open(int index) {
    final node = _wf.nodes[index];
    setState(() => _expanded.add(node));
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final context = _cardKey(node).currentContext;
      if (context != null && context.mounted) {
        Scrollable.ensureVisible(
          context,
          duration: const Duration(milliseconds: 200),
        );
      }
    });
  }

  Iterable<WorkflowIssue> _nodeIssues(int node) =>
      _issues.where((i) => i.nodeIndex == node && i.edgeIndex == null);

  String? _error(int node, Set<WorkflowIssueKind> kinds, {int? field}) {
    for (final issue in _nodeIssues(node)) {
      if (kinds.contains(issue.kind) &&
          (field == null || issue.fieldIndex == field)) {
        return issue.message;
      }
    }
    return null;
  }

  String _where(WorkflowIssue issue) {
    final node = issue.nodeIndex;
    if (node != null) {
      final name = _wf.nodes[node].name;
      final label = name.isEmpty ? 'Node ${node + 1}' : 'Node "$name"';
      final field = issue.fieldIndex;
      if (field == null) return label;
      final fieldName = _wf.nodes[node].form.fields[field].name;
      final fieldLabel = fieldName.isEmpty
          ? 'field ${field + 1}'
          : 'field "$fieldName"';
      final option = issue.optionIndex;
      return option == null
          ? '$label › $fieldLabel'
          : '$label › $fieldLabel › option ${option + 1}';
    }
    final edge = issue.edgeIndex;
    if (edge != null) return 'Edge ${describeWorkflowEdge(_wf.edges[edge])}';
    return 'Workflow';
  }

  @override
  Widget build(BuildContext context) {
    _issues = workflowIssues(
      _wf,
      choices: widget.choices,
      choicesReady: widget.choicesReady,
      selfId: widget.selfId,
    );
    final theme = Theme.of(context);
    final layout = _wf.layout();
    final graphIssues = _issues.where(
      (i) => i.nodeIndex == null && i.edgeIndex == null,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 16),
        Text('Workflow graph', style: theme.textTheme.titleMedium),
        const Text(
          'Nodes run in the order the edges connect them. Edit nodes and edges below; the graph updates as you edit. Tap a node to edit it.',
        ),
        if (widget.showSummary && _issues.isNotEmpty) _summary(theme),
        const SizedBox(height: 8),
        WorkflowGraphView(
          config: _wf.config,
          problemNodes: {
            for (final issue in _issues)
              if (issue.nodeIndex != null) issue.nodeIndex!,
          },
          onTapNode: _open,
        ),
        if (layout.unreached.isNotEmpty)
          Text(
            'Not reached from START: ${layout.unreached.map((i) => _wf.nodes[i].name.isEmpty ? '(unnamed)' : _wf.nodes[i].name).join(', ')}',
          ),
        for (final issue in graphIssues) _errorText(theme, issue.message),
        const SizedBox(height: 16),
        Text('Nodes', style: theme.textTheme.titleMedium),
        for (var i = 0; i < _wf.nodes.length; i++) _nodeCard(i),
        Wrap(
          spacing: 8,
          children: [
            for (final (key, kind, label) in [
              (
                'agent',
                WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT,
                'Add Agent node',
              ),
              (
                'human',
                WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT,
                'Add Human input node',
              ),
              (
                'router',
                WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER,
                'Add Router node',
              ),
              (
                'join',
                WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN,
                'Add Join node',
              ),
            ])
              OutlinedButton.icon(
                key: Key('add-node-$key'),
                onPressed: widget.enabled
                    ? () => _changed(() {
                        final node = _wf.addNode(kind);
                        _expanded.add(node);
                      })
                    : null,
                icon: const Icon(Icons.add),
                label: Text(label),
              ),
          ],
        ),
        const SizedBox(height: 16),
        Text('Edges', style: theme.textTheme.titleMedium),
        const Text(
          'Connect START to the first node. A route label sends a Router\'s matching input down that edge; the default edge catches input that matches no label.',
        ),
        for (var i = 0; i < _wf.edges.length; i++) _edgeCard(i),
        Align(
          alignment: Alignment.centerLeft,
          child: OutlinedButton.icon(
            key: const Key('add-edge'),
            onPressed: widget.enabled
                ? () => _changed(
                    () =>
                        _wf.addEdge(_wf.edges.isEmpty ? startNodeName : '', ''),
                  )
                : null,
            icon: const Icon(Icons.add),
            label: const Text('Add edge'),
          ),
        ),
      ],
    );
  }

  Widget _errorText(ThemeData theme, String message) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 2),
    child: Semantics(
      liveRegion: true,
      child: Text(message, style: TextStyle(color: theme.colorScheme.error)),
    ),
  );

  Widget _summary(ThemeData theme) => Card(
    key: const Key('workflow-issues'),
    color: theme.colorScheme.errorContainer,
    child: Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Fix ${_issues.length} problem${_issues.length == 1 ? '' : 's'} before saving:',
            style: TextStyle(color: theme.colorScheme.onErrorContainer),
          ),
          for (final issue in _issues)
            InkWell(
              onTap: issue.nodeIndex == null
                  ? null
                  : () => _open(issue.nodeIndex!),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: Text(
                  '${_where(issue)}: ${issue.message}',
                  style: TextStyle(color: theme.colorScheme.onErrorContainer),
                ),
              ),
            ),
        ],
      ),
    ),
  );

  Widget _text(
    Object owner,
    String id,
    String label,
    String value,
    void Function(String) set, {
    int lines = 1,
    String? helper,
    String? error,
    String? Function(String?)? validator,
    TextInputType? keyboard,
    bool enabled = true,
  }) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 6),
    child: TextFormField(
      key: ValueKey((_Identity(owner), id)),
      initialValue: value,
      enabled: widget.enabled && enabled,
      maxLines: lines,
      keyboardType: keyboard,
      autocorrect: false,
      autovalidateMode: AutovalidateMode.onUserInteraction,
      validator: validator,
      decoration: InputDecoration(
        labelText: label,
        helperText: helper,
        helperMaxLines: 4,
        errorText: error,
        errorMaxLines: 4,
      ),
      onChanged: (text) => _changed(() => set(text)),
    ),
  );

  /// A whole number the draft only takes when it parses, so a half-typed
  /// value never becomes a different one.
  Widget _whole(
    Object owner,
    String id,
    String label,
    int value,
    void Function(int) set, {
    String? helper,
    String? error,
  }) => _text(
    owner,
    id,
    label,
    '$value',
    (text) {
      final n = _parseWhole(text);
      if (n != null) set(n);
    },
    helper: helper,
    error: error,
    keyboard: TextInputType.number,
    validator: (text) => _parseWhole(text) == null
        ? 'Use a whole number from 0 to $_int32Max.'
        : null,
  );

  int? _parseWhole(String? text) {
    final trimmed = text?.trim() ?? '';
    if (trimmed.isEmpty) return 0;
    final n = RegExp(r'^\d+$').hasMatch(trimmed) ? int.tryParse(trimmed) : null;
    return n == null || n > _int32Max ? null : n;
  }

  Widget _switch(String label, bool value, void Function(bool) set) =>
      SwitchListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(label),
        value: value,
        onChanged: widget.enabled ? (v) => _changed(() => set(v)) : null,
      );

  Widget _dropdown<T>(
    Object owner,
    String id,
    String label,
    T? value,
    List<DropdownMenuItem<T>> items,
    void Function(T) set, {
    String? error,
  }) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 6),
    child: DropdownButtonFormField<T>(
      key: ValueKey((_Identity(owner), id, value)),
      initialValue: value,
      isExpanded: true,
      decoration: InputDecoration(
        labelText: label,
        errorText: error,
        errorMaxLines: 4,
      ),
      items: items,
      onChanged: widget.enabled
          ? (v) {
              if (v != null) set(v);
            }
          : null,
    ),
  );

  // --- nodes -------------------------------------------------------------

  String? _nameError(int index, String? value) {
    final name = value ?? '';
    if (name.isEmpty) return 'Give this node a name.';
    if (name == startNodeName) {
      return '"$startNodeName" is reserved for the entry point. Choose another name.';
    }
    for (var i = 0; i < _wf.nodes.length; i++) {
      if (i != index && _wf.nodes[i].name == name) {
        return 'Another node already uses this name.';
      }
    }
    return null;
  }

  Widget _nodeCard(int index) {
    final node = _wf.nodes[index];
    final theme = Theme.of(context);
    final open = _expanded.contains(node);
    final problems = _nodeIssues(index).length;
    return Card(
      key: _cardKey(node),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          ListTile(
            key: Key('node-header-$index'),
            leading: Icon(workflowKindIcon(node)),
            title: Text(
              '${index + 1}. ${node.name.isEmpty ? '(unnamed)' : node.name}',
            ),
            subtitle: Text(
              problems == 0
                  ? workflowKindLabel(node)
                  : '${workflowKindLabel(node)} · $problems problem${problems == 1 ? '' : 's'}',
              style: problems == 0
                  ? null
                  : TextStyle(color: theme.colorScheme.error),
            ),
            trailing: Icon(open ? Icons.expand_less : Icons.expand_more),
            onTap: () => setState(() {
              if (!_expanded.remove(node)) _expanded.add(node);
            }),
          ),
          // Collapsed bodies stay mounted, so Form validation still reaches
          // every field and a half-typed value is never skipped.
          Offstage(
            offstage: !open,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
              child: _nodeBody(index, node),
            ),
          ),
        ],
      ),
    );
  }

  Widget _nodeBody(int index, WorkflowNode node) {
    final kinds = [
      WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT,
      WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT,
      WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER,
      WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN,
    ];
    final isKnown = kinds.contains(node.kind);
    final isJoin = node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _text(
          node,
          'name',
          'Node name',
          node.name,
          (v) {
            if (_nameError(index, v) == null) _wf.renameNode(index, v);
          },
          helper:
              'Edges refer to this name. Names must be unique; renaming updates the edges.',
          validator: (v) => _nameError(index, v),
        ),
        _dropdown<WorkflowNodeKind>(
          node,
          'kind',
          'Node kind',
          isKnown ? node.kind : WorkflowNodeKind.WORKFLOW_NODE_KIND_UNSPECIFIED,
          [
            if (!isKnown)
              DropdownMenuItem(
                value: WorkflowNodeKind.WORKFLOW_NODE_KIND_UNSPECIFIED,
                child: Text(
                  '${workflowKindLabel(node)} (retained until you choose)',
                ),
              ),
            for (final kind in kinds)
              DropdownMenuItem(
                value: kind,
                child: Text(workflowKindLabel(WorkflowNode(kind: kind))),
              ),
          ],
          (kind) => _changeKind(index, kind),
          error: _error(index, {WorkflowIssueKind.kindUnsupported}),
        ),
        if (node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT)
          ..._agentFields(index, node),
        if (node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT)
          ..._humanInputFields(index, node),
        if (node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 6),
            child: Text(
              'The Router matches its input text (ignoring case and surrounding spaces) against the route labels of its outgoing edges. It needs one default edge.',
            ),
          ),
        if (isJoin)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 6),
            child: Text(
              'A Join waits for every incoming edge, then passes their outputs on together. Edges into a Join cannot carry a route label or be default. A Join has no timeout or retry.',
            ),
          ),
        if (!isJoin) ..._runOptions(index, node),
        const SizedBox(height: 8),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            key: Key('remove-node-$index'),
            onPressed: widget.enabled ? () => _removeNode(index) : null,
            icon: const Icon(Icons.delete_outline),
            label: const Text('Remove node'),
          ),
        ),
      ],
    );
  }

  Future<void> _changeKind(int index, WorkflowNodeKind kind) async {
    final node = _wf.nodes[index];
    if (node.kind == kind) return;
    final loses = <String>[
      if (node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT &&
          kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT &&
          (node.agentId.isNotEmpty || node.parallelWorker))
        'the Agent reference and parallel worker setting',
      if (node.hasForm() &&
          kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT)
        'the form',
      if (kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN &&
          (node.hasRetry() || node.timeoutSeconds != 0))
        'the timeout and retry settings',
    ];
    if (loses.isNotEmpty &&
        !await confirmChange(
          context,
          'Change node kind?',
          'This node will lose ${loses.join(' and ')}.',
          action: 'Change',
        )) {
      return;
    }
    if (!mounted) return;
    _changed(() => _wf.setNodeKind(index, kind));
  }

  Future<void> _removeNode(int index) async {
    final node = _wf.nodes[index];
    final touching = _wf.edges
        .where((e) => e.from == node.name || e.to == node.name)
        .length;
    if (!await confirmChange(
      context,
      'Remove node?',
      touching == 0
          ? 'Remove "${node.name}" from the workflow.'
          : 'Remove "${node.name}" and the $touching edge${touching == 1 ? '' : 's'} connected to it.',
      action: 'Remove',
    )) {
      return;
    }
    if (!mounted) return;
    _changed(() {
      _expanded.remove(node);
      _wf.removeNode(index);
    });
  }

  List<Widget> _agentFields(int index, WorkflowNode node) {
    final selectable = selectableWorkflowAgents(widget.choices, widget.selfId);
    final byId = {for (final a in widget.choices) a.agentId: a};
    String label(String id) {
      final agent = byId[id];
      if (agent == null) return id;
      final name = agent.displayName.isEmpty ? agent.name : agent.displayName;
      return '$id · $name';
    }

    final ids = {
      if (node.agentId.isNotEmpty) node.agentId,
      for (final agent in selectable) agent.agentId,
    };
    return [
      _dropdown<String>(
        node,
        'agent',
        'Agent to run',
        node.agentId.isEmpty ? null : node.agentId,
        [
          for (final id in ids)
            DropdownMenuItem(
              value: id,
              child: Text(
                selectable.any((a) => a.agentId == id) || !widget.choicesReady
                    ? label(id)
                    : '$id (not available)',
                overflow: TextOverflow.ellipsis,
              ),
            ),
        ],
        (id) => _changed(() => _wf.setNodeAgent(index, id)),
        error: _error(index, {
          WorkflowIssueKind.agentMissing,
          WorkflowIssueKind.agentUnknown,
        }),
      ),
      if (!widget.choicesReady)
        const Text('Load editor choices to pick an Agent.'),
      const Text(
        'The Agent stays an independent resource; this workflow becomes its parent. An Agent can have only one parent.',
      ),
      _switch(
        'Parallel worker',
        node.parallelWorker,
        (v) => node.parallelWorker = v,
      ),
      const Text(
        'Runs this node once per item of a JSON list it receives, concurrently, and gathers the results.',
      ),
    ];
  }

  List<Widget> _runOptions(int index, WorkflowNode node) => [
    _whole(
      node,
      'timeout',
      'Timeout (seconds)',
      node.timeoutSeconds,
      (n) => node.timeoutSeconds = n,
      helper: '0 means no per-node timeout.',
      error: _error(index, {WorkflowIssueKind.negativeValue}),
    ),
    _switch(
      'Retry failed runs',
      node.hasRetry(),
      (v) => _wf.setRetry(index, v),
    ),
    if (node.hasRetry()) ...[
      _whole(
        node.retry,
        'attempts',
        'Maximum attempts',
        node.retry.maxAttempts,
        (n) => node.retry.maxAttempts = n,
        helper: 'Including the first attempt. 0 or 1 means no retries.',
      ),
      _whole(
        node.retry,
        'initial',
        'First retry delay (seconds)',
        node.retry.initialDelaySeconds,
        (n) => node.retry.initialDelaySeconds = n,
        helper: '0 uses the engine default.',
      ),
      _whole(
        node.retry,
        'max-delay',
        'Longest delay (seconds)',
        node.retry.maxDelaySeconds,
        (n) => node.retry.maxDelaySeconds = n,
        helper: '0 uses the engine default.',
      ),
      _text(
        node.retry,
        'backoff',
        'Backoff factor',
        node.retry.backoffFactor == 0 ? '0' : '${node.retry.backoffFactor}',
        (v) {
          final n = double.tryParse(v.trim().isEmpty ? '0' : v.trim());
          if (n != null && n.isFinite && n >= 0) node.retry.backoffFactor = n;
        },
        helper:
            'Multiplier applied to the delay on each retry. 0 uses the engine default.',
        keyboard: const TextInputType.numberWithOptions(decimal: true),
        validator: (v) {
          final n = double.tryParse(
            v == null || v.trim().isEmpty ? '0' : v.trim(),
          );
          return n == null || !n.isFinite || n < 0
              ? 'Use a number of 0 or more.'
              : null;
        },
      ),
    ],
  ];

  // --- Human Input -------------------------------------------------------

  List<Widget> _humanInputFields(int index, WorkflowNode node) {
    final theme = Theme.of(context);
    final form = node.form;
    return [
      _text(
        node,
        'question',
        'Question',
        node.question,
        (v) => node.question = v,
        lines: 3,
        helper: 'Asked when the workflow pauses. Every entry point shows it.',
        error: _error(index, {WorkflowIssueKind.questionMissing}),
      ),
      _switch(
        'Present as a form',
        node.hasForm(),
        (v) => _wf.setForm(index, v),
      ),
      const Text(
        'A form is rendered by clients that support it; other entry points get the question followed by the field instructions and can answer in text.',
      ),
      if (node.hasForm()) ...[
        _text(
          form,
          'title',
          'Form title',
          form.title,
          (v) => form.title = v,
          helper: 'Empty uses the question.',
          error: _error(index, {WorkflowIssueKind.formTitle}),
        ),
        for (final issue in _nodeIssues(index).where(
          (i) =>
              i.kind == WorkflowIssueKind.formFields ||
              i.kind == WorkflowIssueKind.formOnWrongKind,
        ))
          _errorText(theme, issue.message),
        for (var f = 0; f < form.fields.length; f++) _fieldCard(index, node, f),
        Wrap(
          spacing: 8,
          children: [
            OutlinedButton.icon(
              key: Key('add-field-text-$index'),
              onPressed: widget.enabled && form.fields.length < maxFormFields
                  ? () => _changed(
                      () => _wf.addFormField(
                        index,
                        HumanInputFormFieldType
                            .HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
                      ),
                    )
                  : null,
              icon: const Icon(Icons.short_text),
              label: const Text('Add text field'),
            ),
            OutlinedButton.icon(
              key: Key('add-field-choice-$index'),
              onPressed: widget.enabled && form.fields.length < maxFormFields
                  ? () => _changed(
                      () => _wf.addFormField(
                        index,
                        HumanInputFormFieldType
                            .HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
                      ),
                    )
                  : null,
              icon: const Icon(Icons.radio_button_checked),
              label: const Text('Add single-choice field'),
            ),
          ],
        ),
        Text('${form.fields.length} of $maxFormFields fields'),
      ],
    ];
  }

  Widget _fieldCard(int node, WorkflowNode owner, int f) {
    final field = owner.form.fields[f];
    final count = owner.form.fields.length;
    final isText =
        field.type == HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT;
    final isChoice =
        field.type ==
        HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE;
    String? err(Set<WorkflowIssueKind> kinds) => _error(node, kinds, field: f);
    return Card(
      key: ValueKey((_Identity(field), 'card')),
      margin: const EdgeInsets.symmetric(vertical: 6),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Field ${f + 1}${field.name.isEmpty ? '' : ' · ${field.name}'}',
              style: Theme.of(context).textTheme.titleSmall,
            ),
            _text(
              field,
              'name',
              'Field name',
              field.name,
              (v) => field.name = v,
              helper:
                  'The key of this field in the submitted answer. Letters, digits and underscores.',
              error: err({WorkflowIssueKind.fieldName}),
            ),
            _text(
              field,
              'label',
              'Label',
              field.label,
              (v) => field.label = v,
              error: err({WorkflowIssueKind.fieldLabel}),
            ),
            _text(
              field,
              'hint',
              'Hint',
              field.hint,
              (v) => field.hint = v,
              error: err({WorkflowIssueKind.fieldHint}),
            ),
            _switch('Required', field.required, (v) => field.required = v),
            _dropdown<HumanInputFormFieldType>(
              field,
              'type',
              'Field type',
              isText || isChoice
                  ? field.type
                  : HumanInputFormFieldType
                        .HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED,
              [
                if (!isText && !isChoice)
                  DropdownMenuItem(
                    value: HumanInputFormFieldType
                        .HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED,
                    child: Text(
                      unknownFieldType(field) == null
                          ? 'Unsupported type'
                          : 'Unknown type (${unknownFieldType(field)}) — choose one',
                    ),
                  ),
                const DropdownMenuItem(
                  value:
                      HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
                  child: Text('Text'),
                ),
                const DropdownMenuItem(
                  value: HumanInputFormFieldType
                      .HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
                  child: Text('Single choice'),
                ),
              ],
              (type) => _changed(() => _wf.setFieldType(node, f, type)),
              error: err({WorkflowIssueKind.fieldType}),
            ),
            if (isText)
              _whole(
                field,
                'max-length',
                'Maximum length (characters)',
                field.maxLength,
                (n) => field.maxLength = n,
                helper: '0 uses $maxFormTextLength, the most allowed.',
                error: err({WorkflowIssueKind.fieldLength}),
              ),
            if (isChoice) ..._optionFields(node, f, field, err),
            if (!isChoice && field.options.isNotEmpty)
              Text(
                err({WorkflowIssueKind.fieldOptions}) ?? '',
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            Wrap(
              children: [
                IconButton(
                  tooltip: 'Move field ${f + 1} up',
                  icon: const Icon(Icons.arrow_upward),
                  onPressed: widget.enabled && f > 0
                      ? () => _changed(() => _wf.moveFormField(node, f, f - 1))
                      : null,
                ),
                IconButton(
                  tooltip: 'Move field ${f + 1} down',
                  icon: const Icon(Icons.arrow_downward),
                  onPressed: widget.enabled && f < count - 1
                      ? () => _changed(() => _wf.moveFormField(node, f, f + 1))
                      : null,
                ),
                TextButton.icon(
                  key: Key('remove-field-$node-$f'),
                  onPressed: widget.enabled
                      ? () => _changed(() => _wf.removeFormField(node, f))
                      : null,
                  icon: const Icon(Icons.remove_circle_outline),
                  label: Text('Remove field ${f + 1}'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  List<Widget> _optionFields(
    int node,
    int f,
    HumanInputFormField field,
    String? Function(Set<WorkflowIssueKind>) err,
  ) {
    final theme = Theme.of(context);
    String? optionError(int o, WorkflowIssueKind kind) {
      for (final issue in _nodeIssues(node)) {
        if (issue.kind == kind &&
            issue.fieldIndex == f &&
            issue.optionIndex == o) {
          return issue.message;
        }
      }
      return null;
    }

    return [
      const SizedBox(height: 8),
      Text('Options', style: theme.textTheme.titleSmall),
      if (err({WorkflowIssueKind.fieldOptions}) case final message?)
        _errorText(theme, message),
      if (field.maxLength != 0)
        _errorText(
          theme,
          err({WorkflowIssueKind.fieldLength}) ??
              'A maximum length applies to text fields only.',
        ),
      for (var o = 0; o < field.options.length; o++)
        Card(
          key: ValueKey((_Identity(field.options[o]), 'card')),
          color: theme.colorScheme.surfaceContainerHighest,
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _text(
                  field.options[o],
                  'value',
                  'Option ${o + 1} value',
                  field.options[o].value,
                  (v) => field.options[o].value = v,
                  helper: 'Submitted when this option is chosen.',
                  error: optionError(o, WorkflowIssueKind.optionValue),
                ),
                _text(
                  field.options[o],
                  'label',
                  'Option ${o + 1} label',
                  field.options[o].label,
                  (v) => field.options[o].label = v,
                  error: optionError(o, WorkflowIssueKind.optionLabel),
                ),
                Wrap(
                  children: [
                    IconButton(
                      tooltip: 'Move option ${o + 1} up',
                      icon: const Icon(Icons.arrow_upward),
                      onPressed: widget.enabled && o > 0
                          ? () => _changed(
                              () => _wf.moveOption(node, f, o, o - 1),
                            )
                          : null,
                    ),
                    IconButton(
                      tooltip: 'Move option ${o + 1} down',
                      icon: const Icon(Icons.arrow_downward),
                      onPressed: widget.enabled && o < field.options.length - 1
                          ? () => _changed(
                              () => _wf.moveOption(node, f, o, o + 1),
                            )
                          : null,
                    ),
                    TextButton.icon(
                      onPressed: widget.enabled
                          ? () => _changed(() => _wf.removeOption(node, f, o))
                          : null,
                      icon: const Icon(Icons.remove_circle_outline),
                      label: Text('Remove option ${o + 1}'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      Align(
        alignment: Alignment.centerLeft,
        child: OutlinedButton.icon(
          key: Key('add-option-$node-$f'),
          onPressed: widget.enabled && field.options.length < maxFormOptions
              ? () => _changed(() => _wf.addOption(node, f))
              : null,
          icon: const Icon(Icons.add),
          label: const Text('Add option'),
        ),
      ),
      Text('${field.options.length} of $maxFormOptions options'),
    ];
  }

  // --- edges -------------------------------------------------------------

  Widget _edgeCard(int index) {
    final edge = _wf.edges[index];
    final theme = Theme.of(context);
    final names = {
      for (final node in _wf.nodes)
        if (node.name.isNotEmpty) node.name,
    };
    DropdownMenuItem<String> item(String name) => DropdownMenuItem(
      value: name,
      child: Text(
        name == startNodeName || names.contains(name)
            ? name
            : '$name (not a node)',
        overflow: TextOverflow.ellipsis,
      ),
    );
    final problems = _issues.where((i) => i.edgeIndex == index);
    return Card(
      key: ValueKey((_Identity(edge), 'card')),
      margin: const EdgeInsets.symmetric(vertical: 6),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Edge ${index + 1}: ${describeWorkflowEdge(edge)}',
              style: theme.textTheme.titleSmall,
            ),
            _dropdown<String>(
              edge,
              'from',
              'From',
              edge.from.isEmpty ? null : edge.from,
              [
                for (final name in {
                  startNodeName,
                  ...names,
                  if (edge.from.isNotEmpty) edge.from,
                })
                  item(name),
              ],
              (v) => _changed(() => edge.from = v),
              error: problems
                  .where((i) => i.kind == WorkflowIssueKind.edgeSourceUnknown)
                  .map((i) => i.message)
                  .firstOrNull,
            ),
            _dropdown<String>(
              edge,
              'to',
              'To',
              edge.to.isEmpty ? null : edge.to,
              [
                for (final name in {...names, if (edge.to.isNotEmpty) edge.to})
                  item(name),
              ],
              (v) => _changed(() => edge.to = v),
              error: problems
                  .where((i) => i.kind == WorkflowIssueKind.edgeTargetUnknown)
                  .map((i) => i.message)
                  .firstOrNull,
            ),
            _text(
              edge,
              'route-${edge.isDefault}',
              'Route label',
              edge.route,
              (v) => _wf.setEdgeRoute(index, v),
              enabled: !edge.isDefault,
              helper: edge.isDefault
                  ? 'A default edge has no route label.'
                  : 'Taken only when the source\'s output matches this text. Empty always fires.',
              error: problems
                  .where(
                    (i) =>
                        i.kind == WorkflowIssueKind.routeDuplicate ||
                        i.kind == WorkflowIssueKind.routeAndDefault,
                  )
                  .map((i) => i.message)
                  .firstOrNull,
            ),
            SwitchListTile(
              key: ValueKey((_Identity(edge), 'default', edge.isDefault)),
              contentPadding: EdgeInsets.zero,
              title: const Text('Default edge'),
              subtitle: const Text(
                'Taken when no other route from the source matches. Setting it clears the route label.',
              ),
              value: edge.isDefault,
              onChanged: widget.enabled
                  ? (v) => _changed(() => _wf.setEdgeDefault(index, v))
                  : null,
            ),
            for (final issue in problems.where(
              (i) => i.kind == WorkflowIssueKind.routedIntoJoin,
            ))
              _errorText(theme, issue.message),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                key: Key('remove-edge-$index'),
                onPressed: widget.enabled
                    ? () => _changed(() => _wf.removeEdge(index))
                    : null,
                icon: const Icon(Icons.remove_circle_outline),
                label: Text('Remove edge ${index + 1}'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
