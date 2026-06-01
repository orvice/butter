package main

import "testing"

func TestResolveHost(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		server string
		want   string
	}{
		{name: "explicit host wins", host: "worker-1.local", server: "butter.example.com:9090", want: "worker-1.local"},
		{name: "host port server", server: "butter.example.com:9090", want: "butter.example.com"},
		{name: "url server", server: "https://butter.example.com:9090", want: "butter.example.com"},
		{name: "ipv4 server", server: "127.0.0.1:9090", want: "127.0.0.1"},
		{name: "bracketed ipv6 server", server: "[::1]:9090", want: "::1"},
		{name: "plain host server", server: "butter.example.com", want: "butter.example.com"},
		{name: "empty server", server: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveHost(tt.host, tt.server); got != tt.want {
				t.Fatalf("resolveHost(%q, %q) = %q, want %q", tt.host, tt.server, got, tt.want)
			}
		})
	}
}
