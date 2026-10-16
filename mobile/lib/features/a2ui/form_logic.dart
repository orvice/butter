import 'protocol.dart';

/// One string per configured field, as the server expects.
Map<String, String> formValues(FormView form, Map<String, Object?> draft) {
  final out = <String, String>{};
  for (final field in form.fields) {
    final raw = draft[field.name];
    if (raw is List) {
      out[field.name] = raw.isNotEmpty && raw.first is String
          ? raw.first as String
          : '';
    } else if (raw is String) {
      out[field.name] = field.type == 'text' ? raw.trim() : raw;
    } else {
      out[field.name] = '';
    }
  }
  return out;
}

/// Client-side validation mirroring the server's field rules.
Map<String, String> validateForm(FormView form, Map<String, String> values) {
  final errors = <String, String>{};
  for (final field in form.fields) {
    final value = values[field.name] ?? '';
    if (value.isEmpty) {
      if (field.required) errors[field.name] = 'Required';
      continue;
    }
    if (field.type == 'text' && field.maxLength != null) {
      final length = value.runes.length;
      if (length > field.maxLength!) {
        errors[field.name] =
            'At most ${field.maxLength} characters ($length now)';
      }
    }
    if (field.type == 'single_choice' &&
        !field.options.any((o) => o.value == value)) {
      errors[field.name] = 'Choose one of the options';
    }
  }
  return errors;
}

Map<String, dynamic> submissionPayload(
  String surfaceId,
  FormView form,
  Map<String, String> values,
) {
  return {
    'butterForm': {
      'version': a2uiVersion,
      'surfaceId': surfaceId,
      'revision': form.revision,
      'token': form.token,
      'values': values,
    },
  };
}

String readableReply(FormView form, Map<String, String> values) {
  final lines = [form.title];
  for (final field in form.fields) {
    final value = values[field.name] ?? '';
    final display = value.isEmpty
        ? '—'
        : field.type == 'single_choice'
        ? () {
            for (final o in field.options) {
              if (o.value == value) return o.label;
            }
            return value;
          }()
        : value;
    lines.add('${field.label}: $display');
  }
  return lines.join('\n');
}
