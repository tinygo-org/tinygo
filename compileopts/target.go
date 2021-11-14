package compileopts

// This file loads a target specification from a JSON file.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	Features         string   `json:"features"`
	GOOS             string   `json:"goos"`
	GOARCH           string   `json:"goarch"`
	BuildTags        []string `json:"build-tags"`
	GC               string   `json:"gc"`
	Scheduler        string   `json:"scheduler"`
	Serial           string   `json:"serial"` // which serial output to use (uart, usb, none)
	Linker           string   `json:"linker"`
	RTLib            string   `json:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	Libc             string   `json:"libc"`
	AutoStackSize    *bool    `json:"automatic-stack-size"` // Determine stack size automatically at compile time.
	DefaultStackSize uint64   `json:"default-stack-size"`   // Default stack size if the size couldn't be determined at compile time.
	CFlags           []string `json:"cflags"`
	LDFlags          []string `json:"ldflags"`
	LinkerScript     string   `json:"linkerscript"`
	ExtraFiles       []string `json:"extra-files"`
	RP2040BootPatch  *bool    `json:"rp2040-boot-patch"`        // Patch RP2040 2nd stage bootloader checksum
	Emulator         []string `json:"emulator" override:"copy"` // inherited Emulator must not be append
	FlashCommand     string   `json:"flash-command"`
	GDB              []string `json:"gdb"`
	PortReset        string   `json:"flash-1200-bps-reset"`
	SerialPort       []string `json:"serial-port"` // serial port IDs in the form "acm:vid:pid" or "usb:vid:pid"
	FlashMethod      string   `json:"flash-method"`
	FlashVolume      string   `json:"msd-volume-name"`
	FlashFilename    string   `json:"msd-firmware-name"`
	UF2FamilyID      string   `json:"uf2-family-id"`
	BinaryFormat     string   `json:"binary-format"`
	OpenOCDInterface string   `json:"openocd-interface"`
	OpenOCDTarget    string   `json:"openocd-target"`
	OpenOCDTransport string   `json:"openocd-transport"`
	OpenOCDCommands  []string `json:"openocd-commands"`
	JLinkDevice      string   `json:"jlink-device"`
	CodeModel        string   `json:"code-model"`
	RelocationModel  string   `json:"relocation-model"`
	WasmAbi          string   `json:"wasm-abi"`
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
					dst.Set(reflect.AppendSlice(dst, src))
				default:
					panic("override mode must be 'copy' or 'append' (default). I don't know how to '" + tag + "'.")
				}
			}
		default:
			panic("unknown field type : " + kind.String())
		}
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
		var llvmarch string
		switch options.GOARCH {
		case "386":
			llvmarch = "i386"
		case "amd64":
			llvmarch = "x86_64"
		case "arm64":
			llvmarch = "aarch64"
		case "arm":
			switch options.GOARM {
			case "5":
				llvmarch = "armv5"
			case "6":
				llvmarch = "armv6"
			case "7":
				llvmarch = "armv7"
			default:
				return nil, fmt.Errorf("invalid GOARM=%s, must be 5, 6, or 7", options.GOARM)
			}
		default:
			llvmarch = options.GOARCH
		}
		llvmos := options.GOOS
		if llvmos == "darwin" {
			// Use macosx* instead of darwin, otherwise darwin/arm64 will refer
			// to iOS!
			llvmos = "macosx10.12.0"
			if llvmarch == "aarch64" {
				// Looks like Apple prefers to call this architecture ARM64
				// instead of AArch64.
				llvmarch = "arm64"
			}
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
	// Successfully loaded this target from a built-in .json file. Make sure
	// it includes all parents as specified in the "inherits" key.
	err = spec.resolveInherits()
	if err != nil {
		return nil, err
	}

	if spec.Scheduler == "asyncify" {
		spec.ExtraFiles = append(spec.ExtraFiles, "src/internal/task/task_asyncify_wasm.S")
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
	switch goarch {
	case "386":
		spec.CPU = "pentium4"
		spec.Features = "+cx8,+fxsr,+mmx,+sse,+sse2,+x87"
	case "amd64":
		spec.CPU = "x86-64"
		spec.Features = "+cx8,+fxsr,+mmx,+sse,+sse2,+x87"
	case "arm":
		spec.CPU = "generic"
		switch strings.Split(triple, "-")[0] {
		case "armv5":
			spec.Features = "+armv5t,+strict-align,-thumb-mode"
		case "armv6":
			spec.Features = "+armv6,+dsp,+fp64,+strict-align,+vfp2,+vfp2sp,-thumb-mode"
		case "armv7":
			spec.Features = "+armv7-a,+dsp,+fp64,+vfp2,+vfp2sp,+vfp3d16,+vfp3d16sp,-thumb-mode"
		}
	case "arm64":
		spec.CPU = "generic"
		spec.Features = "+neon"
	}
	if goos == "darwin" {
		spec.CFlags = append(spec.CFlags, "-isysroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
		spec.LDFlags = append(spec.LDFlags, "-Wl,-dead_strip")
	} else if goos == "linux" {
		spec.Linker = "ld.lld"
		spec.RTLib = "compiler-rt"
		spec.Libc = "musl"
		spec.LDFlags = append(spec.LDFlags, "--gc-sections")
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
			spec.Emulator = []string{"qemu-arm"}
		}
		if goarch == "arm64" && goos == "linux" {
			spec.Emulator = []string{"qemu-aarch64"}
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
