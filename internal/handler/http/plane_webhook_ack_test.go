package http

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	planeruntime "go.orx.me/apps/butter/internal/runtime/plane"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type planeACKReceiver struct {
	err, recordErr   error
	entered, release chan struct{}
	records          atomic.Int32
	committed        atomic.Bool
}

func (s *planeACKReceiver) FindApp(context.Context, string) (*agentsv1.PlaneApp, error) {
	return &agentsv1.PlaneApp{Id: "app", WorkspaceId: "ws"}, nil
}
func (s *planeACKReceiver) Deliver(context.Context, *agentsv1.PlaneApp, http.Header, []byte) (planeruntime.Decision, error) {
	return "", s.err
}
func (s *planeACKReceiver) RecordWebhookRejection(ctx context.Context, _ *agentsv1.PlaneApp, _ []byte, _ error) (string, error) {
	s.records.Add(1)
	if s.entered != nil {
		close(s.entered)
	}
	if s.release != nil {
		select {
		case <-s.release:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if s.recordErr != nil {
		return "", s.recordErr
	}
	s.committed.Store(true)
	return "webhook:stored-receipt", nil
}

type validationOnlyPlaneReceiver struct{ PlaneReceiver }

func TestPlaneWebhookNeverAcknowledgesUnstoredOrUnauthenticatedFailures(t *testing.T) {
	for _, tc := range []struct {
		name          string
		err, storeErr error
		hideRecorder  bool
		status        int
		records       int32
	}{
		{"invalid signature", planeruntime.ErrUnauthorized, nil, false, 401, 0},
		{"signature wins over other classifications", errors.Join(planeruntime.ErrUnauthorized, planeruntime.ErrIdentity), nil, false, 401, 0},
		{"queue unavailable", errors.New("queue unavailable"), nil, false, 503, 0},
		{"diagnostic unavailable", planeruntime.ErrIdentity, errors.New("private database failure"), false, 503, 1},
		{"malformed diagnostic unavailable", planeruntime.ErrMalformed, errors.New("private database failure"), false, 503, 1},
		{"recorder not wired", planeruntime.ErrIdentity, nil, true, 503, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			stub := &planeACKReceiver{err: tc.err, recordErr: tc.storeErr}
			var receiver PlaneReceiver = stub
			if tc.hideRecorder {
				receiver = validationOnlyPlaneReceiver{stub}
			}
			router := gin.New()
			NewPlaneWebhookHandler(receiver).Register(router)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/plane/webhook/app", strings.NewReader(`{}`)))
			if response.Code != tc.status || stub.records.Load() != tc.records || stub.committed.Load() {
				t.Fatalf("false ACK: status=%d records=%d committed=%v", response.Code, stub.records.Load(), stub.committed.Load())
			}
			if strings.Contains(logs.String(), "private database failure") || strings.Contains(response.Body.String(), "private") {
				t.Fatal("archive failure leaked dependency details")
			}
		})
	}
}

func TestPlaneRejectedACKWaitsForPersistenceAndRespectsCancellation(t *testing.T) {
	for _, cancelRequest := range []bool{false, true} {
		t.Run(map[bool]string{false: "persist then ACK", true: "request cancelled"}[cancelRequest], func(t *testing.T) {
			stub := &planeACKReceiver{err: planeruntime.ErrIdentity, entered: make(chan struct{}), release: make(chan struct{})}
			router := gin.New()
			NewPlaneWebhookHandler(stub).Register(router)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			req := httptest.NewRequest(http.MethodPost, "/api/plane/webhook/app", strings.NewReader(`{}`)).WithContext(ctx)
			response := httptest.NewRecorder()
			done := make(chan struct{})
			go func() { defer close(done); router.ServeHTTP(response, req) }()
			select {
			case <-stub.entered:
			case <-time.After(5 * time.Second):
				t.Fatal("receipt persistence did not start")
			}
			select {
			case <-done:
				t.Fatal("ACK returned before persistence")
			default:
			}
			if cancelRequest {
				cancel()
			} else {
				close(stub.release)
			}
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("receipt handler did not settle")
			}
			if cancelRequest {
				if response.Code != 503 || stub.committed.Load() {
					t.Fatal("cancelled write was acknowledged")
				}
			} else if response.Code != 200 || !stub.committed.Load() || !strings.Contains(response.Body.String(), `"decision":"rejected"`) {
				t.Fatal("committed rejection was not acknowledged")
			}
		})
	}
}
