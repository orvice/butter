import 'package:butter_mobile/features/telegram/telegram_draft.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('rejects non-decimal chat id on create', () {
    final draft = TelegramDestinationDraft.create(channelId: 'ch-1')
      ..key = 'topic'
      ..chatId = 'not-a-number'
      ..agentId = 'agent-1';
    expect(draft.validate(ownsModel: false), contains('decimal'));
  });

  test('builds destination with immutable address fields', () {
    final draft = TelegramDestinationDraft.create(channelId: 'ch-1')
      ..key = 'main'
      ..chatId = '-1005550001'
      ..messageThreadId = '77'
      ..agentId = 'tg-agent';
    expect(draft.validate(ownsModel: false), isNull);
    final dest = draft.buildCreate(ownsModel: false);
    expect(dest.config.agentId, 'tg-agent');
    expect(dest.chatId, '-1005550001');
    expect(dest.messageThreadId, '77');
    expect(dest.channelId, 'ch-1');
  });

  test('channel create requires token', () {
    final draft = TelegramChannelDraft.create()
      ..key = 'bot'
      ..name = 'Bot';
    expect(draft.validateCreate(), contains('token'));
  });

  test('decimal id helper accepts negative chat ids', () {
    expect(isTelegramDecimalId('-1001', allowNegative: true), isTrue);
    expect(isTelegramDecimalId('abc', allowNegative: true), isFalse);
  });
}
