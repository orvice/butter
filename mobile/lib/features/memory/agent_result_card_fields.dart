import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';
import '../agents/agent_draft.dart';
import 'agent_result_cards.dart';

class AgentResultCardFields extends StatelessWidget {
  const AgentResultCardFields({
    super.key,
    required this.agent,
    required this.values,
    required this.enabled,
    required this.onChanged,
  });

  final Agent agent;
  final AgentResultCardsFormValues values;
  final bool enabled;
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    if (!supportsResultCards(agent)) return const SizedBox.shrink();
    final unknownGen = agent.hasConfig() && agent.config.hasResultCards()
        ? unknownEnum(agent.config.resultCards, 1)
        : null;
    final unknownPres = agent.hasConfig() && agent.config.hasResultCards()
        ? unknownEnum(agent.config.resultCards, 2)
        : null;
    if (unknownGen != null || unknownPres != null) {
      return const ListTile(
        title: Text('Result cards'),
        subtitle: Text(
          'This agent uses a Card Policy value this app does not edit. It is kept unchanged on save.',
        ),
      );
    }
    void set(void Function(AgentResultCardsFormValues v) fn) {
      fn(values);
      onChanged();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Result cards', style: Theme.of(context).textTheme.titleMedium),
        const Text(
          'Card Policy narrows what child agents may render via render_ui. Presentation is a hint only.',
        ),
        SegmentedButton<CardGenerationChoice>(
          segments: const [
            ButtonSegment(
              value: CardGenerationChoice.allowed,
              label: Text('Allowed'),
            ),
            ButtonSegment(
              value: CardGenerationChoice.off,
              label: Text('Disabled'),
            ),
          ],
          selected: {values.generation},
          onSelectionChanged: enabled
              ? (s) => set((v) => v.generation = s.first)
              : null,
        ),
        SegmentedButton<CardPresentationChoice>(
          segments: const [
            ButtonSegment(
              value: CardPresentationChoice.inherit,
              label: Text('Inherit'),
            ),
            ButtonSegment(
              value: CardPresentationChoice.auto,
              label: Text('Auto'),
            ),
            ButtonSegment(
              value: CardPresentationChoice.preferred,
              label: Text('Preferred'),
            ),
          ],
          selected: {values.presentation},
          onSelectionChanged: enabled
              ? (s) => set((v) => v.presentation = s.first)
              : null,
        ),
      ],
    );
  }
}
