/// An absolute http(s) URL with a host, as the server requires of every
/// endpoint it will call. [what] names the field in the message. An empty
/// value is accepted unless [required].
String? validateHttpUrl(String? value, {required bool required, String? what}) {
  final text = value?.trim() ?? '';
  if (text.isEmpty) {
    return required ? 'Enter ${what ?? 'a URL'}.' : null;
  }
  final uri = Uri.tryParse(text);
  if (uri == null ||
      !(uri.scheme == 'http' || uri.scheme == 'https') ||
      uri.host.isEmpty) {
    return 'Enter an absolute http or https URL, e.g. https://example.com/mcp.';
  }
  return null;
}

/// Whether [text] is a plain-http URL, whose traffic (and any credential in
/// it) is not encrypted.
bool isCleartextUrl(String text) => Uri.tryParse(text.trim())?.scheme == 'http';

/// How a stored credential is treated by an edit.
enum SecretEdit { keep, replace, clear }
