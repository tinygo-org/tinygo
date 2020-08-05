package compileopts

// This file loads a target specification from a JSON file.

import (
	"encoding/json"
	"errors"
	"io"
	"os"
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
	Features         []string `json:"features"`
	GOOS             string   `json:"goos"`
	GOARCH           string   `json:"goarch"`
	BuildTags        []string `json:"build-tags"`
	GC               string   `json:"gc"`
	Scheduler        string   `json:"scheduler"`
	Compiler         string   `json:"compiler"`
	Linker           string   `json:"linker"`
	RTLib            string   `json:"rtlib"` // compiler runtime library (libgcc, compiler-rt)
	Libc             string   `json:"libc"`
	CFlags           []string `json:"cflags"`
	LDFlags          []string `json:"ldflags"`
	LinkerScript     string   `json:"linkerscript"`
	ExtraFiles       []string `json:"extra-files"`
	Emulator         []string `json:"emulator" copy:"replace"` // inherited Emulator must not be appened
	FlashCommand     string   `json:"flash-command"`
	GDB              string   `json:"gdb"`
	PortReset        string   `json:"flash-1200-bps-reset"`
	FlashMethod      string   `json:"flash-method"`
	FlashVolume      string   `json:"msd-volume-name"`
	FlashFilename    string   `json:"msd-firmware-name"`
	UF2FamilyID      string   `json:"uf2-family-id"`
	OpenOCDInterface string   `json:"openocd-interface"`
	OpenOCDTarget    string   `json:"openocd-target"`
	OpenOCDTransport string   `json:"openocd-transport"`
	JLinkDevice      string   `json:"jlink-device"`
	CodeModel        string   `json:"code-model"`
	RelocationModel  string   `json:"relocation-model"`
}

// copyProperties copies all properties that are set in spec2 into itself using reflection.
func (spec *TargetSpec) copyProperties(spec2 *TargetSpec) {
	specType := reflect.TypeOf(spec).Elem()
	specValue := reflect.ValueOf(spec).Elem()
	spec2Value := reflect.ValueOf(spec2).Elem()
	for i := 0; i < specType.NumField(); i++ {
		field := specType.Field(i)
		switch kind := field.Type.Kind(); kind {
		case reflect.String: // for strings, just copy the field of spec2 to spec (if not empty)
			if v := spec2Value.Field(i); v.Len() > 0 {
				specValue.Field(i).Set(v)
			}
		case reflect.Slice: // for slices...
			tag, ok := field.Tag.Lookup("copy")
			if ok && tag == "replace" {
				// ... copy the field of spec2 to spec if the field has a "copy" tag
				// and the tag is "replace" (and is not [])
				if v := spec2Value.Field(i); v.Len() > 0 {
					specValue.Field(i).Set(v)
				}
			} else if ok && tag != "append" {
				panic("copy mode must be 'replace' or 'append' (default). I don't know how to '" + tag + "'.")
			} else {
				// ... else append the field of spec2 to spec
				specValue.Field(i).Set(reflect.AppendSlice(specValue.Field(i), spec2Value.Field(i)))
			}
		default:
			panic("field must be a string or a slice. '" + kind.String() + "' is not expected.")
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
		if tripleSplit[0] == "arm" {
			// LLVM and Clang have a different idea of what "arm" means, so
			// upgrade to a slightly more modern ARM. In fact, when you pass
			// --target=arm--linux-gnueabihf to Clang, it will convert that
			// internally to armv7-unknown-linux-gnueabihf. Changing the
			// architecture to armv7 will keep things consistent.
			tripleSplit[0] = "armv7"
		}
		goos := tripleSplit[2]
		if strings.HasPrefix(goos, "darwin") {
			goos = "darwin"
		}
		goarch := map[string]string{ // map from LLVM arch to Go arch
			"i386":    "386",
			"x86_64":  "amd64",
			"aarch64": "arm64",
			"armv7":   "arm",
		}[tripleSplit[0]]
		if goarch == "" {
			goarch = tripleSplit[0]
		}
		return defaultTarget(goos, goarch, strings.Join(tripleSplit, "-"))
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
		CFlags:      []string{"--target=" + triple},
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
			spec.CFlags = append(spec.CFlags, "--sysroot=/usr/arm-linux-gnueabihf")
			spec.Linker = "arm-linux-gnueabihf-gcc"
			spec.GDB = "arm-linux-gnueabihf-gdb"
			spec.Emulator = []string{"qemu-arm", "-L", "/usr/arm-linux-gnueabihf"}
		}
		if goarch == "arm64" && goos == "linux" {
			spec.CFlags = append(spec.CFlags, "--sysroot=/usr/aarch64-linux-gnu")
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
