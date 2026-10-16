import 'package:butter_mobile/features/a2ui/form_logic.dart';
import 'package:butter_mobile/features/a2ui/protocol.dart';
import 'package:flutter_test/flutter_test.dart';

FormView sampleForm() => FormView(
  interruptId: 'i1',
  token: 'tok',
  revision: 1,
  title: 'Approve',
  question: 'Go?',
  fields: [
    FormField(
      name: 'env',
      label: 'Env',
      type: 'single_choice',
      required: true,
      options: [
        FormOption(value: 'prod', label: 'Production'),
        FormOption(value: 'staging', label: 'Staging'),
      ],
    ),
    FormField(name: 'reason', label: 'Reason', type: 'text', maxLength: 5),
  ],
);

void main() {
  test('validateForm catches required, length and choice rules', () {
    final form = sampleForm();
    expect(validateForm(form, {'env': '', 'reason': ''}), {'env': 'Required'});
    expect(
      validateForm(form, {'env': 'prod', 'reason': '123456'})['reason'],
      contains('At most 5'),
    );
    expect(validateForm(form, {'env': 'nope', 'reason': 'ok'}), {
      'env': 'Choose one of the options',
    });
    expect(validateForm(form, {'env': 'prod', 'reason': 'ok'}), isEmpty);
  });

  test('submissionPayload carries butterForm for one resume entry', () {
    final form = sampleForm();
    final payload = submissionPayload('form-1', form, {
      'env': 'prod',
      'reason': 'hotfix',
    });
    expect(payload['butterForm'], isA<Map>());
    final bf = payload['butterForm'] as Map;
    expect(bf['surfaceId'], 'form-1');
    expect(bf['token'], 'tok');
    expect(bf['revision'], 1);
    expect((bf['values'] as Map)['env'], 'prod');
  });
}
