import 'dart:convert';
import 'dart:math';

import 'package:app_links/app_links.dart';
import 'package:crypto/crypto.dart';
import 'package:url_launcher/url_launcher.dart';

/// Where the server's OAuth callback page sends the browser back to this App.
/// It must be listed in the server's `auth.native_return_uris`, registered as a
/// URL scheme in the Android manifest and iOS Info.plist, and is a development
/// value like the other identifiers (see the README).
const nativeReturnUri = 'ee.orz.butter:/oauth/callback';

/// The OS side of native sign-in: the system browser and incoming links.
abstract interface class NativeOAuthPlatform {
  /// Opens [url] in the system browser (never an embedded WebView). Returns
  /// false if the OS could not open it.
  Future<bool> openBrowser(Uri url);

  /// Links that open the App, including the one that launched it cold.
  Stream<Uri> get links;
}

final class DeviceOAuthPlatform implements NativeOAuthPlatform {
  DeviceOAuthPlatform({AppLinks? links}) : _links = links ?? AppLinks();
  final AppLinks _links;

  @override
  Future<bool> openBrowser(Uri url) =>
      launchUrl(url, mode: LaunchMode.externalApplication);

  @override
  Stream<Uri> get links => _links.uriLinkStream;
}

/// RFC 7636 code verifier: 32 random bytes as 43 unpadded base64url characters.
String generatePkceVerifier() {
  final random = Random.secure();
  return _unpadded(
    base64Url.encode(List<int>.generate(32, (_) => random.nextInt(256))),
  );
}

/// RFC 7636 S256 challenge for [verifier].
String pkceChallenge(String verifier) =>
    _unpadded(base64Url.encode(sha256.convert(utf8.encode(verifier)).bytes));

String _unpadded(String value) => value.replaceAll('=', '');
