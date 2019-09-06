package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// TINYGOROOT is the path to the final location for checking tinygo files. If
// unset (by a -X ldflag), then sourceDir() will fallback to the original build
// directory.
var TINYGOROOT string

// Target specification for a given target. Used for bare metal targets.
//
// The target specification is mostly inspired by Rust:
// https://doc.rust-lang.org/nightly/nightly-rustc/rustc_target/spec/struct.TargetOptions.html
// https://github.com/shepmaster/rust-arduino-blink-led-no-core-with-cargo/blob/master/blink/arduino.json
type TargetSpec struct {
	Inherits   []string `json:"inherits"`
	Triple     string   `json:"llvm-target"`
	CPU        string   `json:"cpu"`
	Features   []string `json:"features"`
	GOOS       string   `json:"goos"`
	GOARCH     string   `json:"goarch"`
	BuildTags  []string `json:"build-tags"`
	GC         string   `json:"gc"`
	Scheduler  string   `json:"scheduler"`
	Compiler   string   `json:"compiler"`
	Linker     string   `json:"linker"`
	RTLib      string   `json:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	CFlags     []string `json:"cflags"`
	LDFlags    []string `json:"ldflags"`
	ExtraFiles []string `json:"extra-files"`
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
	spec.Features = append(spec.Features, spec2.Features...)
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
	if spec2.Scheduler != "" {
		spec.Scheduler = spec2.Scheduler
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

// loadFromGivenStr loads the TargetSpec from the given string that could be:
// - targets/ directory inside the compiler sources
// - a relative or absolute path to custom (project specific) target specification .json file;
//   the Inherits[] could contain the files from target folder (ex. stm32f4disco)
//   as well as path to custom files (ex. myAwesomeProject.json)
func (spec *TargetSpec) loadFromGivenStr(str string) error {
	path := ""
	if strings.HasSuffix(str, ".json") {
		path, _ = filepath.Abs(str)
	} else {
		path = filepath.Join(sourceDir(), "targets", strings.ToLower(str)+".json")
	}
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
		err := subtarget.loadFromGivenStr(name)
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
		if goarch == "arm" {
			target += "-gnueabihf"
		}
		return defaultTarget(goos, goarch, target)
	}

	// See whether there is a target specification for this target (e.g.
	// Arduino).
	spec := &TargetSpec{}
	err := spec.loadFromGivenStr(target)
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
		if strings.HasPrefix(goos, "darwin") {
			goos = "darwin"
		}
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
		Compiler:  "clang",
		Linker:    "cc",
		GDB:       "gdb",
		GDBCmds:   []string{"run"},
	}
	if goos == "darwin" {
		spec.LDFlags = append(spec.LDFlags, "-Wl,-dead_strip")
	} else {
		spec.LDFlags = append(spec.LDFlags, "-no-pie", "-Wl,--gc-sections") // WARNING: clang < 5.0 requires -nopie
	}
	if goarch != runtime.GOARCH {
		// Some educated guesses as to how to invoke helper programs.
		if goarch == "arm" && goos == "linux" {
			spec.Linker = "arm-linux-gnueabihf-gcc"
			spec.GDB = "arm-linux-gnueabihf-gdb"
			spec.Emulator = []string{"qemu-arm", "-L", "/usr/arm-linux-gnueabihf"}
		}
		if goarch == "arm64" && goos == "linux" {
			spec.Linker = "aarch64-linux-gnu-gcc"
			spec.GDB = "aarch64-linux-gnu-gdb"
			spec.Emulator = []string{"qemu-aarch64", "-L", "/usr/aarch64-linux-gnu"}
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

	if TINYGOROOT != "" {
		if !isSourceDir(TINYGOROOT) {
			fmt.Fprintln(os.Stderr, "error: TINYGOROOT was not set to the correct root")
			os.Exit(1)
		}
		return TINYGOROOT
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

// getGoroot returns an appropriate GOROOT from various sources. If it can't be
// found, it returns an empty string.
func getGoroot() string {
	goroot := os.Getenv("GOROOT")
	if goroot != "" {
		// An explicitly set GOROOT always has preference.
		return goroot
	}

	// Check for the location of the 'go' binary and base GOROOT on that.
	binpath, err := exec.LookPath("go")
	if err == nil {
		binpath, err = filepath.EvalSymlinks(binpath)
		if err == nil {
			goroot := filepath.Dir(filepath.Dir(binpath))
			if isGoroot(goroot) {
				return goroot
			}
		}
	}

	// Check what GOROOT was at compile time.
	if isGoroot(runtime.GOROOT()) {
		return runtime.GOROOT()
	}

	// Check for some standard locations, as a last resort.
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			"/usr/local/go", // manually installed
			"/usr/lib/go",   // from the distribution
		}
	case "darwin":
		candidates = []string{
			"/usr/local/go",             // manually installed
			"/usr/local/opt/go/libexec", // from Homebrew
		}
	}

	for _, candidate := range candidates {
		if isGoroot(candidate) {
			return candidate
		}
	}

	// Can't find GOROOT...
	return ""
}

// isGoroot checks whether the given path looks like a GOROOT.
func isGoroot(goroot string) bool {
	_, err := os.Stat(filepath.Join(goroot, "src", "runtime", "internal", "sys", "zversion.go"))
	return err == nil
}

// getGorootVersion returns the major and minor version for a given GOROOT path.
// If the goroot cannot be determined, (0, 0) is returned.
func getGorootVersion(goroot string) (major, minor int, err error) {
	var s string
	var n int
	var trailing string

	if data, err := ioutil.ReadFile(filepath.Join(
		goroot, "src", "runtime", "internal", "sys", "zversion.go")); err == nil {

		r := regexp.MustCompile("const TheVersion = `(.*)`")
		matches := r.FindSubmatch(data)
		if len(matches) != 2 {
			return 0, 0, errors.New("Invalid go version output:\n" + string(data))
		}

		s = string(matches[1])

	} else if data, err := ioutil.ReadFile(filepath.Join(goroot, "VERSION")); err == nil {
		s = string(data)

	} else {
		return 0, 0, err
	}

	if s == "" || s[:2] != "go" {
		return 0, 0, errors.New("could not parse Go version: version does not start with 'go' prefix")
	}

	parts := strings.Split(s[2:], ".")
	if len(parts) < 2 {
		return 0, 0, errors.New("could not parse Go version: version has less than two parts")
	}

	// Ignore the errors, we don't really handle errors here anyway.
	n, err = fmt.Sscanf(s, "go%d.%d%s", &major, &minor, &trailing)
	if n == 2 && err == io.EOF {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}
	return
}

// getClangHeaderPath returns the path to the built-in Clang headers. It tries
// multiple locations, which should make it find the directory when installed in
// various ways.
func getClangHeaderPath(TINYGOROOT string) string {
	// Check whether we're running from the source directory.
	path := filepath.Join(TINYGOROOT, "llvm", "tools", "clang", "lib", "Headers")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	// Check whether we're running from the installation directory.
	path = filepath.Join(TINYGOROOT, "lib", "clang", "include")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	// It looks like we are built with a system-installed LLVM. Do a last
	// attempt: try to use Clang headers relative to the clang binary.
	for _, cmdName := range commands["clang"] {
		binpath, err := exec.LookPath(cmdName)
		if err == nil {
			// This should be the command that will also be used by
			// execCommand. To avoid inconsistencies, make sure we use the
			// headers relative to this command.
			binpath, err = filepath.EvalSymlinks(binpath)
			if err != nil {
				// Unexpected.
				return ""
			}
			// Example executable:
			//     /usr/lib/llvm-8/bin/clang
			// Example include path:
			//     /usr/lib/llvm-8/lib/clang/8.0.1/include/
			llvmRoot := filepath.Dir(filepath.Dir(binpath))
			clangVersionRoot := filepath.Join(llvmRoot, "lib", "clang")
			dirnames, err := ioutil.ReadDir(clangVersionRoot)
			if err != nil || len(dirnames) != 1 {
				// Unexpected.
				return ""
			}
			return filepath.Join(clangVersionRoot, dirnames[0].Name(), "include")
		}
	}

	// Could not find it.
	return ""
}
