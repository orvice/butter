package plane

import (
	"context"
	"testing"
)

type ownerCloser struct {
	events *[]string
	name   string
}

func (c ownerCloser) Close() error { *c.events = append(*c.events, c.name); return nil }
func TestPlaneResourceOwnerRollbackAndCloseOrder(t *testing.T) {
	events := []string{}
	o := NewPlaneResourceOwner(nil, ownerCloser{&events, "mongo"}, ownerCloser{&events, "redis"})
	if e := o.Start(context.Background()); e == nil {
		t.Fatal("expected startup failure")
	}
	if len(events) != 2 || events[0] != "redis" || events[1] != "mongo" {
		t.Fatalf("rollback order=%v", events)
	}
	o.Stop()
	o.Stop()
}
func TestPlaneResourceOwnerRejectsStartAfterStop(t *testing.T) {
	o := NewPlaneResourceOwner(nil)
	o.Stop()
	if e := o.Start(context.Background()); e == nil {
		t.Fatal("expected closed error")
	}
}
