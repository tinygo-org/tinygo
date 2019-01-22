package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
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
	Inherits   []string `json:"inherits"`
	Triple     string   `json:"llvm-target"`
	CPU        string   `json:"cpu"`
	GOOS       string   `json:"goos"`
	GOARCH     string   `json:"goarch"`
	BuildTags  []string `json:"build-tags"`
	GC         string   `json:"gc"`
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
	if spec2.CPU != "" {
		spec.CPU = spec2.CPU
	}
	if spec2.GOOS != "" {
		spec.GOOS = spec2.GOOS
	}
	if spec2.GOARCH != "" {
		spec.GOARCH = spec2.GOARCH
	}
	spec.BuildTags = append(spec.BuildTags, spec2.BuildTags...)
	if spec2.GC != "" {
		spec.GC = spec2.GC
	}
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
		// Configure based on GOOS/GOARCH environment variables (falling back to
		// runtime.GOOS/runtime.GOARCH), and generate a LLVM target based on it.
		goos := os.Getenv("GOOS")
		if goos == "" {
			goos = runtime.GOOS
		}
		goarch := os.Getenv("GOARCH")
		if goarch == "" {
			goarch = runtime.GOARCH
		}
		llvmos := goos
		llvmarch := map[string]string{
			"386":   "i386",
			"amd64": "x86_64",
			"arm64": "aarch64",
		}[goarch]
		if llvmarch == "" {
			llvmarch = goarch
		}
		target = llvmarch + "--" + llvmos
		return defaultTarget(goos, goarch, target)
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
		// Load target from given triple, ignore GOOS/GOARCH environment
		// variables.
		tripleSplit := strings.Split(target, "-")
		if len(tripleSplit) == 1 {
			return nil, errors.New("expected a full LLVM target or a custom target in -target flag")
		}
		goos := tripleSplit[2]
		goarch := map[string]string{ // map from LLVM arch to Go arch
			"i386":    "386",
			"x86_64":  "amd64",
			"aarch64": "arm64",
		}[tripleSplit[0]]
		if goarch == "" {
			goarch = tripleSplit[0]
		}
		return defaultTarget(goos, goarch, target)
	}
}

func defaultTarget(goos, goarch, triple string) (*TargetSpec, error) {
	// No target spec available. Use the default one, useful on most systems
	// with a regular OS.
	spec := TargetSpec{
		Triple:    triple,
		GOOS:      goos,
		GOARCH:    goarch,
		BuildTags: []string{goos, goarch},
		Compiler:  commands["clang"],
		Linker:    "cc",
		LDFlags:   []string{"-no-pie"}, // WARNING: clang < 5.0 requires -nopie
		Objcopy:   "objcopy",
		GDB:       "gdb",
		GDBCmds:   []string{"run"},
	}
	if goarch != runtime.GOARCH {
		// Some educated guesses as to how to invoke helper programs.
		if goarch == "arm" {
			spec.Linker = "arm-linux-gnueabi-gcc"
			spec.Objcopy = "arm-linux-gnueabi-objcopy"
			spec.GDB = "arm-linux-gnueabi-gdb"
		}
		if goarch == "arm64" {
			spec.Linker = "aarch64-linux-gnu-gcc"
			spec.Objcopy = "aarch64-linux-gnu-objcopy"
			spec.GDB = "aarch64-linux-gnu-gdb"
		}
		if goarch == "386" {
			spec.CFlags = []string{"-m32"}
			spec.LDFlags = []string{"-m32"}
		}
	}
	return &spec, nil
}

// Return the TINYGOROOT, or exit with an error.
func sourceDir() string {
	// Use $TINYGOROOT as root, if available.
	root := os.Getenv("TINYGOROOT")
	if root != "" {
		if !isSourceDir(root) {
			fmt.Fprintln(os.Stderr, "error: $TINYGOROOT was not set to the correct root")
			os.Exit(1)
		}
		return root
	}

	// Find root from executable path.
	path, err := os.Executable()
	if err != nil {
		// Very unlikely. Bail out if it happens.
		panic("could not get executable path: " + err.Error())
	}
	root = filepath.Dir(filepath.Dir(path))
	if isSourceDir(root) {
		return root
	}

	// Fallback: use the original directory from where it was built
	// https://stackoverflow.com/a/32163888/559350
	_, path, _, _ = runtime.Caller(0)
	root = filepath.Dir(path)
	if isSourceDir(root) {
		return root
	}

	fmt.Fprintln(os.Stderr, "error: could not autodetect root directory, set the TINYGOROOT environment variable to override")
	os.Exit(1)
	panic("unreachable")
}

// isSourceDir returns true if the directory looks like a TinyGo source directory.
func isSourceDir(root string) bool {
	_, err := os.Stat(filepath.Join(root, "src/runtime/internal/sys/zversion.go"))
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(root, "src/device/arm/arm.go"))
	return err == nil
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
