import '../../gen/agents/v1/agent.pb.dart';
import 'agent_draft.dart';

/// An unrecognized enum from a newer server is kept on the wire as an unknown
/// field while the getter answers the default. Only an explicit choice by the
/// user replaces it.
int? unknownKind(WorkflowNode node) => unknownEnum(node, 2);
int? unknownFieldType(HumanInputFormField field) => unknownEnum(field, 5);

/// The reserved entry sentinel that may appear as an edge's `from`.
const startNodeName = 'START';

enum WorkflowIssueKind {
  noNodes,
  noEntry,
  nodeNameMissing,
  nodeNameReserved,
  nodeNameDuplicate,
  kindUnsupported,
  agentMissing,
  agentUnknown,
  questionMissing,
  routerNeedsDefault,
  negativeValue,
  edgeSourceUnknown,
  edgeTargetUnknown,
  routeAndDefault,
  routeDuplicate,
  routedIntoJoin,
  parallelOnWrongKind,
  formOnWrongKind,
  formTitle,
  formFields,
  fieldName,
  fieldLabel,
  fieldHint,
  fieldType,
  fieldLength,
  fieldOptions,
  optionValue,
  optionLabel,
}

/// Limits the server applies to a Human Input form.
const maxFormFields = 20;
const maxFormTextLength = 2000;
const maxFormOptions = 50;
const maxFormNameLength = 64;
const maxFormLabelRunes = 200;
const maxFormHintRunes = 500;
const maxFormTitleRunes = 200;
final _fieldName = RegExp(r'^[A-Za-z_][A-Za-z0-9_]*$');

/// One problem with the draft, located at the node, edge, form field or
/// option it belongs to. A graph-level problem has no location.
final class WorkflowIssue {
  const WorkflowIssue(
    this.kind,
    this.message, {
    this.nodeIndex,
    this.edgeIndex,
    this.fieldIndex,
    this.optionIndex,
  });
  final WorkflowIssueKind kind;
  final String message;
  final int? nodeIndex;
  final int? edgeIndex;
  final int? fieldIndex;
  final int? optionIndex;
}

/// Nodes grouped by their distance from START, for the graph view. Built
/// from the same [WorkflowConfig] the editor changes, so view and edits
/// cannot disagree. [layers] hold node indexes; [unreached] lists nodes no
/// path from START reaches.
final class WorkflowLayout {
  WorkflowLayout._(this.layers, this.unreached);

  factory WorkflowLayout.of(WorkflowConfig config) {
    final indexByName = <String, int>{};
    for (var i = 0; i < config.nodes.length; i++) {
      indexByName.putIfAbsent(config.nodes[i].name, () => i);
    }
    final depth = <int, int>{};
    var frontier = <int>[];
    for (final edge in config.edges) {
      final to = indexByName[edge.to];
      if (edge.from == startNodeName && to != null && !depth.containsKey(to)) {
        depth[to] = 0;
        frontier.add(to);
      }
    }
    for (var level = 0; frontier.isNotEmpty; level++) {
      final next = <int>[];
      for (final from in frontier) {
        for (final edge in config.edges) {
          if (indexByName[edge.from] != from) continue;
          final to = indexByName[edge.to];
          if (to != null && !depth.containsKey(to)) {
            depth[to] = level + 1;
            next.add(to);
          }
        }
      }
      frontier = next;
    }
    final layers = <List<int>>[];
    for (var i = 0; i < config.nodes.length; i++) {
      final d = depth[i];
      if (d == null) continue;
      while (layers.length <= d) {
        layers.add([]);
      }
      layers[d].add(i);
    }
    return WorkflowLayout._(layers, [
      for (var i = 0; i < config.nodes.length; i++)
        if (!depth.containsKey(i)) i,
    ]);
  }

  final List<List<int>> layers;
  final List<int> unreached;
}

/// Structured edits of a Workflow Agent's graph, applied in place to the
/// [AgentDraft]'s message. Only user edits touch fields, so unknown fields on
/// nodes, edges and forms survive a save.
final class WorkflowDraft {
  WorkflowDraft(this._draft);
  final AgentDraft _draft;

  WorkflowConfig get config {
    if (!_draft.config.hasWorkflow()) _draft.config.workflow = WorkflowConfig();
    return _draft.config.workflow;
  }

  List<WorkflowNode> get nodes => config.nodes;

  WorkflowNode addNode(WorkflowNodeKind kind) {
    final prefix = switch (kind) {
      WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT => 'agent',
      WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT => 'ask',
      WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER => 'router',
      WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN => 'join',
      _ => 'node',
    };
    var n = 1;
    while (nodes.any((node) => node.name == '${prefix}_$n')) {
      n++;
    }
    final node = WorkflowNode(name: '${prefix}_$n', kind: kind);
    nodes.add(node);
    return node;
  }

