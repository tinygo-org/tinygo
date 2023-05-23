package goenv

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Version of TinyGo.
// Update this value before release of new version of software.
const Version = "0.29.0-dev"

var (
	// This variable is set at build time using -ldflags parameters.
	// See: https://stackoverflow.com/a/11355611
	GitSha1 string
)

// GetGorootVersion returns the major and minor version for a given GOROOT path.
// If the goroot cannot be determined, (0, 0) is returned.
func GetGorootVersion() (major, minor int, err error) {
	s, err := GorootVersionString()
	if err != nil {
		return 0, 0, err
	}

	if s == "" || s[:2] != "go" {
		return 0, 0, errors.New("could not parse Go version: version does not start with 'go' prefix")
	}

	parts := strings.Split(s[2:], ".")
	if len(parts) < 2 {
		return 0, 0, errors.New("could not parse Go version: version has less than two parts")
	}

	// Ignore the errors, we don't really handle errors here anyway.
	var trailing string
	n, err := fmt.Sscanf(s, "go%d.%d%s", &major, &minor, &trailing)
	if n == 2 && err == io.EOF {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}
	return
}

// GorootVersionString returns the version string as reported by the Go
// toolchain. It is usually of the form `go1.x.y` but can have some variations
// (for beta releases, for example).
func GorootVersionString() (string, error) {
	err := readGoEnvVars()
	return goEnvVars.GOVERSION, err
}
