//
//  Generated code. Do not modify.
//  source: agents/v1/agent_file.proto
//

import "package:connectrpc/connect.dart" as connect;
import "agent_file.pb.dart" as agentsv1agent_file;
import "agent_file.connect.spec.dart" as specs;

extension type AgentFileServiceClient (connect.Transport _transport) {
  Future<agentsv1agent_file.ListAgentFileSpacesResponse> listAgentFileSpaces(
    agentsv1agent_file.ListAgentFileSpacesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.listAgentFileSpaces,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.GetAgentFileSpaceResponse> getAgentFileSpace(
    agentsv1agent_file.GetAgentFileSpaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.getAgentFileSpace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.CreateAgentFileSpaceResponse> createAgentFileSpace(
    agentsv1agent_file.CreateAgentFileSpaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.createAgentFileSpace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.UpdateAgentFileSpaceResponse> updateAgentFileSpace(
    agentsv1agent_file.UpdateAgentFileSpaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.updateAgentFileSpace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.DeleteAgentFileSpaceResponse> deleteAgentFileSpace(
    agentsv1agent_file.DeleteAgentFileSpaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.deleteAgentFileSpace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.ListAgentFilesResponse> listAgentFiles(
    agentsv1agent_file.ListAgentFilesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.listAgentFiles,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.GetAgentFileResponse> getAgentFile(
    agentsv1agent_file.GetAgentFileRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.getAgentFile,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.WriteAgentFileResponse> writeAgentFile(
    agentsv1agent_file.WriteAgentFileRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.writeAgentFile,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.DeleteAgentFileResponse> deleteAgentFile(
    agentsv1agent_file.DeleteAgentFileRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.deleteAgentFile,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  Future<agentsv1agent_file.SearchAgentFilesResponse> searchAgentFiles(
    agentsv1agent_file.SearchAgentFilesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.AgentFileService.searchAgentFiles,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
