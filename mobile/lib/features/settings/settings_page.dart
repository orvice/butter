import 'package:flutter/material.dart';

import '../../app/app_preferences.dart';
import '../../core/management_page.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key, required this.preferences});
  final AppPreferences preferences;
  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  late ThemeMode _theme = widget.preferences.theme;
  late bool _identifiers = widget.preferences.showIdentifiers;
  bool _dirty = false;
  String? _notice;
  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: widget.preferences,
    builder: (context, _) => Scaffold(
      appBar: AppBar(title: const Text('Appearance and display')),
      body: SafeArea(
        child: DraftGuard(
          dirty: _dirty,
          busy: widget.preferences.busy,
          child: ListView(
            padding: const EdgeInsets.all(24),
            children: [
              DropdownButtonFormField<ThemeMode>(
                initialValue: _theme,
                decoration: const InputDecoration(labelText: 'Theme'),
                items: [
                  for (final theme in ThemeMode.values)
                    DropdownMenuItem(value: theme, child: Text(theme.name)),
                ],
                onChanged: widget.preferences.busy
                    ? null
                    : (value) => setState(() {
                        _theme = value!;
                        _dirty = true;
                        _notice = null;
                      }),
              ),
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Show resource identifiers'),
                subtitle: const Text(
                  'Show Agent IDs in lists. Management details always show IDs.',
                ),
                value: _identifiers,
                onChanged: widget.preferences.busy
                    ? null
                    : (value) => setState(() {
                        _identifiers = value!;
                        _dirty = true;
                        _notice = null;
                      }),
              ),
              const Text(
                'Text follows the system font size and screen-reader settings.',
              ),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: widget.preferences.busy
                    ? null
                    : () async {
                        final ok = await widget.preferences.save(
                          theme: _theme,
                          showIdentifiers: _identifiers,
                        );
                        if (mounted && ok) {
                          setState(() {
                            _dirty = false;
                            _notice = 'Appearance saved.';
                          });
                        }
                      },
                child: const Text('Save appearance'),
              ),
              if (widget.preferences.busy)
                const LinearProgressIndicator(
                  semanticsLabel: 'Saving appearance',
                ),
              if (widget.preferences.error != null)
                Semantics(
                  liveRegion: true,
                  child: Text(widget.preferences.error!),
                ),
              if (_notice != null)
                Semantics(liveRegion: true, child: Text(_notice!)),
              const SizedBox(height: 24),
              const Text(
                'Development client: language support and formal tablet acceptance are not yet confirmed. These settings do not choose a release language or OS minimum.',
              ),
            ],
          ),
        ),
      ),
    ),
  );
}
