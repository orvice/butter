import 'dart:convert';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import 'agent_draft.dart';
import 'associations.dart';
import 'box_leaf_fields.dart';
import 'structured_fields.dart';
import 'workflow_draft.dart';
import 'workflow_editor.dart';
import '../memory/agent_memory_config.dart';
import '../memory/agent_memory_fields.dart';
import '../memory/agent_result_card_fields.dart';
import '../memory/agent_result_cards.dart';

class AgentFormPage extends StatefulWidget {
  const AgentFormPage({super.key, required this.client, this.base});
  final ClientContext client;
  final Agent? base;
  @override
  State<AgentFormPage> createState() => _AgentFormPageState();
}

class _AgentFormPageState extends State<AgentFormPage> {
  final _form = GlobalKey<FormState>();
  late final _draft = AgentDraft(base: widget.base);
  Agent get _agent => _draft.agent;
  late final _workflow = WorkflowDraft(_draft);
  bool _dirty = false;
  bool _saved = false;
  bool _showIssues = false;
  late final _model = TextEditingController(text: _agent.config.model);
  late final AgentMemoryFormValues _memoryValues = memoryFormFromConfig(
    _agent.hasConfig() && _agent.config.hasMemory()
        ? _agent.config.memory
        : null,
  );
  late final AgentResultCardsFormValues _resultCardsValues =
      resultCardsFromConfig(
        _agent.hasConfig() && _agent.config.hasResultCards()
            ? _agent.config.resultCards
            : null,
      );

  @override
  void dispose() {
    _model.dispose();
    super.dispose();
  }

  void _change(VoidCallback change) => setState(() {
    change();
    _dirty = true;
  });

  Future<void> _save() async {
    final fieldsOk = _form.currentState!.validate();
    final graphOk =
        _agent.type != AgentType.AGENT_TYPE_WORKFLOW ||
        workflowIssues(
          _workflow,
          choices: widget.client.agentChoices,
          choicesReady: widget.client.agentEditorReady,
          selfId: _agent.agentId,
        ).isEmpty;
    if (!fieldsOk || !graphOk) {
      // Nothing is sent: list every problem where the user can fix it.
      setState(() => _showIssues = true);
      return;
    }
    FocusScope.of(context).unfocus();
    applyMemoryForm(_draft, _memoryValues);
    applyResultCardsForm(_draft, _resultCardsValues);
    final saved = await widget.client.saveAgent(
      widget.base == null ? _draft.buildForCreate() : _draft.build(),
      create: widget.base == null,
      previous: widget.base,
    );
    if (!mounted || !saved) return;
    setState(() {
      _saved = true;
      _dirty = false;
    });
    if (Navigator.canPop(context)) Navigator.pop(context);
  }

