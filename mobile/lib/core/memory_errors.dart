import 'package:connectrpc/connect.dart';

enum MemoryOperation {
  loadConfig,
  saveConfig,
  deleteConfig,
  testConnection,
  listMemories,
  searchMemories,
  deleteMemory,
}

String memoryFailureMessage(Object failure, MemoryOperation operation) {
  if (failure is ConnectException) {
    final message = failure.message.toLowerCase();
    switch (failure.code) {
      case Code.permissionDenied:
        return 'You need the owner or admin role in this Workspace to change memory settings.';
      case Code.notFound:
        if (operation == MemoryOperation.deleteConfig) {
          return 'This Workspace has no memory connection.';
        }
        if (operation == MemoryOperation.deleteMemory) {
          return 'That memory was not found or belongs to another workspace.';
        }
        return 'Nothing matched that request. Refresh and try again.';
      case Code.failedPrecondition:
        if (message.contains('rejected') || message.contains('api key')) {
          return 'The mem0 server rejected the API key. Check the key and try again.';
        }
        if (message.contains('not configured') ||
            message.contains('connect a mem0')) {
          return 'Connect an enabled mem0 server in Memory settings first.';
        }
        if (message.contains('encryption')) {
          return 'The server cannot store API keys right now.';
        }
        return 'The server refused this change. Check the form and try again.';
      case Code.invalidArgument:
        return 'The server rejected these values. Check the form and try again.';
      case Code.unavailable:
        return 'The mem0 server did not answer. Try again later.';
      case Code.unauthenticated:
        return 'Session expired. Sign in again.';
      default:
        break;
    }
  }
  if (failure is FormatException) return failure.message;
  return switch (operation) {
    MemoryOperation.loadConfig => 'Memory settings could not be loaded.',
    MemoryOperation.saveConfig => 'Memory settings were not saved.',
    MemoryOperation.deleteConfig => 'The memory connection was not removed.',
    MemoryOperation.testConnection => 'The connection test did not complete.',
    MemoryOperation.listMemories => 'Memories could not be loaded.',
    MemoryOperation.searchMemories => 'Search did not complete.',
    MemoryOperation.deleteMemory => 'The memory was not deleted.',
  };
}

/// Maps a mem0 probe outcome to a safe message. Raw server text is never shown.
String classifyMem0ProbeError(String raw) {
  final lower = raw.toLowerCase();
  if (lower.contains('401') ||
      lower.contains('403') ||
      lower.contains('unauthorized') ||
      lower.contains('invalid api key') ||
      lower.contains('rejected')) {
    return 'The mem0 server rejected the API key.';
  }
  if (lower.contains('timeout') ||
      lower.contains('connection refused') ||
      lower.contains('no such host') ||
      lower.contains('did not answer')) {
    return 'Could not reach the mem0 server.';
  }
  return 'The connection check failed.';
}
