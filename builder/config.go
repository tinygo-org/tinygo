package builder

import (
	"fmt"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// NewConfig builds a new Config object from a set of compiler options. It also
// loads some information from the environment while doing that. For example, it
// uses the currently active GOPATH (from the goenv package) to determine the Go
// version to use.
func NewConfig(options *compileopts.Options) (*compileopts.Config, error) {
	spec, err := compileopts.LoadTarget(options)
	if err != nil {
		return nil, err
	}

	if options.OpenOCDCommands != nil {
		// Override the OpenOCDCommands from the target spec if specified on
		// the command-line
		spec.OpenOCDCommands = options.OpenOCDCommands
	}

	major, minor, err := goenv.GetGorootVersion()
	if err != nil {
		return nil, err
	}
	if major != 1 || minor < 19 || minor > 23 {
		// Note: when this gets updated, also update the Go compatibility matrix:
		// https://github.com/tinygo-org/tinygo-site/blob/dev/content/docs/reference/go-compat-matrix.md
		return nil, fmt.Errorf("requires go version 1.19 through 1.23, got go%d.%d", major, minor)
	}

	return &compileopts.Config{
		Options:        options,
		Target:         spec,
		GoMinorVersion: minor,
		TestConfig:     options.TestConfig,
	}, nil
}
