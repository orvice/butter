package a2ui

import (
	"time"

	"google.golang.org/adk/v2/session"
)

// Render serializes card validation/reservation across all leaves in this run.
// write stages a validated card in the adapter's event delta; persistence still
// belongs to the runner. A failed write leaves the shared view unchanged.
func (r *Run) Render(state session.ReadonlyState, batch Batch, invocationID string, write func(*Card) error) (Result, error) {
	r.cardsMu.Lock()
	defer r.cardsMu.Unlock()
	if r.cards == nil {
		r.cards = Cards(state)
	}
	res, err := Apply(r.cards, batch, Anchor{
		ThreadID: r.ThreadID, RunID: r.RunID, MessageID: r.MessageID,
		InvocationID: invocationID, At: time.Now().UTC(),
	}, DefaultLimits)
	if err != nil {
		return Result{}, err
	}
	if err := write(res.Card); err != nil {
		return Result{}, err
	}
	r.cards[res.Card.ID] = res.Card
	return res, nil
}
