package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Target specification for a given target. Used for bare metal targets.
//
// The target specification is mostly inspired by Rust:
// https://doc.rust-lang.org/nightly/nightly-rustc/rustc_target/spec/struct.TargetOptions.html
// https://github.com/shepmaster/rust-arduino-blink-led-no-core-with-cargo/blob/master/blink/arduino.json
type TargetSpec struct {
	Triple      string   `json:"llvm-target"`
	BuildTags   []string `json:"build-tags"`
	Linker      string   `json:"linker"`
	PreLinkArgs []string `json:"pre-link-args"`
	Objcopy     string   `json:"objcopy"`
	Flasher     string   `json:"flash"`
}

// Load a target specification
func LoadTarget(target string) (*TargetSpec, error) {
	spec := &TargetSpec{
		Triple:      target,
		BuildTags:   []string{runtime.GOOS, runtime.GOARCH},
		Linker:      "cc",
		PreLinkArgs: []string{"-no-pie"}, // WARNING: clang < 5.0 requires -nopie
		Objcopy:     "objcopy",
	}

	// See whether there is a target specification for this target (e.g.
	// Arduino).
	path := filepath.Join(sourceDir(), "targets", strings.ToLower(target)+".json")
	if fp, err := os.Open(path); err == nil {
		defer fp.Close()
		err := json.NewDecoder(fp).Decode(spec)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		// Expected a 'file not found' error, got something else.
		return nil, err
	} else {
		// No target spec available. This is fine.
	}

	return spec, nil
}

// Return the source directory of this package, or "." when it cannot be
// recovered.
func sourceDir() string {
	// https://stackoverflow.com/a/32163888/559350
	_, path, _, _ := runtime.Caller(0)
	return filepath.Dir(path)
}
