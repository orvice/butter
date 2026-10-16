import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../core/repo_binding_errors.dart';
import '../../gen/agents/v1/repobinding.pb.dart';
import 'repo_binding_controller.dart';
import 'repo_binding_draft.dart';

Future<void> openRepoBinding(BuildContext context, ClientContext client) async {
  final repo = client.repoBinding;
  if (repo == null) return;
  unawaited(repo.load());
  unawaited(repo.loadGitHosts());
  await openManagementPage(context, client, RepoBindingPage(client: client));
}

class RepoBindingPage extends StatelessWidget {
  const RepoBindingPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final repo = client.repoBinding;
    if (repo == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: repo,
      builder: (context, _) => Scaffold(
        appBar: AppBar(title: const Text('Git repository')),
        body: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · Repository binding',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'Bind this Workspace to a platform Git host. Agent Content can live in Git after onboarding. The personal access token is write-only and never shown again.',
            ),
            if (!client.canManageWorkspaceResources) ...[
              const SizedBox(height: 12),
              const Text(
                'You can view the binding. Changing it needs the owner or admin role.',
              ),
            ],
            if (repo.loadState == RepoBindingLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (repo.loadError != null) ErrorLine(repo.loadError!),
            if (repo.unconfirmed) ...[
              const SizedBox(height: 8),
              ErrorLine(repo.writeError ?? repoBindingUnconfirmed),
            ],
            const SizedBox(height: 12),
            _BindingForm(client: client, repo: repo),
            if (repo.hasBinding) ...[
              const SizedBox(height: 24),
              _StatusCard(repo: repo),
              const SizedBox(height: 16),
              _CredentialSection(
                repo: repo,
                canManage: client.canManageWorkspaceResources,
              ),
              if (repo.hasActiveRevision) ...[
                const SizedBox(height: 16),
                _CacheBrowser(
                  repo: repo,
                  canManage: client.canManageWorkspaceResources,
                ),
              ] else ...[
                const SizedBox(height: 16),
                _OnboardingSection(
                  repo: repo,
                  canManage: client.canManageWorkspaceResources,
                ),
              ],
              const SizedBox(height: 16),
              _OperationsSection(
                repo: repo,
                canManage: client.canManageWorkspaceResources,
              ),
              const SizedBox(height: 16),
              _DetachSection(
                repo: repo,
                canManage: client.canManageWorkspaceResources,
              ),
            ],
            if (repo.webhookSecret != null) _WebhookBanner(repo: repo),
          ],
        ),
      ),
    );
  }
}

class ErrorLine extends StatelessWidget {
  const ErrorLine(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 8),
    child: Text(
      text,
      style: TextStyle(color: Theme.of(context).colorScheme.error),
    ),
  );
}

class _BindingForm extends StatefulWidget {
  const _BindingForm({required this.client, required this.repo});
  final ClientContext client;
  final RepoBindingController repo;

  @override
  State<_BindingForm> createState() => _BindingFormState();
}

class _BindingFormState extends State<_BindingForm> {
  late RepoBindingDraft _draft;
  String? _previousHost;

  RepoBindingWriteMode get _effectiveWriteMode =>
      _draft.writeMode ==
          RepoBindingWriteMode.REPO_BINDING_WRITE_MODE_UNSPECIFIED
      ? RepoBindingWriteMode.REPO_BINDING_WRITE_MODE_DIRECT_COMMIT
      : _draft.writeMode;

  @override
  void initState() {
    super.initState();
    _draft = widget.repo.binding != null
        ? RepoBindingDraft.fromStored(widget.repo.binding!)
        : RepoBindingDraft.empty();
    _previousHost = _draft.gitHostId;
  }

