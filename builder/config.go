package builder

import (
	"errors"
	"fmt"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// NewConfig builds a new Config object from a set of compiler options. It also
// loads some information from the environment while doing that. For example, it
// uses the currently active GOPATH (from the goenv package) to determine the Go
// version to use.
func NewConfig(options *compileopts.Options) (*compileopts.Config, error) {
	spec, err := compileopts.LoadTarget(options.Target)
	if err != nil {
		return nil, err
	}
	goroot := goenv.Get("GOROOT")
	if goroot == "" {
		return nil, errors.New("cannot locate $GOROOT, please set it manually")
	}
	major, minor, err := getGorootVersion(goroot)
	if err != nil {
		return nil, fmt.Errorf("could not read version from GOROOT (%v): %v", goroot, err)
	}
	if major != 1 || minor < 11 || minor > 14 {
		return nil, fmt.Errorf("requires go version 1.11, 1.12, 1.13, or 1.14, got go%d.%d", major, minor)
	}
	clangHeaderPath := getClangHeaderPath(goenv.Get("TINYGOROOT"))
	return &compileopts.Config{
		Options:        options,
		Target:         spec,
		GoMinorVersion: minor,
		ClangHeaders:   clangHeaderPath,
		TestConfig:     options.TestConfig,
	}, nil
}
