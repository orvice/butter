package sessionshare

import (
	"context"
	"testing"
)

func TestAllowMarksOnlyTheDerivedContext(t *testing.T) {
	base := context.Background()
	if Allowed(base) {
		t.Fatal("a plain context must not allow sharing")
	}
	shared := Allow(base)
	if !Allowed(shared) {
		t.Fatal("Allow must mark the context it returns")
	}
	if Allowed(base) {
		t.Fatal("Allow must not mark the parent context")
	}
	type otherKey struct{}
	if !Allowed(context.WithValue(shared, otherKey{}, "x")) {
		t.Fatal("the mark must survive derived contexts")
	}
}
