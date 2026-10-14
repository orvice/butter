import '../../gen/agents/v1/agent.pb.dart';

/// What an edit does to the stored API key. `ModelProvider.api_key` is a plain
/// proto3 string and Update replaces the whole object, so "keep" means sending
/// the key the server returned, "replace" a new one and "clear" an empty one.
enum KeyAction { keep, replace, clear }

const providerTypes = ['openai', 'gemini'];

/// `context_window_tokens` is a protobuf uint32.
const maxContextWindowTokens = 4294967295;

/// One model row as the user edited it. [original] is the server's message for
/// that row, so a row keeps its own unknown fields wherever it ends up.
final class ModelRowInput {
  const ModelRowInput({
    this.original,
    required this.name,
    required this.alias,
    required this.contextWindow,
  });
  final ModelConfig? original;
  final String name;
  final String alias;
  final String contextWindow;
}

/// Builds the complete message to store. Starts from a deep copy of [base]
/// (the provider as the server returned it) and changes only what the form
/// edits, so the Workspace ID and unknown fields survive. [base] is null for a
/// new provider.
ModelProvider buildModelProvider({
  ModelProvider? base,
  String name = '',
  required String type,
  required String baseUrl,
  required KeyAction key,
  String newKey = '',
  required List<ModelRowInput> rows,
}) {
  final provider = base?.deepCopy() ?? (ModelProvider()..name = name.trim());
  provider
    ..type = type
    ..baseUrl = baseUrl.trim();
  switch (key) {
    case KeyAction.keep:
      break;
    case KeyAction.replace:
      provider.apiKey = newKey.trim();
    case KeyAction.clear:
      provider.clearApiKey();
  }
  final models = [
    for (final row in rows)
      (row.original?.deepCopy() ?? ModelConfig())
        ..name = row.name.trim()
        ..alias = row.alias.trim()
        ..contextWindowTokens = parseContextWindow(row.contextWindow) ?? 0,
  ];
  provider.models
    ..clear()
    ..addAll(models);
  return provider;
}

/// Empty and 0 both mean "unset": the backend then uses model metadata. Null
/// means the text is not a valid uint32.
int? parseContextWindow(String text) {
  final value = text.trim();
  if (value.isEmpty) return 0;
  if (!RegExp(r'^\d+$').hasMatch(value)) return null;
  final parsed = int.tryParse(value);
  if (parsed == null || parsed > maxContextWindowTokens) return null;
  return parsed;
}

String? validateProviderName(String? value) =>
    value == null || value.trim().isEmpty
    ? 'Enter a name for this provider.'
    : null;

String? validateBaseUrl(String? value) {
  final text = value?.trim() ?? '';
  if (text.isEmpty) return null;
  final url = Uri.tryParse(text);
  if (url == null ||
      !url.hasAuthority ||
      url.host.isEmpty ||
      (url.scheme != 'https' && url.scheme != 'http')) {
    return 'Enter an absolute http or https address, or leave this empty.';
  }
  return null;
}

String? validateModelId(String? value) => value == null || value.trim().isEmpty
    ? 'Enter the model ID the provider uses.'
    : null;

String? validateContextWindow(String? value) =>
    parseContextWindow(value ?? '') == null
    ? 'Enter a whole number from 0 to 4,294,967,295, or leave this empty.'
    : null;

/// "128,000" for a token count.
String formatTokens(int value) {
  final digits = value.toString();
  final out = StringBuffer();
  for (var i = 0; i < digits.length; i++) {
    if (i > 0 && (digits.length - i) % 3 == 0) out.write(',');
    out.write(digits[i]);
  }
  return out.toString();
}
