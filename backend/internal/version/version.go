// Package version provides version information for the IBN Backend
package version

import (
	"fmt"
	"runtime"
)

const (
	// Version is the current version of the IBN Backend
	Version = "1.0.0"

	// BuildDate is the date when the binary was built (set during build)
	BuildDate = "2025-12-01"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "300adab"
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