  @override
  void didUpdateWidget(covariant _BindingForm oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.repo.binding != widget.repo.binding &&
        widget.repo.binding != null &&
        !widget.repo.writing &&
        !widget.repo.unconfirmed) {
      _draft = RepoBindingDraft.fromStored(widget.repo.binding!);
      _previousHost = _draft.gitHostId;
    }
  }

  Future<void> _save() async {
    if ((_previousHost ?? '').isNotEmpty &&
        _draft.gitHostId != _previousHost &&
        _draft.credentialSet) {
      final ok = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Change Git host?'),
          content: const Text(
            'Saving a different Git host clears the stored personal access token on the server. You will need to set a new token.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Save'),
            ),
          ],
        ),
      );
      if (ok != true) return;
    }
    if (await widget.repo.saveBinding(_draft) && mounted) {
      _previousHost = _draft.gitHostId;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Binding saved')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final repo = widget.repo;
    final canEdit =
        widget.client.canManageWorkspaceResources &&
        !repo.writing &&
        !repo.unconfirmed;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Configuration',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            if (repo.hostsState == RepoBindingLoad.failed)
              ErrorLine(repo.hostsError ?? 'Git hosts unavailable'),
            DropdownButtonFormField<String>(
              key: ValueKey('git-host-${_draft.gitHostId}'),
              initialValue: _draft.gitHostId.isEmpty ? null : _draft.gitHostId,
              decoration: const InputDecoration(labelText: 'Git host'),
              items: [
                for (final h in repo.gitHosts)
                  DropdownMenuItem(
                    value: h.id,
                    child: Text(h.name.isEmpty ? h.id : h.name),
                  ),
              ],
              onChanged: canEdit
                  ? (v) => setState(() => _draft.gitHostId = v ?? '')
                  : null,
            ),
            TextFormField(
              key: const Key('repository'),
              initialValue: _draft.repository,
              decoration: const InputDecoration(
                labelText: 'Repository (owner/repo)',
              ),
              enabled: canEdit,
              onChanged: (v) => _draft.repository = v,
            ),
            TextFormField(
              key: const Key('branch'),
              initialValue: _draft.branch,
              decoration: const InputDecoration(labelText: 'Branch'),
              enabled: canEdit,
              onChanged: (v) => _draft.branch = v,
            ),
            TextFormField(
              key: const Key('root-path'),
              initialValue: _draft.rootPath,
              decoration: const InputDecoration(
                labelText: 'Root path (optional)',
                helperText:
                    'Directory containing agents/; empty means repository root.',
              ),
              enabled: canEdit,
              onChanged: (v) => _draft.rootPath = v,
            ),
            const SizedBox(height: 8),
            Text('Write mode', style: Theme.of(context).textTheme.titleSmall),
            RadioGroup<RepoBindingWriteMode>(
              groupValue: _effectiveWriteMode,
              onChanged: (RepoBindingWriteMode? v) {
                if (!canEdit || v == null) return;
                setState(() => _draft.writeMode = v);
              },
              child: Column(
                children: const [
                  RadioListTile<RepoBindingWriteMode>(
                    contentPadding: EdgeInsets.zero,
                    title: Text('Direct commit'),
                    value: RepoBindingWriteMode
                        .REPO_BINDING_WRITE_MODE_DIRECT_COMMIT,
                  ),
                  RadioListTile<RepoBindingWriteMode>(
                    contentPadding: EdgeInsets.zero,
                    title: Text('Pull / merge request'),
                    value: RepoBindingWriteMode
                        .REPO_BINDING_WRITE_MODE_CHANGE_REQUEST,
                  ),
                ],
              ),
            ),
            if (repo.overlaps.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                'Other Workspaces share this location: ${repo.overlaps.map((o) => o.workspaceName.isEmpty ? o.workspaceId : o.workspaceName).join(', ')}',
              ),
            ],
            const SizedBox(height: 8),
            FilledButton(
              key: const Key('save-binding'),
              onPressed: canEdit ? _save : null,
              child: const Text('Save binding'),
            ),
          ],
        ),
      ),
    );
  }
}

class _StatusCard extends StatelessWidget {
  const _StatusCard({required this.repo});
  final RepoBindingController repo;

  @override
  Widget build(BuildContext context) {
    final b = repo.binding!;
    final state = b.hasStatus()
        ? b.status.state
        : RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_UNVALIDATED;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Connection', style: Theme.of(context).textTheme.titleMedium),
            Text('Status: ${connectionStateLabel(state)}'),
            if (b.observedCommitSha.isNotEmpty)
              Text('Observed revision: ${b.observedCommitSha}'),
            if (b.activeCommitSha.isNotEmpty)
              Text('Active revision: ${b.activeCommitSha}'),
            if (b.hasStatus() && b.status.error.isNotEmpty)
              ErrorLine(b.status.error),
            if (b.lastSyncError.isNotEmpty)
              ErrorLine('Last sync: ${b.lastSyncError}'),
            if (b.lastPublicationError.isNotEmpty)
              ErrorLine('Last publish: ${b.lastPublicationError}'),
            for (final e in b.publicationErrors) ErrorLine(e),
          ],
        ),
      ),
    );
  }
}

class _CredentialSection extends StatefulWidget {
  const _CredentialSection({required this.repo, required this.canManage});
  final RepoBindingController repo;
  final bool canManage;

