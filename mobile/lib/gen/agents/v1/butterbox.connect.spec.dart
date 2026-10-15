//
//  Generated code. Do not modify.
//  source: agents/v1/butterbox.proto
//

import "package:connectrpc/connect.dart" as connect;
import "butterbox.pb.dart" as agentsv1butterbox;

/// ButterBoxService manages the workspace's registered ButterBoxes.
abstract final class ButterBoxService {
  /// Fully-qualified name of the ButterBoxService service.
  static const name = 'agents.v1.ButterBoxService';

  static const listButterBoxes = connect.Spec(
    '/$name/ListButterBoxes',
    connect.StreamType.unary,
    agentsv1butterbox.ListButterBoxesRequest.new,
    agentsv1butterbox.ListButterBoxesResponse.new,
  );

  static const getButterBox = connect.Spec(
    '/$name/GetButterBox',
    connect.StreamType.unary,
    agentsv1butterbox.GetButterBoxRequest.new,
    agentsv1butterbox.GetButterBoxResponse.new,
  );

  static const createButterBox = connect.Spec(
    '/$name/CreateButterBox',
    connect.StreamType.unary,
    agentsv1butterbox.CreateButterBoxRequest.new,
    agentsv1butterbox.CreateButterBoxResponse.new,
  );

  static const updateButterBox = connect.Spec(
    '/$name/UpdateButterBox',
    connect.StreamType.unary,
    agentsv1butterbox.UpdateButterBoxRequest.new,
    agentsv1butterbox.UpdateButterBoxResponse.new,
  );

  static const deleteButterBox = connect.Spec(
    '/$name/DeleteButterBox',
    connect.StreamType.unary,
    agentsv1butterbox.DeleteButterBoxRequest.new,
    agentsv1butterbox.DeleteButterBoxResponse.new,
  );

  /// SetButterBoxToken sets, rotates, or clears (empty token) the box access
  /// token. The token is write-only: encrypted at rest, never read back.
  static const setButterBoxToken = connect.Spec(
    '/$name/SetButterBoxToken',
    connect.StreamType.unary,
    agentsv1butterbox.SetButterBoxTokenRequest.new,
    agentsv1butterbox.SetButterBoxTokenResponse.new,
  );

  /// GetButterBoxStatus probes the box and reports reachability and the
  /// number of active pi sessions.
  static const getButterBoxStatus = connect.Spec(
    '/$name/GetButterBoxStatus',
    connect.StreamType.unary,
    agentsv1butterbox.GetButterBoxStatusRequest.new,
    agentsv1butterbox.GetButterBoxStatusResponse.new,
  );

  /// ListButterBoxModels reports the box's pi model catalog (the models pi is
  /// configured with on the box).
  static const listButterBoxModels = connect.Spec(
    '/$name/ListButterBoxModels',
    connect.StreamType.unary,
    agentsv1butterbox.ListButterBoxModelsRequest.new,
    agentsv1butterbox.ListButterBoxModelsResponse.new,
  );

  /// ListButterBoxCursorModels reports the box's Cursor model catalog (from
  /// the box's CursorService, which asks Cursor's API with the box's key).
  static const listButterBoxCursorModels = connect.Spec(
    '/$name/ListButterBoxCursorModels',
    connect.StreamType.unary,
    agentsv1butterbox.ListButterBoxCursorModelsRequest.new,
    agentsv1butterbox.ListButterBoxCursorModelsResponse.new,
  );
}
