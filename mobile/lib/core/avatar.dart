import 'dart:typed_data';

/// The server's image whitelist, read from the bytes themselves, never from a
/// file suffix. Null for anything else.
String? sniffImageContentType(Uint8List bytes) {
  bool starts(List<int> signature, [int offset = 0]) =>
      bytes.length >= offset + signature.length &&
      Iterable<int>.generate(
        signature.length,
      ).every((i) => bytes[offset + i] == signature[i]);
  if (starts([137, 80, 78, 71, 13, 10, 26, 10])) return 'image/png';
  if (starts([255, 216, 255])) return 'image/jpeg';
  if (starts([71, 73, 70, 56, 55, 97]) || starts([71, 73, 70, 56, 57, 97])) {
    return 'image/gif';
  }
  if (starts([82, 73, 70, 70]) && starts([87, 69, 66, 80], 8)) {
    return 'image/webp';
  }
  return null;
}

/// Matches the upload service's image whitelist; never trusts a file suffix.
String avatarContentType(Uint8List bytes) {
  if (bytes.isEmpty || bytes.length > 10 * 1024 * 1024) {
    throw const FormatException(
      'Choose an image no larger than 10 MiB. The server may set a smaller limit.',
    );
  }
  return sniffImageContentType(bytes) ??
      (throw const FormatException('Choose a PNG, JPEG, GIF or WebP image.'));
}
