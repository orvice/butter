import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/model_providers/provider_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:connectrpc/connect.dart' show Code, ConnectException;
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

Matcher connectCode(Code code) =>
    isA<ConnectException>().having((e) => e.code, 'code', code);

/// POSTs Connect JSON by hand, so the header can be left out.
Future<(int, String)> rawList(
  String base,
  String token, {
  String? workspace,
}) async {
  final client = HttpClient();
  try {
    final request = await client.postUrl(
      Uri.parse('$base/api/agents.v1.ModelProviderService/ListModelProviders'),
    );
    request.headers
      ..contentType = ContentType.json
      ..set('Authorization', 'Bearer $token');
    if (workspace != null) request.headers.set('X-Workspace-ID', workspace);
    request.write('{}');
    final response = await request.close();
    return (response.statusCode, await utf8.decodeStream(response));
  } finally {
    client.close(force: true);
  }
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Model providers against the real Butter service at "$prefix"',
      skip: server.isEmpty
          ? 'Run tool/contract.sh against the real Go handlers.'
          : false,
      () {
        final base = '$server$prefix';
        // Each prefix run owns its names, so a rerun against one server is
        // not confused by the other's data.
        final tag = prefix.isEmpty ? 'root' : 'prefix';
        late ButterApi api;
        late String token;

        setUp(() async {
          api = ButterApi(ServerAddress.parse(base, allowHttp: true));
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });

        test(
          'create, read, edit and delete keep fields and credential semantics',
          () async {
            final name = 'crud-$tag';
            final created = await api.saveModelProvider(
              token,
              'alpha',
              buildModelProvider(
                name: name,
                type: 'gemini',
                baseUrl: ' https://llm.example/v1 ',
                key: KeyAction.replace,
                newKey: ' sk-first ',
                rows: const [
                  ModelRowInput(
                    name: 'gem-a',
                    alias: 'fast',
                    contextWindow: '1000',
                  ),
                  ModelRowInput(name: 'gem-b', alias: '', contextWindow: ''),
                ],
              ),
              create: true,
            );
            expect(created.workspaceId, 'alpha');
            expect(created.type, 'gemini');
            expect(created.baseUrl, 'https://llm.example/v1');
            expect(created.apiKey, 'sk-first');

            final listed = await api.modelProviders(token, 'alpha');
            expect(listed.map((p) => p.name), contains(name));
            final read = await api.modelProvider(token, 'alpha', name);
            expect(read.models.map((m) => m.name), ['gem-a', 'gem-b']);
            expect(read.models[0].alias, 'fast');
            expect(read.models[0].contextWindowTokens, 1000);
            // Absent and 0 are the same to the server: no limit set.
            expect(read.models[1].contextWindowTokens, 0);

            // Keep: the key and the untouched model stay exactly as stored.
            final kept = await api.saveModelProvider(
              token,
              'alpha',
              buildModelProvider(
                base: read,
                type: read.type,
                baseUrl: read.baseUrl,
                key: KeyAction.keep,
                rows: [
                  ModelRowInput(
                    original: read.models[0],
                    name: 'gem-a',
                    alias: 'quick',
                    contextWindow: '2000',
                  ),
                  ModelRowInput(
                    original: read.models[1],
                    name: 'gem-b',
                    alias: '',
                    contextWindow: '',
                  ),
                ],
              ),
              create: false,
            );
            expect(kept.apiKey, 'sk-first');
            expect(kept.models[0].alias, 'quick');
            expect(kept.models[0].contextWindowTokens, 2000);
            expect(kept.models[1], read.models[1]);

            // Replace, then clear.
            final replaced = await api.saveModelProvider(
              token,
              'alpha',
              buildModelProvider(
                base: kept,
                type: kept.type,
                baseUrl: kept.baseUrl,
                key: KeyAction.replace,
                newKey: 'sk-second',
                rows: [
                  for (final m in kept.models)
                    ModelRowInput(
                      original: m,
                      name: m.name,
                      alias: m.alias,
                      contextWindow: m.contextWindowTokens == 0
                          ? ''
                          : '${m.contextWindowTokens}',
                    ),
                ],
              ),
              create: false,
            );
            expect(replaced.apiKey, 'sk-second');
            final cleared = await api.saveModelProvider(
              token,
              'alpha',
              buildModelProvider(
                base: replaced,
                type: replaced.type,
                baseUrl: '',
                key: KeyAction.clear,
                rows: [
                  for (final m in replaced.models)
                    ModelRowInput(
                      original: m,
                      name: m.name,
                      alias: m.alias,
                      contextWindow: m.contextWindowTokens == 0
                          ? ''
                          : '${m.contextWindowTokens}',
                    ),
                ],
              ),
              create: false,
            );
            expect(cleared.apiKey, isEmpty);
            expect(cleared.baseUrl, isEmpty);
            expect(cleared.models.map((m) => m.name), ['gem-a', 'gem-b']);

            await api.deleteModelProvider(token, 'alpha', name);
            await expectLater(
              api.modelProvider(token, 'alpha', name),
              throwsA(connectCode(Code.notFound)),
            );
            await expectLater(
              api.deleteModelProvider(token, 'alpha', name),
              throwsA(connectCode(Code.notFound)),
            );
            await expectLater(
              api.saveModelProvider(
                token,
                'alpha',
                ModelProvider(name: name, type: 'openai'),
                create: false,
              ),
              throwsA(connectCode(Code.notFound)),
            );
          },
        );

        test('a duplicate name and a bad base URL are refused', () async {
          final name = 'dup-$tag';
          final provider = ModelProvider(
            name: name,
            type: 'openai',
            models: [ModelConfig(name: 'm')],
          );
          await api.saveModelProvider(token, 'alpha', provider, create: true);
          addTearDown(() => api.deleteModelProvider(token, 'alpha', name));
          await expectLater(
            api.saveModelProvider(token, 'alpha', provider, create: true),
            throwsA(connectCode(Code.alreadyExists)),
          );
          await expectLater(
            api.saveModelProvider(
              token,
              'alpha',
              provider.deepCopy()..baseUrl = 'ftp://nope',
              create: false,
            ),
            throwsA(connectCode(Code.invalidArgument)),
          );
          expect(
            (await api.modelProvider(token, 'alpha', name)).baseUrl,
            isEmpty,
          );
        });

        test('providers belong to one Workspace', () async {
          final name = 'scoped-$tag';
          await api.saveModelProvider(
            token,
            'alpha',
            ModelProvider(
              name: name,
              type: 'openai',
              models: [ModelConfig(name: 'm')],
            ),
            create: true,
          );
          addTearDown(() => api.deleteModelProvider(token, 'alpha', name));
          expect(
            (await api.modelProviders(token, 'beta')).map((p) => p.name),
            isNot(contains(name)),
          );
          await expectLater(
            api.modelProvider(token, 'beta', name),
            throwsA(connectCode(Code.notFound)),
          );
          await expectLater(
            api.modelProviders(token, 'forbidden'),
            throwsA(connectCode(Code.permissionDenied)),
          );
          final (status, body) = await rawList(base, token);
          expect(status, 400); // Connect maps failed_precondition to 400
          expect(jsonDecode(body)['code'], 'failed_precondition');
          final (okStatus, _) = await rawList(base, token, workspace: 'alpha');
          expect(okStatus, 200);
        });

        test(
          'Telegram references block removal, and the App says which',
          () async {
            final name = 'routed-$tag';
            ModelProvider routed() => ModelProvider(
              name: name,
              type: 'openai',
              models: [
                ModelConfig(name: 'real-a', alias: 'guarded-default'),
                ModelConfig(name: 'real-b', alias: 'guarded-choice'),
                ModelConfig(name: 'free'),
              ],
            );
            final context = ClientContext(
              storage: MemoryStorage(),
              allowHttp: true,
            );
            addTearDown(context.dispose);
            await context.setServer(base);
            await context.login('mobile-test', 'contract-test-password');
            await context.selectWorkspace('alpha');
            expect(
              await context.saveModelProvider(routed(), create: true),
              true,
            );
            addTearDown(() async {
              // Beta is not referenced, so nothing blocks the cleanup there.
              await api.deleteModelProvider(token, 'beta', name);
            });

            await context.loadModelProvider(name);
            final stored = context.modelProvider!;
            final drop = buildModelProvider(
              base: stored,
              type: stored.type,
              baseUrl: stored.baseUrl,
              key: KeyAction.keep,
              rows: [
                for (final m in stored.models.skip(1))
                  ModelRowInput(
                    original: m,
                    name: m.name,
                    alias: m.alias,
                    contextWindow: '',
                  ),
              ],
            );
            expect(await context.saveModelProvider(drop, create: false), false);
            expect(context.error, contains('contract-destination'));
            expect(context.error, contains('guarded-default'));
            expect(context.error, startsWith('These models cannot be removed'));

            expect(await context.deleteModelProvider(name), false);
            expect(context.error, contains('contract-destination'));
            expect(
              context.error,
              startsWith('This provider cannot be deleted'),
            );
            expect(
              (await api.modelProvider(token, 'alpha', name)).models,
              hasLength(3),
            );

            // The same aliases in Workspace beta are unreferenced.
            await api.saveModelProvider(token, 'beta', routed(), create: true);
            await api.deleteModelProvider(token, 'beta', name);
            await api.saveModelProvider(token, 'beta', routed(), create: true);
          },
        );
      },
    );
  }
}
