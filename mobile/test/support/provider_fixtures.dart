import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';

/// A model row with an unknown varint field 15, as a newer server might send.
ModelConfig modelWithUnknown(
  String name, {
  String alias = '',
  int window = 0,
  int marker = 3,
}) => ModelConfig.fromBuffer([
  ...ModelConfig(
    name: name,
    alias: alias,
    contextWindowTokens: window,
  ).writeToBuffer(),
  0x78,
  marker,
]);

/// A provider with an unknown varint field 99 and rows carrying field 15.
ModelProvider providerWithUnknown(
  String name, {
  String apiKey = 'sk-secret-value',
  String baseUrl = 'https://llm.example/v1',
  List<ModelConfig>? models,
}) => ModelProvider.fromBuffer([
  ...ModelProvider(
    name: name,
    type: 'openai',
    apiKey: apiKey,
    baseUrl: baseUrl,
    workspaceId: 'alpha',
    models:
        models ??
        [
          modelWithUnknown('gpt-4o', alias: 'fast', window: 128000, marker: 1),
          modelWithUnknown('gpt-mini', marker: 2),
          modelWithUnknown('gpt-nano', alias: 'tiny', marker: 3),
        ],
  ).writeToBuffer(),
  0x98,
  0x06,
  7,
]);

/// The value of unknown varint [field], or null when it is absent.
int? unknownVarint(dynamic message, int field) {
  final set = message.unknownFields;
  if (!set.hasField(field)) return null;
  return set.getField(field).varints.single.toInt();
}