  @override
  State<_CredentialSection> createState() => _CredentialSectionState();
}

class _CredentialSectionState extends State<_CredentialSection> {
  final _pat = TextEditingController();

  @override
  void dispose() {
    _pat.dispose();
    super.dispose();
  }

  Future<void> _setPat() async {
    final pat = _pat.text;
    _pat.clear();
    if (await widget.repo.setCredential(pat) && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Token saved; validate the connection')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final b = widget.repo.binding!;
    final busy = widget.repo.writing || widget.repo.unconfirmed;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Personal access token',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            Text(
              b.credentialSet ? 'Token: set (never shown)' : 'Token: not set',
            ),
            if (widget.canManage) ...[
              TextField(
                key: const Key('pat-field'),
                controller: _pat,
                obscureText: true,
                enabled: !busy,
                decoration: InputDecoration(
                  labelText: b.credentialSet ? 'Replace token' : 'Set token',
                ),
              ),
              FilledButton(
                key: const Key('save-pat'),
                onPressed: busy ? null : _setPat,
                child: Text(b.credentialSet ? 'Replace token' : 'Set token'),
              ),
              OutlinedButton(
                key: const Key('validate-binding'),
                onPressed: busy ? null : () => widget.repo.validate(),
                child: const Text('Validate connection'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _OnboardingSection extends StatefulWidget {
  const _OnboardingSection({required this.repo, required this.canManage});
  final RepoBindingController repo;
  final bool canManage;

  @override
  State<_OnboardingSection> createState() => _OnboardingSectionState();
}

class _OnboardingSectionState extends State<_OnboardingSection> {
  RepoBindingOnboardingMode _mode =
      RepoBindingOnboardingMode.REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT;

  Future<void> _run() async {
    final label =
        _mode ==
            RepoBindingOnboardingMode
                .REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT
        ? 'Export current Agent Content to Git? This writes database content to the repository, validates it, and publishes the Active Revision when successful.'
        : 'Import repository content? Only directories matching existing Agent IDs are adopted.';
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Activate Agent Content from Git'),
        content: Text(label),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Continue'),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    final result = await widget.repo.onboard(_mode);
    if (result != null && mounted) {
      final msg = result.published
          ? 'Onboarding succeeded (${result.agentsExported > 0 ? '${result.agentsExported} exported' : '${result.agentsImported} imported'})'
          : 'Onboarding finished with validation errors';
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.canManage) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Text('Onboarding needs the owner or admin role.'),
        ),
      );
    }
    final busy = widget.repo.writing || widget.repo.unconfirmed;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Onboarding', style: Theme.of(context).textTheme.titleMedium),
            const Text('Choose how to reconcile Agent Content with Git once.'),
            RadioGroup<RepoBindingOnboardingMode>(
              groupValue: _mode,
              onChanged: (RepoBindingOnboardingMode? v) {
                if (busy || v == null) return;
                setState(() => _mode = v);
              },
              child: Column(
                children: const [
                  RadioListTile<RepoBindingOnboardingMode>(
                    contentPadding: EdgeInsets.zero,
                    title: Text('Export current content'),
                    subtitle: Text('Write DB content to Git, then publish.'),
                    value: RepoBindingOnboardingMode
                        .REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
                  ),
                  RadioListTile<RepoBindingOnboardingMode>(
                    contentPadding: EdgeInsets.zero,
                    title: Text('Import repository content'),
                    subtitle: Text('Adopt matching Agent IDs from Git.'),
                    value: RepoBindingOnboardingMode
                        .REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY,
                  ),
                ],
              ),
            ),
            FilledButton(
              key: const Key('onboard'),
              onPressed: busy ? null : _run,
              child: const Text('Activate Agent Content from Git'),
            ),
          ],
        ),
      ),
    );
  }
}

class _OperationsSection extends StatelessWidget {
  const _OperationsSection({required this.repo, required this.canManage});
  final RepoBindingController repo;
  final bool canManage;

