package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// Target specification for a given target. Used for bare metal targets.
//
// The target specification is mostly inspired by Rust:
// https://doc.rust-lang.org/nightly/nightly-rustc/rustc_target/spec/struct.TargetOptions.html
// https://github.com/shepmaster/rust-arduino-blink-led-no-core-with-cargo/blob/master/blink/arduino.json
type TargetSpec struct {
	Inherits         []string `json:"inherits"`
	Triple           string   `json:"llvm-target"`
	CPU              string   `json:"cpu"`
	Features         []string `json:"features"`
	GOOS             string   `json:"goos"`
	GOARCH           string   `json:"goarch"`
	BuildTags        []string `json:"build-tags"`
	GC               string   `json:"gc"`
	Scheduler        string   `json:"scheduler"`
	Compiler         string   `json:"compiler"`
	Linker           string   `json:"linker"`
	RTLib            string   `json:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	CFlags           []string `json:"cflags"`
	LDFlags          []string `json:"ldflags"`
	ExtraFiles       []string `json:"extra-files"`
	Emulator         []string `json:"emulator"`
	FlashCommand     string   `json:"flash-command"`
	OCDDaemon        []string `json:"ocd-daemon"`
	GDB              string   `json:"gdb"`
	PortReset        string   `json:"flash-1200-bps-reset"`
	FlashMethod      string   `json:"flash-method"`
	FlashVolume      string   `json:"msd-volume-name"`
	FlashFilename    string   `json:"msd-firmware-name"`
	OpenOCDInterface string   `json:"openocd-interface"`
	OpenOCDTarget    string   `json:"openocd-target"`
	OpenOCDTransport string   `json:"openocd-transport"`
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
	if spec2.FlashCommand != "" {
		spec.FlashCommand = spec2.FlashCommand
	}
	if len(spec2.OCDDaemon) != 0 {
		spec.OCDDaemon = spec2.OCDDaemon
	}
	if spec2.GDB != "" {
		spec.GDB = spec2.GDB
	}
	if spec2.PortReset != "" {
		spec.PortReset = spec2.PortReset
	}
	if spec2.FlashMethod != "" {
		spec.FlashMethod = spec2.FlashMethod
	}
	if spec2.FlashVolume != "" {
		spec.FlashVolume = spec2.FlashVolume
	}
	if spec2.FlashFilename != "" {
		spec.FlashFilename = spec2.FlashFilename
	}
	if spec2.OpenOCDInterface != "" {
		spec.OpenOCDInterface = spec2.OpenOCDInterface
	}
	if spec2.OpenOCDTarget != "" {
		spec.OpenOCDTarget = spec2.OpenOCDTarget
	}
	if spec2.OpenOCDTransport != "" {
		spec.OpenOCDTransport = spec2.OpenOCDTransport
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
		path = filepath.Join(goenv.Get("TINYGOROOT"), "targets", strings.ToLower(str)+".json")
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
		goos := goenv.Get("GOOS")
		goarch := goenv.Get("GOARCH")
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
		if len(tripleSplit) < 3 {
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
		Triple:      triple,
		GOOS:        goos,
		GOARCH:      goarch,
		BuildTags:   []string{goos, goarch},
		Compiler:    "clang",
		Linker:      "cc",
		GDB:         "gdb",
		PortReset:   "false",
		FlashMethod: "native",
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

// OpenOCDConfiguration returns a list of command line arguments to OpenOCD.
// This list of command-line arguments is based on the various OpenOCD-related
// flags in the target specification.
func (spec *TargetSpec) OpenOCDConfiguration() (args []string, err error) {
	if spec.OpenOCDInterface == "" {
		return nil, errors.New("OpenOCD programmer not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(spec.OpenOCDInterface) {
		return nil, fmt.Errorf("OpenOCD programmer has an invalid name: %#v", spec.OpenOCDInterface)
	}
	if spec.OpenOCDTarget == "" {
		return nil, errors.New("OpenOCD chip not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(spec.OpenOCDTarget) {
		return nil, fmt.Errorf("OpenOCD target has an invalid name: %#v", spec.OpenOCDTarget)
	}
	if spec.OpenOCDTransport != "" && spec.OpenOCDTransport != "swd" {
		return nil, fmt.Errorf("unknown OpenOCD transport: %#v", spec.OpenOCDTransport)
	}
	args = []string{"-f", "interface/" + spec.OpenOCDInterface + ".cfg"}
	if spec.OpenOCDTransport != "" {
		args = append(args, "-c", "transport select "+spec.OpenOCDTransport)
	}
	args = append(args, "-f", "target/"+spec.OpenOCDTarget+".cfg")
	return args, nil
}

// getGorootVersion returns the major and minor version for a given GOROOT path.
// If the goroot cannot be determined, (0, 0) is returned.
func getGorootVersion(goroot string) (major, minor int, err error) {
	s, err := getGorootVersionString(goroot)
	if err != nil {
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
	var trailing string
	n, err := fmt.Sscanf(s, "go%d.%d%s", &major, &minor, &trailing)
	if n == 2 && err == io.EOF {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}
	return
}

// getGorootVersionString returns the version string as reported by the Go
// toolchain for the given GOROOT path. It is usually of the form `go1.x.y` but
// can have some variations (for beta releases, for example).
func getGorootVersionString(goroot string) (string, error) {
	if data, err := ioutil.ReadFile(filepath.Join(
		goroot, "src", "runtime", "internal", "sys", "zversion.go")); err == nil {

		r := regexp.MustCompile("const TheVersion = `(.*)`")
		matches := r.FindSubmatch(data)
		if len(matches) != 2 {
			return "", errors.New("Invalid go version output:\n" + string(data))
		}

		return string(matches[1]), nil

	} else if data, err := ioutil.ReadFile(filepath.Join(goroot, "VERSION")); err == nil {
		return string(data), nil

	} else {
		return "", err
	}
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
