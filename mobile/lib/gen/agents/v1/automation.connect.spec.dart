//
//  Generated code. Do not modify.
//  source: agents/v1/automation.proto
//

import "package:connectrpc/connect.dart" as connect;
import "automation.pb.dart" as agentsv1automation;

abstract final class AutomationService {
  /// Fully-qualified name of the AutomationService service.
  static const name = 'agents.v1.AutomationService';

  static const listAutomations = connect.Spec(
    '/$name/ListAutomations',
    connect.StreamType.unary,
    agentsv1automation.ListAutomationsRequest.new,
    agentsv1automation.ListAutomationsResponse.new,
  );

  static const getAutomation = connect.Spec(
    '/$name/GetAutomation',
    connect.StreamType.unary,
    agentsv1automation.GetAutomationRequest.new,
    agentsv1automation.GetAutomationResponse.new,
  );

  static const createAutomation = connect.Spec(
    '/$name/CreateAutomation',
    connect.StreamType.unary,
    agentsv1automation.CreateAutomationRequest.new,
    agentsv1automation.CreateAutomationResponse.new,
  );

  static const updateAutomation = connect.Spec(
    '/$name/UpdateAutomation',
    connect.StreamType.unary,
    agentsv1automation.UpdateAutomationRequest.new,
    agentsv1automation.UpdateAutomationResponse.new,
  );

  static const deleteAutomation = connect.Spec(
    '/$name/DeleteAutomation',
    connect.StreamType.unary,
    agentsv1automation.DeleteAutomationRequest.new,
    agentsv1automation.DeleteAutomationResponse.new,
  );

  static const runAutomationNow = connect.Spec(
    '/$name/RunAutomationNow',
    connect.StreamType.unary,
    agentsv1automation.RunAutomationNowRequest.new,
    agentsv1automation.RunAutomationNowResponse.new,
  );

  static const listAutomationRuns = connect.Spec(
    '/$name/ListAutomationRuns',
    connect.StreamType.unary,
    agentsv1automation.ListAutomationRunsRequest.new,
    agentsv1automation.ListAutomationRunsResponse.new,
  );

  static const getAutomationRun = connect.Spec(
    '/$name/GetAutomationRun',
    connect.StreamType.unary,
    agentsv1automation.GetAutomationRunRequest.new,
    agentsv1automation.GetAutomationRunResponse.new,
  );

  static const listAutomationStepRuns = connect.Spec(
    '/$name/ListAutomationStepRuns',
    connect.StreamType.unary,
    agentsv1automation.ListAutomationStepRunsRequest.new,
    agentsv1automation.ListAutomationStepRunsResponse.new,
  );
}
