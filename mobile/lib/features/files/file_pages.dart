import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent_file.pb.dart';
import 'file_controller.dart';
import 'file_rules.dart';

/// Opens the Workspace's file spaces. The list reads fresh from the server.
Future<void> openFileSpaces(BuildContext context, ClientContext client) async {
  final files = client.fileSpaces;
  if (files == null) return;
  unawaited(files.load());
  await openManagementPage(context, client, FileSpacesPage(client: client));
}

String _when(AgentFile file) => file.hasUpdatedAt()
    ? file.updatedAt.toDateTime().toLocal().toString().split('.').first
    : 'unknown time';

String fileSummary(AgentFile file) =>
    '${formatBytes(file.sizeBytes.toInt())} · version ${file.version} · ${_when(file)}';

class FileSpacesPage extends StatelessWidget {
  const FileSpacesPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final files = client.fileSpaces;
    if (files == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: files,
      title: 'File spaces',
      child: ListenableBuilder(
        listenable: files,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · File spaces',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A file space holds text files that Agents can read or change when it is mounted on them in the Agent editor. Each save of a file makes a new version.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-space'),
                  onPressed: files.writing
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          SpaceFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('New file space'),
                ),
                TextButton.icon(
                  key: const Key('refresh-spaces'),
                  onPressed: files.listState == ConfigLoad.loading
                      ? null
                      : files.load,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (files.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (files.listState == ConfigLoad.failed)
              ErrorLine(files.listError ?? 'File spaces could not be loaded.'),
            if (files.listState == ConfigLoad.ready && files.items.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No file spaces yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text('Create one to keep files an Agent can use.'),
            ],
            for (final space in files.items)
              ListTile(
                key: Key('space-${space.id}'),
                contentPadding: EdgeInsets.zero,
                title: Text(space.name),
                subtitle: Text(
                  [
                    if (space.description.isNotEmpty) space.description,
                    'ID: ${space.id}',
                  ].join('\n'),
                ),
                isThreeLine: space.description.isNotEmpty,
                onTap: () => openManagementPage(
                  context,
                  client,
                  SpaceDetailPage(client: client, id: space.id),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Creates or renames a space. An edit starts from a fresh read of the stored
/// record and keeps every field it does not show (the server replaces the
/// whole record on update).
class SpaceFormPage extends StatefulWidget {
  const SpaceFormPage({super.key, required this.client, this.id});
  final ClientContext client;
  final String? id;
  @override
  State<SpaceFormPage> createState() => _SpaceFormPageState();
}

class _SpaceFormPageState extends State<SpaceFormPage> {
  final _name = TextEditingController();
  final _description = TextEditingController();
  AgentFileSpace? _base;
  bool _initialised = false;
  bool _dirty = false;
  String? _problem;

  @override
  void initState() {
    super.initState();
    if (widget.id != null) {
      unawaited(widget.client.fileSpaces?.loadItem(widget.id!));
    } else {
      _initialised = true;
    }
  }

  @override
  void dispose() {
    _name.dispose();
    _description.dispose();
    super.dispose();
  }

  Future<void> _save(FileSpaceController files) async {
    final name = _name.text.trim();
    if (name.isEmpty) {
      setState(() => _problem = 'Enter a name.');
      return;
    }
    setState(() => _problem = null);
    final base = _base;
    final ok = base == null
        ? await files.create(
            AgentFileSpace(name: name, description: _description.text),
          )
        : await files.update(
            base.deepCopy()
              ..name = name
              ..description = _description.text,
          );
    if (ok && mounted) {
      setState(() => _dirty = false);
      if (Navigator.canPop(context)) Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final files = widget.client.fileSpaces;
    if (files == null) return const SizedBox.shrink();
    return DraftGuard(
      dirty: _dirty,
      busy: files.writing,
      child: ConfigScaffold(
        controller: files,
        title: widget.id == null ? 'New file space' : 'Edit file space',
        child: ListenableBuilder(
          listenable: files,
          builder: (context, _) {
            if (widget.id != null && !_initialised) {
              final item = files.item?.id == widget.id ? files.item : null;
              if (item != null) {
                _base = item.deepCopy();
                _name.text = item.name;
                _description.text = item.description;
                _initialised = true;
              } else {
                return ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    if (files.itemState == ConfigLoad.loading)
                      const LinearProgressIndicator(semanticsLabel: 'Loading'),
                    if (files.itemError != null) ErrorLine(files.itemError!),
                    TextButton(
                      onPressed: files.itemState == ConfigLoad.loading
                          ? null
                          : () => files.loadItem(widget.id!),
                      child: const Text('Load file space'),
                    ),
                  ],
                );
              }
            }
            final kept = _base?.metadata.length ?? 0;
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                TextField(
                  key: const Key('space-name'),
                  controller: _name,
                  enabled: !files.writing,
                  decoration: const InputDecoration(labelText: 'Name'),
                  onChanged: (_) => setState(() => _dirty = true),
                ),
                const SizedBox(height: 12),
                TextField(
                  key: const Key('space-description'),
                  controller: _description,
                  enabled: !files.writing,
                  maxLines: 3,
                  decoration: const InputDecoration(labelText: 'Description'),
                  onChanged: (_) => setState(() => _dirty = true),
                ),
                if (kept > 0)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      '$kept metadata ${kept == 1 ? 'entry' : 'entries'} set elsewhere are kept as stored.',
                    ),
                  ),
                if (_problem != null) ErrorLine(_problem!),
                const SizedBox(height: 16),
                FilledButton(
                  key: const Key('save-space'),
                  onPressed: files.writing || files.unconfirmed
                      ? null
                      : () => _save(files),
                  child: Text(widget.id == null ? 'Create' : 'Save'),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class SpaceDetailPage extends StatefulWidget {
  const SpaceDetailPage({super.key, required this.client, required this.id});
  final ClientContext client;
  final String id;
  @override
  State<SpaceDetailPage> createState() => _SpaceDetailPageState();
}

class _SpaceDetailPageState extends State<SpaceDetailPage> {
  final _prefix = TextEditingController();
  final _query = TextEditingController();
  List<AgentFileSearchResult>? _results;
  String? _searchedFor;
  String? _searchError;
  bool _searching = false;
  int _searchSeq = 0;
  bool _checking = false;
  FileSpaceController? _files;

  @override
  void initState() {
    super.initState();
    _files = widget.client.fileSpaces;
    final files = _files;
    if (files != null) {
      unawaited(files.loadItem(widget.id));
      unawaited(files.openSpace(widget.id));
    }
  }

  @override
  void dispose() {
    _files?.closeSpace(notify: false);
    _prefix.dispose();
    _query.dispose();
    super.dispose();
  }

  Future<void> _search(FileSpaceController files) async {
    final seq = ++_searchSeq;
    setState(() {
      _searching = true;
      _searchError = null;
    });
    final answer = await files.search(widget.id, _query.text);
    if (!mounted || seq != _searchSeq || answer.stale) return;
    setState(() {
      _searching = false;
      _results = answer.results;
      _searchedFor = answer.error == null ? _query.text.trim() : null;
      _searchError = answer.error;
    });
  }

  Future<void> _delete(FileSpaceController files, AgentFileSpace space) async {
    setState(() => _checking = true);
    final refs = await files.agentsMounting(space.id);
    if (!mounted) return;
    setState(() => _checking = false);
    final count =
        files.filesPrefix.isEmpty && files.filesState == ConfigLoad.ready
        ? '${files.files.length} file${files.files.length == 1 ? '' : 's'} and all their versions'
        : 'every file in it and all their versions';
    final mounted_ = refs == null
        ? ''
        : refs.agents.isNotEmpty
        ? ' Agents that mount it (${refs.agents.join(', ')}) will fail when they use it.'
        : refs.complete
        ? ''
        : ' Butter could not check which Agents mount it; Agents that do will fail when they use it.';
    final ok = await confirmChange(
      context,
      'Delete file space?',
      'Delete "${space.name}"? This permanently deletes $count. It cannot be undone.$mounted_',
      action: 'Delete space',
    );
    if (!ok || !mounted) return;
    if (await files.delete(space.id) && mounted && Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final files = widget.client.fileSpaces;
    if (files == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: files,
      title: 'File space',
      child: ListenableBuilder(
        listenable: files,
        builder: (context, _) {
          final space = files.item?.id == widget.id
              ? files.item
              : files.byId(widget.id);
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (space != null) ...[
                SelectableText(
                  space.name,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                if (space.description.isNotEmpty)
                  SelectableText(space.description),
                SelectableText('ID: ${space.id}'),
              ] else if (files.itemState == ConfigLoad.failed)
                ErrorLine(files.itemError ?? 'This file space is unavailable.'),
              const SizedBox(height: 12),
              Wrap(
                spacing: 12,
                children: [
                  FilledButton.icon(
                    key: const Key('new-file'),
                    onPressed: files.writing
                        ? null
                        : () => openManagementPage(
                            context,
                            widget.client,
                            FileEditPage(
                              client: widget.client,
                              space: widget.id,
                            ),
                          ),
                    icon: const Icon(Icons.note_add_outlined),
                    label: const Text('New file'),
                  ),
                  TextButton.icon(
                    key: const Key('refresh-files'),
                    onPressed: files.filesState == ConfigLoad.loading
                        ? null
                        : () => files.loadFiles(),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                  if (space != null)
                    OutlinedButton(
                      key: const Key('edit-space'),
                      onPressed: files.writing
                          ? null
                          : () => openManagementPage(
                              context,
                              widget.client,
                              SpaceFormPage(
                                client: widget.client,
                                id: space.id,
                              ),
                            ),
                      child: const Text('Edit space'),
                    ),
                  if (space != null)
                    OutlinedButton(
                      key: const Key('delete-space'),
                      onPressed: files.writing || _checking
                          ? null
                          : () => _delete(files, space),
                      child: Text(_checking ? 'Checking…' : 'Delete space'),
                    ),
                ],
              ),
              const Divider(height: 32),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      key: const Key('files-prefix'),
                      controller: _prefix,
                      decoration: const InputDecoration(
                        labelText: 'Only paths under',
                        hintText: '/notes',
                      ),
                      onSubmitted: (v) => files.loadFiles(prefix: v),
                    ),
                  ),
                  const SizedBox(width: 8),
                  OutlinedButton(
                    key: const Key('apply-prefix'),
                    onPressed: files.filesState == ConfigLoad.loading
                        ? null
                        : () => files.loadFiles(prefix: _prefix.text),
                    child: const Text('Filter'),
                  ),
                ],
              ),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      key: const Key('files-search'),
                      controller: _query,
                      decoration: const InputDecoration(
                        labelText: 'Search file contents',
                      ),
                      onSubmitted: (_) => _search(files),
                    ),
                  ),
                  const SizedBox(width: 8),
                  OutlinedButton(
                    key: const Key('run-search'),
                    onPressed: _searching ? null : () => _search(files),
                    child: const Text('Search'),
                  ),
                ],
              ),
              if (_searching)
                const LinearProgressIndicator(semanticsLabel: 'Searching'),
              if (_searchError != null) ErrorLine(_searchError!),
              if (_results != null) ...[
                const SizedBox(height: 8),
                Text(
                  _results!.isEmpty
                      ? 'No file contains "$_searchedFor".'
                      : '${_results!.length} match${_results!.length == 1 ? '' : 'es'} for "$_searchedFor":',
                  key: const Key('search-summary'),
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                for (final hit in _results!)
                  ListTile(
                    key: Key('hit-${hit.file.path}'),
                    contentPadding: EdgeInsets.zero,
                    title: Text(hit.file.path),
                    subtitle: Text(hit.snippets.take(2).join('\n')),
                    onTap: () => openManagementPage(
                      context,
                      widget.client,
                      FileViewPage(
                        client: widget.client,
                        space: widget.id,
                        path: hit.file.path,
                      ),
                    ),
                  ),
              ],
              const Divider(height: 32),
              Text(
                files.filesPrefix.isEmpty
                    ? 'Files'
                    : 'Files under ${files.filesPrefix}',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              if (files.filesState == ConfigLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              if (files.filesState == ConfigLoad.failed)
                ErrorLine(files.filesError ?? 'Files could not be loaded.'),
              if (files.filesState == ConfigLoad.ready && files.files.isEmpty)
                Text(
                  files.filesPrefix.isEmpty
                      ? 'No files yet.'
                      : 'No files under that path.',
                  key: const Key('no-files'),
                ),
              for (final file in files.files)
                ListTile(
                  key: Key('file-${file.path}'),
                  contentPadding: EdgeInsets.zero,
                  title: Text(file.path),
                  subtitle: Text(fileSummary(file)),
                  onTap: () => openManagementPage(
                    context,
                    widget.client,
                    FileViewPage(
                      client: widget.client,
                      space: widget.id,
                      path: file.path,
                    ),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}

enum _ConflictChoice { review, overwrite, cancel }

Future<_ConflictChoice> _askConflict(
  BuildContext context,
  FileConflict conflict,
) async {
  final (title, review, over) = switch (conflict.kind) {
    FileConflictKind.exists => (
      'A file is already there',
      'Open the existing file',
      'Replace it with my text',
    ),
    FileConflictKind.changed => (
      'The file changed',
      'Review the latest version',
      'Save over it anyway',
    ),
    FileConflictKind.deleted => (
      'The file was deleted',
      'Check the file',
      'Create it again',
    ),
  };
  final choice = await showDialog<_ConflictChoice>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(title),
      content: Text(
        '${conflict.message}\n\nThe server cannot refuse a save that is out of date, so Butter checked just before saving. Saving anyway adds a new version on top; earlier versions stay in the history.',
      ),
      actions: [
        TextButton(
          key: const Key('conflict-cancel'),
          onPressed: () => Navigator.pop(context, _ConflictChoice.cancel),
          child: const Text('Keep editing'),
        ),
        TextButton(
          key: const Key('conflict-review'),
          onPressed: () => Navigator.pop(context, _ConflictChoice.review),
          child: Text(review),
        ),
        FilledButton(
          key: const Key('conflict-overwrite'),
          onPressed: () => Navigator.pop(context, _ConflictChoice.overwrite),
          child: Text(over),
        ),
      ],
    ),
  );
  return choice ?? _ConflictChoice.cancel;
}

/// One file: its latest version, or any earlier one by number.
class FileViewPage extends StatefulWidget {
  const FileViewPage({
    super.key,
    required this.client,
    required this.space,
    required this.path,
  });
  final ClientContext client;
  final String space;
  final String path;
  @override
  State<FileViewPage> createState() => _FileViewPageState();
}

class _FileViewPageState extends State<FileViewPage> {
  final _jump = TextEditingController();
  FileDocument? _latest;
  FileDocument? _shown;
  String? _error;
  bool _missing = false;
  bool _tooLarge = false;
  bool _loading = true;
  int _seq = 0;

  @override
  void initState() {
    super.initState();
    unawaited(_read());
  }

  @override
  void dispose() {
    _jump.dispose();
    super.dispose();
  }

  FileSpaceController? get _files => widget.client.fileSpaces;

  /// Reads the latest version, or [version] of it.
  Future<void> _read({int version = 0}) async {
    final files = _files;
    if (files == null) return;
    // A listed size beyond the limit is not worth a read.
    final listed = files.files.where((f) => f.path == widget.path);
    if (version == 0 &&
        listed.isNotEmpty &&
        listed.first.sizeBytes.toInt() > editableFileLimit) {
      setState(() {
        _tooLarge = true;
        _loading = false;
      });
      return;
    }
    final seq = ++_seq;
    setState(() {
      _loading = true;
      _error = null;
      _missing = false;
    });
    final read = await files.readFile(
      widget.space,
      widget.path,
      version: version,
      latest: _latest?.version,
    );
    if (!mounted || seq != _seq || read.stale) {
      if (mounted && seq == _seq) setState(() => _loading = false);
      return;
    }
    setState(() {
      _loading = false;
      final doc = read.document;
      if (doc == null) {
        _error = read.error;
        _missing = read.notFound && version == 0;
        if (version == 0) {
          _latest = null;
          _shown = null;
        }
        return;
      }
      if (utf8Length(doc.content) > editableFileLimit) {
        _tooLarge = true;
        _shown = null;
        return;
      }
      _tooLarge = false;
      _shown = doc;
      if (doc.isLatest) _latest = doc;
    });
  }

  Future<void> _delete() async {
    final files = _files;
    if (files == null) return;
    final ok = await confirmChange(
      context,
      'Delete file?',
      'Delete ${widget.path} from "${files.byId(widget.space)?.name ?? widget.space}"? This deletes the file and all its versions. It cannot be undone.',
      action: 'Delete file',
    );
    if (!ok || !mounted) return;
    if (await files.deleteFile(widget.space, widget.path) &&
        mounted &&
        Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  Future<void> _verifyDelete() async {
    final files = _files;
    if (files == null) return;
    final result = await files.verifyWrite(widget.space, widget.path);
    if (!mounted) return;
    if (result == FileVerification.applied && Navigator.canPop(context)) {
      Navigator.pop(context);
    } else {
      unawaited(_read());
    }
  }

  Future<void> _restore(FileDocument version) async {
    final files = _files;
    final latest = _latest;
    if (files == null || latest == null) return;
    final ok = await confirmChange(
      context,
      'Restore version ${version.version}?',
      'Save the text of version ${version.version} of ${widget.path} as a new version ${latest.version + 1}. Version ${latest.version} stays in the history.',
      action: 'Restore',
    );
    if (!ok || !mounted) return;
    var overwrite = false;
    for (var attempt = 0; attempt < 2; attempt++) {
      final saved = await files.saveFile(
        space: widget.space,
        path: widget.path,
        content: version.content,
        baseVersion: latest.version,
        contentType: latest.file.contentType.isEmpty
            ? defaultFileContentType
            : latest.file.contentType,
        metadata: Map.of(latest.file.metadata),
        overwrite: overwrite,
      );
      if (!mounted) return;
      if (saved != null) {
        unawaited(_read());
        return;
      }
      final conflict = files.conflict;
      if (conflict == null) return;
      final choice = await _askConflict(context, conflict);
      if (!mounted) return;
      files.clearConflict();
      if (choice == _ConflictChoice.review) {
        unawaited(_read());
        return;
      }
      if (choice == _ConflictChoice.cancel) return;
      overwrite = true;
    }
  }

  Future<void> _edit(FileDocument doc) async {
    await openManagementPage(
      context,
      widget.client,
      FileEditPage(client: widget.client, space: widget.space, existing: doc),
    );
    if (mounted) unawaited(_read());
  }

  @override
  Widget build(BuildContext context) {
    final files = _files;
    if (files == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: files,
      title: widget.path,
      child: ListenableBuilder(
        listenable: files,
        builder: (context, _) {
          final doc = _shown;
          final latest = _latest;
          final children = <Widget>[
            SelectableText(
              widget.path,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            if (_loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (_error != null) ErrorLine(_error!),
            if (_tooLarge)
              const Text(
                'This file is larger than 1 MiB, so it is not opened on this device. Use the web dashboard to view or edit it.',
                key: Key('file-too-large'),
              ),
          ];
          if (_missing) {
            children.add(
              const Text(
                'This file no longer exists in this space.',
                key: Key('file-missing'),
              ),
            );
          }
          if (files.fileWriteUnconfirmed) {
            children.add(
              OutlinedButton(
                key: const Key('check-delete'),
                onPressed: _verifyDelete,
                child: const Text('Check whether the change was applied'),
              ),
            );
          }
          if (doc != null) {
            final viewingOld = !doc.isLatest;
            children.addAll([
              Text(
                'Version ${doc.version} of ${doc.latest}'
                '${viewingOld ? ' (not the latest)' : ''}',
                key: const Key('file-version'),
                style: Theme.of(context).textTheme.titleSmall,
              ),
              Text(
                '${formatBytes(doc.file.sizeBytes.toInt())} · ${doc.file.contentType.isEmpty ? 'no content type' : doc.file.contentType} · ${_when(doc.file)}',
              ),
              if (doc.isLatest && doc.file.metadata.isNotEmpty)
                Text(
                  'Metadata: ${(doc.file.metadata.keys.toList()..sort()).join(', ')}',
                ),
              if (viewingOld)
                const Text(
                  'Earlier versions show their text and content type only.',
                ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  OutlinedButton(
                    key: const Key('older-version'),
                    onPressed: doc.version > 1 && !_loading
                        ? () => _read(version: doc.version - 1)
                        : null,
                    child: const Text('Older'),
                  ),
                  OutlinedButton(
                    key: const Key('newer-version'),
                    onPressed: viewingOld && !_loading
                        ? () => _read(version: doc.version + 1)
                        : null,
                    child: const Text('Newer'),
                  ),
                  OutlinedButton(
                    key: const Key('latest-version'),
                    onPressed: viewingOld && !_loading ? () => _read() : null,
                    child: const Text('Latest'),
                  ),
                ],
              ),
              Row(
                children: [
                  SizedBox(
                    width: 120,
                    child: TextField(
                      key: const Key('version-jump'),
                      controller: _jump,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Go to version',
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  OutlinedButton(
                    key: const Key('go-version'),
                    onPressed: _loading
                        ? null
                        : () {
                            final v = int.tryParse(_jump.text.trim());
                            if (v == null || v < 1 || v > doc.latest) {
                              setState(
                                () => _error =
                                    'Enter a version from 1 to ${doc.latest}.',
                              );
                              return;
                            }
                            _read(version: v >= doc.latest ? 0 : v);
                          },
                    child: const Text('Go'),
                  ),
                ],
              ),
              const Divider(height: 24),
              if (viewingOld && latest != null)
                FilledButton(
                  key: const Key('restore-version'),
                  onPressed: files.writing || files.unconfirmed
                      ? null
                      : () => _restore(doc),
                  child: Text('Restore version ${doc.version}'),
                ),
              if (!viewingOld)
                Wrap(
                  spacing: 8,
                  children: [
                    FilledButton(
                      key: const Key('edit-file'),
                      onPressed: files.writing || files.unconfirmed
                          ? null
                          : () => _edit(doc),
                      child: const Text('Edit'),
                    ),
                    OutlinedButton(
                      key: const Key('delete-file'),
                      onPressed: files.writing || files.unconfirmed
                          ? null
                          : _delete,
                      child: const Text('Delete file'),
                    ),
                  ],
                ),
              const SizedBox(height: 12),
              SelectableText(
                doc.content.isEmpty ? '(empty file)' : doc.content,
                key: const Key('file-content-view'),
                style: const TextStyle(fontFamily: 'monospace'),
              ),
            ]);
          } else if (_missing) {
            // Nothing else to offer for a file that is gone.
          } else if (_error != null && !_loading) {
            children.add(
              TextButton(
                key: const Key('retry-read'),
                onPressed: _read,
                child: const Text('Read again'),
              ),
            );
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: children,
          );
        },
      ),
    );
  }
}

/// Creates a file, or edits the latest version of one. The content is plain
/// UTF-8 text, sent to the server as such.
class FileEditPage extends StatefulWidget {
  const FileEditPage({
    super.key,
    required this.client,
    required this.space,
    this.existing,
  });
  final ClientContext client;
  final String space;
  final FileDocument? existing;
  @override
  State<FileEditPage> createState() => _FileEditPageState();
}

class _FileEditPageState extends State<FileEditPage> {
  late final _path = TextEditingController(text: widget.existing?.file.path);
  late final _content = TextEditingController(text: widget.existing?.content);
  bool _dirty = false;
  String? _problem;
  String? _verifyNote;
  String? _lastSent;
  int _bytes = 0;

  @override
  void initState() {
    super.initState();
    _bytes = utf8Length(_content.text);
  }

  @override
  void dispose() {
    _path.dispose();
    _content.dispose();
    super.dispose();
  }

  Future<void> _save(
    FileSpaceController files, {
    bool overwrite = false,
  }) async {
    final path = normalizeFilePath(_path.text);
    if (path == null) {
      setState(() => _problem = 'Enter a file path such as /notes/todo.md.');
      return;
    }
    setState(() {
      _problem = null;
      _verifyNote = null;
    });
    final existing = widget.existing;
    _lastSent = _content.text;
    final saved = await files.saveFile(
      space: widget.space,
      path: path,
      content: _content.text,
      baseVersion: existing?.version,
      contentType: existing == null || existing.file.contentType.isEmpty
          ? defaultFileContentType
          : existing.file.contentType,
      metadata: existing == null ? const {} : Map.of(existing.file.metadata),
      overwrite: overwrite,
    );
    if (!mounted) return;
    if (saved != null) {
      setState(() => _dirty = false);
      if (Navigator.canPop(context)) Navigator.pop(context);
      return;
    }
    final conflict = files.conflict;
    if (conflict == null) return;
    final choice = await _askConflict(context, conflict);
    if (!mounted) return;
    files.clearConflict();
    switch (choice) {
      case _ConflictChoice.overwrite:
        // An existing file that was deleted is created again; one that
        // changed is replaced. Either way the person chose it.
        await _save(files, overwrite: true);
      case _ConflictChoice.review:
        if (Navigator.canPop(context)) Navigator.pop(context);
        await openManagementPage(
          context,
          widget.client,
          FileViewPage(
            client: widget.client,
            space: widget.space,
            path: conflict.path,
          ),
        );
      case _ConflictChoice.cancel:
        break;
    }
  }

  Future<void> _verify(FileSpaceController files) async {
    final path = normalizeFilePath(_path.text);
    final sent = _lastSent;
    if (path == null || sent == null) return;
    final result = await files.verifyWrite(
      widget.space,
      path,
      expectedContent: sent,
    );
    if (!mounted) return;
    setState(() {
      switch (result) {
        case FileVerification.applied:
          _dirty = false;
          _verifyNote =
              'The file now holds your text, so the save went through.';
        case FileVerification.notApplied:
          _verifyNote =
              'The file does not hold your text, so the save did not go through. Your text is still here; save again if you want to.';
        case FileVerification.unknown:
          _verifyNote = 'Could not read the file to check. Try again.';
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final files = widget.client.fileSpaces;
    if (files == null) return const SizedBox.shrink();
    final existing = widget.existing;
    final normalized = normalizeFilePath(_path.text);
    final over = _bytes > defaultFileMaxBytes;
    return DraftGuard(
      dirty: _dirty,
      busy: files.writing,
      child: ConfigScaffold(
        controller: files,
        title: existing == null ? 'New file' : 'Edit file',
        child: ListenableBuilder(
          listenable: files,
          builder: (context, _) => ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (existing == null) ...[
                TextField(
                  key: const Key('file-path'),
                  controller: _path,
                  enabled: !files.writing,
                  autocorrect: false,
                  decoration: const InputDecoration(
                    labelText: 'Path',
                    hintText: '/notes/todo.md',
                  ),
                  onChanged: (_) => setState(() => _dirty = true),
                ),
                Text(
                  normalized == null
                      ? 'This is not a usable path.'
                      : 'Saved as $normalized',
                  key: const Key('file-path-normalized'),
                ),
              ] else
                SelectableText(
                  existing.file.path,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              if (existing != null)
                Text(
                  'Editing version ${existing.version}. Saving adds version ${existing.version + 1}; the content type${existing.file.metadata.isEmpty ? '' : ' and ${existing.file.metadata.length} metadata entries'} are kept.',
                ),
              const SizedBox(height: 12),
              TextField(
                key: const Key('file-content'),
                controller: _content,
                enabled: !files.writing,
                minLines: 8,
                maxLines: 20,
                autocorrect: false,
                keyboardType: TextInputType.multiline,
                style: const TextStyle(fontFamily: 'monospace'),
                decoration: const InputDecoration(
                  labelText: 'Content',
                  alignLabelWithHint: true,
                  border: OutlineInputBorder(),
                ),
                onChanged: (text) => setState(() {
                  _dirty = true;
                  _bytes = utf8Length(text);
                }),
              ),
              Text(
                '${formatBytes(_bytes)} ($_bytes bytes) of the server\'s usual 256 KiB limit',
                key: const Key('file-bytes'),
                style: over
                    ? TextStyle(color: Theme.of(context).colorScheme.error)
                    : null,
              ),
              if (over)
                const Text(
                  'This is over the usual limit. The server may refuse it; a refused save changes nothing.',
                  key: Key('file-over-limit'),
                ),
              if (_problem != null) ErrorLine(_problem!),
              if (_verifyNote != null)
                Text(_verifyNote!, key: const Key('verify-note')),
              const SizedBox(height: 16),
              Wrap(
                spacing: 12,
                children: [
                  FilledButton(
                    key: const Key('save-file'),
                    onPressed: files.writing || files.unconfirmed
                        ? null
                        : () => _save(files),
                    child: const Text('Save'),
                  ),
                  if (files.unconfirmed && _lastSent != null)
                    OutlinedButton(
                      key: const Key('verify-file'),
                      onPressed: () => _verify(files),
                      child: const Text('Check whether it was saved'),
                    ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
