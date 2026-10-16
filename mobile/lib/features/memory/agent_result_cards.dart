import '../../gen/agents/v1/agent.pb.dart';
import '../agents/agent_draft.dart';

enum CardGenerationChoice { allowed, off }

enum CardPresentationChoice { inherit, auto, preferred }

class AgentResultCardsFormValues {
  AgentResultCardsFormValues({
    this.generation = CardGenerationChoice.allowed,
    this.presentation = CardPresentationChoice.inherit,
  });

  CardGenerationChoice generation;
  CardPresentationChoice presentation;
}

AgentResultCardsFormValues resultCardsFromConfig(ResultCardConfig? config) {
  if (config == null) return AgentResultCardsFormValues();
  return AgentResultCardsFormValues(
    generation:
        config.generation ==
            ResultCardGeneration.RESULT_CARD_GENERATION_DISABLED
        ? CardGenerationChoice.off
        : CardGenerationChoice.allowed,
    presentation: switch (config.presentation) {
      ResultCardPresentation.RESULT_CARD_PRESENTATION_AUTO =>
        CardPresentationChoice.auto,
      ResultCardPresentation.RESULT_CARD_PRESENTATION_PREFERRED =>
        CardPresentationChoice.preferred,
      _ => CardPresentationChoice.inherit,
    },
  );
}

bool supportsResultCards(Agent agent) => !isBoxLeaf(agent);

void applyResultCardsForm(AgentDraft draft, AgentResultCardsFormValues values) {
  if (!supportsResultCards(draft.agent)) {
    if (draft.agent.hasConfig()) draft.config.clearResultCards();
    return;
  }
  final rc = ResultCardConfig();
  var any = false;
  if (values.generation == CardGenerationChoice.off) {
    rc.generation = ResultCardGeneration.RESULT_CARD_GENERATION_DISABLED;
    any = true;
  }
  switch (values.presentation) {
    case CardPresentationChoice.inherit:
      break;
    case CardPresentationChoice.auto:
      rc.presentation = ResultCardPresentation.RESULT_CARD_PRESENTATION_AUTO;
      any = true;
    case CardPresentationChoice.preferred:
      rc.presentation =
          ResultCardPresentation.RESULT_CARD_PRESENTATION_PREFERRED;
      any = true;
  }
  if (!any) {
    draft.config.clearResultCards();
  } else {
    draft.config.resultCards = rc;
  }
}
