package memory

import (
	"context"
	"time"
)

func (s *Store) ReserveBudget(_ context.Context, key string, spacing time.Duration) (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.Clock()
	if next := s.budgets[key]; next.After(now) {
		return next.Sub(now), nil
	}
	s.budgets[key] = now.Add(spacing)
	return 0, nil
}
func (s *Store) DeferBudget(_ context.Context, key string, delay time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	until := s.Clock().Add(delay)
	if until.After(s.budgets[key]) {
		s.budgets[key] = until
	}
	return nil
}
