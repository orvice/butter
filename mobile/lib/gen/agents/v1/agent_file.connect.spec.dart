//
//  Generated code. Do not modify.
//  source: agents/v1/agent_file.proto
//

import "package:connectrpc/connect.dart" as connect;
import "agent_file.pb.dart" as agentsv1agent_file;

abstract final class AgentFileService {
  /// Fully-qualified name of the AgentFileService service.
  static const name = 'agents.v1.AgentFileService';

  static const listAgentFileSpaces = connect.Spec(
    '/$name/ListAgentFileSpaces',
    connect.StreamType.unary,
    agentsv1agent_file.ListAgentFileSpacesRequest.new,
    agentsv1agent_file.ListAgentFileSpacesResponse.new,
  );

  static const getAgentFileSpace = connect.Spec(
    '/$name/GetAgentFileSpace',
    connect.StreamType.unary,
    agentsv1agent_file.GetAgentFileSpaceRequest.new,
    agentsv1agent_file.GetAgentFileSpaceResponse.new,
  );

  static const createAgentFileSpace = connect.Spec(
    '/$name/CreateAgentFileSpace',
    connect.StreamType.unary,
    agentsv1agent_file.CreateAgentFileSpaceRequest.new,
    agentsv1agent_file.CreateAgentFileSpaceResponse.new,
  );

  static const updateAgentFileSpace = connect.Spec(
    '/$name/UpdateAgentFileSpace',
    connect.StreamType.unary,
    agentsv1agent_file.UpdateAgentFileSpaceRequest.new,
    agentsv1agent_file.UpdateAgentFileSpaceResponse.new,
  );

  static const deleteAgentFileSpace = connect.Spec(
    '/$name/DeleteAgentFileSpace',
    connect.StreamType.unary,
    agentsv1agent_file.DeleteAgentFileSpaceRequest.new,
    agentsv1agent_file.DeleteAgentFileSpaceResponse.new,
  );

  static const listAgentFiles = connect.Spec(
    '/$name/ListAgentFiles',
    connect.StreamType.unary,
    agentsv1agent_file.ListAgentFilesRequest.new,
    agentsv1agent_file.ListAgentFilesResponse.new,
  );

  static const getAgentFile = connect.Spec(
    '/$name/GetAgentFile',
    connect.StreamType.unary,
    agentsv1agent_file.GetAgentFileRequest.new,
    agentsv1agent_file.GetAgentFileResponse.new,
  );

  static const writeAgentFile = connect.Spec(
    '/$name/WriteAgentFile',
    connect.StreamType.unary,
    agentsv1agent_file.WriteAgentFileRequest.new,
    agentsv1agent_file.WriteAgentFileResponse.new,
  );

  static const deleteAgentFile = connect.Spec(
    '/$name/DeleteAgentFile',
    connect.StreamType.unary,
    agentsv1agent_file.DeleteAgentFileRequest.new,
    agentsv1agent_file.DeleteAgentFileResponse.new,
  );

  static const searchAgentFiles = connect.Spec(
    '/$name/SearchAgentFiles',
    connect.StreamType.unary,
    agentsv1agent_file.SearchAgentFilesRequest.new,
    agentsv1agent_file.SearchAgentFilesResponse.new,
  );
}
