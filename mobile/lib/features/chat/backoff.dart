import 'dart:async';

/// Paces thread reads while waiting a run out: 1s, 2s, 4s, then 5s (Web parity).
const pollDelaysMs = [1000, 2000, 4000, 5000];

/// Backoff spaces out polling reads; [hurry] makes the next read immediate.
class Backoff {
  var _attempt = 0;
  var _hurried = false;
  void Function()? _cut;

  Future<void> wait({required bool Function() live}) async {
    if (!live()) return;
    if (_hurried) {
      _hurried = false;
      return;
    }
    final delay = pollDelaysMs[_attempt.clamp(0, pollDelaysMs.length - 1)];
    _attempt++;
    final done = Completer<void>();
    final timer = Timer(Duration(milliseconds: delay), () {
      _cut = null;
      if (!done.isCompleted) done.complete();
    });
    _cut = () {
      timer.cancel();
      if (!done.isCompleted) done.complete();
    };
    await done.future;
  }

  void hurry() {
    _attempt = 0;
    if (_cut != null) {
      _cut!();
    } else {
      _hurried = true;
    }
  }
}
