import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/butterbox.pb.dart';
import '../butterboxes/box_controller.dart';
import 'agent_draft.dart';
import 'leaf_draft.dart';

const _continuityNote =
    'Each conversation keeps one session on the box. Pointing this Agent at another ButterBox or working directory starts a fresh session on the next message; earlier sessions are never moved or deleted on the box. Model, provider, thinking level and mode apply to sessions created after the change.';

/// The structured configuration of a PI or CURSOR leaf: which ButterBox runs
/// it, the working directory, the model and the run limit. A leaf's
/// instructions, tools, skills, memory and cards live on the box, so none of
/// those fields exist here. The phone only talks to the Butter server, which
/// lists the boxes and asks each box for its catalog.
class BoxLeafFields extends StatefulWidget {
  const BoxLeafFields({
    super.key,
    required this.draft,
    required this.boxes,
    required this.enabled,
    required this.onChanged,
  });
  final AgentDraft draft;
  final ButterBoxController boxes;
  final bool enabled;
  final VoidCallback onChanged;

  @override
  State<BoxLeafFields> createState() => _BoxLeafFieldsState();
}

class _BoxLeafFieldsState extends State<BoxLeafFields> {
  late final _dir = TextEditingController();
  late final _model = TextEditingController();
  late final _provider = TextEditingController();
  late final _seconds = TextEditingController();
  late RunLimitMode _limit;

  bool get _pi => widget.draft.agent.type == AgentType.AGENT_TYPE_PI;
  PiAgentConfig get _piConfig => widget.draft.config.ensurePi();
  CursorAgentConfig get _cursorConfig => widget.draft.config.ensureCursor();

  @override
  void initState() {
    super.initState();
    if (_pi) {
      final c = _piConfig;
      _dir.text = c.workingDir;
      _model.text = c.model;
      _provider.text = c.provider;
      _limit = piRunLimit(c);
      if (_limit == RunLimitMode.seconds) _seconds.text = '${c.maxRunSeconds}';
    } else {
      final c = _cursorConfig;
      _dir.text = c.workingDir;
      _model.text = c.model;
      _limit = cursorRunLimit(c);
      if (_limit == RunLimitMode.seconds) _seconds.text = '${c.maxRunSeconds}';
    }
    if (widget.boxes.listState == BoxLoad.idle) widget.boxes.loadBoxes();
  }

  @override
  void dispose() {
    _dir.dispose();
    _model.dispose();
    _provider.dispose();
    _seconds.dispose();
    super.dispose();
  }

  String get _boxId => _pi ? _piConfig.butterboxId : _cursorConfig.butterboxId;

  void _changed(VoidCallback change) {
    setState(change);
    widget.onChanged();
  }

  void _pickBox(String id) => _changed(() {
    if (_pi) {
      _piConfig
        ..butterboxId = id
        ..clearProvider()
        ..clearModel();
    } else {
      _cursorConfig
        ..butterboxId = id
        ..clearModel();
    }
    // A model chosen on another box would not be honest here.
    _model.clear();
    _provider.clear();
  });

