import 'package:butter_mobile/features/daemons/daemon_pages.dart';
import 'package:flutter_test/flutter_test.dart';

// The worker token's expiry is typed in hours, like the Web dashboard's, and
// only text that clearly means a positive length of time becomes a lifetime.
// Blank is the one way to ask for a token that does not expire.
void main() {
  Duration? ok(String text) {
    final r = parseExpiryHours(text);
    expect(r.error, isNull, reason: '"$text"');
    return r.ttl;
  }

  void refused(String text) {
    final r = parseExpiryHours(text);
    expect(r.error, isNotNull, reason: '"$text" must be refused');
    expect(r.ttl, isNull, reason: '"$text" must not become any lifetime');
  }

  test('blank means no expiry', () {
    expect(ok(''), isNull);
    expect(ok('   '), isNull);
  });

  test('one hour, twelve hours and fractions of an hour', () {
    expect(ok('1'), const Duration(hours: 1));
    expect(ok('12'), const Duration(hours: 12));
    expect(ok('0.5'), const Duration(minutes: 30));
    expect(ok('.25'), const Duration(minutes: 15));
    expect(ok('1.5'), const Duration(hours: 1, minutes: 30));
    expect(ok(' 2 '), const Duration(hours: 2));
    expect(ok('24'), const Duration(days: 1));
    expect(ok('720'), const Duration(days: 30));
    expect(ok('0.1'), const Duration(minutes: 6));
    expect(ok('0.0003'), const Duration(milliseconds: 1080));
    expect(ok('5.'), const Duration(hours: 5));
  });

  test('a lifetime over a year is allowed: there is no day-only limit', () {
    expect(ok('87600'), const Duration(hours: 87600));
    expect(
      ok('$maxTokenLifetimeHours'),
      const Duration(hours: maxTokenLifetimeHours),
    );
  });

  test('zero, negative and non-numbers are refused, never long-lived', () {
    for (final text in [
      '0',
      '0.0',
      '.0',
      '-1',
      '-0.5',
      '+1',
      'NaN',
      'nan',
      'Infinity',
      '-Infinity',
      'inf',
      '1e3',
      '1E3',
      '0x10',
      '1,5',
      '1 2',
      '1h',
      'one',
      '.',
      '..5',
      '1.2.3',
      '--1',
      '١٢',
    ]) {
      refused(text);
    }
  });

  test('a lifetime the server cannot hold, or that rounds to nothing', () {
    refused('${maxTokenLifetimeHours + 1}');
    refused('99999999999999999999999');
    refused('9' * 400);
    refused('0.0000000001');
  });
}
