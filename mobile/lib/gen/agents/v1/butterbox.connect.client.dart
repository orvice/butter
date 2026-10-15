//
//  Generated code. Do not modify.
//  source: agents/v1/butterbox.proto
//

import "package:connectrpc/connect.dart" as connect;
import "butterbox.pb.dart" as agentsv1butterbox;
import "butterbox.connect.spec.dart" as specs;

/// ButterBoxService manages the workspace's registered ButterBoxes.
extension type ButterBoxServiceClient (connect.Transport _transport) {
  Future<agentsv1butterbox.ListButterBoxesResponse> listButterBoxes(
    agentsv1butterbox.ListButterBoxesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.listButterBoxes,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1butterbox.GetButterBoxResponse> getButterBox(
    agentsv1butterbox.GetButterBoxRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.getButterBox,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1butterbox.CreateButterBoxResponse> createButterBox(
    agentsv1butterbox.CreateButterBoxRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.createButterBox,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1butterbox.UpdateButterBoxResponse> updateButterBox(
    agentsv1butterbox.UpdateButterBoxRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.updateButterBox,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1butterbox.DeleteButterBoxResponse> deleteButterBox(
    agentsv1butterbox.DeleteButterBoxRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.deleteButterBox,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SetButterBoxToken sets, rotates, or clears (empty token) the box access
  /// token. The token is write-only: encrypted at rest, never read back.
  Future<agentsv1butterbox.SetButterBoxTokenResponse> setButterBoxToken(
    agentsv1butterbox.SetButterBoxTokenRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.setButterBoxToken,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetButterBoxStatus probes the box and reports reachability and the
  /// number of active pi sessions.
  Future<agentsv1butterbox.GetButterBoxStatusResponse> getButterBoxStatus(
    agentsv1butterbox.GetButterBoxStatusRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.getButterBoxStatus,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListButterBoxModels reports the box's pi model catalog (the models pi is
  /// configured with on the box).
  Future<agentsv1butterbox.ListButterBoxModelsResponse> listButterBoxModels(
    agentsv1butterbox.ListButterBoxModelsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.listButterBoxModels,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListButterBoxCursorModels reports the box's Cursor model catalog (from
  /// the box's CursorService, which asks Cursor's API with the box's key).
  Future<agentsv1butterbox.ListButterBoxCursorModelsResponse> listButterBoxCursorModels(
    agentsv1butterbox.ListButterBoxCursorModelsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ButterBoxService.listButterBoxCursorModels,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
