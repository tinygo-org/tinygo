package compileopts

// This file loads a target specification from a JSON file.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	ABI              string   `json:"target-abi"` // rougly equivalent to -mabi= flag
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
	RP2040BootPatch  *bool    `json:"rp2040-boot-patch"` // Patch RP2040 2nd stage bootloader checksum
	Emulator         string   `json:"emulator"`
	FlashCommand     string   `json:"flash-command"`
	GDB              []string `json:"gdb"`
	PortReset        string   `json:"flash-1200-bps-reset"`
	SerialPort       []string `json:"serial-port"` // serial port IDs in the form "vid:pid"
	FlashMethod      string   `json:"flash-method"`
	FlashVolume      string   `json:"msd-volume-name"`
	FlashFilename    string   `json:"msd-firmware-name"`
	UF2FamilyID      string   `json:"uf2-family-id"`
	BinaryFormat     string   `json:"binary-format"`
	OpenOCDInterface string   `json:"openocd-interface"`
	OpenOCDTarget    string   `json:"openocd-target"`
	OpenOCDTransport string   `json:"openocd-transport"`
	OpenOCDCommands  []string `json:"openocd-commands"`
	OpenOCDVerify    *bool    `json:"openocd-verify"` // enable verify when flashing with openocd
	JLinkDevice      string   `json:"jlink-device"`
	CodeModel        string   `json:"code-model"`
	RelocationModel  string   `json:"relocation-model"`
	WasmAbi          string   `json:"wasm-abi"`
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
		path = filepath.Join(goenv.Get("GOCACHE"), "targets", strings.ToLower(str)+".json")
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
		default:
			llvmarch = options.GOARCH
		}
		llvmvendor := "unknown"
		llvmos := options.GOOS
		if llvmos == "darwin" {
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

	tinygoRootBundle := filepath.Join(goenv.Get("TINYGOROOT"), "bundle")
	cacheDir := goenv.Get("GOCACHE")

	for _, dir := range []string{"targets"} {

	destDir := filepath.Join(cacheDir, dir)
	tinygoBundleEntries, err := ioutil.ReadDir(tinygoRootBundle)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(destDir, 0777)
	if err != nil {
		return nil, err
	}
	for _, be := range tinygoBundleEntries {
		bundleName := be.Name()
		srcDir := filepath.Join(tinygoRootBundle, bundleName, dir)
		entries, err := ioutil.ReadDir(srcDir)
		if err != nil {
		// return nil, err
			continue
		}
		// Create all symlinks.
		for _, e := range entries {
			if _, err := os.Stat(filepath.Join(destDir, e.Name())); err == nil {
				continue
			}
			err := Symlink(filepath.Join(srcDir, e.Name()), filepath.Join(destDir, e.Name()))
			if err != nil {
				return nil, err
			}
		}
	}
	{
		tinygoSrcDir := filepath.Join(goenv.Get("TINYGOROOT"), dir)
		entries, err := ioutil.ReadDir(tinygoSrcDir)
		if err != nil {
			return nil, err
		}
		// Create all symlinks.
		for _, e := range entries {
			if _, err := os.Stat(filepath.Join(destDir, e.Name())); err == nil {
				continue
			}
			err := Symlink(filepath.Join(tinygoSrcDir, e.Name()), filepath.Join(destDir, e.Name()))
			if err != nil {
				return nil, err
			}
		}
	}

	}
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
		spec.Features = "+neon"
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
		spec.LDFlags = append(spec.LDFlags,
			"-m", "i386pep",
			"-Bdynamic",
			"--image-base", "0x400000",
			"--gc-sections",
			"--no-insert-timestamp",
			"--no-dynamicbase",
		)
	} else {
		spec.LDFlags = append(spec.LDFlags, "-no-pie", "-Wl,--gc-sections") // WARNING: clang < 5.0 requires -nopie
	}
	if goarch != "wasm" {
		suffix := ""
		if goos == "windows" {
			// Windows uses a different calling convention from other operating
			// systems so we need separate assembly files.
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

// Symlink creates a symlink or something similar. On Unix-like systems, it
// always creates a symlink. On Windows, it tries to create a symlink and if
// that fails, creates a hardlink or directory junction instead.
//
// Note that while Windows 10 does support symlinks and allows them to be
// created using os.Symlink, it requires developer mode to be enabled.
// Therefore provide a fallback for when symlinking is not possible.
// Unfortunately this fallback only works when TinyGo is installed on the same
// filesystem as the TinyGo cache and the Go installation (which is usually the
// C drive).
func Symlink(oldname, newname string) error {
	symlinkErr := os.Symlink(oldname, newname)
	if runtime.GOOS == "windows" && symlinkErr != nil {
		// Fallback for when developer mode is disabled.
		// Note that we return the symlink error even if something else fails
		// later on. This is because symlinks are the easiest to support
		// (they're also used on Linux and MacOS) and enabling them is easy:
		// just enable developer mode.
		st, err := os.Stat(oldname)
		if err != nil {
			return symlinkErr
		}
		if st.IsDir() {
			// Make a directory junction. There may be a way to do this
			// programmatically, but it involves a lot of magic. Use the mklink
			// command built into cmd instead (mklink is a builtin, not an
			// external command).
			err := exec.Command("cmd", "/k", "mklink", "/J", newname, oldname).Run()
			if err != nil {
				return symlinkErr
			}
		} else {
			// Try making a hard link.
			err := os.Link(oldname, newname)
			if err != nil {
				// Making a hardlink failed. Try copying the file as a last
				// fallback.
				inf, err := os.Open(oldname)
				if err != nil {
					return err
				}
				defer inf.Close()
				outf, err := os.Create(newname)
				if err != nil {
					return err
				}
				defer outf.Close()
				_, err = io.Copy(outf, inf)
				if err != nil {
					os.Remove(newname)
					return err
				}
				// File was copied.
			}
		}
		return nil // success
	}
	return symlinkErr
}
