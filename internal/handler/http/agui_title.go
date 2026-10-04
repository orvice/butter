package http

import (
	"context"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"google.golang.org/adk/v2/session"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// aguiTitleTimeout bounds titling a thread after its run. The title model call
// has its own shorter limit; this also covers the session read and the store
// write around it.
const aguiTitleTimeout = 30 * time.Second

// AGUISessionTitler titles a thread's session after a successful run.
// SessionServiceServer.TitleSession implements it: it gives an untitled session
// a title from its first turn and never replaces an existing one.
type AGUISessionTitler interface {
	TitleSession(ctx context.Context, appName, userID, sessionID string) (*agentsv1.SessionInfo, bool, error)
}

// SetSessionTitler wires server-side thread titles. Without a titler, a thread
// keeps no title until a client asks for one.
func (h *AGUIHandler) SetSessionTitler(titler AGUISessionTitler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessionTitler = titler
}

func (h *AGUIHandler) getSessionTitler() AGUISessionTitler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessionTitler
}

// titleThread titles the run's thread in the background. It is called after a
// successful run on a thread that had no title, for a caller the run already
// authorized. The work runs on a context detached from the request, so it
// outlives the response, bounded by aguiTitleTimeout. It never holds the
// thread lease, so the next run on the thread does not wait for it.
func (h *AGUIHandler) titleThread(ctx context.Context, ctxInfo *agentsv1.ContextInfo) {
	titler := h.getSessionTitler()
	if titler == nil {
		return
	}
	userID, sessionID := ctxInfo.GetUserId(), ctxInfo.GetSessionId()
	detached := context.WithoutCancel(ctx)
	h.titles.Add(1)
	go func() {
		defer h.titles.Done()
		ctx, cancel := context.WithTimeout(detached, aguiTitleTimeout)
		defer cancel()
		if _, _, err := titler.TitleSession(ctx, aguiAppName, userID, sessionID); err != nil {
			log.FromContext(ctx).Warn("agui thread title failed", "session_id", sessionID, "err", err)
		}
	}()
}

// aguiThreadTitled reports whether the thread's session had a title when the
// run started, as the Mongo store's sessions expose it. It only spares the
// titler a session read: the titler checks again, a legacy state["title"]
// included, and never replaces a title.
func aguiThreadTitled(sess session.Session) bool {
	titled, ok := sess.(interface{ Title() string })
	return ok && strings.TrimSpace(titled.Title()) != ""
}
