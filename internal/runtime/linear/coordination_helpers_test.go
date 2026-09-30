package linear

import "github.com/google/uuid"

// Steal simulates another holder taking the session: the current holder's
// context is cancelled and its lease is gone.
func (c *MemoryCoordinator) Steal(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	if s.cancel != nil {
		s.cancel()
	}
	s.holder = "stolen-" + uuid.NewString()
	s.cancel = nil
}
