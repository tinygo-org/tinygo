package compileopts

// This file loads a target specification from a YAML file.

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
	"gopkg.in/yaml.v2"
)

// Target specification for a given target. Used for bare metal targets.
//
// The target specification is mostly inspired by Rust:
// https://doc.rust-lang.org/nightly/nightly-rustc/rustc_target/spec/struct.TargetOptions.html
// https://github.com/shepmaster/rust-arduino-blink-led-no-core-with-cargo/blob/master/blink/arduino.json
type TargetSpec struct {
	Inherits         []string `yaml:"inherits"`
	Triple           string   `yaml:"llvm-target"`
	CPU              string   `yaml:"cpu"`
	Features         []string `yaml:"features"`
	GOOS             string   `yaml:"goos"`
	GOARCH           string   `yaml:"goarch"`
	BuildTags        []string `yaml:"build-tags"`
	GC               string   `yaml:"gc"`
	Scheduler        string   `yaml:"scheduler"`
	Serial           string   `yaml:"serial"` // which serial output to use (uart, usb, none)
	Linker           string   `yaml:"linker"`
	RTLib            string   `yaml:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	Libc             string   `yaml:"libc"`
	AutoStackSize    *bool    `yaml:"automatic-stack-size"` // Determine stack size automatically at compile time.
	DefaultStackSize uint64   `yaml:"default-stack-size"`   // Default stack size if the size couldn't be determined at compile time.
	CFlags           []string `yaml:"cflags"`
	LDFlags          []string `yaml:"ldflags"`
	LinkerScript     string   `yaml:"linkerscript"`
	ExtraFiles       []string `yaml:"extra-files"`
	RP2040BootPatch  *bool    `yaml:"rp2040-boot-patch"`        // Patch RP2040 2nd stage bootloader checksum
	Emulator         []string `yaml:"emulator" override:"copy"` // inherited Emulator must not be append
	FlashCommand     string   `yaml:"flash-command"`
	GDB              []string `yaml:"gdb"`
	PortReset        string   `yaml:"flash-1200-bps-reset"`
	SerialPort       []string `yaml:"serial-port"` // serial port IDs in the form "acm:vid:pid" or "usb:vid:pid"
	FlashMethod      string   `yaml:"flash-method"`
	FlashVolume      string   `yaml:"msd-volume-name"`
	FlashFilename    string   `yaml:"msd-firmware-name"`
	UF2FamilyID      string   `yaml:"uf2-family-id"`
	BinaryFormat     string   `yaml:"binary-format"`
	OpenOCDInterface string   `yaml:"openocd-interface"`
	OpenOCDTarget    string   `yaml:"openocd-target"`
	OpenOCDTransport string   `yaml:"openocd-transport"`
	OpenOCDCommands  []string `yaml:"openocd-commands"`
	JLinkDevice      string   `yaml:"jlink-device"`
	CodeModel        string   `yaml:"code-model"`
	RelocationModel  string   `yaml:"relocation-model"`
	WasmAbi          string   `yaml:"wasm-abi"`
}

// overrideProperties overrides all properties that are set in child into itself using reflection.
func (spec *TargetSpec) overrideProperties(child *TargetSpec) {
	specType := reflect.TypeOf(spec).Elem()
	specValue := reflect.ValueOf(spec).Elem()
	childValue := reflect.ValueOf(child).Elem()

	for i := 0; i < specType.NumField(); i++ {
		field := specType.Field(i)
		src := childValue.Field(i)
		dst := specValue.Field(i)

		switch kind := field.Type.Kind(); kind {
		case reflect.String: // for strings, just copy the field of child to spec if not empty
			if src.Len() > 0 {
				dst.Set(src)
			}
		case reflect.Uint, reflect.Uint32, reflect.Uint64: // for Uint, copy if not zero
			if src.Uint() != 0 {
				dst.Set(src)
			}
		case reflect.Ptr: // for pointers, copy if not nil
			if !src.IsNil() {
				dst.Set(src)
			}
		case reflect.Slice: // for slices...
			if src.Len() > 0 { // ... if not empty ...
				switch tag := field.Tag.Get("override"); tag {
				case "copy":
					// copy the field of child to spec
					dst.Set(src)
				case "append", "":
					// or append the field of child to spec
					dst.Set(reflect.AppendSlice(src, dst))
				default:
					panic("override mode must be 'copy' or 'append' (default). I don't know how to '" + tag + "'.")
				}
			}
		default:
			panic("unknown field type : " + kind.String())
		}
	}
}

// load reads a target specification from the YAML in the given io.Reader. It
// may load more targets specified using the "inherits" property.
func (spec *TargetSpec) load(r io.Reader) error {
	err := yaml.NewDecoder(r).Decode(spec)
	if err != nil {
		return err
	}

	return nil
}

// loadFromGivenStr loads the TargetSpec from the given string that could be:
// - targets/ directory inside the compiler sources
// - a relative or absolute path to custom (project specific) target specification .yaml file;
//   the Inherits[] could contain the files from target folder (ex. stm32f4disco)
//   as well as path to custom files (ex. myAwesomeProject.yaml)
func (spec *TargetSpec) loadFromGivenStr(str string) error {
	path := ""
	if strings.HasSuffix(str, ".yaml") {
		path, _ = filepath.Abs(str)
	} else {
		path = filepath.Join(goenv.Get("TINYGOROOT"), "targets", strings.ToLower(str)+".yaml")
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
		newSpec.overrideProperties(subtarget)
	}

	// When all properties are loaded, make sure they are properly inherited.
	newSpec.overrideProperties(spec)
	*spec = *newSpec

	return nil
}

// Load a target specification.
func LoadTarget(options *Options) (*TargetSpec, error) {
	if options.Target == "" {
		// Configure based on GOOS/GOARCH environment variables (falling back to
		// runtime.GOOS/runtime.GOARCH), and generate a LLVM target based on it.
		llvmos := options.GOOS
		llvmarch := map[string]string{
			"386":   "i386",
			"amd64": "x86_64",
			"arm64": "aarch64",
			"arm":   "armv7",
		}[options.GOARCH]
		if llvmarch == "" {
			llvmarch = options.GOARCH
		}
		// Target triples (which actually have four components, but are called
		// triples for historical reasons) have the form:
		//   arch-vendor-os-environment
		target := llvmarch + "-unknown-" + llvmos
		if options.GOARCH == "arm" {
			target += "-gnueabihf"
		}
		return defaultTarget(options.GOOS, options.GOARCH, target)
	}

	// See whether there is a target specification for this target (e.g.
	// Arduino).
	spec := &TargetSpec{}
	err := spec.loadFromGivenStr(options.Target)
	if err != nil {
		return nil, err
	}
	// Successfully loaded this target from a built-in .yaml file. Make sure
	// it includes all parents as specified in the "inherits" key.
	err = spec.resolveInherits()
	if err != nil {
		return nil, err
	}
	return spec, nil
}

// WindowsBuildNotSupportedErr is being thrown, when goos is windows and no target has been specified.
var WindowsBuildNotSupportedErr = errors.New("Building Windows binaries is currently not supported. Try specifying a different target")

func defaultTarget(goos, goarch, triple string) (*TargetSpec, error) {
	if goos == "windows" {
		return nil, WindowsBuildNotSupportedErr
	}
	// No target spec available. Use the default one, useful on most systems
	// with a regular OS.
	spec := TargetSpec{
		Triple:           triple,
		GOOS:             goos,
		GOARCH:           goarch,
		BuildTags:        []string{goos, goarch},
		Scheduler:        "tasks",
		Linker:           "cc",
		DefaultStackSize: 1024 * 64, // 64kB
		GDB:              []string{"gdb"},
		PortReset:        "false",
	}
	if goarch == "386" {
		spec.CPU = "pentium4"
	}
	if goos == "darwin" {
		spec.CFlags = append(spec.CFlags, "-isysroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
		spec.LDFlags = append(spec.LDFlags, "-Wl,-dead_strip")
	} else {
		spec.LDFlags = append(spec.LDFlags, "-no-pie", "-Wl,--gc-sections") // WARNING: clang < 5.0 requires -nopie
	}
	if goarch != "wasm" {
		spec.ExtraFiles = append(spec.ExtraFiles, "src/runtime/gc_"+goarch+".S")
		spec.ExtraFiles = append(spec.ExtraFiles, "src/internal/task/task_stack_"+goarch+".S")
	}
	if goarch != runtime.GOARCH {
		// Some educated guesses as to how to invoke helper programs.
		spec.GDB = []string{"gdb-multiarch"}
		if goarch == "arm" && goos == "linux" {
			spec.CFlags = append(spec.CFlags, "--sysroot=/usr/arm-linux-gnueabihf")
			spec.Linker = "arm-linux-gnueabihf-gcc"
			spec.Emulator = []string{"qemu-arm", "-L", "/usr/arm-linux-gnueabihf"}
		}
		if goarch == "arm64" && goos == "linux" {
			spec.CFlags = append(spec.CFlags, "--sysroot=/usr/aarch64-linux-gnu")
			spec.Linker = "aarch64-linux-gnu-gcc"
			spec.Emulator = []string{"qemu-aarch64", "-L", "/usr/aarch64-linux-gnu"}
		}
		if goarch == "386" && runtime.GOARCH == "amd64" {
			spec.CFlags = append(spec.CFlags, "-m32")
			spec.LDFlags = append(spec.LDFlags, "-m32")
		}
	}
	return &spec, nil
}

// LookupGDB looks up a gdb executable.
func (spec *TargetSpec) LookupGDB() (string, error) {
	if len(spec.GDB) == 0 {
		return "", errors.New("gdb not configured in the target specification")
	}
	for _, d := range spec.GDB {
		_, err := exec.LookPath(d)
		if err == nil {
			return d, nil
		}
	}
	return "", errors.New("no gdb found configured in the target specification (" + strings.Join(spec.GDB, ", ") + ")")
}
