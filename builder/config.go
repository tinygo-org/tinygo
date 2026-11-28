package builder

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// requiresCgo returns whether the given target needs CGo to build. These are
// the targets whose libc is provided by the host system through cgo (macOS
// libSystem, Windows mingw-w64) rather than by a bundled library, plus every
// baremetal target: TinyGo's own device/machine sources use cgo there (for
// example src/device/arm/arm.go), so CGO_ENABLED=0 would drop those files
// from the go list file set and break the build.
func requiresCgo(spec *compileopts.TargetSpec) bool {
	switch spec.Libc {
	case "darwin-libSystem", "mingw-w64":
		return true
	}
	for _, tag := range spec.BuildTags {
		if tag == "baremetal" || tag == "nintendoswitch" {
			return true
		}
	}
	return false
}

// targetName returns a human-readable name for the target being built, for
// use in error messages. options.Target is empty for GOOS/GOARCH builds, so
// fall back to the spec triple or GOOS/GOARCH.
func targetName(options *compileopts.Options, spec *compileopts.TargetSpec) string {
	if options.Target != "" {
		return options.Target
	}
	if spec.Triple != "" {
		return spec.Triple
	}
	return spec.GOOS + "/" + spec.GOARCH
}

// NewConfig builds a new Config object from a set of compiler options. It also
// loads some information from the environment while doing that. For example, it
// uses the currently active GOPATH (from the goenv package) to determine the Go
// version to use.
// The options must provide a semaphore with positive capacity.
func NewConfig(options *compileopts.Options) (*compileopts.Config, error) {
	if cap(options.Semaphore) == 0 {
		return nil, errors.New("cannot run 0 jobs at a time")
	}

	spec, err := compileopts.LoadTarget(options)
	if err != nil {
		return nil, err
	}

	if options.OpenOCDCommands != nil {
		// Override the OpenOCDCommands from the target spec if specified on
		// the command-line
		spec.OpenOCDCommands = options.OpenOCDCommands
	}

	// Check that we support this Go toolchain version.
	gorootMajor, gorootMinor, err := goenv.GetGorootVersion()
	if err != nil {
		return nil, err
	}

	if options.GoCompatibility {
		if gorootMajor != 1 || gorootMinor < compileopts.MinimumGoMinorVersion || gorootMinor > compileopts.MaximumGoMinorVersion {
			// Note: when this gets updated, also update the Go compatibility matrix:
			// https://github.com/tinygo-org/tinygo-site/blob/dev/content/docs/reference/go-compat-matrix.md
			return nil, fmt.Errorf("requires go version 1.%d through 1.%d, got go%d.%d", compileopts.MinimumGoMinorVersion, compileopts.MaximumGoMinorVersion, gorootMajor, gorootMinor)
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

	// Some targets cannot work without CGo: macOS and Windows, whose libc is
	// provided by the host system through cgo, and baremetal targets, whose
	// own device/machine sources use cgo. Reject CGO_ENABLED=0 there instead
	// of failing later with a confusing compile or linker error.
	if goenv.Get("CGO_ENABLED") == "0" && requiresCgo(spec) {
		return nil, fmt.Errorf("CGO_ENABLED=0 is not supported for target %s (libc %s): it requires CGo", targetName(options, spec), spec.Libc)
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
