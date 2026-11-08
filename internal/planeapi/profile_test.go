package planeapi

import "testing"

func TestBuildCloudProfile(t *testing.T) {
	profile, err := Build(DefaultCloudConfig())
	if err != nil {
		t.Fatalf("Build(DefaultCloudConfig()) = %v", err)
	}
	if profile.Name != "cloud" || profile.APIBaseURL != "https://api.plane.so" ||
		profile.OAuthURL != "https://api.plane.so" || profile.WebURL != "https://app.plane.so" {
		t.Fatalf("profile = %+v; want the fixed Cloud origins", profile)
	}
	if profile.Edition != EditionCloud {
		t.Fatalf("edition = %q, want cloud", profile.Edition)
	}
	// Gate A (#496) has not passed for the first-version profile.
	if profile.Verified {
		t.Fatal("the default Cloud profile must not claim verification")
	}
}

func TestBuildProfileDefaults(t *testing.T) {
	profile, err := Build(Config{Name: "cloud", APIBaseURL: "https://api.plane.so/"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if profile.OAuthURL != "https://api.plane.so" {
		t.Fatalf("oauth url = %q; want the API origin", profile.OAuthURL)
	}
	if profile.Edition != EditionCloud {
		t.Fatalf("edition = %q; want cloud by default", profile.Edition)
	}
	if profile.Verified {
		t.Fatal("verified must default to false")
	}
}

func TestBuildProfileRejectsUnsafeOrigins(t *testing.T) {
	for _, tc := range []struct{ name, api string }{
		{"plain http", "http://api.plane.so"},
		{"path", "https://api.plane.so/api"},
		{"query", "https://api.plane.so?tenant=1"},
		{"fragment", "https://api.plane.so#x"},
		{"credentials", "https://user:pass@api.plane.so"},
		{"missing host", "https://"},
		{"other scheme", "ftp://api.plane.so"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Build(Config{Name: "cloud", APIBaseURL: tc.api}); err == nil {
				t.Fatalf("Build(%q) succeeded; want rejection", tc.api)
			}
		})
	}
	// http is only tolerated for loopback local development.
	if _, err := Build(Config{Name: "local", APIBaseURL: "http://127.0.0.1:8080"}); err != nil {
		t.Fatalf("loopback http = %v; want accepted", err)
	}
}

func TestBuildProfileRequiresNameAndKnownEdition(t *testing.T) {
	if _, err := Build(Config{APIBaseURL: "https://api.plane.so"}); err == nil {
		t.Fatal("Build without a name succeeded")
	}
	if _, err := Build(Config{Name: "Cloud", APIBaseURL: "https://api.plane.so"}); err == nil {
		t.Fatal("Build accepted an uppercase profile name")
	}
	if _, err := Build(Config{Name: "cloud", APIBaseURL: "https://api.plane.so", Edition: "enterprise"}); err == nil {
		t.Fatal("Build accepted an unknown edition")
	}
}

func TestValidateOrigin(t *testing.T) {
	for _, good := range []string{"https://butter.example.com", "https://butter.example.com:8443", "http://localhost:3000"} {
		if err := ValidateOrigin(good); err != nil {
			t.Fatalf("ValidateOrigin(%q) = %v; want nil", good, err)
		}
	}
	for _, bad := range []string{"", "http://butter.example.com", "https://butter.example.com/x", "//butter.example.com"} {
		if err := ValidateOrigin(bad); err == nil {
			t.Fatalf("ValidateOrigin(%q) succeeded; want rejection", bad)
		}
	}
	// A trailing slash is normalised rather than rejected.
	if err := ValidateOrigin("https://butter.example.com/"); err != nil {
		t.Fatalf("trailing slash = %v; want accepted", err)
	}
}