  void _applyLimit() {
    final n = parseRunSeconds(_seconds.text);
    // A seconds choice without a valid number is not applied; the form
    // refuses to save it instead.
    final mode = _limit == RunLimitMode.seconds && n == null
        ? RunLimitMode.serverDefault
        : _limit;
    if (_pi) {
      final c = _piConfig;
      applyRunLimit(
        mode,
        n ?? 0,
        clear: c.clearMaxRunSeconds,
        set: (v) => c.maxRunSeconds = v,
      );
    } else {
      final c = _cursorConfig;
      applyRunLimit(
        mode,
        n ?? 0,
        clear: c.clearMaxRunSeconds,
        set: (v) => c.maxRunSeconds = v,
      );
    }
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: widget.boxes,
    builder: (context, _) {
      final boxes = widget.boxes;
      final id = _boxId;
      final known = boxes.byId(id);
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 12),
          Text(
            _pi ? 'Pi configuration' : 'Cursor configuration',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 4),
          Text(
            _pi
                ? 'Butter chooses where this Agent runs. Pi manages its tools and instructions on the box, from the working directory\'s AGENTS.md and .pi configuration, so none are set here.'
                : 'Butter chooses where this Agent runs. Cursor reads its rules and tools on the box, from .cursor/rules and mcp.json, and the Cursor API key lives in the box\'s CURSOR_API_KEY, so none are set here.',
          ),
          const SizedBox(height: 12),
          _boxPicker(boxes, id, known),
          _textField(
            'leaf-working-dir',
            'Working directory',
            _dir,
            (v) =>
                _pi ? _piConfig.workingDir = v : _cursorConfig.workingDir = v,
            helper:
                'Absolute, or relative to the box\'s sandbox root. Empty uses the box server\'s working directory.',
          ),
          const SizedBox(height: 12),
          _loadModelsRow(boxes, id),
          if (_pi) ..._piFields(boxes, id) else ..._cursorFields(boxes, id),
          const SizedBox(height: 12),
          _runLimitFields(),
          const SizedBox(height: 12),
          const Text(_continuityNote),
        ],
      );
    },
  );

  Widget _textField(
    String key,
    String label,
    TextEditingController controller,
    void Function(String) write, {
    String? helper,
  }) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: TextFormField(
      key: Key(key),
      controller: controller,
      enabled: widget.enabled,
      autocorrect: false,
      enableSuggestions: false,
      decoration: InputDecoration(
        labelText: label,
        helperText: helper,
        helperMaxLines: 4,
      ),
      onChanged: (text) => _changed(() => write(text.trim())),
    ),
  );

  Widget _boxPicker(ButterBoxController boxes, String id, ButterBox? known) {
    final unknownCurrent = id.isNotEmpty && known == null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        DropdownButtonFormField<String>(
          key: ValueKey('leaf-box-$id-${boxes.boxes.length}'),
          initialValue: id.isEmpty ? null : id,
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'ButterBox'),
          items: [
            for (final b in boxes.boxes)
              DropdownMenuItem(
                value: b.id,
                // A disabled box cannot take a new binding; an existing
                // binding to it stays valid.
                enabled: b.enabled || b.id == id,
                child: Text(b.enabled ? b.name : '${b.name} (disabled)'),
              ),
            if (unknownCurrent)
              DropdownMenuItem(value: id, child: Text('Unavailable box ($id)')),
          ],
          validator: (v) =>
              v == null || v.isEmpty ? 'Select a ButterBox.' : null,
          onChanged: widget.enabled
              ? (v) {
                  if (v != null && v != id) _pickBox(v);
                }
              : null,
        ),
        if (boxes.listState == BoxLoad.loading)
          const Padding(
            padding: EdgeInsets.only(top: 4),
            child: Text('Loading ButterBoxes…'),
          ),
        if (boxes.listState == BoxLoad.failed)
          Row(
            children: [
              Expanded(
                child: Text(
                  boxes.listError ?? 'ButterBoxes could not be loaded.',
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
              TextButton(
                onPressed: boxes.loadBoxes,
                child: const Text('Retry'),
              ),
            ],
          ),
        if (boxes.listState == BoxLoad.ready && boxes.boxes.isEmpty)
          const Padding(
            padding: EdgeInsets.only(top: 4),
            child: Text(
              'No ButterBoxes are registered in this Workspace. Register one first.',
            ),
          ),
        if (unknownCurrent && boxes.listState == BoxLoad.ready)
          const Padding(
            padding: EdgeInsets.only(top: 4),
            child: Text(
              'The ButterBox this Agent points at is not in this Workspace. Choose another to repoint it, or leave it unchanged.',
            ),
          ),
      ],
    );
  }

  Widget _loadModelsRow(ButterBoxController boxes, String id) {
    final catalog = _pi ? boxes.piCatalogs[id] : boxes.cursorCatalogs[id];
    final loading = catalog?.state == BoxLoad.loading;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Align(
          alignment: Alignment.centerLeft,
          child: OutlinedButton.icon(
            key: const Key('leaf-load-models'),
            onPressed: !widget.enabled || id.isEmpty || loading
                ? null
                : () =>
                      _pi ? boxes.loadPiModels(id) : boxes.loadCursorModels(id),
            icon: const Icon(Icons.cloud_download_outlined),
            label: Text(
              catalog?.state == BoxLoad.ready
                  ? 'Reload models from this box'
                  : 'Load models from this box',
            ),
          ),
        ),
        if (loading)
          Semantics(
            liveRegion: true,
            child: Text(
              _pi
                  ? 'Asking the box for its models…'
                  : 'Asking the box for its Cursor models. This can take up to a minute…',
            ),
          ),
        if (catalog?.state == BoxLoad.failed)
          Semantics(
            liveRegion: true,
            child: Text(
              catalog!.error ?? 'The model catalog could not be loaded.',
              key: const Key('leaf-catalog-error'),
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
        if (catalog?.state == BoxLoad.ready && catalog!.models.isEmpty)
          const Text('The box reported no models. Enter one by hand.'),
      ],
    );
  }

  List<Widget> _piFields(ButterBoxController boxes, String id) {
    final catalog = boxes.piCatalogs[id];
    final models = catalog?.state == BoxLoad.ready
        ? catalog!.models
        : const <ButterBoxModel>[];
    final c = _piConfig;
    ButterBoxModel? selected;
    for (final m in models) {
      if (m.id == c.model && m.provider == c.provider) selected = m;
    }
    final level = c.thinkingLevel;
    final retainedLevel = level.isNotEmpty && !piThinkingLevels.contains(level);
    return [
      if (models.isNotEmpty)
        DropdownButtonFormField<String>(
          key: ValueKey('leaf-pi-catalog-${selected?.id}-${models.length}'),
          initialValue: selected == null
              ? null
              : '${selected.provider}\u0000${selected.id}',
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'Model from catalog'),
          items: [
            for (final m in models)
              DropdownMenuItem(
                value: '${m.provider}\u0000${m.id}',
                child: Text(
                  [
                    if (m.provider.isNotEmpty) m.provider,
                    m.name.isEmpty || m.name == m.id
                        ? m.id
                        : '${m.name} (${m.id})',
                  ].join(' · '),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
          ],
          onChanged: widget.enabled
              ? (v) {
                  final m = models.firstWhere(
                    (m) => '${m.provider}\u0000${m.id}' == v,
                  );
                  _changed(() {
                    c
                      ..provider = m.provider
                      ..model = m.id;
                    _provider.text = m.provider;
                    _model.text = m.id;
                  });
                }
              : null,
        ),
      _textField(
        'leaf-provider',
        'Provider',
        _provider,
        (v) => c.provider = v,
        helper: 'Optional. Empty uses pi\'s default on the box.',
      ),
      _textField(
        'leaf-model',
        'Model',
        _model,
        (v) => c.model = v,
        helper:
            'Optional model ID or pattern as pi accepts it. Pick from the catalog, or enter it by hand.',
      ),
      DropdownButtonFormField<String>(
        key: ValueKey('leaf-thinking-$level'),
        initialValue: level,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Thinking level'),
        items: [
          const DropdownMenuItem(value: '', child: Text('Default')),
          for (final l in piThinkingLevels)
            DropdownMenuItem(value: l, child: Text(l)),
          if (retainedLevel)
            DropdownMenuItem(value: level, child: Text('$level (retained)')),
        ],
        onChanged: widget.enabled
            ? (v) {
                if (v != null) _changed(() => c.thinkingLevel = v);
              }
            : null,
      ),
      if (selected != null && !selected.reasoning && level.isNotEmpty)
        const Text(
          'This model does not report reasoning support, so the thinking level may have no effect.',
        ),
    ];
  }

  List<Widget> _cursorFields(ButterBoxController boxes, String id) {
    final catalog = boxes.cursorCatalogs[id];
    final models = catalog?.state == BoxLoad.ready
        ? catalog!.models
        : const <ButterBoxCursorModel>[];
    final c = _cursorConfig;
    final inCatalog = models.any((m) => m.id == c.model);
    final mode = c.mode;
    final retainedMode = mode.isNotEmpty && !cursorModes.contains(mode);
    return [
      if (models.isNotEmpty)
        DropdownButtonFormField<String>(
          key: ValueKey('leaf-cursor-catalog-${c.model}-${models.length}'),
          initialValue: inCatalog ? c.model : null,
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'Model from catalog'),
          items: [
            for (final m in models)
              DropdownMenuItem(
                value: m.id,
                child: Text(
                  m.name.isEmpty || m.name == m.id
                      ? m.id
                      : '${m.name} (${m.id})',
                  overflow: TextOverflow.ellipsis,
                ),
              ),
          ],
          onChanged: widget.enabled
              ? (v) {
                  if (v == null) return;
                  _changed(() {
                    c.model = v;
                    _model.text = v;
                  });
                }
              : null,
        ),
      _textField(
        'leaf-model',
        'Model',
        _model,
        (v) => c.model = v,
        helper:
            'Optional Cursor model ID, e.g. composer-2.5. Empty uses the bridge default. Pick from the catalog, or enter it by hand.',
      ),
      DropdownButtonFormField<String>(
        key: ValueKey('leaf-mode-$mode'),
        initialValue: mode,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Mode'),
        items: [
          const DropdownMenuItem(value: '', child: Text('Default (agent)')),
          for (final m in cursorModes)
            DropdownMenuItem(value: m, child: Text(m)),
          if (retainedMode)
            DropdownMenuItem(value: mode, child: Text('$mode (retained)')),
        ],
        onChanged: widget.enabled
            ? (v) {
                if (v != null) _changed(() => c.mode = v);
              }
            : null,
      ),
    ];
  }

  Widget _runLimitFields() => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      DropdownButtonFormField<RunLimitMode>(
        key: ValueKey('leaf-run-limit-${_limit.name}'),
        initialValue: _limit,
        isExpanded: true,
        decoration: const InputDecoration(labelText: 'Maximum run time'),
        items: const [
          DropdownMenuItem(
            value: RunLimitMode.serverDefault,
            child: Text('Server default (1800 seconds)'),
          ),
          DropdownMenuItem(
            value: RunLimitMode.unlimited,
            child: Text('Unlimited'),
          ),
          DropdownMenuItem(
            value: RunLimitMode.seconds,
            child: Text('Set seconds'),
          ),
        ],
        onChanged: widget.enabled
            ? (v) {
                if (v == null) return;
                _changed(() {
                  _limit = v;
                  _applyLimit();
                });
              }
            : null,
      ),
      const Padding(
        padding: EdgeInsets.only(top: 4),
        child: Text(
          'A turn that runs longer is aborted on the box. Unlimited removes the bound, so a runaway run can hold a session and a process slot on the box.',
        ),
      ),
      if (_limit == RunLimitMode.seconds) _textFieldWithValidator(),
    ],
  );

  Widget _textFieldWithValidator() => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: TextFormField(
      key: const Key('leaf-run-seconds'),
      controller: _seconds,
      enabled: widget.enabled,
      keyboardType: TextInputType.number,
      decoration: const InputDecoration(
        labelText: 'Maximum run time (seconds)',
      ),
      validator: validateRunSeconds,
      onChanged: (_) => _changed(_applyLimit),
    ),
  );
}
