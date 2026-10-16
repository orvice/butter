import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/agent_file.pb.dart';
import '../files/file_rules.dart';
import 'agent_draft.dart';

/// An existing resource an Agent can reference: its stable id and a label.
class Choice {
  const Choice(this.id, this.label);
  final String id;
  final String label;
}

/// Whether a list of choices could be read. Associations are optional context
/// for the editor, so one list failing never blocks the rest: the selection
/// stays, and the editor says plainly why nothing can be added.
enum ChoiceStatus { loaded, denied, unavailable }

class ChoiceSet {
  const ChoiceSet(this.status, [this.items = const []]);
  const ChoiceSet.unloaded() : this(ChoiceStatus.unavailable);
  final ChoiceStatus status;
  final List<Choice> items;
  bool get loaded => status == ChoiceStatus.loaded;
  bool contains(String id) => items.any((c) => c.id == id);
  String? labelOf(String id) {
    for (final c in items) {
      if (c.id == id) return c.label;
    }
    return null;
  }
}

/// What the runtime does with each association, by Agent type. MCP servers,
/// file mounts and Skills are LLM-only. Remote Agents attach to an LLM or to a
/// Loop / Sequential / Parallel parent; a Workflow ignores them and Pi/Cursor
/// leaves are refused them. The editor offers only what takes effect.
bool offersMcp(Agent a) => isLLM(a);
bool offersMounts(Agent a) => isLLM(a);
bool offersSkills(Agent a) => isLLM(a);
bool offersRemote(Agent a) =>
    isLLM(a) ||
    a.type == AgentType.AGENT_TYPE_LOOP ||
    a.type == AgentType.AGENT_TYPE_SEQUENTIAL ||
    a.type == AgentType.AGENT_TYPE_PARALLEL;

/// Checkable list of existing resources, by id. Selected ids missing from the
/// loaded list stay selected (and flagged) until the user removes them.
class ReferenceChecklist extends StatelessWidget {
  const ReferenceChecklist({
    super.key,
    required this.keyPrefix,
    required this.title,
    required this.help,
    required this.emptyText,
    required this.selected,
    required this.choices,
    required this.enabled,
    required this.onChanged,
    this.addable = true,
    this.missingNote = 'Not found in this Workspace; retained until removed.',
  });
  final String keyPrefix;
  final String title;
  final String help;
  final String emptyText;
  final List<String> selected;
  final ChoiceSet choices;
  final bool enabled;
  final bool addable;
  final String missingNote;
  final void Function(List<String>) onChanged;

  @override
  Widget build(BuildContext context) {
    final canAdd = addable && choices.loaded;
    final ids = [
      ...selected,
      if (canAdd)
        for (final c in choices.items)
          if (!selected.contains(c.id)) c.id,
    ];
    final note = !addable
        ? 'The runtime ignores these for this Agent type. Remove them to clear; nothing can be added.'
        : switch (choices.status) {
            ChoiceStatus.loaded => null,
            ChoiceStatus.denied =>
              'Your role cannot list these in this Workspace. The current selection is kept; it can be removed but nothing can be added.',
            ChoiceStatus.unavailable =>
              'The list could not be loaded. The current selection is kept; reload the editor choices to add more.',
          };
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 16),
        Text(title, style: Theme.of(context).textTheme.titleMedium),
        Text(help),
        if (note != null) Text(note, key: Key('$keyPrefix-note')),
        for (final id in ids)
          CheckboxListTile(
            key: ValueKey('$keyPrefix-$id'),
            contentPadding: EdgeInsets.zero,
            title: Text(choices.labelOf(id) ?? id),
            subtitle: Text(
              choices.loaded && !choices.contains(id) && addable
                  ? missingNote
                  : (choices.labelOf(id) != null && choices.labelOf(id) != id
                        ? id
                        : ''),
            ),
            value: selected.contains(id),
            onChanged: !enabled
                ? null
                : (value) => onChanged([
                    for (final old in selected)
                      if (old != id) old,
                    if (value == true) id,
                  ]),
          ),
        if (ids.isEmpty && addable && choices.loaded) Text(emptyText),
      ],
    );
  }
}

/// The file spaces an Agent can read or write, each with its own virtual path
/// and permission. Edits go into the mount messages in place, so a mount keeps
/// any field this app does not know about.
class FileMountFields extends StatelessWidget {
  const FileMountFields({
    super.key,
    required this.mounts,
    required this.spaces,
    required this.enabled,
    required this.onChanged,
    this.addable = true,
  });
  final List<AgentFileMount> mounts;
  final ChoiceSet spaces;
  final bool enabled;
  final bool addable;
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    final canAdd = addable && spaces.loaded && spaces.items.isNotEmpty;
    final note = !addable
        ? 'The runtime ignores these for this Agent type. Remove them to clear; nothing can be added.'
        : switch (spaces.status) {
            ChoiceStatus.loaded =>
              spaces.items.isEmpty && mounts.isEmpty
                  ? 'No file spaces in this Workspace yet. Create one under Files.'
                  : null,
            ChoiceStatus.denied =>
              'Your role cannot list file spaces in this Workspace. Current mounts are kept; they can be edited or removed but nothing can be added.',
            ChoiceStatus.unavailable =>
              'File spaces could not be loaded. Current mounts are kept; reload the editor choices to add more.',
          };
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 16),
        Text('File mounts', style: Theme.of(context).textTheme.titleMedium),
        const Text(
          'Give this Agent access to a file space under a virtual path. The longest matching path wins. The model is shown the file tools whatever the permission; a call the permission does not allow is refused when it runs.',
        ),
        if (note != null) Text(note, key: const Key('mounts-note')),
        for (var i = 0; i < mounts.length; i++)
          _MountRow(
            key: ObjectKey(mounts[i]),
            index: i,
            mount: mounts[i],
            all: mounts,
            spaces: spaces,
            enabled: enabled,
            editable: addable,
            onChanged: onChanged,
            onRemove: () {
              mounts.removeAt(i);
              onChanged();
            },
          ),
        if (canAdd)
          DropdownButtonFormField<String>(
            key: ValueKey('add-mount-${mounts.length}'),
            isExpanded: true,
            decoration: const InputDecoration(labelText: 'Mount a file space'),
            items: [
              for (final s in spaces.items)
                DropdownMenuItem(
                  value: s.id,
                  child: Text(
                    s.label == s.id ? s.id : '${s.label} (${s.id})',
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
            ],
            onChanged: !enabled
                ? null
                : (id) {
                    if (id == null) return;
                    final name = spaces.labelOf(id) ?? id;
                    mounts.add(
                      AgentFileMount(
                        spaceId: id,
                        mountPath: suggestMountPath(name, [
                          for (final m in mounts)
                            normalizeMountPath(m.mountPath),
                        ]),
                        // Explicit: an unspecified permission reads as READ,
                        // but a new mount should say so.
                        permission: AgentFileMountPermission
                            .AGENT_FILE_MOUNT_PERMISSION_READ,
                      ),
                    );
                    onChanged();
                  },
          ),
      ],
    );
  }
}

