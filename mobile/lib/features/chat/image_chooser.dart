import 'package:image_picker/image_picker.dart';

import 'run_request.dart';

/// Picks images from the gallery for one chat turn. Selection is local only;
/// nothing is sent or stored until the turn is sent.
final class ImageChooser {
  const ImageChooser();

  /// Up to [remaining] images. Each is size-checked before it is read, then
  /// format-checked by its bytes. Returns what could be added and, when
  /// something could not, the first reason.
  Future<({List<ChatImage> images, String? problem})> choose(
    int remaining,
  ) async {
    if (remaining <= 0) {
      return (
        images: const <ChatImage>[],
        problem: 'A message holds at most $maxImagesPerMessage images.',
      );
    }
    final picker = ImagePicker();
    final List<XFile> files;
    if (remaining == 1) {
      final file = await picker.pickImage(
        source: ImageSource.gallery,
        requestFullMetadata: false,
      );
      files = [?file];
    } else {
      files = await picker.pickMultiImage(
        requestFullMetadata: false,
        limit: remaining,
      );
    }
    final images = <ChatImage>[];
    String? problem;
    for (final file in files.take(remaining)) {
      try {
        if (await file.length() > maxImageBytes) {
          throw const ChatInputError('Choose images no larger than 10 MiB.');
        }
        images.add(ChatImage.fromBytes(file.name, await file.readAsBytes()));
      } on ChatInputError catch (error) {
        problem ??= error.message;
      }
    }
    if (files.length > remaining) {
      problem ??= 'A message holds at most $maxImagesPerMessage images.';
    }
    return (images: images, problem: problem);
  }
}
