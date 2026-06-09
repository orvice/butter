package app

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
)

func TestStartH2CServerServesRouterOverHTTP2(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server, err := StartH2CServer("127.0.0.1:0", func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
	})
	if err != nil {
		t.Fatalf("start h2c server: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	resp, err := client.Get("http://" + server.Addr + "/ping")
	if err != nil {
		t.Fatalf("get /ping over h2c: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if resp.ProtoMajor != 2 {
		t.Fatalf("expected HTTP/2 response, got %s", resp.Proto)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "pong" {
		t.Fatalf("unexpected response: status=%d body=%q", resp.StatusCode, string(body))
	}
}
