import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import 'avatar_selector.dart';
import '../../core/management_page.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key, required this.client});
  final ClientContext client;
  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final _profileForm = GlobalKey<FormState>();
  final _passwordForm = GlobalKey<FormState>();
  late final _name = TextEditingController(
    text: widget.client.user!.displayName.isEmpty
        ? widget.client.user!.username
        : widget.client.user!.displayName,
  );
  final _current = TextEditingController();
  final _next = TextEditingController();
  final _confirm = TextEditingController();
  Uint8List? _avatar;
  bool _clearAvatar = false;
  bool _picking = false;
  bool _dirty = false;
  String? _notice;
  String? _localError;
  @override
  void dispose() {
    _name.dispose();
    _current.dispose();
    _next.dispose();
    _confirm.dispose();
    super.dispose();
  }

  Future<void> _pick() async {
    if (_picking || widget.client.busy) return;
    final scope = widget.client.scope;
    setState(() {
      _picking = true;
      _localError = null;
    });
    try {
      final bytes = await AvatarSelector().choose();
      if (!mounted || scope != widget.client.scope || bytes == null) return;
      setState(() {
        _avatar = bytes;
        _clearAvatar = false;
        _dirty = true;
        _notice = null;
      });
    } catch (failure) {
      if (mounted && scope == widget.client.scope) {
        setState(() => _localError = errorMessage(failure));
      }
    } finally {
      if (mounted) setState(() => _picking = false);
    }
  }

  @override
  Widget build(BuildContext context) => ManagementPage(
    client: widget.client,
    title: 'My profile',
    child: AnimatedBuilder(
      animation: widget.client,
      builder: (context, _) {
        final user = widget.client.user;
        final busy = widget.client.busy || _picking;
        return DraftGuard(
          dirty: _dirty,
          busy: busy,
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              Text('Username: ${user?.username ?? ''}'),
              SelectableText('User ID: ${user?.id ?? ''}'),
              if (user?.email.isNotEmpty == true)
                Text('Email: ${user!.email} (read-only)'),
              if (_localError != null)
                Semantics(liveRegion: true, child: Text(_localError!)),
              if (_notice != null)
                Semantics(liveRegion: true, child: Text(_notice!)),
              const SizedBox(height: 16),
              Semantics(
                label: 'Profile avatar',
                image: true,
                child: SizedBox(
                  height: 96,
                  child: _avatar != null
                      ? Image.memory(
                          _avatar!,
                          errorBuilder: (_, _, _) => const Icon(Icons.person),
                        )
                      : !_clearAvatar && user?.avatarUrl.isNotEmpty == true
                      ? Image.network(
                          user!.avatarUrl,
                          errorBuilder: (_, _, _) => const Icon(Icons.person),
                        )
                      : const Icon(Icons.person),
                ),
              ),
              Wrap(
                spacing: 12,
                children: [
                  TextButton.icon(
                    onPressed: busy ? null : _pick,
                    icon: const Icon(Icons.photo_library),
                    label: const Text('Choose avatar'),
                  ),
                  TextButton(
                    onPressed: busy
                        ? null
                        : () async {
                            if (await confirmChange(
                                  context,
                                  'Remove avatar?',
                                  'The stored avatar URL will be cleared when you save your profile.',
                                  action: 'Remove',
                                ) &&
                                mounted) {
                              setState(() {
                                _avatar = null;
                                _clearAvatar = true;
                                _dirty = true;
                              });
                            }
                          },
                    child: const Text('Remove avatar'),
                  ),
                ],
              ),
              const Text(
                'PNG, JPEG, GIF or WebP, up to 10 MiB (the server may set a smaller limit). Selecting a photo does not upload it until Save. Uploads require server avatar storage; ask an administrator if unavailable.',
              ),
              Form(
                key: _profileForm,
                onChanged: () => setState(() {
                  _dirty = true;
                  _notice = null;
                }),
                child: Column(
                  children: [
                    TextFormField(
                      controller: _name,
                      enabled: !busy,
                      decoration: const InputDecoration(
                        labelText: 'Display name',
                      ),
                      validator: (value) =>
                          value == null || value.trim().isEmpty
                          ? 'Enter a display name.'
                          : null,
                    ),
                    const SizedBox(height: 16),
                    FilledButton(
                      onPressed: busy
                          ? null
                          : () async {
                              if (!_profileForm.currentState!.validate()) {
                                return;
                              }
                              FocusScope.of(context).unfocus();
                              final ok = await widget.client.saveProfile(
                                _name.text,
                                avatar: _avatar,
                                avatarUrl: _clearAvatar ? '' : null,
                              );
                              if (!mounted || !ok) return;
                              setState(() {
                                _avatar = null;
                                _clearAvatar = false;
                                _dirty =
                                    _current.text.isNotEmpty ||
                                    _next.text.isNotEmpty ||
                                    _confirm.text.isNotEmpty;
                                _notice = 'Profile saved.';
                              });
                            },
                      child: const Text('Save profile'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 32),
              Text(
                'Change my password',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              Form(
                key: _passwordForm,
                onChanged: () => setState(() {
                  _dirty = true;
                  _notice = null;
                }),
                child: Column(
                  children: [
                    _passwordField(
                      _current,
                      'Current password',
                      busy,
                      AutofillHints.password,
                    ),
                    _passwordField(
                      _next,
                      'New password',
                      busy,
                      AutofillHints.newPassword,
                    ),
                    TextFormField(
                      controller: _confirm,
                      enabled: !busy,
                      obscureText: true,
                      autocorrect: false,
                      enableSuggestions: false,
                      decoration: const InputDecoration(
                        labelText: 'Confirm new password',
                      ),
                      validator: (value) =>
                          value != _next.text || value == null || value.isEmpty
                          ? 'Passwords must match.'
                          : null,
                    ),
                    const SizedBox(height: 16),
                    FilledButton(
                      onPressed: busy
                          ? null
                          : () async {
                              if (!_passwordForm.currentState!.validate()) {
                                return;
                              }
                              FocusScope.of(context).unfocus();
                              final current = _current.text;
                              final next = _next.text;
                              // Passwords are never retained after a submission, including failure.
                              _current.clear();
                              _next.clear();
                              _confirm.clear();
                              final ok = await widget.client.changePassword(
                                current,
                                next,
                              );
                              if (mounted && ok) {
                                setState(() => _notice = 'Password changed.');
                              }
                            },
                      child: const Text('Change password'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              OutlinedButton.icon(
                onPressed: busy
                    ? null
                    : () async {
                        if (await confirmChange(
                              context,
                              'Sign out?',
                              'Clears this account and all in-memory drafts. This does not stop server runs.',
                              action: 'Sign out',
                            ) &&
                            context.mounted) {
                          await widget.client.logout();
                        }
                      },
                icon: const Icon(Icons.logout),
                label: const Text('Sign out'),
              ),
            ],
          ),
        );
      },
    ),
  );

  Widget _passwordField(
    TextEditingController controller,
    String label,
    bool busy,
    String hint,
  ) => TextFormField(
    controller: controller,
    enabled: !busy,
    obscureText: true,
    autocorrect: false,
    enableSuggestions: false,
    autofillHints: [hint],
    decoration: InputDecoration(labelText: label),
    validator: (value) => value == null || value.isEmpty ? 'Required.' : null,
  );
}
