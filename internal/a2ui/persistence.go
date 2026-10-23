package a2ui

import (
	"context"
	"maps"
	"sync"

	"google.golang.org/adk/v2/session"
)

// PersistCards keeps reserved revisions monotonic when parallel leaves yield
// their events out of reservation order. Each card delta is a full snapshot,
// so a later revision subsumes an earlier one, which must not overwrite it.
func PersistCards(s session.Service) session.Service { return &cardService{Service: s} }

type cardService struct{ session.Service }
type persistedCards struct {
	mu        sync.Mutex
	revisions map[string]int
}

func (s *cardService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	run, ok := RunFrom(ctx)
	if !ok {
		return s.Service.AppendEvent(ctx, sess, event)
	}
	run.persisted.mu.Lock()
	defer run.persisted.mu.Unlock()
	if run.persisted.revisions == nil {
		run.persisted.revisions = make(map[string]int)
		for id, card := range Cards(sess.State()) {
			run.persisted.revisions[id] = card.Revision
		}
	}
	delta := maps.Clone(event.Actions.StateDelta)
	for key, value := range delta {
		id, cardKey := CardIDFromKey(key)
		if cardKey {
			if card, valid := DecodeCard(value); valid && card.Revision <= run.persisted.revisions[id] {
				delete(delta, key)
			}
		}
	}
	event.Actions.StateDelta = delta
	if err := s.Service.AppendEvent(ctx, sess, event); err != nil {
		return err
	}
	for key, value := range delta {
		if id, ok := CardIDFromKey(key); ok {
			if card, valid := DecodeCard(value); valid {
				run.persisted.revisions[id] = card.Revision
			}
		}
	}
	return nil
}
