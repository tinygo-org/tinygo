package main

import (
	"encoding/json"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aykevl/go-llvm"
)

// Target specification for a given target. Used for bare metal targets.
//
// The target specification is mostly inspired by Rust:
// https://doc.rust-lang.org/nightly/nightly-rustc/rustc_target/spec/struct.TargetOptions.html
// https://github.com/shepmaster/rust-arduino-blink-led-no-core-with-cargo/blob/master/blink/arduino.json
type TargetSpec struct {
	Inherits   []string `json:"inherits"`
	Triple     string   `json:"llvm-target"`
	BuildTags  []string `json:"build-tags"`
	Compiler   string   `json:"compiler"`
	Linker     string   `json:"linker"`
	RTLib      string   `json:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	CFlags     []string `json:"cflags"`
	LDFlags    []string `json:"ldflags"`
	ExtraFiles []string `json:"extra-files"`
	Objcopy    string   `json:"objcopy"`
	Emulator   []string `json:"emulator"`
	Flasher    string   `json:"flash"`
	OCDDaemon  []string `json:"ocd-daemon"`
	GDB        string   `json:"gdb"`
	GDBCmds    []string `json:"gdb-initial-cmds"`
}

// copyProperties copies all properties that are set in spec2 into itself.
func (spec *TargetSpec) copyProperties(spec2 *TargetSpec) {
	// TODO: simplify this using reflection? Inherits and BuildTags are special
	// cases, but the rest can simply be copied if set.
	spec.Inherits = append(spec.Inherits, spec2.Inherits...)
	if spec2.Triple != "" {
		spec.Triple = spec2.Triple
	}
	spec.BuildTags = append(spec.BuildTags, spec2.BuildTags...)
	if spec2.Compiler != "" {
		spec.Compiler = spec2.Compiler
	}
	if spec2.Linker != "" {
		spec.Linker = spec2.Linker
	}
	if spec2.RTLib != "" {
		spec.RTLib = spec2.RTLib
	}
	spec.CFlags = append(spec.CFlags, spec2.CFlags...)
	spec.LDFlags = append(spec.LDFlags, spec2.LDFlags...)
	spec.ExtraFiles = append(spec.ExtraFiles, spec2.ExtraFiles...)
	if spec2.Objcopy != "" {
		spec.Objcopy = spec2.Objcopy
	}
	if len(spec2.Emulator) != 0 {
		spec.Emulator = spec2.Emulator
	}
	if spec2.Flasher != "" {
		spec.Flasher = spec2.Flasher
	}
	if len(spec2.OCDDaemon) != 0 {
		spec.OCDDaemon = spec2.OCDDaemon
	}
	if spec2.GDB != "" {
		spec.GDB = spec2.GDB
	}
	if len(spec2.GDBCmds) != 0 {
		spec.GDBCmds = spec2.GDBCmds
	}
}

// load reads a target specification from the JSON in the given io.Reader. It
// may load more targets specified using the "inherits" property.
func (spec *TargetSpec) load(r io.Reader) error {
	err := json.NewDecoder(r).Decode(spec)
	if err != nil {
		return err
	}

	return nil
}

// loadFromName loads the given target from the targets/ directory inside the
// compiler sources.
func (spec *TargetSpec) loadFromName(name string) error {
	path := filepath.Join(sourceDir(), "targets", strings.ToLower(name)+".json")
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	return spec.load(fp)
}

// resolveInherits loads inherited targets, recursively.
func (spec *TargetSpec) resolveInherits() error {
	// First create a new spec with all the inherited properties.
	newSpec := &TargetSpec{}
	for _, name := range spec.Inherits {
		subtarget := &TargetSpec{}
		err := subtarget.loadFromName(name)
		if err != nil {
			return err
		}
		err = subtarget.resolveInherits()
		if err != nil {
			return err
		}
		newSpec.copyProperties(subtarget)
	}

	// When all properties are loaded, make sure they are properly inherited.
	newSpec.copyProperties(spec)
	*spec = *newSpec

	return nil
}

// Load a target specification.
func LoadTarget(target string) (*TargetSpec, error) {
	if target == "" {
		target = llvm.DefaultTargetTriple()
	}

	// See whether there is a target specification for this target (e.g.
	// Arduino).
	spec := &TargetSpec{}
	err := spec.loadFromName(target)
	if err == nil {
		// Successfully loaded this target from a built-in .json file. Make sure
		// it includes all parents as specified in the "inherits" key.
		err = spec.resolveInherits()
		if err != nil {
			return nil, err
		}
		return spec, nil
	} else if !os.IsNotExist(err) {
		// Expected a 'file not found' error, got something else. Report it as
		// an error.
		return nil, err
	} else {
		// No target spec available. Use the default one, useful on most systems
		// with a regular OS.
		*spec = TargetSpec{
			Triple:    target,
			BuildTags: []string{runtime.GOOS, runtime.GOARCH},
			Linker:    "cc",
			LDFlags:   []string{"-no-pie"}, // WARNING: clang < 5.0 requires -nopie
			Objcopy:   "objcopy",
			GDB:       "gdb",
			GDBCmds:   []string{"run"},
		}
		return spec, nil
	}
}

// Return the source directory of this package, or "." when it cannot be
// recovered.
func sourceDir() string {
	// https://stackoverflow.com/a/32163888/559350
	_, path, _, _ := runtime.Caller(0)
	return filepath.Dir(path)
}

func getGopath() string {
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		return gopath
	}

	// fallback
	home := getHomeDir()
	return filepath.Join(home, "go")
}

func getHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic("cannot get current user: " + err.Error())
	}
	if u.HomeDir == "" {
		// This is very unlikely, so panic here.
		// Not the nicest solution, however.
		panic("could not find home directory")
	}
	return u.HomeDir
}
