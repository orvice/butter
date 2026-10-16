/// Side effects of a followed run besides its visible reply (#462 hooks here).
abstract class RunEffects {
  const RunEffects();

  /// Each `CUSTOM butter.a2ui` envelope, in order.
  void a2ui(Object? value);

  /// Shared AG-UI state after snapshot or delta.
  void state(Map<String, dynamic>? next);
}

/// No-op effects until A2UI is wired (#462).
final class NullRunEffects extends RunEffects {
  const NullRunEffects();
  @override
  void a2ui(Object? value) {}
  @override
  void state(Map<String, dynamic>? next) {}
}
