import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/workspace.pb.dart';

const workspaceRoles = ['member', 'admin', 'owner'];

Future<void> openWorkspaceDetails(
  BuildContext context,
  ClientContext client,
  String id,
) async {
  final loading = client.openWorkspace(id);
  // openWorkspace synchronously advances the intentional management scope.
  // Only replacements after that transition should suppress navigation.
  final scope = client.scope;
  await loading;
  if (context.mounted && client.scope == scope && client.user != null) {
    await openManagementPage(
      context,
      client,
      WorkspaceDetailPage(client: client, id: id),
    );
  }
}

class WorkspaceDetailPage extends StatefulWidget {
  const WorkspaceDetailPage({
    super.key,
    required this.client,
    required this.id,
  });
  final ClientContext client;
  final String id;
  @override
  State<WorkspaceDetailPage> createState() => _WorkspaceDetailPageState();
}

class _WorkspaceDetailPageState extends State<WorkspaceDetailPage> {
  @override
  Widget build(BuildContext context) => ManagementPage(
    client: widget.client,
    title: 'Workspace details',
    child: AnimatedBuilder(
      animation: widget.client,
      builder: (context, _) {
        final client = widget.client;
        final space = client.managedWorkspace;
        if (space == null || space.id != widget.id) {
          return Center(
            child: TextButton(
              onPressed: client.busy
                  ? null
                  : () => client.openWorkspace(widget.id),
              child: const Text('Load Workspace'),
            ),
          );
        }
        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(space.name, style: Theme.of(context).textTheme.headlineSmall),
            SelectableText('ID: ${space.id}'),
            Text('Slug: ${space.slug}'),
            Text(space.description),
            TextButton.icon(
              onPressed: client.busy
                  ? null
                  : () => client.openWorkspace(widget.id),
              icon: const Icon(Icons.refresh),
              label: const Text('Refresh details'),
            ),
            if (client.canManageWorkspace)
              Wrap(
                spacing: 12,
                children: [
                  FilledButton(
                    onPressed: client.busy
                        ? null
                        : () => openManagementPage(
                            context,
                            client,
                            WorkspaceFormPage(client: client, workspace: space),
                          ),
                    child: const Text('Edit Workspace'),
                  ),
                  OutlinedButton(
                    onPressed: client.busy
                        ? null
                        : () async {
                            final confirmed = await confirmChange(
                              context,
                              'Delete ${space.name}?',
                              'Permanently removes this Workspace and its memberships. Its resources will no longer be accessible through this Workspace. This cannot be undone. A lost reply does not mean nothing was deleted.',
                              action: 'Delete Workspace',
                            );
                            if (!confirmed || !context.mounted) return;
                            if (await client.deleteWorkspace() &&
                                context.mounted) {
                              Navigator.pop(context);
                            }
                          },
                    child: const Text('Delete Workspace'),
                  ),
                ],
              )
            else
              const Text(
                'Read-only. Only a Workspace owner can manage this Workspace and its members.',
              ),
            const SizedBox(height: 24),
            Text('Members', style: Theme.of(context).textTheme.titleLarge),
            const Text(
              'Members are identified by their existing user IDs; this is not user administration.',
            ),
            if (client.canManageWorkspace)
              TextButton.icon(
                onPressed: client.busy
                    ? null
                    : () => openManagementPage(
                        context,
                        client,
                        MemberFormPage(client: client),
                      ),
                icon: const Icon(Icons.person_add),
                label: const Text('Add member'),
              ),
            for (final member in client.members)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SelectableText(member.userId),
                    Text('Role: ${member.role}'),
                    if (client.canManageWorkspace)
                      Wrap(
                        spacing: 12,
                        children: [
                          TextButton(
                            onPressed:
                                client.busy ||
                                    !workspaceRoles.contains(member.role)
                                ? null
                                : () => openManagementPage(
                                    context,
                                    client,
                                    MemberFormPage(
                                      client: client,
                                      member: member,
                                    ),
                                  ),
                            child: Text('Change role for ${member.userId}'),
                          ),
                          TextButton(
                            onPressed: client.busy
                                ? null
                                : () async {
                                    final confirmed = await confirmChange(
                                      context,
                                      'Remove ${member.userId}?',
                                      'Removes access to ${space.name}.${member.userId == client.user?.id ? ' You will lose access yourself.' : ''}',
                                      action: 'Remove member',
                                    );
                                    if (confirmed && context.mounted) {
                                      await client.removeMember(member.userId);
                                    }
                                  },
                            child: Text('Remove ${member.userId}'),
                          ),
                        ],
                      ),
                  ],
                ),
              ),
            if (client.members.isEmpty) const Text('No members.'),
          ],
        );
      },
    ),
  );
}

class WorkspaceFormPage extends StatefulWidget {
  const WorkspaceFormPage({super.key, required this.client, this.workspace});
  final ClientContext client;
  final Workspace? workspace;
  @override
  State<WorkspaceFormPage> createState() => _WorkspaceFormPageState();
}