  List<WorkflowEdge> get edges => config.edges;

  /// Edges follow a rename, so the graph never silently loses a connection.
  void renameNode(int index, String name) {
    final node = nodes[index];
    final old = node.name;
    if (old == name) return;
    node.name = name;
    for (final edge in edges) {
      if (edge.from == old) edge.from = name;
      if (edge.to == old) edge.to = name;
    }
  }

  /// Removing a node removes only the edges that touched it.
  void removeNode(int index) {
    final node = nodes.removeAt(index);
    edges.removeWhere((e) => e.from == node.name || e.to == node.name);
    if (node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT) {
      _unreference(node.agentId);
    }
  }

  /// The runtime builds AGENT-node targets from `child_agent_ids`, so a
  /// reference always adds its child. A child leaves only when the user
  /// removes its last reference; children the user never touched stay.
  void setNodeAgent(int index, String agentId) {
    final node = nodes[index];
    final old = node.agentId;
    if (agentId.isEmpty) {
      node.clearAgentId();
    } else {
      node.agentId = agentId;
      if (!_draft.agent.childAgentIds.contains(agentId)) {
        _draft.agent.childAgentIds.add(agentId);
      }
    }
    if (old != agentId) _unreference(old);
  }

  void _unreference(String agentId) {
    if (agentId.isEmpty) return;
    final stillUsed = nodes.any(
      (n) =>
          n.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT &&
          n.agentId == agentId,
    );
    if (!stillUsed) _draft.agent.childAgentIds.remove(agentId);
  }

  /// Clears only what the server rejects on the new kind: the form outside
  /// HUMAN_INPUT, the agent binding and parallel worker outside AGENT, and
  /// retry/timeout on JOIN, which takes no node options.
  void setNodeKind(int index, WorkflowNodeKind kind) {
    final node = nodes[index];
    if (node.kind == kind) return;
    final leftAgent = node.kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT;
    final old = node.agentId;
    node.unknownFields.clearField(2);
    node.kind = kind;
    if (kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT) {
      node.clearForm();
    }
    if (kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT) {
      node
        ..clearParallelWorker()
        ..clearAgentId();
      if (leftAgent) _unreference(old);
    }
    if (kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN) {
      node
        ..clearTimeoutSeconds()
        ..clearRetry();
    }
  }

  WorkflowEdge addEdge(String from, String to) {
    final edge = WorkflowEdge(from: from, to: to);
    edges.add(edge);
    return edge;
  }

  void removeEdge(int index) => edges.removeAt(index);

  void setEdgeRoute(int index, String route) {
    final edge = edges[index];
    edge.route = route;
    if (route.isNotEmpty) edge.clearIsDefault();
  }

  void setEdgeDefault(int index, bool isDefault) {
    final edge = edges[index];
    if (isDefault) {
      edge.clearRoute();
      edge.isDefault = true;
    } else {
      edge.clearIsDefault();
    }
  }

  void setRetry(int index, bool enabled) {
    final node = nodes[index];
    if (!enabled) {
      node.clearRetry();
    } else if (!node.hasRetry()) {
      node.retry = WorkflowRetryConfig();
    }
  }

  // --- Human Input form -------------------------------------------------

  /// Turning the form off removes only the form; the plain question stays.
  void setForm(int node, bool enabled) {
    final target = nodes[node];
    if (!enabled) {
      target.clearForm();
    } else if (!target.hasForm()) {
      target.form = HumanInputForm();
    }
  }

  List<HumanInputFormField> _fields(int node) => nodes[node].form.fields;

  HumanInputFormField addFormField(int node, HumanInputFormFieldType type) {
    final fields = _fields(node);
    var n = 1;
    while (fields.any((f) => f.name == 'field_$n')) {
      n++;
    }
    final field = HumanInputFormField(name: 'field_$n', type: type);
    fields.add(field);
    return field;
  }

  void removeFormField(int node, int field) => _fields(node).removeAt(field);

  void moveFormField(int node, int from, int to) {
    final fields = _fields(node);
    fields.insert(to, fields.removeAt(from));
  }

  /// Text fields carry a length limit and single-choice fields carry
  /// options; each is cleared when the field leaves its type.
  void setFieldType(int node, int field, HumanInputFormFieldType type) {
    final target = _fields(node)[field];
    if (target.type == type && unknownFieldType(target) == null) return;
    target.unknownFields.clearField(5);
    target.type = type;
    if (type !=
        HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE) {
      target.options.clear();
    }
    if (type != HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT) {
      target.clearMaxLength();
    }
  }

  HumanInputFormOption addOption(int node, int field) {
    final options = _fields(node)[field].options;
    var n = 1;
    while (options.any((o) => o.value == 'option_$n')) {
      n++;
    }
    final option = HumanInputFormOption(value: 'option_$n');
    options.add(option);
    return option;
  }

