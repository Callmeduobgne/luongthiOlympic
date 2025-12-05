// Package version provides version information for the IBN Backend
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the IBN Backend (set via ldflags during build)
	Version = "dev"

	// BuildDate is the date when the binary was built (set via ldflags during build)
	BuildDate = "unknown"

	// GitCommit is the git commit hash (set via ldflags during build)
	GitCommit = "unknown"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("IBN Backend v%s (commit: %s, built: %s, go: %s, os: %s, arch: %s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.OS, i.Arch)
}