  @override
  Widget build(BuildContext context) {
    if (!canManage) return const SizedBox.shrink();
    final busy = repo.writing || repo.unconfirmed;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            FilledButton(
              key: const Key('sync-repo'),
              onPressed: busy ? null : () => repo.syncRepository(),
              child: const Text('Sync cache'),
            ),
            OutlinedButton(
              key: const Key('publish-repo'),
              onPressed: busy ? null : () => repo.publish(),
              child: const Text('Publish Active Revision'),
            ),
            OutlinedButton(
              onPressed: busy ? null : () => repo.configureWebhook(),
              child: const Text('Configure webhook secret'),
            ),
            if (repo.binding!.status.state ==
                RepoBindingConnectionState
                    .REPO_BINDING_CONNECTION_STATE_DIVERGED)
              OutlinedButton(
                onPressed: busy ? null : () => repo.acceptBaseline(),
                child: const Text('Accept new baseline'),
              ),
          ],
        ),
      ),
    );
  }
}

class _CacheBrowser extends StatelessWidget {
  const _CacheBrowser({required this.repo, required this.canManage});
  final RepoBindingController repo;
  final bool canManage;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    'Cached repository tree',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                if (canManage)
                  TextButton(
                    onPressed: repo.writing
                        ? null
                        : () => repo.loadEntries(path: ''),
                    child: const Text('Refresh list'),
                  ),
              ],
            ),
            if (repo.entriesState == RepoBindingLoad.loading)
              const LinearProgressIndicator(),
            if (repo.entriesError != null) ErrorLine(repo.entriesError!),
            for (final e in repo.entries)
              ListTile(
                key: Key('entry-${e.path}'),
                leading: Icon(
                  e.kind == RepoCacheEntryKind.REPO_CACHE_ENTRY_KIND_DIRECTORY
                      ? Icons.folder_outlined
                      : Icons.description_outlined,
                ),
                title: Text(e.path.isEmpty ? '(root)' : e.path),
                onTap: () {
                  if (e.kind ==
                      RepoCacheEntryKind.REPO_CACHE_ENTRY_KIND_DIRECTORY) {
                    repo.loadEntries(path: e.path);
                  } else {
                    repo.loadFile(e.path);
                  }
                },
              ),
            if (repo.fileContent != null && repo.selectedFilePath != null) ...[
              const Divider(),
              Text(
                repo.selectedFilePath!,
                style: Theme.of(context).textTheme.titleSmall,
              ),
              SelectableText(repo.fileContent!),
            ],
          ],
        ),
      ),
    );
  }
}

class _DetachSection extends StatelessWidget {
  const _DetachSection({required this.repo, required this.canManage});
  final RepoBindingController repo;
  final bool canManage;

  Future<void> _detach(
    BuildContext context, {
    required bool keepDatabase,
  }) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove repository binding?'),
        content: Text(
          keepDatabase
              ? 'Keep database Agent Content as-is and remove the binding. Remote Git is not changed.'
              : 'Materialize the Active Revision into the database, reload agents, then remove the binding and stored token.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Remove binding'),
          ),
        ],
      ),
    );
    if (ok != true) return;
    final recovery = keepDatabase
        ? RepoBindingDetachRecovery.REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE
        : RepoBindingDetachRecovery.REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED;
    final n = await repo.detach(recovery: recovery);
    if (context.mounted && n != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Binding removed ($n agents materialized)')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!canManage) return const SizedBox.shrink();
    final busy = repo.writing || repo.unconfirmed;
    return Card(
      color: Theme.of(
        context,
      ).colorScheme.errorContainer.withValues(alpha: 0.3),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Safe detach', style: Theme.of(context).textTheme.titleMedium),
            const Text(
              'Removing the binding does not delete remote Git history. Without a valid Active Revision snapshot, detach is refused unless you keep database content.',
            ),
            const SizedBox(height: 8),
            FilledButton(
              key: const Key('detach-binding'),
              onPressed: busy
                  ? null
                  : () => _detach(context, keepDatabase: false),
              child: const Text('Remove binding (materialize)'),
            ),
            OutlinedButton(
              onPressed: busy
                  ? null
                  : () => _detach(context, keepDatabase: true),
              child: const Text('Remove binding (keep database content)'),
            ),
          ],
        ),
      ),
    );
  }
}

class _WebhookBanner extends StatelessWidget {
  const _WebhookBanner({required this.repo});
  final RepoBindingController repo;

  @override
  Widget build(BuildContext context) {
    final w = repo.webhookSecret!;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Webhook secret (shown once)',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            SelectableText('URL: ${w.callbackUrl}'),
            SelectableText('Secret: ${w.secret}'),
            TextButton(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: w.secret));
              },
              child: const Text('Copy secret'),
            ),
            FilledButton(
              onPressed: repo.dismissWebhookSecret,
              child: const Text('Dismiss'),
            ),
          ],
        ),
      ),
    );
  }
}
