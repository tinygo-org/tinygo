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
	Inherits         []string `json:"inherits,omitempty"`
	Triple           string   `json:"llvm-target,omitempty"`
	CPU              string   `json:"cpu,omitempty"`
	ABI              string   `json:"target-abi,omitempty"` // rougly equivalent to -mabi= flag
	Features         string   `json:"features,omitempty"`
	GOOS             string   `json:"goos,omitempty"`
	GOARCH           string   `json:"goarch,omitempty"`
	BuildTags        []string `json:"build-tags,omitempty"`
	GC               string   `json:"gc,omitempty"`
	Scheduler        string   `json:"scheduler,omitempty"`
	Serial           string   `json:"serial,omitempty"` // which serial output to use (uart, usb, none)
	Linker           string   `json:"linker,omitempty"`
	RTLib            string   `json:"rtlib,omitempty"` // compiler runtime library (libgcc, compiler-rt)
	Libc             string   `json:"libc,omitempty"`
	AutoStackSize    *bool    `json:"automatic-stack-size,omitempty"` // Determine stack size automatically at compile time.
	DefaultStackSize uint64   `json:"default-stack-size,omitempty"`   // Default stack size if the size couldn't be determined at compile time.
	CFlags           []string `json:"cflags,omitempty"`
	LDFlags          []string `json:"ldflags,omitempty"`
	LinkerScript     string   `json:"linkerscript,omitempty"`
	ExtraFiles       []string `json:"extra-files,omitempty"`
	RP2040BootPatch  *bool    `json:"rp2040-boot-patch,omitempty"` // Patch RP2040 2nd stage bootloader checksum
	Emulator         string   `json:"emulator,omitempty"`
	FlashCommand     string   `json:"flash-command,omitempty"`
	GDB              []string `json:"gdb,omitempty"`
	PortReset        string   `json:"flash-1200-bps-reset,omitempty"`
	SerialPort       []string `json:"serial-port,omitempty"` // serial port IDs in the form "vid:pid"
	FlashMethod      string   `json:"flash-method,omitempty"`
	FlashVolume      []string `json:"msd-volume-name,omitempty"`
	FlashFilename    string   `json:"msd-firmware-name,omitempty"`
	UF2FamilyID      string   `json:"uf2-family-id,omitempty"`
	BinaryFormat     string   `json:"binary-format,omitempty"`
	OpenOCDInterface string   `json:"openocd-interface,omitempty"`
	OpenOCDTarget    string   `json:"openocd-target,omitempty"`
	OpenOCDTransport string   `json:"openocd-transport,omitempty"`
	OpenOCDCommands  []string `json:"openocd-commands,omitempty"`
	OpenOCDVerify    *bool    `json:"openocd-verify,omitempty"` // enable verify when flashing with openocd
	JLinkDevice      string   `json:"jlink-device,omitempty"`
	CodeModel        string   `json:"code-model,omitempty"`
	RelocationModel  string   `json:"relocation-model,omitempty"`
}

// overrideProperties overrides all properties that are set in child into itself using reflection.
func (spec *TargetSpec) overrideProperties(child *TargetSpec) error {
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
		case reflect.Slice: // for slices, append the field and check for duplicates
			dst.Set(reflect.AppendSlice(dst, src))
			for i := 0; i < dst.Len(); i++ {
				v := dst.Index(i).String()
				for j := i + 1; j < dst.Len(); j++ {
					w := dst.Index(j).String()
					if v == w {
						return fmt.Errorf("duplicate value '%s' in field %s", v, field.Name)
					}
				}
			}
		default:
			return fmt.Errorf("unknown field type: %s", kind)
		}
	}
	return nil
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
//   - targets/ directory inside the compiler sources
//   - a relative or absolute path to custom (project specific) target specification .json file;
//     the Inherits[] could contain the files from target folder (ex. stm32f4disco)
//     as well as path to custom files (ex. myAwesomeProject.json)
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
		err = newSpec.overrideProperties(subtarget)
		if err != nil {
			return err
		}
	}

	// When all properties are loaded, make sure they are properly inherited.
	err := newSpec.overrideProperties(spec)
	if err != nil {
		return err
	}
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
		case "wasm":
			llvmarch = "wasm32"
		default:
			llvmarch = options.GOARCH
		}
		llvmvendor := "unknown"
		llvmos := options.GOOS
		switch llvmos {
		case "darwin":
			// Use macosx* instead of darwin, otherwise darwin/arm64 will refer
			// to iOS!
			llvmos = "macosx10.12.0"
			if llvmarch == "aarch64" {
				// Looks like Apple prefers to call this architecture ARM64
				// instead of AArch64.
				llvmarch = "arm64"
				llvmos = "macosx11.0.0"
			}
			llvmvendor = "apple"
		case "wasip1":
			llvmos = "wasi"
		}
		// Target triples (which actually have four components, but are called
		// triples for historical reasons) have the form:
		//   arch-vendor-os-environment
		target := llvmarch + "-" + llvmvendor + "-" + llvmos
		if options.GOOS == "windows" {
			target += "-gnu"
		} else if options.GOARCH == "arm" {
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
		return nil, fmt.Errorf("%s : %w", options.Target, err)
	}

	if spec.Scheduler == "asyncify" {
		spec.ExtraFiles = append(spec.ExtraFiles, "src/internal/task/task_asyncify_wasm.S")
	}

	return spec, nil
}

