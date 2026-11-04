package workstation

import (
	"context"
	"net/url"
	"strings"

	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const PlaneCloudHost = "plane-cloud"

type Host struct {
	ID         string `bson:"id"`
	Name       string `bson:"name"`
	APIBaseURL string `bson:"api_base_url"`
	WebBaseURL string `bson:"web_base_url"`
	Profile    string `bson:"profile"`
	Revision   int64  `bson:"revision"`
}
type HostRepository interface {
	ListHosts(context.Context) ([]Host, error)
	GetHost(context.Context, string) (Host, error)
	SaveHost(context.Context, Host, int64) error
	DeleteHost(context.Context, string, int64) error
}

func CloudHost() Host {
	return Host{ID: PlaneCloudHost, Name: "Plane Cloud", APIBaseURL: "https://api.plane.so", WebBaseURL: "https://app.plane.so", Profile: "plane-v2", Revision: 1}
}
func (h Host) Public() *v1.WorkStationHost {
	return &v1.WorkStationHost{Id: h.ID, Name: h.Name, ApiBaseUrl: h.APIBaseURL, WebBaseUrl: h.WebBaseURL, Profile: h.Profile, Revision: h.Revision}
}
func ApprovedURL(raw string) (string, error) {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(u.EscapedPath(), "%\\") || strings.Contains(u.Path, "..") {
		return "", ErrInvalid
	}
	u.Host = strings.ToLower(u.Host)
	if u.Port() == "443" {
		u.Host = u.Hostname()
		if strings.Contains(u.Host, ":") {
			u.Host = "[" + u.Host + "]"
		}
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}
func ValidateHost(h *Host) error {
	if !validName(h.Name) || h.Profile != "plane-v2" {
		return ErrInvalid
	}
	var err error
	h.APIBaseURL, err = ApprovedURL(h.APIBaseURL)
	if err != nil {
		return err
	}
	h.WebBaseURL, err = ApprovedURL(h.WebBaseURL)
	return err
}
