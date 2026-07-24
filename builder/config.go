package builder

import (
	"errors"
	"fmt"
	"runtime"

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

	// Version range supported by TinyGo.
	const minorMin = 25 // when updating the min version, also update .github/workflows/compat.yml
	const minorMax = 27

	// Check that we support this Go toolchain version.
	gorootMajor, gorootMinor, err := goenv.GetGorootVersion()
	if err != nil {
		return nil, err
	}

	if options.GoCompatibility {
		if gorootMajor != 1 || gorootMinor < minorMin || gorootMinor > minorMax {
			// Note: when this gets updated, also update the Go compatibility matrix:
			// https://github.com/tinygo-org/tinygo-site/blob/dev/content/docs/reference/go-compat-matrix.md
			return nil, fmt.Errorf("requires go version 1.%d through 1.%d, got go%d.%d", minorMin, minorMax, gorootMajor, gorootMinor)
		}
	}

	// Check that the Go toolchain version isn't too new, if we haven't been
	// compiled with the latest Go version.
	// This may be a bit too aggressive: if the newer version doesn't change the
	// Go language we will most likely be able to compile it.
	buildMajor, buildMinor, _, err := goenv.Parse(runtime.Version())
	if err != nil {
		return nil, err
	}
	if buildMajor != 1 || buildMinor < gorootMinor {
		return nil, fmt.Errorf("cannot compile with Go toolchain version go%d.%d (TinyGo was built using toolchain version %s)", gorootMajor, gorootMinor, runtime.Version())
	}

	config := &compileopts.Config{
		Options:        options,
		Target:         spec,
		GoMinorVersion: gorootMinor,
		TestConfig:     options.TestConfig,
	}
	requestedPanicUnwind := options.PanicUnwind
	if requestedPanicUnwind == "" {
		requestedPanicUnwind = spec.PanicUnwind
	}
	if requestedPanicUnwind == "explicit" && config.Scheduler() == "asyncify" {
		return nil, errors.New("explicit panic unwinding cannot be used with the asyncify scheduler")
	}
	if config.PanicUnwind() == "explicit" && !config.SupportsExplicitUnwind() {
		return nil, fmt.Errorf("explicit panic unwinding is not supported on %s", config.Triple())
	}
	if config.PanicUnwind() == "explicit" && config.Scheduler() == "threads" {
		return nil, errors.New("explicit panic unwinding is not supported with the threads scheduler")
	}
	return config, nil
}
