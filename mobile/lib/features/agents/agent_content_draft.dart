import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/repobinding.pb.dart';

/// In-memory edit of the three Git-managed Agent Content text fields.
class AgentContentDraft {
  AgentContentDraft.fromAgent(
    this.agentId,
    Agent agent, {
    this.baseRevision = '',
  }) : description = agent.description,
       prompt = agent.config.instruction,
       globalPrompt = agent.config.globalInstruction,
       _baselineDescription = agent.description,
       _baselinePrompt = agent.config.instruction,
       _baselineGlobal = agent.config.globalInstruction;

  final String agentId;
  String description;
  String prompt;
  String globalPrompt;

  /// Cached Active Revision SHA when the draft was loaded (informational).
  final String baseRevision;

  String _baselineDescription;
  String _baselinePrompt;
  String _baselineGlobal;

  bool get dirty =>
      description != _baselineDescription ||
      prompt != _baselinePrompt ||
      globalPrompt != _baselineGlobal;

  void reconcileFrom(Agent agent, {String revision = ''}) {
    description = agent.description;
    prompt = agent.config.instruction;
    globalPrompt = agent.config.globalInstruction;
  }

  /// File actions for [CommitAgentContent], empty when nothing changed.
  List<ContentFileAction> commitActions() {
    final actions = <ContentFileAction>[];
    void add(String file, String value, String baseline) {
      if (value == baseline) return;
      actions.add(
        ContentFileAction(
          path: 'agents/$agentId/$file',
          content: value,
          operation: value.isEmpty
              ? ContentFileOperation.CONTENT_FILE_OPERATION_DELETE
              : ContentFileOperation.CONTENT_FILE_OPERATION_PUT,
        ),
      );
    }

    add('description.md', description, _baselineDescription);
    add('prompt.md', prompt, _baselinePrompt);
    add('global-prompt.md', globalPrompt, _baselineGlobal);
    return actions;
  }

  void markSaved() {
    _baselineDescription = description;
    _baselinePrompt = prompt;
    _baselineGlobal = globalPrompt;
  }
}
