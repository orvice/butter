import 'package:flutter/foundation.dart';
import 'package:image_picker/image_picker.dart';

import '../../core/avatar.dart';

/// One local gallery selection. No credentials, persistence or network writes.
final class AvatarSelector {
  Future<Uint8List?> choose() async {
    final file = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      requestFullMetadata: false,
    );
    if (file == null) return null;
    if (await file.length() > 10 * 1024 * 1024) {
      throw const FormatException('Choose an image no larger than 10 MiB.');
    }
    final bytes = await file.readAsBytes();
    avatarContentType(bytes);
    return bytes;
  }
}