  Widget _text(
    String key,
    String label,
    String value,
    void Function(String) onChanged, {
    int lines = 1,
    String? helper,
    String? Function(String?)? validator,
    bool enabled = true,
    TextEditingController? controller,
  }) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: TextFormField(
      key: Key(key),
      initialValue: controller == null ? value : null,
      controller: controller,
      enabled: enabled && !widget.client.busy && !_saved,
      maxLines: lines,
      autocorrect: false,
      decoration: InputDecoration(
        labelText: label,
        helperText: helper,
        helperMaxLines: 5,
      ),
      validator: validator,
      onChanged: (text) => _change(() => onChanged(text)),
    ),
  );

  Widget _toggle(String label, bool value, void Function(bool) change) =>
      SwitchListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(label),
        value: value,
        onChanged: widget.client.busy || _saved
            ? null
            : (v) => _change(() => change(v)),
      );

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: widget.client,
    builder: (context, _) {
      final client = widget.client;
      final supported = canEditAgent(_agent);
      return DraftGuard(
        dirty: _dirty,
        busy: client.busy,
        child: ManagementPage(
          client: client,
          title: widget.base == null ? 'Create Agent' : 'Edit Agent',
          child: !supported
              ? const Center(
                  child: Text(
                    'This Agent cannot be safely edited here (unsupported type, blocked lifecycle, legacy embedded children, or a Pi/Cursor leaf that holds fields a leaf cannot have). Use the Web dashboard.',
                  ),
                )
              : Form(
                  key: _form,
                  // Keep every field mounted: validation must include invalid
                  // fields scrolled off-screen, not just visible lazy rows.
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (!client.agentEditorReady)
                          TextButton(
                            onPressed: client.busy
                                ? null
                                : client.prepareAgentEditor,
                            child: const Text('Load editor choices'),
                          ),
                        if (client.agentContentGitOwned ||
                            (widget.base == null &&
                                client.agentRepositoryBound))
                          const Text(
                            'Repository-bound Workspace: prompt and description save through Agent Content. Owner/admin permission and the observed Agent version are checked by the server.',
                          ),
                        _text(
                          'agent-id',
                          'Agent ID',
                          _agent.agentId,
                          (v) => _agent.agentId = v.trim(),
                          enabled: widget.base == null,
                          helper:
                              'Immutable ID; references never use the display name.',
                          validator: (v) =>
                              v == null ||
                                  !RegExp(
                                    r'^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$',
                                  ).hasMatch(v.trim()) ||
                                  v.trim().length > 64
                              ? 'Use 1–64 lowercase letters, digits or internal hyphens.'
                              : null,
                        ),
                        _text(
                          'agent-name',
                          'Runtime name',
                          _agent.name,
                          (v) => _agent.name = v.trim(),
                          enabled: widget.base == null,
                          helper: 'Server-controlled after creation.',
                          validator: (v) =>
                              v == null ||
                                  v.trim().isEmpty ||
                                  v.trim() == 'user'
                              ? 'Enter a runtime name other than user.'
                              : null,
                        ),
                        _text(
                          'display-name',
                          'Display name',
                          _agent.displayName,
                          (v) => _agent.displayName = v,
                        ),
                        _text(
                          'description',
                          'Description',
                          _agent.description,
                          (v) => _agent.description = v,
                          lines: 3,
                        ),
                        DropdownButtonFormField<AgentType>(
                          key: const Key('agent-type'),
                          initialValue: _agent.type,
                          isExpanded: true,
                          decoration: const InputDecoration(
                            labelText: 'Agent type',
                          ),
                          items: [
                            for (final type in {
                              if (_agent.type ==
                                  AgentType.AGENT_TYPE_UNSPECIFIED)
                                _agent.type,
                              ...editableAgentTypes,
                            })
                              DropdownMenuItem(
                                value: type,
                                child: Text(
                                  type.name.replaceFirst('AGENT_TYPE_', ''),
                                ),
                              ),
                          ],
                          // Existing resources keep their type here; changing type can discard
                          // future type-specific fields. Create selects any supported type.
                          onChanged:
                              client.busy || widget.base != null || _saved
                              ? null
                              : (v) {
                                  if (v != null) {
                                    _change(() {
                                      _draft.setType(v);
                                      _model.text = _agent.config.model;
                                    });
                                  }
                                },
                        ),
                        if (isLLM(_agent)) ..._llmFields(),
                        if (_agent.type == AgentType.AGENT_TYPE_LOOP)
                          _text(
                            'max-iterations',
                            'Max iterations',
                            '${_agent.config.maxIterations}',
                            (v) {
                              final n = parseWholeNumber(v, 4294967295);
                              if (n != null) _draft.config.maxIterations = n;
                            },
                            helper: '0 runs until escalation or termination.',
                            validator: (v) =>
                                validateWholeNumber(v, 4294967295),
                          ),
                        if (_agent.type == AgentType.AGENT_TYPE_WORKFLOW)
                          WorkflowEditor(
                            workflow: _workflow,
                            selfId: _agent.agentId,
                            choices: client.agentChoices,
                            choicesReady: client.agentEditorReady,
                            enabled: !client.busy && !_saved,
                            showSummary: _showIssues,
                            onChanged: () => _change(() {}),
                          )
                        else if (isBoxLeaf(_agent))
                          if (client.boxes case final boxes?)
                            BoxLeafFields(
                              key: ValueKey(_agent.type),
                              draft: _draft,
                              boxes: boxes,
                              enabled: !client.busy && !_saved,
                              onChanged: () => _change(() {}),
                            )
                          else
                            const Text('ButterBoxes are unavailable.')
                        else ...[
                          if (!isLLM(_agent))
                            const Padding(
                              padding: EdgeInsets.symmetric(vertical: 12),
                              child: Text(
                                'Children remain independent resources. Sequential and Loop execute in the listed order; Parallel runs children concurrently. ContextGuard belongs on each child LLM.',
                              ),
                            ),
                          ChildAgentFields(
                            agent: _agent,
                            choices: client.agentChoices,
                            enabled:
                                !client.busy &&
                                !_saved &&
                                client.agentEditorReady,
                            onChanged: () => _change(() {}),
                          ),
                        ],
                        StringMapFields(
                          title: 'Labels',
                          values: _agent.labels,
                          enabled: !client.busy && !_saved,
                          onChanged: () => _change(() {}),
                        ),
                        StringMapFields(
                          title: 'Metadata',
                          values: _agent.metadata,
                          enabled: !client.busy && !_saved,
                          onChanged: () => _change(() {}),
                        ),
                        _toggle(
                          'Expose A2A',
                          _agent.enableA2a,
                          (v) => _agent.enableA2a = v,
                        ),
                        _toggle(
                          'Expose OpenAI-compatible API',
                          _agent.enableOpenaiApi,
                          (v) => _agent.enableOpenaiApi = v,
                        ),
                        _toggle(
                          'Expose AG-UI',
                          _agent.enableAgui,
                          (v) => _agent.enableAgui = v,
                        ),
                        AgentMemoryFields(
                          agent: _agent,
                          values: _memoryValues,
                          workspaceMemoryReady:
                              widget.client.memory?.configEnabled ?? false,
                          enabled: !client.busy && !_saved,
                          onChanged: () => _change(() {}),
                        ),
                        AgentResultCardFields(
                          agent: _agent,
                          values: _resultCardsValues,
                          enabled: !client.busy && !_saved,
                          onChanged: () => _change(() {}),
                        ),
                        ..._associations(),
                        const Text(
                          'Other associations and configuration are retained unchanged. No Agent run or lifecycle action is performed by Save.',
                        ),
                        const SizedBox(height: 16),
                        FilledButton(
                          key: const Key('save-agent'),
                          onPressed:
                              client.busy ||
                                  !client.agentEditorReady ||
                                  _saved ||
                                  client.writeNotice != null
                              ? null
                              : _save,
                          child: Text(_saved ? 'Saved' : 'Save Agent'),
                        ),
                        if (client.writeNotice != null)
                          const Text(
                            'Leave this draft and reload the Agent to reconcile the outcome before editing again.',
                          ),
                      ],
                    ),
                  ),
                ),
        ),
      );
    },
  );

  /// MCP servers, Remote Agents, file mounts and Skills, for the Agent types
  /// whose runtime uses them. For a type that ignores what the draft holds,
  /// the entries stay visible with removal only, because a missing id would
  /// still make the server refuse the save.
  List<Widget> _associations() {
    final client = widget.client;
    if (isBoxLeaf(_agent)) return const [];
    final enabled = !client.busy && !_saved && client.agentEditorReady;
    final config = _agent.config;
    void touch() => _change(() {});
    final showRetainedMcp =
        !offersMcp(_agent) && config.mcpServerIds.isNotEmpty;
    final showRetainedRemote =
        widget.base != null &&
        !offersRemote(_agent) &&
        config.remoteAgentIds.isNotEmpty;
    final showRetainedMounts =
        widget.base != null &&
        !offersMounts(_agent) &&
        config.fileMounts.isNotEmpty;
    final showRetainedSkills =
        widget.base != null &&
        !offersSkills(_agent) &&
        config.skills.isNotEmpty;
    return [
      if (offersMcp(_agent) || showRetainedMcp)
        ReferenceChecklist(
          keyPrefix: 'mcp',
          title: 'MCP servers',
          help:
              'Tools from the selected existing servers are offered to this Agent. Server setup stays under MCP servers.',
          emptyText: 'No MCP servers in this Workspace.',
          selected: config.mcpServerIds,
          choices: client.agentMcp,
          enabled: enabled,
          addable: offersMcp(_agent),
          missingNote:
              'Not found in this Workspace. The server refuses to save a missing MCP server; remove it.',
          onChanged: (ids) => _change(() {
            _draft.config.mcpServerIds
              ..clear()
              ..addAll(ids);
          }),
        ),
      if (offersRemote(_agent) || showRetainedRemote)
        ReferenceChecklist(
          keyPrefix: 'remote',
          title: 'Remote Agents',
          help:
              'Selected existing Remote Agents join this Agent as sub-agents.',
          emptyText: 'No Remote Agents in this Workspace.',
          selected: config.remoteAgentIds,
          choices: client.agentRemote,
          enabled: enabled,
          addable: offersRemote(_agent),
          missingNote:
              'Not found in this Workspace. The server refuses to save a missing Remote Agent; remove it.',
          onChanged: (ids) => _change(() {
            _draft.config.remoteAgentIds
              ..clear()
              ..addAll(ids);
          }),
        ),
      if (offersMounts(_agent) || showRetainedMounts)
        FileMountFields(
          mounts: config.fileMounts,
          spaces: client.agentSpaces,
          enabled: enabled,
          addable: offersMounts(_agent),
          onChanged: touch,
        ),
      if (offersSkills(_agent) || showRetainedSkills)
        ReferenceChecklist(
          keyPrefix: 'skill',
          title: 'Existing Skills',
          help:
              'Associate existing Workspace Skills. Skill creation and content editing remain in Web.',
          emptyText: 'No existing Skills in this Workspace.',
          selected: config.skills,
          choices: client.agentSkills,
          enabled: enabled,
          addable: offersSkills(_agent),
          missingNote:
              'Not in the current Workspace catalog; retained until removed.',
          onChanged: (names) => _change(() {
            _draft.config.skills
              ..clear()
              ..addAll(names);
          }),
        ),
    ];
  }

  List<Widget> _llmFields() {
    final config = _agent.config;
    final unknownContents = unknownEnum(config, 25);
    return [
      _text(
        'model',
        'Model / alias',
        config.model,
        (v) => _draft.config.model = v.trim(),
        controller: _model,
        helper:
            'Uses the model alias (or model ID when no alias is configured). Existing or custom values are kept.',
      ),
      if (widget.client.agentModelChoices.isNotEmpty)
        DropdownButtonFormField<String>(
          key: ValueKey('model-choice-${config.model}'),
          initialValue: widget.client.agentModelChoices.contains(config.model)
              ? config.model
              : null,
          isExpanded: true,
          decoration: const InputDecoration(
            labelText: 'Configured model choices',
          ),
          items: [
            for (final model in widget.client.agentModelChoices)
              DropdownMenuItem(value: model, child: Text(model)),
          ],
          onChanged: widget.client.busy || _saved
              ? null
              : (v) {
                  if (v != null) {
                    _change(() {
                      _draft.config.model = v;
                      _model.text = v;
                    });
                  }
                },
        ),
      _text(
        'instruction',
        'Instruction',
        config.instruction,
        (v) => _draft.config.instruction = v,
        lines: 5,
      ),
      _text(
        'global-instruction',
        'Global instruction',
        config.globalInstruction,
        (v) => _draft.config.globalInstruction = v,
        lines: 4,
      ),
      _text(
        'input-schema',
        'Input schema (JSON)',
        config.inputSchemaJson,
        (v) => _draft.config.inputSchemaJson = v,
        lines: 3,
        validator: validateSchema,
      ),
      _text(
        'output-schema',
        'Output schema (JSON)',
        config.outputSchemaJson,
        (v) => _draft.config.outputSchemaJson = v,
        lines: 3,
        validator: validateSchema,
        helper:
            'Stored configuration only. Saving a schema does not mean the runtime or provider enforces structured output.',
      ),
      _text(
        'output-key',
        'Output key',
        config.outputKey,
        (v) => _draft.config.outputKey = v,
      ),
      if (unknownContents != null)
        Text(
          'Include contents: unknown value $unknownContents retained unchanged',
        )
      else
        DropdownButtonFormField<LLMIncludeContents>(
          key: const Key('include-contents'),
          initialValue: config.includeContents,
          isExpanded: true,
          decoration: const InputDecoration(
            labelText: 'Include conversation contents',
          ),
          items: [
            for (final value in LLMIncludeContents.values)
              DropdownMenuItem(
                value: value,
                child: Text(
                  value.name.replaceFirst('LLM_INCLUDE_CONTENTS_', ''),
                ),
              ),
          ],
          onChanged: widget.client.busy || _saved
              ? null
              : (v) {
                  if (v != null) {
                    _change(() => _draft.config.includeContents = v);
                  }
                },
        ),
      _toggle(
        'Disallow transfer to parent',
        config.disallowTransferToParent,
        (v) => _draft.config.disallowTransferToParent = v,
      ),
      _toggle(
        'Disallow transfer to peers',
        config.disallowTransferToPeers,
        (v) => _draft.config.disallowTransferToPeers = v,
      ),
      ..._guardFields(),
    ];
  }

  List<Widget> _guardFields() {
    final config = _agent.config;
    final guard = config.contextGuard;
    final unknown = unknownEnum(guard, 1);
    final mode = !config.hasContextGuard()
        ? GuardMode.off
        : switch (guard.strategy) {
            ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_THRESHOLD =>
              GuardMode.threshold,
            ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW =>
              GuardMode.slidingWindow,
            _ => GuardMode.retained,
          };
    return [
      const SizedBox(height: 12),
      if (unknown != null)
        Text('ContextGuard: unknown strategy $unknown retained unchanged')
      else
        DropdownButtonFormField<GuardMode>(
          key: const Key('guard-mode'),
          initialValue: mode,
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'ContextGuard'),
          items: [
            if (mode == GuardMode.retained)
              const DropdownMenuItem(
                value: GuardMode.retained,
                child: Text('Unspecified (retained)'),
              ),
            const DropdownMenuItem(value: GuardMode.off, child: Text('Off')),
            const DropdownMenuItem(
              value: GuardMode.threshold,
              child: Text('Threshold'),
            ),
            const DropdownMenuItem(
              value: GuardMode.slidingWindow,
              child: Text('Sliding Window'),
            ),
          ],
          onChanged: widget.client.busy || _saved
              ? null
              : (v) {
                  if (v != null && v != GuardMode.retained) {
                    _change(() => _draft.setGuard(v));
                  }
                },
        ),
      if (unknown == null && mode == GuardMode.threshold)
        _text(
          'guard-tokens',
          'Agent Context Override (input tokens)',
          '${guard.maxTokens}',
          (v) {
            final n = parseWholeNumber(v, 2147483647);
            if (n != null) _draft.config.contextGuard.maxTokens = n;
          },
          helper:
              '0 inherits Model Context Capacity or embedded model metadata. This is not maximum output tokens.',
          validator: (v) => validateWholeNumber(v, 2147483647),
        ),
      if (unknown == null && mode == GuardMode.slidingWindow)
        _text(
          'guard-turns',
          'Sliding Window content-entry limit',
          '${guard.maxTurns}',
          (v) {
            final n = parseWholeNumber(v, 2147483647);
            if (n != null) _draft.config.contextGuard.maxTurns = n;
          },
          helper:
              '0 uses the backend default of 20 entries; no input-token override applies.',
          validator: (v) => validateWholeNumber(v, 2147483647),
        ),
    ];
  }
}

int? parseWholeNumber(String? value, int maximum) {
  final text = value?.trim() ?? '';
  if (text.isEmpty) return 0;
  if (!RegExp(r'^\d+$').hasMatch(text)) return null;
  final number = int.tryParse(text);
  return number == null || number > maximum ? null : number;
}

String? validateWholeNumber(String? value, int maximum) =>
    parseWholeNumber(value, maximum) == null
    ? 'Use a whole number from 0 to $maximum.'
    : null;
String? validateSchema(String? value) {
  if (value == null || value.trim().isEmpty) return null;
  try {
    if (jsonDecode(value) is Map<String, dynamic>) return null;
  } catch (_) {
    /* Form error only. */
  }
  return 'Enter a JSON schema object, or leave this empty.';
}
