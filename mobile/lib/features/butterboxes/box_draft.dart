/// Validation shared by the ButterBox form. The server validates again; these
/// only keep an obviously wrong value from being sent.
String? validateBoxName(String? value) =>
    value == null || value.trim().isEmpty ? 'Enter a name.' : null;

/// An absolute http(s) URL with a host, as the server requires. A trailing
/// slash is normalized by the server.
String? validateBoxUrl(String? value) {
  final text = value?.trim() ?? '';
  final uri = Uri.tryParse(text);
  if (text.isEmpty ||
      uri == null ||
      !(uri.scheme == 'http' || uri.scheme == 'https') ||
      uri.host.isEmpty) {
    return 'Enter an absolute http or https URL, e.g. https://box.example.com.';
  }
  return null;
}
