package telegramapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testToken = "123456:AAH-super-secret-token"

func newTestClient(t *testing.T, handler http.HandlerFunc) *HTTPClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(testToken, WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
}

func TestGetMeReturnsBotIdentity(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"ok":true,"result":{"id":42,"username":"butterbot","first_name":"Butter",`+
			`"can_join_groups":true,"can_read_all_group_messages":false,"supports_inline_queries":true}}`)
	})

	identity, err := client.GetMe(t.Context())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if identity.ID != 42 || identity.Username != "butterbot" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
	if identity.CanReadAllGroupMessages {
		t.Error("expected group privacy to be reported as enabled")
	}
	if !identity.SupportsInlineQueries {
		t.Error("expected inline query support to be reported")
	}
	if want := "/bot" + testToken + "/getMe"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestGetMeMapsUnauthorizedToken(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"ok":false,"error_code":401,"description":"Unauthorized"}`)
	})

	_, err := client.GetMe(t.Context())
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

// A rejected request must never hand the caller back the Bot Token, which
// travels in the URL path and would otherwise reach logs and error details.
func TestErrorsRedactTheBotToken(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"ok":false,"error_code":400,"description":"bad token %s in request"}`, testToken)
	})

	_, err := client.GetMe(t.Context())
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), testToken) {
		t.Fatalf("error leaked the bot token: %v", err)
	}
	if !strings.Contains(err.Error(), "[REDACTED]") {
		t.Errorf("expected the token to be replaced with a redaction marker, got %q", err.Error())
	}
}

func TestRetryAfterReportsTelegramBackoff(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":7}}`)
	})

	_, err := client.GetMe(t.Context())
	wait, ok := RetryAfter(err)
	if !ok {
		t.Fatalf("RetryAfter did not recognize the 429: %v", err)
	}
	if wait != 7*time.Second {
		t.Errorf("wait = %v, want 7s", wait)
	}
}

func TestGetMeRejectsEmptyToken(t *testing.T) {
	client := New("")
	_, err := client.GetMe(t.Context())
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestIDRoundTripPreservesInt64Precision(t *testing.T) {
	// Telegram chat IDs for supergroups exceed 2^53, so the decimal-string
	// representation is what keeps JSON clients honest.
	const supergroup int64 = -1002233445566
	formatted := FormatID(supergroup)
	if formatted != "-1002233445566" {
		t.Fatalf("FormatID = %q", formatted)
	}
	parsed, err := ParseID(" -1002233445566 ")
	if err != nil {
		t.Fatalf("ParseID: %v", err)
	}
	if parsed != supergroup {
		t.Errorf("ParseID = %d, want %d", parsed, supergroup)
	}
}
