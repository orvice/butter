import 'package:butter_mobile/features/agents/agent_draft.dart';
import 'package:butter_mobile/features/agents/workflow_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

const agentKind = WorkflowNodeKind.WORKFLOW_NODE_KIND_AGENT;
const humanKind = WorkflowNodeKind.WORKFLOW_NODE_KIND_HUMAN_INPUT;
const routerKind = WorkflowNodeKind.WORKFLOW_NODE_KIND_ROUTER;
const joinKind = WorkflowNodeKind.WORKFLOW_NODE_KIND_JOIN;

WorkflowDraft newWorkflow({Agent? base}) {
  final draft = AgentDraft(base: base);
  if (base == null) draft.setType(AgentType.AGENT_TYPE_WORKFLOW);
  return WorkflowDraft(draft);
}

void main() {
  test('a draft switched to Workflow drops LLM-only fields and, when switched '
      'back, the graph', () {
    final draft = AgentDraft(
      base: Agent(
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(model: 'fast', skills: ['skill']),
      ),
    );
    draft.setType(AgentType.AGENT_TYPE_WORKFLOW);
    final workflow = WorkflowDraft(draft);
    workflow.addNode(agentKind);
    workflow.setNodeAgent(0, 'writer');
    workflow.addNode(routerKind);
    expect(draft.build().childAgentIds, ['writer']);
    expect(draft.build().config.hasModel(), isFalse);
    expect(draft.build().config.skills, ['skill']);
    expect(draft.build().config.workflow.nodes, hasLength(2));
    draft.setType(AgentType.AGENT_TYPE_LLM);
    expect(draft.build().config.hasWorkflow(), isFalse);
    expect(draft.build().childAgentIds, isEmpty);
  });

  test('renaming or removing a node keeps edges pointing at real nodes', () {
    final wf = newWorkflow();
    final a = wf.addNode(agentKind);
    final r = wf.addNode(routerKind);
    final b = wf.addNode(agentKind);
    wf.addEdge(startNodeName, a.name);
    wf.addEdge(a.name, r.name);
    wf.addEdge(r.name, b.name);
    wf.renameNode(1, 'decide');
    expect(
      [for (final e in wf.config.edges) '${e.from}>${e.to}'],
      ['START>agent_1', 'agent_1>decide', 'decide>agent_2'],
    );
    wf.removeNode(1);
    expect(
      [for (final e in wf.config.edges) '${e.from}>${e.to}'],
      ['START>agent_1'],
    );
    expect(wf.nodes.map((n) => n.name), ['agent_1', 'agent_2']);
  });

  test('a route label and the default flag are mutually exclusive', () {
    final wf = newWorkflow();
    wf.addNode(routerKind);
    wf.addNode(agentKind);
    wf.addEdge('router_1', 'agent_1');
    wf.setEdgeRoute(0, 'approve');
    expect(wf.config.edges[0].route, 'approve');
    wf.setEdgeDefault(0, true);
    expect(wf.config.edges[0].isDefault, isTrue);
    expect(
      wf.config.edges[0].hasRoute() && wf.config.edges[0].route != '',
      isFalse,
    );
    wf.setEdgeRoute(0, 'reject');
    expect(wf.config.edges[0].isDefault, isFalse);
    expect(wf.config.edges[0].route, 'reject');
    wf.removeEdge(0);
    expect(wf.config.edges, isEmpty);
  });

  test('AGENT node references keep child_agent_ids in step, and only drop '
      'children the user unreferenced', () {
    final base = Agent(
      agentId: 'wf',
      type: AgentType.AGENT_TYPE_WORKFLOW,
      childAgentIds: ['legacy-child'],
      config: AgentConfig(workflow: WorkflowConfig()),
    );
    final draft = AgentDraft(base: base);
    final wf = WorkflowDraft(draft);
    wf.addNode(agentKind);
    wf.addNode(agentKind);
    wf.setNodeAgent(0, 'writer');
    wf.setNodeAgent(1, 'writer');
    expect(draft.build().childAgentIds, ['legacy-child', 'writer']);
    wf.setNodeAgent(1, 'editor');
    expect(draft.build().childAgentIds, ['legacy-child', 'writer', 'editor']);
    wf.removeNode(0);
    expect(draft.build().childAgentIds, ['legacy-child', 'editor']);
    wf.setNodeKind(0, routerKind);
    expect(draft.build().childAgentIds, ['legacy-child']);
    expect(wf.nodes[0].hasAgentId(), isFalse);
  });

  test('changing a node kind clears only fields the new kind rejects', () {
    final wf = newWorkflow();
    final node = wf.addNode(agentKind);
    node
      ..parallelWorker = true
      ..timeoutSeconds = 30
      ..retry = WorkflowRetryConfig(maxAttempts: 3);
    node.mergeFromBuffer([0xd0, 0x07, 5]); // future node field 122
    wf.setNodeKind(0, humanKind);
    expect(node.parallelWorker, isFalse);
    expect(node.timeoutSeconds, 30);
    expect(node.retry.maxAttempts, 3);
    node.question = 'Approve?';
    node.form = HumanInputForm(title: 'T');
    wf.setNodeKind(0, routerKind);
    expect(node.hasForm(), isFalse);
    expect(node.question, 'Approve?');
    wf.setNodeKind(0, joinKind);
    expect(node.hasTimeoutSeconds(), isFalse);
    expect(node.hasRetry(), isFalse);
    expect(node.unknownFields.getField(122)?.varints.single.toInt(), 5);
  });

  group('validation mirrors the server rules and locates each error', () {
    // classify -> decide -(approve)-> approver, decide -(default)-> rejecter
    WorkflowDraft valid() {
      final wf = newWorkflow();
      wf.addNode(agentKind);
      wf.addNode(routerKind);
      wf.addNode(agentKind);
      wf.addNode(agentKind);
      wf.renameNode(0, 'classify');
      wf.renameNode(1, 'decide');
      wf.renameNode(2, 'approver');
      wf.renameNode(3, 'rejecter');
      wf.setNodeAgent(0, 'classify-agent');
      wf.setNodeAgent(2, 'approve-agent');
      wf.setNodeAgent(3, 'reject-agent');
      wf.addEdge(startNodeName, 'classify');
      wf.addEdge('classify', 'decide');
      wf.addEdge('decide', 'approver');
      wf.setEdgeRoute(2, 'approve');
      wf.addEdge('decide', 'rejecter');
      wf.setEdgeDefault(3, true);
      return wf;
    }

    const known = {'classify-agent', 'approve-agent', 'reject-agent'};

    test('a complete graph has no issues', () {
      expect(valid().validate(knownAgentIds: known), isEmpty);
    });

    void expectIssue(
      WorkflowDraft wf,
      WorkflowIssueKind kind, {
      int? node,
      int? edge,
    }) {
      final issues = wf.validate(knownAgentIds: known);
      final match = issues.where((i) => i.kind == kind).toList();
      expect(
        match,
        hasLength(1),
        reason: '$kind in ${[for (final i in issues) i.kind]}',
      );
      expect(match.single.nodeIndex, node);
      expect(match.single.edgeIndex, edge);
      expect(match.single.message, isNotEmpty);
    }

    test('graph-level rules', () {
      expectIssue(newWorkflow(), WorkflowIssueKind.noNodes);
      final noEntry = valid()..removeEdge(0);
      expectIssue(noEntry, WorkflowIssueKind.noEntry);
    });

    test('node rules point at the node', () {
      expectIssue(
        valid()..renameNode(0, ''),
        WorkflowIssueKind.nodeNameMissing,
        node: 0,
      );
      expectIssue(
        valid()..renameNode(0, 'START'),
        WorkflowIssueKind.nodeNameReserved,
        node: 0,
      );
      expectIssue(
        valid()..renameNode(3, 'approver'),
        WorkflowIssueKind.nodeNameDuplicate,
        node: 3,
      );
      expectIssue(
        valid()..setNodeAgent(0, ''),
        WorkflowIssueKind.agentMissing,
        node: 0,
      );
      expectIssue(
        valid()..setNodeAgent(0, 'gone'),
        WorkflowIssueKind.agentUnknown,
        node: 0,
      );
      final human = valid()..setNodeKind(3, humanKind);
      expectIssue(human, WorkflowIssueKind.questionMissing, node: 3);
      final unspecified = valid()
        ..setNodeKind(3, WorkflowNodeKind.WORKFLOW_NODE_KIND_UNSPECIFIED);
      expectIssue(unspecified, WorkflowIssueKind.kindUnsupported, node: 3);
      final noDefault = valid()..removeEdge(3);
      expectIssue(noDefault, WorkflowIssueKind.routerNeedsDefault, node: 1);
      final negative = valid();
      negative.nodes[0].timeoutSeconds = -1;
      expectIssue(negative, WorkflowIssueKind.negativeValue, node: 0);
    });

    test('edge rules point at the edge', () {
      expectIssue(
        valid()
          ..renameNode(0, 'x')
          ..edges[0].to = 'nowhere',
        WorkflowIssueKind.edgeTargetUnknown,
        edge: 0,
      );
      expectIssue(
        valid()..edges[1].from = 'nowhere',
        WorkflowIssueKind.edgeSourceUnknown,
        edge: 1,
      );
      final both = valid();
      both.edges[2].isDefault = true;
      expectIssue(both, WorkflowIssueKind.routeAndDefault, edge: 2);
      final folded = valid()
        ..addEdge('decide', 'rejecter')
        ..setEdgeRoute(4, '  APPROVE ');
      expectIssue(folded, WorkflowIssueKind.routeDuplicate, edge: 4);
      final join = valid()
        ..addNode(joinKind)
        ..addEdge('decide', 'join_1')
        ..setEdgeRoute(4, 'other');
      expectIssue(join, WorkflowIssueKind.routedIntoJoin, edge: 4);
    });

    test('references to missing Agents are only checked once choices are '
        'known', () {
      expect(valid().validate(), isEmpty);
    });
  });

  group('Human Input forms', () {
    const text = HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT;
    const choice =
        HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE;

    test('fields and options are added, reordered and removed in order', () {
      final wf = newWorkflow();
      wf.addNode(humanKind);
      wf.setForm(0, true);
      wf.addFormField(0, text);
      wf.addFormField(0, choice);
      wf.addFormField(0, text);
      final form = wf.nodes[0].form;
      expect(form.fields.map((f) => f.name), ['field_1', 'field_2', 'field_3']);
      wf.moveFormField(0, 2, 0);
      expect(form.fields.map((f) => f.name), ['field_3', 'field_1', 'field_2']);
      wf.addOption(0, 2);
      wf.addOption(0, 2);
      wf.moveOption(0, 2, 1, 0);
      expect(form.fields[2].options.map((o) => o.value), [
        'option_2',
        'option_1',
      ]);
      wf.removeOption(0, 2, 0);
      wf.removeFormField(0, 0);
      expect(form.fields.map((f) => f.name), ['field_1', 'field_2']);
      expect(form.fields[1].options.map((o) => o.value), ['option_1']);
    });

    test('switching a field type drops what the other type rejects and '
        'keeps the rest, including unknown fields', () {
      final wf = newWorkflow();
      wf.addNode(humanKind);
      wf.setForm(0, true);
      wf.addFormField(0, text);
      final field = wf.nodes[0].form.fields[0];
      field
        ..label = 'Reason'
        ..required = true
        ..maxLength = 200;
      field.mergeFromBuffer([0xd0, 0x07, 3]); // future field 122
      wf.setFieldType(0, 0, choice);
      expect(field.hasMaxLength(), isFalse);
      wf.addOption(0, 0);
      wf.setFieldType(0, 0, text);
      expect(field.options, isEmpty);
      expect(field.label, 'Reason');
      expect(field.required, isTrue);
      expect(field.unknownFields.getField(122)?.varints.single.toInt(), 3);
    });

    test('turning the form off keeps the question', () {
      final wf = newWorkflow();
      wf.addNode(humanKind);
      wf.nodes[0].question = 'Approve?';
      wf.setForm(0, true);
      expect(wf.nodes[0].hasForm(), isTrue);
      wf.setForm(0, false);
      expect(wf.nodes[0].hasForm(), isFalse);
      expect(wf.nodes[0].question, 'Approve?');
    });
  });

  group('Human Input form validation locates the field or option', () {
    const text = HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT;
    const choice =
        HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE;

    // ask(Human Input) with a required single choice and a bounded text.
    WorkflowDraft valid() {
      final wf = newWorkflow();
      wf.addNode(humanKind);
      wf.nodes[0].question = 'Deploy?';
      wf.addEdge(startNodeName, 'ask_1');
      wf.setForm(0, true);
      wf.addFormField(0, choice);
      wf.addFormField(0, text);
      final env = wf.nodes[0].form.fields[0]
        ..name = 'env'
        ..label = 'Environment'
        ..required = true;
      wf.addOption(0, 0);
      wf.addOption(0, 0);
      env.options[0]
        ..value = 'prod'
        ..label = 'Production';
      env.options[1]
        ..value = 'staging'
        ..label = 'Staging';
      wf.nodes[0].form.fields[1]
        ..label = 'Reason'
        ..maxLength = 200;
      return wf;
    }

    WorkflowIssue only(WorkflowDraft wf, WorkflowIssueKind kind) {
      final issues = wf.validate();
      final match = issues.where((i) => i.kind == kind).toList();
      expect(
        match,
        hasLength(1),
        reason: '${[for (final i in issues) i.kind]}',
      );
      expect(match.single.nodeIndex, 0);
      return match.single;
    }

    test('a complete form and a plain question both pass', () {
      final wf = valid();
      expect(wf.validate(), isEmpty);
      wf.setForm(0, false);
      expect(wf.validate(), isEmpty);
      // An empty form (no title, no fields) keeps the plain question.
      wf.setForm(0, true);
      expect(wf.validate(), isEmpty);
    });

    test('form-level rules', () {
      final noFields = valid();
      noFields.nodes[0].form.fields.clear();
      noFields.nodes[0].form.title = 'Only a title';
      expect(only(noFields, WorkflowIssueKind.formFields).fieldIndex, isNull);
      final many = valid();
      for (var i = 0; i < 18; i++) {
        many.addFormField(0, text);
      }
      expect(
        many.validate().where((i) => i.kind == WorkflowIssueKind.formFields),
        isEmpty,
      );
      many.addFormField(0, text);
      expect(only(many, WorkflowIssueKind.formFields).fieldIndex, isNull);
      final title = valid();
      title.nodes[0].form.title = 'x' * 201;
      expect(only(title, WorkflowIssueKind.formTitle).fieldIndex, isNull);
    });

    test('field rules point at the field', () {
      for (final bad in ['', '2 words', 'a-b', 'x' * 65]) {
        final wf = valid();
        wf.nodes[0].form.fields[1].name = bad;
        expect(
          only(wf, WorkflowIssueKind.fieldName).fieldIndex,
          1,
          reason: bad,
        );
      }
      final dup = valid();
      dup.nodes[0].form.fields[1].name = 'env';
      expect(only(dup, WorkflowIssueKind.fieldName).fieldIndex, 1);
      final label = valid();
      label.nodes[0].form.fields[1].label = '  ';
      expect(only(label, WorkflowIssueKind.fieldLabel).fieldIndex, 1);
      final hint = valid();
      hint.nodes[0].form.fields[0].hint = 'h' * 501;
      expect(only(hint, WorkflowIssueKind.fieldHint).fieldIndex, 0);
      final length = valid();
      length.nodes[0].form.fields[1].maxLength = 2001;
      expect(only(length, WorkflowIssueKind.fieldLength).fieldIndex, 1);
      final unspecified = valid();
      unspecified.nodes[0].form.fields[1].type =
          HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED;
      expect(only(unspecified, WorkflowIssueKind.fieldType).fieldIndex, 1);
    });

    test('option rules point at the option', () {
      final none = valid();
      none.nodes[0].form.fields[0].options.clear();
      expect(only(none, WorkflowIssueKind.fieldOptions).fieldIndex, 0);
      final dup = valid();
      dup.nodes[0].form.fields[0].options[1].value = 'prod';
      final issue = only(dup, WorkflowIssueKind.optionValue);
      expect((issue.fieldIndex, issue.optionIndex), (0, 1));
      final blank = valid();
      blank.nodes[0].form.fields[0].options[0].label = ' ';
      final label = only(blank, WorkflowIssueKind.optionLabel);
      expect((label.fieldIndex, label.optionIndex), (0, 0));
      final many = valid();
      for (var i = 0; i < 49; i++) {
        many.addOption(0, 0);
      }
      expect(
        many.validate().where((i) => i.kind == WorkflowIssueKind.fieldOptions),
        isNotEmpty,
      );
    });

    test('a form on any other node kind is refused', () {
      final wf = valid();
      wf.nodes[0].kind = routerKind;
      expect(only(wf, WorkflowIssueKind.formOnWrongKind).fieldIndex, isNull);
    });
  });

  group('graph layout is derived from the edit model', () {
    test('nodes sit in layers by distance from START; edits move them', () {
      final wf = newWorkflow();
      for (final kind in [
        agentKind,
        routerKind,
        agentKind,
        agentKind,
        joinKind,
      ]) {
        wf.addNode(kind);
      }
      // agent_1 > router_1 > {agent_2, agent_3} > join_1
      wf.addEdge(startNodeName, 'agent_1');
      wf.addEdge('agent_1', 'router_1');
      wf.addEdge('router_1', 'agent_2');
      wf.addEdge('router_1', 'agent_3');
      wf.addEdge('agent_2', 'join_1');
      wf.addEdge('agent_3', 'join_1');
      var layout = wf.layout();
      expect(layout.layers, [
        [0],
        [1],
        [2, 3],
        [4],
      ]);
      expect(layout.unreached, isEmpty);
      // A shortcut edge pulls the join up; a node nobody reaches is listed.
      wf.addEdge('agent_1', 'join_1');
      wf.addNode(agentKind);
      layout = wf.layout();
      expect(layout.layers.first, [0]);
      expect(layout.layers[1], containsAll([1, 4]));
      expect(layout.unreached, [5]);
    });

    test('a cycle does not loop the layout and a broken edge is ignored', () {
      final wf = newWorkflow();
      wf.addNode(agentKind);
      wf.addNode(agentKind);
      wf.addEdge(startNodeName, 'agent_1');
      wf.addEdge('agent_1', 'agent_2');
      wf.addEdge('agent_2', 'agent_1');
      wf.addEdge('agent_2', 'gone');
      expect(wf.layout().layers, [
        [0],
        [1],
      ]);
    });
  });

  test('an unrelated edit keeps unknown kinds, unknown fields and the legacy '
      'agent reference on the wire', () {
    final config = WorkflowConfig(
      nodes: [
        WorkflowNode(name: 'known', kind: agentKind, agentId: 'a')
          ..mergeFromBuffer([0xd0, 0x07, 4]), // future node field 122
        WorkflowNode(name: 'future'), // kind 99, below
        // ignore: deprecated_member_use
        WorkflowNode(name: 'legacy', kind: agentKind, agent: 'by-name'),
      ],
      edges: [
        WorkflowEdge(from: 'START', to: 'known')
          ..mergeFromBuffer([0xd0, 0x07, 6]),
      ],
    );
    config.nodes[1].mergeFromBuffer([0x10, 99]);
    final base = Agent(
      agentId: 'wf',
      type: AgentType.AGENT_TYPE_WORKFLOW,
      childAgentIds: ['a'],
      config: AgentConfig(workflow: config),
    );
    final draft = AgentDraft(base: base);
    final wf = WorkflowDraft(draft);
    wf.nodes[0].timeoutSeconds = 12;
    wf.setRetry(0, true);
    expect(wf.nodes[0].hasRetry(), isTrue);
    wf.setRetry(0, false);
    wf.nodes[0].clearTimeoutSeconds();
    final saved = Agent.fromBuffer(draft.build().writeToBuffer());
    expect(saved.config.workflow.writeToBuffer(), config.writeToBuffer());
    expect(
      saved.config.workflow.nodes[1].unknownFields
          .getField(2)
          ?.varints
          .single
          .toInt(),
      99,
    );
    final kinds = [for (final i in wf.validate()) i.kind];
    expect(kinds, contains(WorkflowIssueKind.kindUnsupported));
    expect(kinds, contains(WorkflowIssueKind.agentMissing));
    // Validation never rewrites the draft.
    expect(
      draft.build().config.workflow.writeToBuffer(),
      config.writeToBuffer(),
    );
  });

  test('choosing a kind or field type replaces an unknown enum value '
      'instead of writing both', () {
    final wf = newWorkflow();
    wf.addNode(humanKind);
    wf.nodes[0].mergeFromBuffer([0x10, 99]);
    wf.setForm(0, true);
    wf.addFormField(
      0,
      HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
    );
    final field = wf.nodes[0].form.fields[0];
    field.mergeFromBuffer([0x28, 77]);
    expect(unknownKind(wf.nodes[0]), isNotNull);
    expect(unknownFieldType(field), isNotNull);
    wf.setFieldType(
      0,
      0,
      HumanInputFormFieldType.HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
    );
    final choice = WorkflowNode.fromBuffer(wf.nodes[0].writeToBuffer());
    expect(unknownFieldType(choice.form.fields[0]), isNull);
    wf.setNodeKind(0, routerKind);
    final saved = WorkflowNode.fromBuffer(wf.nodes[0].writeToBuffer());
    expect(unknownKind(saved), isNull);
    expect(saved.kind, routerKind);
  });
}
