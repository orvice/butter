package eventcommit_test

import (
	"context"
	"errors"
	"testing"

	"go.orx.me/apps/butter/internal/runtime/eventcommit"
	"google.golang.org/adk/v2/session"
)

type gatedService struct {
	session.Service
	entered chan struct{}
	release chan struct{}
	err     error
}

func (s gatedService) AppendEvent(ctx context.Context, sess session.Session, event *session.Event) error {
	close(s.entered)
	select {
	case <-s.release:
	case <-ctx.Done():
		return ctx.Err()
	}
	if s.err != nil {
		return s.err
	}
	return s.Service.AppendEvent(ctx, sess, event)
}

func TestReceiptWaitsForActualStoreAndPropagatesFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			ctx, cancel := eventcommit.Track(t.Context())
			defer cancel()
			base := session.InMemoryService()
			created, err := base.Create(ctx, &session.CreateRequest{AppName: "app", UserID: "u", SessionID: "s"})
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New("store failed")
			gate := gatedService{Service: base, entered: make(chan struct{}), release: make(chan struct{})}
			if fail {
				gate.err = failure
			}
			svc := eventcommit.Service(gate)
			event := session.NewEvent(ctx, "inv")
			wait, discard := eventcommit.Expect(ctx, event.ID)
			defer discard()
			appendDone := make(chan error, 1)
			go func() { appendDone <- svc.AppendEvent(ctx, created.Session, event) }()
			<-gate.entered
			waitDone := make(chan error, 1)
			go func() { waitDone <- wait(ctx) }()
			select {
			case err := <-waitDone:
				t.Fatalf("acknowledged before persistence: %v", err)
			default:
			}
			close(gate.release)
			appendErr, waitErr := <-appendDone, <-waitDone
			if fail {
				if !errors.Is(appendErr, failure) || !errors.Is(waitErr, failure) {
					t.Fatalf("append=%v wait=%v", appendErr, waitErr)
				}
			} else if appendErr != nil || waitErr != nil {
				t.Fatalf("append=%v wait=%v", appendErr, waitErr)
			}
		})
	}
}

func TestRunnerExitReleasesUncommittedLeaf(t *testing.T) {
	ctx, cancel := eventcommit.Track(t.Context())
	wait, discard := eventcommit.Expect(ctx, "never-persisted")
	defer discard()
	cancel()
	if err := wait(t.Context()); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait=%v", err)
	}
}
