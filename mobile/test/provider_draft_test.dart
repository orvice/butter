import 'package:butter_mobile/features/model_providers/provider_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/provider_fixtures.dart';

ModelRowInput row(
  ModelConfig m, {
  String? name,
  String? alias,
  String? window,
}) => ModelRowInput(
  original: m,
  name: name ?? m.name,
  alias: alias ?? m.alias,
  contextWindow:
      window ?? (m.contextWindowTokens == 0 ? '' : '${m.contextWindowTokens}'),
);

void main() {
  group('buildModelProvider', () {
    final base = providerWithUnknown('openai');

    test('changes only the edited field and keeps everything else', () {
      final draft = buildModelProvider(
        base: base,
        type: base.type,
        baseUrl: 'https://other.example/v1',
        key: KeyAction.keep,
        rows: [for (final m in base.models) row(m)],
      );
      expect(draft.name, 'openai');
      expect(draft.apiKey, 'sk-secret-value');
      expect(draft.workspaceId, 'alpha');
      expect(draft.baseUrl, 'https://other.example/v1');
      expect(unknownVarint(draft, 99), 7);
      expect([for (final m in draft.models) unknownVarint(m, 15)], [1, 2, 3]);
      // The loaded message itself is never mutated by building a draft.
      expect(base.baseUrl, 'https://llm.example/v1');
    });

    test('keep, replace and clear are three different outcomes', () {
      ModelProvider build(KeyAction action, [String key = '']) =>
          buildModelProvider(
            base: base,
            type: 'openai',
            baseUrl: base.baseUrl,
            key: action,
            newKey: key,
            rows: [for (final m in base.models) row(m)],
          );
      expect(build(KeyAction.keep).apiKey, 'sk-secret-value');
      expect(build(KeyAction.replace, '  sk-new\n').apiKey, 'sk-new');
      expect(build(KeyAction.clear).apiKey, '');
      expect(build(KeyAction.clear).hasApiKey(), isFalse);
    });

    test('removing or editing a row never moves unknown fields to another', () {
      final models = base.models;
      final draft = buildModelProvider(
        base: base,
        type: 'openai',
        baseUrl: base.baseUrl,
        key: KeyAction.keep,
        rows: [
          row(models[2], alias: 'renamed'),
          row(models[0]),
          const ModelRowInput(
            name: ' fresh ',
            alias: '',
            contextWindow: '4294967295',
          ),
        ],
      );
      expect(
        [for (final m in draft.models) m.name],
        ['gpt-nano', 'gpt-4o', 'fresh'],
      );
      expect(
        [for (final m in draft.models) unknownVarint(m, 15)],
        [3, 1, null],
      );
      expect(draft.models[0].alias, 'renamed');
      expect(draft.models[1].contextWindowTokens, 128000);
      expect(draft.models[2].contextWindowTokens, 4294967295);
    });

    test('a new provider carries only what was entered', () {
      final draft = buildModelProvider(
        name: ' fresh ',
        type: 'gemini',
        baseUrl: ' ',
        key: KeyAction.replace,
        newKey: 'k',
        rows: const [ModelRowInput(name: 'm', alias: '', contextWindow: '')],
      );
      expect(draft.name, 'fresh');
      expect(draft.baseUrl, '');
      expect(draft.apiKey, 'k');
      expect(draft.workspaceId, '');
      expect(draft.models.single.contextWindowTokens, 0);
    });
  });

  group('validation', () {
    test('context window is a uint32; empty and 0 mean unset', () {
      expect(parseContextWindow(''), 0);
      expect(parseContextWindow(' 0 '), 0);
      expect(parseContextWindow('4294967295'), 4294967295);
      expect(parseContextWindow('4294967296'), isNull);
      expect(parseContextWindow('-1'), isNull);
      expect(parseContextWindow('1.5'), isNull);
      expect(parseContextWindow('1e3'), isNull);
      expect(parseContextWindow('99999999999999999999999'), isNull);
      expect(validateContextWindow('4294967296'), isNotNull);
      expect(validateContextWindow(''), isNull);
    });

    test('base URL must be absolute http(s) when present', () {
      expect(validateBaseUrl(''), isNull);
      expect(validateBaseUrl('https://a.example/v1'), isNull);
      expect(validateBaseUrl('http://localhost:11434/v1'), isNull);
      expect(validateBaseUrl('ftp://a.example'), isNotNull);
      expect(validateBaseUrl('a.example/v1'), isNotNull);
      expect(validateBaseUrl('/v1'), isNotNull);
    });

    test('required fields and token formatting', () {
      expect(validateProviderName(' '), isNotNull);
      expect(validateModelId(''), isNotNull);
      expect(formatTokens(0), '0');
      expect(formatTokens(128000), '128,000');
      expect(formatTokens(4294967295), '4,294,967,295');
    });
  });
}