class _WorkspaceFormPageState extends State<WorkspaceFormPage> {
  final _form = GlobalKey<FormState>();
  late final _name = TextEditingController(text: widget.workspace?.name);
  late final _slug = TextEditingController(text: widget.workspace?.slug);
  late final _description = TextEditingController(
    text: widget.workspace?.description,
  );
  bool _dirty = false;
  String? _notice;
  @override
  void dispose() {
    _name.dispose();
    _slug.dispose();
    _description.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ManagementPage(
    client: widget.client,
    title: widget.workspace == null ? 'Create Workspace' : 'Edit Workspace',
    child: AnimatedBuilder(
      animation: widget.client,
      builder: (context, _) => DraftGuard(
        dirty: _dirty,
        busy: widget.client.busy,
        child: Form(
          key: _form,
          onChanged: () => setState(() {
            _dirty = true;
            _notice = null;
          }),
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              TextFormField(
                controller: _name,
                enabled: !widget.client.busy,
                decoration: const InputDecoration(labelText: 'Name'),
                validator: _required,
              ),
              TextFormField(
                controller: _slug,
                enabled: !widget.client.busy,
                decoration: const InputDecoration(labelText: 'Slug'),
                validator: _required,
              ),
              TextFormField(
                controller: _description,
                enabled: !widget.client.busy,
                decoration: const InputDecoration(labelText: 'Description'),
                minLines: 2,
                maxLines: 6,
              ),
              const SizedBox(height: 20),
              FilledButton(
                onPressed:
                    widget.client.busy ||
                        (widget.workspace != null &&
                            !widget.client.canManageWorkspace)
                    ? null
                    : () async {
                        if (!_form.currentState!.validate()) return;
                        FocusScope.of(context).unfocus();
                        // Retain timestamps and protobuf unknown fields rather than rebuilding.
                        final draft =
                            widget.workspace?.deepCopy() ?? Workspace();
                        draft
                          ..name = _name.text.trim()
                          ..slug = _slug.text.trim()
                          ..description = _description.text;
                        final ok = await widget.client.saveWorkspace(draft);
                        if (!mounted || !ok) return;
                        setState(() {
                          _dirty = false;
                          _notice = 'Workspace saved.';
                        });
                        Navigator.pop(this.context);
                      },
                child: const Text('Save Workspace'),
              ),
              if (_notice != null) Text(_notice!),
              const Text(
                'No concurrent-edit protection is provided by this API. Refresh before editing again.',
              ),
            ],
          ),
        ),
      ),
    ),
  );
}

String? _required(String? value) =>
    value == null || value.trim().isEmpty ? 'Required.' : null;

class MemberFormPage extends StatefulWidget {
  const MemberFormPage({super.key, required this.client, this.member});
  final ClientContext client;
  final WorkspaceMember? member;
  @override
  State<MemberFormPage> createState() => _MemberFormPageState();
}

class _MemberFormPageState extends State<MemberFormPage> {
  final _form = GlobalKey<FormState>();
  late final _userId = TextEditingController(text: widget.member?.userId);
  late String _role = widget.member?.role ?? 'member';
  bool _dirty = false;
  @override
  void dispose() {
    _userId.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ManagementPage(
    client: widget.client,
    title: widget.member == null ? 'Add member' : 'Change member role',
    child: AnimatedBuilder(
      animation: widget.client,
      builder: (context, _) => DraftGuard(
        dirty: _dirty,
        busy: widget.client.busy,
        child: Form(
          key: _form,
          onChanged: () => setState(() => _dirty = true),
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              TextFormField(
                controller: _userId,
                enabled: widget.member == null && !widget.client.busy,
                decoration: const InputDecoration(
                  labelText: 'Existing user ID',
                ),
                validator: _required,
              ),
              DropdownButtonFormField<String>(
                initialValue: _role,
                decoration: const InputDecoration(labelText: 'Workspace role'),
                items: [
                  for (final role in workspaceRoles)
                    DropdownMenuItem(value: role, child: Text(role)),
                ],
                onChanged: widget.client.busy
                    ? null
                    : (value) => setState(() {
                        _role = value!;
                        _dirty = true;
                      }),
              ),
              const Text(
                'Workspace admin is not a global admin and cannot manage Workspace metadata or members on this server.',
              ),
              const SizedBox(height: 20),
              FilledButton(
                onPressed:
                    widget.client.busy || !widget.client.canManageWorkspace
                    ? null
                    : () async {
                        if (!_form.currentState!.validate()) return;
                        final confirmed = await confirmChange(
                          context,
                          '${widget.member == null ? 'Add' : 'Change role for'} ${_userId.text.trim()}?',
                          'Assign $_role in ${widget.client.managedWorkspace!.name}.${_userId.text.trim() == widget.client.user?.id ? ' Changing your own role may remove your management access.' : ''}',
                        );
                        if (!confirmed || !context.mounted) return;
                        final ok = widget.member == null
                            ? await widget.client.addMember(
                                _userId.text.trim(),
                                _role,
                              )
                            : await widget.client.updateMember(
                                _userId.text.trim(),
                                _role,
                              );
                        if (ok && mounted) {
                          setState(() => _dirty = false);
                          Navigator.pop(this.context);
                        }
                      },
                child: const Text('Save member'),
              ),
            ],
          ),
        ),
      ),
    ),
  );
}
