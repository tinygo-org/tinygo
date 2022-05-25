package goenv

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// Version of TinyGo.
// Update this value before release of new version of software.
const Version = "0.24.0-dev"

var (
	// This variable is set at build time using -ldflags parameters.
	// See: https://stackoverflow.com/a/11355611
	GitSha1 string
)

// parseGorootVersion returns the major and minor version for a given Go version
// string of the form `goX.Y.Z`.
// Returns (0, 0) if the version cannot be determined.
func parseGorootVersion(version string) (int, int, error) {
	var (
		maj, min int
		pch      string
	)
	n, err := fmt.Sscanf(version, "go%d.%d%s", &maj, &min, &pch)
	if n == 2 && io.EOF == err {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if nil != err {
		return 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}
	return maj, min, nil
}

// GetGorootVersion returns the major and minor version for a given GOROOT path.
// If the version cannot be determined, (0, 0) is returned.
func GetGorootVersion(goroot string) (int, int, error) {
	const errPrefix = "could not parse Go version"
	s, err := GorootVersionString(goroot)
	if err != nil {
		return 0, 0, err
	}

	if "" == s {
		return 0, 0, fmt.Errorf("%s: version string is empty", errPrefix)
	}

	if strings.HasPrefix(s, "devel") {
		maj, min, err := getGorootApiVersion(goroot)
		if nil != err {
			return 0, 0, fmt.Errorf("%s: invalid GOROOT API version: %s", errPrefix, err)
		}
		return maj, min, nil
	}

	if !strings.HasPrefix(s, "go") {
		return 0, 0, fmt.Errorf("%s: version does not start with 'go' prefix", errPrefix)
	}

	parts := strings.Split(s[2:], ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("%s: version has less than two parts", errPrefix)
	}

	return parseGorootVersion(s)
}

// getGorootApiVersion returns the major and minor version of the Go API files
// defined for a given GOROOT path.
// If the version cannot be determined, (0, 0) is returned.
func getGorootApiVersion(goroot string) (int, int, error) {
	info, err := ioutil.ReadDir(filepath.Join(goroot, "api"))
	if nil != err {
		return 0, 0, fmt.Errorf("could not read API feature directory: %s", err)
	}
	maj, min := -1, -1
	for _, f := range info {
		if !strings.HasPrefix(f.Name(), "go") || f.IsDir() {
			continue
		}
		vers := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		part := strings.Split(vers[2:], ".")
		if len(part) < 2 {
			continue
		}
		vmaj, vmin, err := parseGorootVersion(vers)
		if nil != err {
			continue
		}
		if vmaj >= maj && vmin > min {
			maj, min = vmaj, vmin
		}
	}
	if maj < 0 || min < 0 {
		return 0, 0, errors.New("no valid API feature files")
	}
	return maj, min, nil
}

// GorootVersionString returns the version string as reported by the Go
// toolchain for the given GOROOT path. It is usually of the form `go1.x.y` but
// can have some variations (for beta releases, for example).
func GorootVersionString(goroot string) (string, error) {
	if data, err := ioutil.ReadFile(filepath.Join(goroot, "VERSION")); err == nil {
		return string(data), nil

	} else if data, err := ioutil.ReadFile(filepath.Join(
		goroot, "src", "internal", "buildcfg", "zbootstrap.go")); err == nil {

		r := regexp.MustCompile("const version = `(.*)`")
		matches := r.FindSubmatch(data)
		if len(matches) != 2 {
			return "", errors.New("Invalid go version output:\n" + string(data))
		}

		return string(matches[1]), nil

	} else {
		return "", err
	}
}
