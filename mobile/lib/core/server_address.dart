/// The single server's HTTPS origin and optional deployment prefix.
final class ServerAddress {
  ServerAddress._(this.uri);
  final Uri uri;

  static ServerAddress parse(String input, {bool allowHttp = false}) {
    var text = input.trim();
    if (text.isEmpty || RegExp(r'[\s\\]').hasMatch(text)) {
      throw const FormatException('Enter a valid server address.');
    }
    if (!text.contains('://')) text = 'https://$text';
    // Uri normalizes dot segments: reject them before that can happen.
    final rawPath = text.split('://').last.split('/').skip(1);
    if (rawPath.any(
      (segment) => ['.', '..'].contains(Uri.decodeComponent(segment)),
    )) {
      throw const FormatException('Dot path segments are not supported.');
    }
    final uri = Uri.parse(text);
    if (uri.host.isEmpty ||
        uri.userInfo.isNotEmpty ||
        uri.hasQuery ||
        uri.hasFragment ||
        (uri.scheme != 'https' && !(allowHttp && uri.scheme == 'http'))) {
      throw const FormatException(
        'Use HTTPS without credentials, query or fragment.',
      );
    }
    return ServerAddress._(
      uri.replace(path: uri.path.replaceFirst(RegExp(r'/+$'), '')),
    );
  }

  Uri endpoint(String path) => Uri.parse('$uri/$path');
  @override
  String toString() => uri.toString();
}
