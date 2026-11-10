// Package version reports the build the window runs.
package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Version is set by the release build.
var Version = "0.1.0-dev"

// String is Version plus the short VCS revision, when the build has one.
func String() string {
	rev := revision()
	if rev == "" {
		return Version
	}
	return fmt.Sprintf("%s (%s)", Version, rev)
}

func revision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key != "vcs.revision" {
			continue
		}
		if len(s.Value) > 7 {
			return s.Value[:7]
		}
		return s.Value
	}
	return ""
}

// IsDev reports a build that is not a tagged release.
func IsDev() bool { return strings.Contains(Version, "-dev") }
