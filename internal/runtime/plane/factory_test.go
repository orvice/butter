package plane

import (
	"context"
	"testing"
)

func TestPlaneRuntimeFactoryFailsClosed(t *testing.T) {
	if _, e := (PlaneRuntimeFactory{}).Build(PlaneRuntimeDependencies{}); e != ErrRuntimeDependency {
		t.Fatalf("err=%v", e)
	}
}
func TestRuntimeCoordinatorStopIsIdempotent(t *testing.T) {
	c := NewRuntimeCoordinator(nil)
	if e := c.Start(context.Background()); e != ErrRuntimeDependency {
		t.Fatal(e)
	}
	c.Stop()
	c.Stop()
}
