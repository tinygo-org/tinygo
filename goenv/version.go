package goenv

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Version of TinyGo.
// Update this value before release of new version of software.
const version = "0.34.0-dev"

var (
	// This variable is set at build time using -ldflags parameters.
	// See: https://stackoverflow.com/a/11355611
	GitSha1 string
)

// Return TinyGo version, either in the form 0.30.0 or as a development version
// (like 0.30.0-dev-abcd012).
func Version() string {
	v := version
	if strings.HasSuffix(version, "-dev") && GitSha1 != "" {
		v += "-" + GitSha1
	}
	return v
}

// GetGorootVersion returns the major and minor version for a given GOROOT path.
// If the goroot cannot be determined, (0, 0) is returned.
func GetGorootVersion() (major, minor int, err error) {
	s, err := GorootVersionString()
	if err != nil {
		return 0, 0, err
	}
	major, minor, _, err = Parse(s)
	return major, minor, err
}

// Parse parses the Go version (like "go1.3.2") in the parameter and return the
// major, minor, and patch version: 1, 3, and 2 in this example.
// If there is an error, (0, 0, 0) and an error will be returned.
func Parse(version string) (major, minor, patch int, err error) {
	if version == "" || version[:2] != "go" {
		return 0, 0, 0, errors.New("could not parse Go version: version does not start with 'go' prefix")
	}

	parts := strings.Split(version[2:], ".")
	if len(parts) < 2 {
		return 0, 0, 0, errors.New("could not parse Go version: version has less than two parts")
	}

	// Ignore the errors, we don't really handle errors here anyway.
	var trailing string
	n, err := fmt.Sscanf(version, "go%d.%d.%d%s", &major, &minor, &patch, &trailing)
	if n == 2 {
		n, err = fmt.Sscanf(version, "go%d.%d%s", &major, &minor, &trailing)
	}
	if n >= 2 && err == io.EOF {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}

	return major, minor, patch, nil
}

// WantGoVersion returns true if Go version s is >= major and minor.
// Returns false if s is not a valid Go version string. See [Parse] for more information.
func WantGoVersion(s string, major, minor int) bool {
	ma, mi, _, err := Parse(s)
	if err != nil {
		return false
	}
	return ma > major || (ma == major && mi >= minor)
}

// Compare compares two Go version strings.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
// If either a or b is not a valid Go version, it is treated as "go0.0"
// and compared lexicographically.
// See [Parse] for more information.
func Compare(a, b string) int {
	aMajor, aMinor, aPatch, _ := Parse(a)
	bMajor, bMinor, bPatch, _ := Parse(b)
	switch {
	case aMajor < bMajor:
		return -1
	case aMajor > bMajor:
		return +1
	case aMinor < bMinor:
		return -1
	case aMinor > bMinor:
		return +1
	case aPatch < bPatch:
		return -1
	case aPatch > bPatch:
		return +1
	default:
		return strings.Compare(a, b)
	}
}

// GorootVersionString returns the version string as reported by the Go
// toolchain. It is usually of the form `go1.x.y` but can have some variations
// (for beta releases, for example).
func GorootVersionString() (string, error) {
	err := readGoEnvVars()
	return goEnvVars.GOVERSION, err
}
