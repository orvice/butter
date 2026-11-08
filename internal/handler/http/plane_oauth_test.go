package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type planeCallbackSpy struct{ state, installation, oauthErr string }

func (s *planeCallbackSpy) CompletePlaneInstall(_ context.Context, state, installation, oauthErr string) string {
	s.state, s.installation, s.oauthErr = state, installation, oauthErr
	return "/plane-apps/app/edit?plane_install=success"
}

func TestPlaneCallbackPassesInstallationRatherThanAuthorizationCode(t *testing.T) {
	spy := &planeCallbackSpy{}
	r := gin.New()
	NewPlaneOAuthHandler(spy).Register(r)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodGet,
		"/api/plane/oauth/callback?state=nonce&code=not-a-bot-token&app_installation_id=inst&error=", nil))
	if response.Code != http.StatusFound || spy.state != "nonce" || spy.installation != "inst" ||
		response.Header().Get("Location") != "/plane-apps/app/edit?plane_install=success" {
		t.Fatalf("callback=%+v status=%d location=%s", spy, response.Code, response.Header().Get("Location"))
	}
}