func defaultTarget(goos, goarch, triple string) (*TargetSpec, error) {
	// No target spec available. Use the default one, useful on most systems
	// with a regular OS.
	spec := TargetSpec{
		Triple:           triple,
		GOOS:             goos,
		GOARCH:           goarch,
		BuildTags:        []string{goos, goarch},
		GC:               "precise",
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
		spec.CFlags = append(spec.CFlags, "-fno-unwind-tables", "-fno-asynchronous-unwind-tables")
		switch strings.Split(triple, "-")[0] {
		case "armv5":
			spec.Features = "+armv5t,+strict-align,-aes,-bf16,-d32,-dotprod,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-mve.fp,-neon,-sha2,-thumb-mode,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp"
		case "armv6":
			spec.Features = "+armv6,+dsp,+fp64,+strict-align,+vfp2,+vfp2sp,-aes,-d32,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fullfp16,-neon,-sha2,-thumb-mode,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp"
		case "armv7":
			spec.Features = "+armv7-a,+d32,+dsp,+fp64,+neon,+vfp2,+vfp2sp,+vfp3,+vfp3d16,+vfp3d16sp,+vfp3sp,-aes,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fullfp16,-sha2,-thumb-mode,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp"
		}
	case "arm64":
		spec.CPU = "generic"
		if goos == "darwin" {
			spec.Features = "+neon"
		} else { // windows, linux
			spec.Features = "+neon,-fmv"
		}
	case "wasm":
		spec.CPU = "generic"
		spec.Features = "+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext"
		spec.BuildTags = append(spec.BuildTags, "tinygo.wasm")
		spec.CFlags = append(spec.CFlags,
			"-mbulk-memory",
			"-mnontrapping-fptoint",
			"-msign-ext",
		)
	}
	if goos == "darwin" {
		spec.Linker = "ld.lld"
		spec.Libc = "darwin-libSystem"
		arch := strings.Split(triple, "-")[0]
		platformVersion := strings.TrimPrefix(strings.Split(triple, "-")[2], "macosx")
		spec.LDFlags = append(spec.LDFlags,
			"-flavor", "darwin",
			"-dead_strip",
			"-arch", arch,
			"-platform_version", "macos", platformVersion, platformVersion,
		)
	} else if goos == "linux" {
		spec.Linker = "ld.lld"
		spec.RTLib = "compiler-rt"
		spec.Libc = "musl"
		spec.LDFlags = append(spec.LDFlags, "--gc-sections")
	} else if goos == "windows" {
		spec.Linker = "ld.lld"
		spec.Libc = "mingw-w64"
		// Note: using a medium code model, low image base and no ASLR
		// because Go doesn't really need those features. ASLR patches
		// around issues for unsafe languages like C/C++ that are not
		// normally present in Go (without explicitly opting in).
		// For more discussion:
		// https://groups.google.com/g/Golang-nuts/c/Jd9tlNc6jUE/m/Zo-7zIP_m3MJ?pli=1
		switch goarch {
		case "amd64":
			spec.LDFlags = append(spec.LDFlags,
				"-m", "i386pep",
				"--image-base", "0x400000",
			)
		case "arm64":
			spec.LDFlags = append(spec.LDFlags,
				"-m", "arm64pe",
			)
		}
		spec.LDFlags = append(spec.LDFlags,
			"-Bdynamic",
			"--gc-sections",
			"--no-insert-timestamp",
			"--no-dynamicbase",
		)
	} else if goos == "wasip1" {
		spec.GC = "" // use default GC
		spec.Scheduler = "asyncify"
		spec.Linker = "wasm-ld"
		spec.RTLib = "compiler-rt"
		spec.Libc = "wasi-libc"
		spec.DefaultStackSize = 1024 * 64 // 64kB
		spec.LDFlags = append(spec.LDFlags,
			"--stack-first",
			"--no-demangle",
		)
		spec.Emulator = "wasmtime --mapdir=/tmp::{tmpDir} {}"
		spec.ExtraFiles = append(spec.ExtraFiles,
			"src/runtime/asm_tinygowasm.S",
			"src/internal/task/task_asyncify_wasm.S",
		)
	} else {
		spec.LDFlags = append(spec.LDFlags, "-no-pie", "-Wl,--gc-sections") // WARNING: clang < 5.0 requires -nopie
	}
	if goarch != "wasm" {
		suffix := ""
		if goos == "windows" && goarch == "amd64" {
			// Windows uses a different calling convention on amd64 from other
			// operating systems so we need separate assembly files.
			suffix = "_windows"
		}
		spec.ExtraFiles = append(spec.ExtraFiles, "src/runtime/asm_"+goarch+suffix+".S")
		spec.ExtraFiles = append(spec.ExtraFiles, "src/internal/task/task_stack_"+goarch+suffix+".S")
	}
	if goarch != runtime.GOARCH {
		// Some educated guesses as to how to invoke helper programs.
		spec.GDB = []string{"gdb-multiarch"}
		if goos == "linux" {
			switch goarch {
			case "386":
				// amd64 can _usually_ run 32-bit programs, so skip the emulator in that case.
				if runtime.GOARCH != "amd64" {
					spec.Emulator = "qemu-i386 {}"
				}
			case "amd64":
				spec.Emulator = "qemu-x86_64 {}"
			case "arm":
				spec.Emulator = "qemu-arm {}"
			case "arm64":
				spec.Emulator = "qemu-aarch64 {}"
			}
		}
	}
	if goos != runtime.GOOS {
		if goos == "windows" {
			spec.Emulator = "wine {}"
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