  void removeOption(int node, int field, int option) =>
      _fields(node)[field].options.removeAt(option);

  void moveOption(int node, int field, int from, int to) {
    final options = _fields(node)[field].options;
    options.insert(to, options.removeAt(from));
  }

  WorkflowLayout layout() => WorkflowLayout.of(config);

  /// The rules the server enforces on save, checked where the user edits.
  /// [knownAgentIds] are the Agents an AGENT node may reference; null skips
  /// that check (the server still applies it).
  List<WorkflowIssue> validate({Set<String>? knownAgentIds}) {
    final issues = <WorkflowIssue>[];
    if (nodes.isEmpty) {
      issues.add(
        const WorkflowIssue(
          WorkflowIssueKind.noNodes,
          'Add at least one node to the workflow.',
        ),
      );
    }
    final kinds = <String, WorkflowNodeKind>{};
    for (var i = 0; i < nodes.length; i++) {
      final node = nodes[i];
      void add(WorkflowIssueKind kind, String message) =>
          issues.add(WorkflowIssue(kind, message, nodeIndex: i));
      if (node.name.isEmpty) {
        add(WorkflowIssueKind.nodeNameMissing, 'Give this node a name.');
      } else if (node.name == startNodeName) {
        add(
          WorkflowIssueKind.nodeNameReserved,
          '"$startNodeName" is reserved for the entry point. Choose another name.',
        );
      } else if (kinds.containsKey(node.name)) {
        add(
          WorkflowIssueKind.nodeNameDuplicate,
          'Another node already uses this name.',
        );
      } else {
        kinds[node.name] = node.kind;
      }
      switch (node.kind) {
        case WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT:
          if (node.agentId.isEmpty) {
            add(WorkflowIssueKind.agentMissing, 'Choose the Agent to run.');
          } else if (knownAgentIds != null &&
              !knownAgentIds.contains(node.agentId)) {
            add(
              WorkflowIssueKind.agentUnknown,
              'This Agent is not available to reference in this Workspace.',
            );
          }
        case WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT:
          if (node.question.trim().isEmpty) {
            add(
              WorkflowIssueKind.questionMissing,
              'Enter the question to ask the person.',
            );
          }
        case WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER:
        case WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN:
          break;
        default:
          add(
            WorkflowIssueKind.kindUnsupported,
            'Choose Agent, Human input, Router or Join.',
          );
      }
      if (node.parallelWorker &&
          node.kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT) {
        add(
          WorkflowIssueKind.parallelOnWrongKind,
          'Parallel worker applies only to Agent nodes.',
        );
      }
      if (_hasForm(node)) {
        if (node.kind != WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT) {
          add(
            WorkflowIssueKind.formOnWrongKind,
            'A form is supported only on Human input nodes.',
          );
        } else {
          issues.addAll(_validateForm(i, node.form));
        }
      }
      if (node.timeoutSeconds < 0 ||
          (node.hasRetry() &&
              (node.retry.maxAttempts < 0 ||
                  node.retry.initialDelaySeconds < 0 ||
                  node.retry.maxDelaySeconds < 0 ||
                  node.retry.backoffFactor < 0))) {
        add(
          WorkflowIssueKind.negativeValue,
          'Timeout and retry values must not be negative.',
        );
      }
    }

    var hasEntry = false;
    final defaults = <String>{};
    final labels = <String>{};
    for (var i = 0; i < edges.length; i++) {
      final edge = edges[i];
      void add(WorkflowIssueKind kind, String message) =>
          issues.add(WorkflowIssue(kind, message, edgeIndex: i));
      if (edge.from == startNodeName) {
        hasEntry = true;
      } else if (!kinds.containsKey(edge.from)) {
        add(WorkflowIssueKind.edgeSourceUnknown, 'Choose an existing source.');
      }
      if (!kinds.containsKey(edge.to)) {
        add(WorkflowIssueKind.edgeTargetUnknown, 'Choose an existing target.');
      }
      if (edge.isDefault && edge.route.isNotEmpty) {
        add(
          WorkflowIssueKind.routeAndDefault,
          'Use either a route label or the default flag, not both.',
        );
      }
      if (edge.isDefault) defaults.add(edge.from);
      if (edge.route.isNotEmpty &&
          !labels.add('${edge.from}\u0000${edge.route.trim().toLowerCase()}')) {
        add(
          WorkflowIssueKind.routeDuplicate,
          'Another route from this node matches the same text (ignoring case and surrounding spaces).',
        );
      }
      if ((edge.route.isNotEmpty || edge.isDefault) &&
          kinds[edge.to] == WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN) {
        add(
          WorkflowIssueKind.routedIntoJoin,
          'A routed or default edge cannot target a Join: it waits for every predecessor, and a branch skipped by routing never arrives.',
        );
      }
    }
    if (nodes.isNotEmpty && !hasEntry) {
      issues.add(
        const WorkflowIssue(
          WorkflowIssueKind.noEntry,
          'Add an edge from START to the first node.',
        ),
      );
    }
    for (var i = 0; i < nodes.length; i++) {
      if (nodes[i].kind == WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER &&
          !defaults.contains(nodes[i].name)) {
        issues.add(
          WorkflowIssue(
            WorkflowIssueKind.routerNeedsDefault,
            'A Router needs an outgoing default edge for input that matches no route.',
            nodeIndex: i,
          ),
        );
      }
    }
    return issues;
  }

