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
	Triple    string   `json:"llvm-target"`
	BuildTags []string `json:"build-tags"`
}

// Load a target specification
func LoadTarget(target string) (*TargetSpec, error) {
	spec := &TargetSpec{
		Triple:    target,
		BuildTags: []string{runtime.GOOS, runtime.GOARCH},
	}

	// See whether there is a target specification for this target (e.g.
	// Arduino).
	path := filepath.Join("targets", strings.ToLower(target)+".json")
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
