import 'package:butter_mobile/core/server_address.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('one HTTPS server supports deployment prefixes and default scheme', () {
    expect(
      ServerAddress.parse(' butter.example.org/butter/ ').uri.toString(),
      'https://butter.example.org/butter',
    );
    expect(
      ServerAddress.parse('https://butter.example.org').endpoint('api'),
      Uri.parse('https://butter.example.org/api'),
    );
  });
  test('credentials and insecure or ambiguous addresses are rejected', () {
    for (final address in [
      '',
      'http://butter.example.org',
      'https://user:pass@butter.example.org',
      'https://butter.example.org?token=secret',
      'https://butter.example.org#x',
      'file:///etc/passwd',
      'https://butter.example.org/a/../b',
    ]) {
      expect(() => ServerAddress.parse(address), throwsFormatException);
    }
    expect(
      ServerAddress.parse('http://127.0.0.1:8080', allowHttp: true).uri.scheme,
      'http',
    );
  });
}
