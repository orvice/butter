package http

import (
	"strings"
	"testing"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

func drainObserver(o *aguiObserver) (frames []string, ended, cutOff bool) {
	got, ended, cutOff := o.take()
	for _, f := range got {
		frames = append(frames, string(f))
	}
	return frames, ended, cutOff
}

// Each observer gets every event emitted while it is attached, one SSE frame
// each, then the end of the stream. An observer that detaches stops
// receiving; the others carry on.
func TestAGUIFanout_ObserversFollowTheRun(t *testing.T) {
	f := newAGUIFanout()
	first := f.attach()
	second := f.attach()
	if err := f.emit(aguievents.NewRunStartedEvent("t-1", "run-1")); err != nil {
		t.Fatal(err)
	}
	second.detach()
	if err := f.emit(aguievents.NewRunFinishedEventWithOptions("t-1", "run-1", aguievents.WithSuccessOutcome())); err != nil {
		t.Fatal(err)
	}
	f.close()

	frames, ended, cutOff := drainObserver(first)
	if len(frames) != 2 || !ended || cutOff {
		t.Fatalf("first observer: frames %q, ended %v, cut off %v", frames, ended, cutOff)
	}
	for i, typ := range []string{"RUN_STARTED", "RUN_FINISHED"} {
		if !strings.Contains(frames[i], "data: ") || !strings.Contains(frames[i], `"type":"`+typ+`"`) ||
			!strings.HasSuffix(frames[i], "\n\n") {
			t.Fatalf("frame %d = %q, want one SSE frame of %s", i, frames[i], typ)
		}
	}
	if frames, ended, _ := drainObserver(second); len(frames) != 1 || ended {
		t.Fatalf("detached observer: frames %q, ended %v; want only what came before it left", frames, ended)
	}
	// Attaching after the run ended yields an ended stream at once.
	if _, ended, _ := drainObserver(f.attach()); !ended {
		t.Fatal("an observer of a closed fan-out did not end")
	}
}

// An observer that falls too far behind is cut off rather than buffered
// without end; the run's other observers are not affected.
func TestAGUIFanout_SlowObserverIsCutOff(t *testing.T) {
	f := newAGUIFanout()
	slow := f.attach()
	big := strings.Repeat("x", 1<<20)
	for range aguiObserverMaxQueue/(1<<20) + 1 {
		if err := f.emit(aguievents.NewTextMessageContentEvent("m-1", big)); err != nil {
			t.Fatal(err)
		}
	}
	if frames, _, cutOff := drainObserver(slow); !cutOff || len(frames) != 0 {
		t.Fatalf("frames %d, cut off %v; want the observer cut off", len(frames), cutOff)
	}
	fresh := f.attach()
	if err := f.emit(aguievents.NewTextMessageContentEvent("m-1", "small")); err != nil {
		t.Fatal(err)
	}
	if frames, _, cutOff := drainObserver(fresh); cutOff || len(frames) != 1 {
		t.Fatalf("another observer: frames %d, cut off %v", len(frames), cutOff)
	}
}
