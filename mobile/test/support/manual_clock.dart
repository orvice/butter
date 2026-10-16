import 'dart:async';

/// A clock and timers the test moves by hand, so a controller's polling and
/// deadlines are exercised without waiting and without flakiness.
///
/// [now] is what the controller under test reads as the time; [schedule] is its
/// timer factory. [advance] moves time forward and fires every due timer in
/// order. The callbacks may start real (fake-HTTP) requests: the test waits for
/// those itself, since only the clock is simulated.
class ManualClock {
  ManualClock(this.now);

  DateTime now;
  final _timers = <_ManualTimer>[];

  Timer schedule(Duration after, void Function() callback) {
    final timer = _ManualTimer(now.add(after), callback);
    _timers.add(timer);
    return timer;
  }

  /// Timers that are still going to fire.
  int get pending => _timers.where((t) => t.isActive).length;

  /// Moves time forward by [by], firing due timers in time order. A timer set
  /// by a callback is fired too when it falls due inside the span.
  Future<void> advance(Duration by) async {
    final target = now.add(by);
    while (true) {
      final due =
          _timers.where((t) => t.isActive && !t.at.isAfter(target)).toList()
            ..sort((a, b) => a.at.compareTo(b.at));
      if (due.isEmpty) break;
      final next = due.first;
      if (next.at.isAfter(now)) now = next.at;
      next.fire();
      // Let the callback run up to its first await.
      await Future<void>.delayed(Duration.zero);
    }
    now = target;
  }
}

class _ManualTimer implements Timer {
  _ManualTimer(this.at, this._callback);

  final DateTime at;
  final void Function() _callback;
  bool _active = true;
  int _ticks = 0;

  void fire() {
    if (!_active) return;
    _active = false;
    _ticks++;
    _callback();
  }

  @override
  void cancel() => _active = false;

  @override
  bool get isActive => _active;

  @override
  int get tick => _ticks;
}
