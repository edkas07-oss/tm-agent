package buildinfo

import (
	"fmt"
	"runtime"
)

// Default build variables replaced at compile time via -ldflags.
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// String formats the build info into a concise string.
func String() string {
	return fmt.Sprintf("tm-agent v%s (commit: %s, built: %s, %s/%s)",
		Version, GitCommit, BuildDate, runtime.GOOS, runtime.GOARCH)
}
