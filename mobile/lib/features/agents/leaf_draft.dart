import '../../gen/agents/v1/agent.pb.dart';

/// How a box leaf bounds one turn's run. The three are different on the wire:
/// absent (the server's default, currently 1800 seconds), an explicit 0
/// (unlimited), and an explicit positive number.
enum RunLimitMode { serverDefault, unlimited, seconds }

const maxRunSecondsLimit = 2147483647;

/// pi's thinking levels, as the server documents them. A stored value outside
/// this list is retained and shown, never rewritten.
const piThinkingLevels = [
  'off',
  'minimal',
  'low',
  'medium',
  'high',
  'xhigh',
  'max',
];

/// Cursor modes the server accepts. Empty means the default (agent).
const cursorModes = ['agent', 'plan'];

RunLimitMode runLimitModeOf({required bool present, required int seconds}) =>
    !present
    ? RunLimitMode.serverDefault
    : seconds == 0
    ? RunLimitMode.unlimited
    : RunLimitMode.seconds;

RunLimitMode piRunLimit(PiAgentConfig c) =>
    runLimitModeOf(present: c.hasMaxRunSeconds(), seconds: c.maxRunSeconds);

RunLimitMode cursorRunLimit(CursorAgentConfig c) =>
    runLimitModeOf(present: c.hasMaxRunSeconds(), seconds: c.maxRunSeconds);

/// Writes the chosen limit into whichever config holds it. [seconds] is used
/// only for [RunLimitMode.seconds].
void applyRunLimit(
  RunLimitMode mode,
  int seconds, {
  required void Function() clear,
  required void Function(int) set,
}) {
  switch (mode) {
    case RunLimitMode.serverDefault:
      clear();
    case RunLimitMode.unlimited:
      set(0);
    case RunLimitMode.seconds:
      set(seconds);
  }
}

/// A positive whole number of seconds. 0 is "unlimited", chosen explicitly.
int? parseRunSeconds(String? value) {
  final text = value?.trim() ?? '';
  if (!RegExp(r'^\d+$').hasMatch(text)) return null;
  final n = int.tryParse(text);
  return n == null || n < 1 || n > maxRunSecondsLimit ? null : n;
}

String? validateRunSeconds(String? value) => parseRunSeconds(value) == null
    ? 'Use a whole number of seconds from 1 to $maxRunSecondsLimit, or choose Unlimited.'
    : null;