const _permissionLabels = {
  AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ: 'Read only',
  AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ_WRITE:
      'Read and write',
  AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE:
      'Read, write and delete',
};

class _MountRow extends StatefulWidget {
  const _MountRow({
    super.key,
    required this.index,
    required this.mount,
    required this.all,
    required this.spaces,
    required this.enabled,
    required this.editable,
    required this.onChanged,
    required this.onRemove,
  });
  final int index;
  final AgentFileMount mount;
  final List<AgentFileMount> all;
  final ChoiceSet spaces;
  final bool enabled;
  final bool editable;
  final VoidCallback onChanged;
  final VoidCallback onRemove;
  @override
  State<_MountRow> createState() => _MountRowState();
}

class _MountRowState extends State<_MountRow> {
  late final _path = TextEditingController(text: widget.mount.mountPath);
  @override
  void dispose() {
    _path.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final mount = widget.mount;
    final i = widget.index;
    final name = widget.spaces.labelOf(mount.spaceId);
    final unknown = unknownEnum(mount, 3);
    final permission = mount.permission;
    final normalized = normalizeMountPath(_path.text);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              name == null || name == mount.spaceId
                  ? 'Space ${mount.spaceId}'
                  : '$name (${mount.spaceId})',
              key: Key('mount-$i-space'),
              style: Theme.of(context).textTheme.titleSmall,
            ),
            if (widget.spaces.loaded && name == null && widget.editable)
              Text(
                'This space was not found in the Workspace. The server accepts the mount, but every file call through it fails until the space exists or the mount is removed.',
                key: Key('mount-$i-missing'),
              ),
            TextFormField(
              key: Key('mount-$i-path'),
              controller: _path,
              enabled: widget.enabled && widget.editable,
              autocorrect: false,
              decoration: InputDecoration(
                labelText: 'Virtual path',
                helperText:
                    'Mounted at $normalized. An empty path mounts at /.',
              ),
              validator: (_) {
                final mine = normalizeMountPath(_path.text);
                final clash = widget.all.any(
                  (m) =>
                      !identical(m, mount) &&
                      normalizeMountPath(m.mountPath) == mine,
                );
                return clash
                    ? 'Another mount already uses $mine. The runtime would drop one of them.'
                    : null;
              },
              onChanged: (text) {
                mount.mountPath = text.trim();
                setState(() {});
                widget.onChanged();
              },
            ),
            if (unknown != null)
              Text(
                'Permission: unknown value $unknown retained unchanged. Choose one below to replace it.',
                key: Key('mount-$i-unknown'),
              ),
            DropdownButtonFormField<AgentFileMountPermission>(
              key: ValueKey('mount-$i-permission-$unknown-${permission.value}'),
              initialValue: unknown != null
                  ? null
                  : (_permissionLabels.containsKey(permission)
                        ? permission
                        : AgentFileMountPermission
                              .AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED),
              isExpanded: true,
              decoration: const InputDecoration(labelText: 'Permission'),
              items: [
                if (unknown == null &&
                    !_permissionLabels.containsKey(permission))
                  const DropdownMenuItem(
                    value: AgentFileMountPermission
                        .AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED,
                    child: Text('Unspecified (read only)'),
                  ),
                for (final e in _permissionLabels.entries)
                  DropdownMenuItem(value: e.key, child: Text(e.value)),
              ],
              onChanged: !widget.enabled || !widget.editable
                  ? null
                  : (v) {
                      if (v == null) return;
                      mount.permission = v;
                      mount.unknownFields.clearField(3);
                      widget.onChanged();
                    },
            ),
            if (permission.value >= 2 && unknown == null)
              Text(
                permission ==
                        AgentFileMountPermission
                            .AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE
                    ? 'This Agent can change and delete files in ${name ?? mount.spaceId} when it runs.'
                    : 'This Agent can change files in ${name ?? mount.spaceId} when it runs.',
                key: Key('mount-$i-warning'),
              ),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                key: Key('mount-$i-remove'),
                onPressed: widget.enabled ? widget.onRemove : null,
                icon: const Icon(Icons.remove_circle_outline),
                label: Text(
                  'Remove mount ${mount.mountPath.isEmpty ? '/' : mount.mountPath}',
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
