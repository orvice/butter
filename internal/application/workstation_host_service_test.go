package application

import (
	"connectrpc.com/connect"
	"context"
	authrepo "go.orx.me/apps/butter/internal/repo/auth"
	wsmemory "go.orx.me/apps/butter/internal/workstation/memory"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

func TestPlaneHostAuthorityAndRevision(t *testing.T) {
	s := NewWorkStationHostServiceServer()
	s.SetRepo(wsmemory.New())
	if _, err := s.ListHosts(context.Background(), connect.NewRequest(&v1.WorkStationHostServiceListHostsRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatal(err)
	}
	for _, role := range []string{"member", "owner", "admin"} {
		ctx := stationContext(role)
		if _, err := s.ListHosts(ctx, connect.NewRequest(&v1.WorkStationHostServiceListHostsRequest{})); err != nil {
			t.Fatal(err)
		}
		if _, err := s.CreateHost(ctx, connect.NewRequest(&v1.WorkStationHostServiceCreateHostRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatal("workspace role changed global host", role, err)
		}
	}
	admin := authrepo.WithAuthenticated(context.Background(), &v1.User{Id: "global", Role: "admin"}, nil)
	created, err := s.CreateHost(admin, connect.NewRequest(&v1.WorkStationHostServiceCreateHostRequest{Host: &v1.WorkStationHost{Name: "Plane", ApiBaseUrl: "https://plane.example/", WebBaseUrl: "https://plane.example", Profile: "plane-v2"}}))
	if err != nil {
		t.Fatal(err)
	}
	h := created.Msg.Host
	h.Name = "renamed"
	if _, err := s.UpdateHost(admin, connect.NewRequest(&v1.WorkStationHostServiceUpdateHostRequest{Host: h, ExpectedRevision: h.Revision + 1})); connect.CodeOf(err) != connect.CodeAborted {
		t.Fatal("stale host update", err)
	}
	if _, err := s.DeleteHost(admin, connect.NewRequest(&v1.WorkStationHostServiceDeleteHostRequest{Id: "plane-cloud", ExpectedRevision: 1})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatal("built-in host deleted", err)
	}
	for _, endpoint := range []string{"http://plane.example", "https://user@plane.example", "https://plane.example/?key=x", "https://plane.example/%2e%2e/"} {
		if _, err := s.CreateHost(admin, connect.NewRequest(&v1.WorkStationHostServiceCreateHostRequest{Host: &v1.WorkStationHost{Name: "Unsafe", ApiBaseUrl: endpoint, WebBaseUrl: "https://plane.example", Profile: "plane-v2"}})); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatal(endpoint, err)
		}
	}
}