  /// Like the server, an empty form (no title, no fields) is a plain question.
  static bool _hasForm(WorkflowNode node) =>
      node.hasForm() &&
      (node.form.title.trim().isNotEmpty || node.form.fields.isNotEmpty);

  static List<WorkflowIssue> _validateForm(int node, HumanInputForm form) {
    final issues = <WorkflowIssue>[];
    void add(
      WorkflowIssueKind kind,
      String message, {
      int? field,
      int? option,
    }) => issues.add(
      WorkflowIssue(
        kind,
        message,
        nodeIndex: node,
        fieldIndex: field,
        optionIndex: option,
      ),
    );
    if (form.title.runes.length > maxFormTitleRunes) {
      add(
        WorkflowIssueKind.formTitle,
        'The form title is limited to $maxFormTitleRunes characters.',
      );
    }
    if (form.fields.isEmpty) {
      add(WorkflowIssueKind.formFields, 'A form needs at least one field.');
    } else if (form.fields.length > maxFormFields) {
      add(
        WorkflowIssueKind.formFields,
        'A form may have at most $maxFormFields fields.',
      );
    }
    final names = <String>{};
    for (var f = 0; f < form.fields.length; f++) {
      final field = form.fields[f];
      void addField(WorkflowIssueKind kind, String message) =>
          add(kind, message, field: f);
      if (field.name.isEmpty ||
          field.name.length > maxFormNameLength ||
          !_fieldName.hasMatch(field.name)) {
        addField(
          WorkflowIssueKind.fieldName,
          'Use up to $maxFormNameLength letters, digits or underscores, starting with a letter or underscore.',
        );
      } else if (!names.add(field.name)) {
        addField(
          WorkflowIssueKind.fieldName,
          'Another field already uses this name.',
        );
      }
      if (field.label.trim().isEmpty ||
          field.label.runes.length > maxFormLabelRunes) {
        addField(
          WorkflowIssueKind.fieldLabel,
          'Enter a label of up to $maxFormLabelRunes characters.',
        );
      }
      if (field.hint.runes.length > maxFormHintRunes) {
        addField(
          WorkflowIssueKind.fieldHint,
          'The hint is limited to $maxFormHintRunes characters.',
        );
      }
      switch (field.type) {
        case HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT:
          if (field.maxLength < 0 || field.maxLength > maxFormTextLength) {
            addField(
              WorkflowIssueKind.fieldLength,
              'Use a maximum length from 0 to $maxFormTextLength (0 uses $maxFormTextLength).',
            );
          }
          if (field.options.isNotEmpty) {
            addField(
              WorkflowIssueKind.fieldOptions,
              'A text field has no options.',
            );
          }
        case HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE:
          if (field.maxLength != 0) {
            addField(
              WorkflowIssueKind.fieldLength,
              'A maximum length applies to text fields only.',
            );
          }
          if (field.options.isEmpty) {
            addField(
              WorkflowIssueKind.fieldOptions,
              'A single-choice field needs at least one option.',
            );
          } else if (field.options.length > maxFormOptions) {
            addField(
              WorkflowIssueKind.fieldOptions,
              'A single-choice field may have at most $maxFormOptions options.',
            );
          }
          final values = <String>{};
          for (var o = 0; o < field.options.length; o++) {
            final option = field.options[o];
            if (option.value.trim().isEmpty ||
                option.value.runes.length > maxFormLabelRunes) {
              add(
                WorkflowIssueKind.optionValue,
                'Enter a value of up to $maxFormLabelRunes characters.',
                field: f,
                option: o,
              );
            } else if (!values.add(option.value)) {
              add(
                WorkflowIssueKind.optionValue,
                'Another option already uses this value.',
                field: f,
                option: o,
              );
            }
            if (option.label.trim().isEmpty ||
                option.label.runes.length > maxFormLabelRunes) {
              add(
                WorkflowIssueKind.optionLabel,
                'Enter a label of up to $maxFormLabelRunes characters.',
                field: f,
                option: o,
              );
            }
          }
        default:
          addField(
            WorkflowIssueKind.fieldType,
            'Choose Text or Single choice for this field.',
          );
      }
    }
    return issues;
  }
}
