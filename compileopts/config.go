// Package compileopts contains the configuration for a single to-be-built
// binary.
package compileopts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// Config keeps all configuration affecting the build in a single struct.
type Config struct {
	Options        *Options
	Target         *TargetSpec
	GoMinorVersion int
	ClangHeaders   string // Clang built-in header include path
	TestConfig     TestConfig
}

// Triple returns the LLVM target triple, like armv6m-unknown-unknown-eabi.
func (c *Config) Triple() string {
	return c.Target.Triple
}

// CPU returns the LLVM CPU name, like atmega328p or arm7tdmi. It may return an
// empty string if the CPU name is not known.
func (c *Config) CPU() string {
	return c.Target.CPU
}

// Features returns a list of features this CPU supports. For example, for a
// RISC-V processor, that could be "+a,+c,+m". For many targets, an empty list
// will be returned.
func (c *Config) Features() string {
	if c.Target.Features == "" {
		return c.Options.LLVMFeatures
	}
	if c.Options.LLVMFeatures == "" {
		return c.Target.Features
	}
	return c.Target.Features + "," + c.Options.LLVMFeatures
}

// GOOS returns the GOOS of the target. This might not always be the actual OS:
// for example, bare-metal targets will usually pretend to be linux to get the
// standard library to compile.
func (c *Config) GOOS() string {
	return c.Target.GOOS
}

// GOARCH returns the GOARCH of the target. This might not always be the actual
// archtecture: for example, the AVR target is not supported by the Go standard
// library so such targets will usually pretend to be linux/arm.
func (c *Config) GOARCH() string {
	return c.Target.GOARCH
}

// GOARM will return the GOARM environment variable given to the compiler when
// building a program.
func (c *Config) GOARM() string {
	return c.Options.GOARM
}

// BuildTags returns the complete list of build tags used during this build.
func (c *Config) BuildTags() []string {
	tags := append(c.Target.BuildTags, []string{"tinygo", "math_big_pure_go", "gc." + c.GC(), "scheduler." + c.Scheduler(), "serial." + c.Serial()}...)
	for i := 1; i <= c.GoMinorVersion; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	if extraTags := strings.Fields(c.Options.Tags); len(extraTags) != 0 {
		tags = append(tags, extraTags...)
	}
	return tags
}

// CgoEnabled returns true if (and only if) CGo is enabled. It is true by
// default and false if CGO_ENABLED is set to "0".
func (c *Config) CgoEnabled() bool {
	return goenv.Get("CGO_ENABLED") == "1"
}

// GC returns the garbage collection strategy in use on this platform. Valid
// values are "none", "leaking", and "conservative".
func (c *Config) GC() string {
	if c.Options.GC != "" {
		return c.Options.GC
	}
	if c.Target.GC != "" {
		return c.Target.GC
	}
	return "conservative"
}

// NeedsStackObjects returns true if the compiler should insert stack objects
// that can be traced by the garbage collector.
func (c *Config) NeedsStackObjects() bool {
	switch c.GC() {
	case "conservative":
		for _, tag := range c.BuildTags() {
			if tag == "tinygo.wasm" {
				return true
			}
		}

		return false
	default:
		return false
	}
}

// Scheduler returns the scheduler implementation. Valid values are "none",
//"coroutines" and "tasks".
func (c *Config) Scheduler() string {
	if c.Options.Scheduler != "" {
		return c.Options.Scheduler
	}
	if c.Target.Scheduler != "" {
		return c.Target.Scheduler
	}
	// Fall back to coroutines, which are supported everywhere.
	return "coroutines"
}

// Serial returns the serial implementation for this build configuration: uart,
// usb (meaning USB-CDC), or none.
func (c *Config) Serial() string {
	if c.Options.Serial != "" {
		return c.Options.Serial
	}
	if c.Target.Serial != "" {
		return c.Target.Serial
	}
	return "none"
}

// OptLevels returns the optimization level (0-2), size level (0-2), and inliner
// threshold as used in the LLVM optimization pipeline.
func (c *Config) OptLevels() (optLevel, sizeLevel int, inlinerThreshold uint) {
	switch c.Options.Opt {
	case "none", "0":
		return 0, 0, 0 // -O0
	case "1":
		return 1, 0, 0 // -O1
	case "2":
		return 2, 0, 225 // -O2
	case "s":
		return 2, 1, 225 // -Os
	case "z":
		return 2, 2, 5 // -Oz, default
	default:
		// This is not shown to the user: valid choices are already checked as
		// part of Options.Verify(). It is here as a sanity check.
		panic("unknown optimization level: -opt=" + c.Options.Opt)
	}
}

// FuncImplementation picks an appropriate func value implementation for the
// target.
func (c *Config) FuncImplementation() string {
	switch c.Scheduler() {
	case "tasks", "asyncify":
		// A func value is implemented as a pair of pointers:
		//     {context, function pointer}
		// where the context may be a pointer to a heap-allocated struct
		// containing the free variables, or it may be undef if the function
		// being pointed to doesn't need a context. The function pointer is a
		// regular function pointer.
		return "doubleword"
	case "none", "coroutines":
		// As "doubleword", but with the function pointer replaced by a unique
		// ID per function signature. Function values are called by using a
		// switch statement and choosing which function to call.
		// Pick the switch implementation with the coroutines scheduler, as it
		// allows the use of blocking inside a function that is used as a func
		// value.
		return "switch"
	default:
		panic("unknown scheduler type")
	}
}

// PanicStrategy returns the panic strategy selected for this target. Valid
// values are "print" (print the panic value, then exit) or "trap" (issue a trap
// instruction).
func (c *Config) PanicStrategy() string {
	return c.Options.PanicStrategy
}

// AutomaticStackSize returns whether goroutine stack sizes should be determined
// automatically at compile time, if possible. If it is false, no attempt is
// made.
func (c *Config) AutomaticStackSize() bool {
	if c.Target.AutoStackSize != nil && c.Scheduler() == "tasks" {
		return *c.Target.AutoStackSize
	}
	return false
}

// RP2040BootPatch returns whether the RP2040 boot patch should be applied that
// calculates and patches in the checksum for the 2nd stage bootloader.
func (c *Config) RP2040BootPatch() bool {
	if c.Target.RP2040BootPatch != nil {
		return *c.Target.RP2040BootPatch
	}
	return false
}

// MuslArchitecture returns the architecture name as used in musl libc. It is
// usually the same as the first part of the LLVM triple, but not always.
func MuslArchitecture(triple string) string {
	arch := strings.Split(triple, "-")[0]
	if strings.HasPrefix(arch, "arm") || strings.HasPrefix(arch, "thumb") {
		arch = "arm"
	}
	return arch
}

// LibcPath returns the path to the libc directory. The libc path will be either
// a precompiled libc shipped with a TinyGo build, or a libc path in the cache
// directory (which might not yet be built).
func (c *Config) LibcPath(name string) (path string, precompiled bool) {
	// Try to load a precompiled library.
	precompiledDir := filepath.Join(goenv.Get("TINYGOROOT"), "pkg", c.Triple(), name)
	if _, err := os.Stat(precompiledDir); err == nil {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return precompiledDir, true
	}

	// No precompiled library found. Determine the path name that will be used
	// in the build cache.
	var outname string
	if c.CPU() != "" {
		outname = name + "-" + c.Triple() + "-" + c.CPU()
	} else {
		outname = name + "-" + c.Triple()
	}
	return filepath.Join(goenv.Get("GOCACHE"), outname), false
}

// CFlags returns the flags to pass to the C compiler. This is necessary for CGo
// preprocessing.
func (c *Config) CFlags() []string {
	var cflags []string
	for _, flag := range c.Target.CFlags {
		cflags = append(cflags, strings.ReplaceAll(flag, "{root}", goenv.Get("TINYGOROOT")))
	}
	switch c.Target.Libc {
	case "picolibc":
		root := goenv.Get("TINYGOROOT")
		picolibcDir := filepath.Join(root, "lib", "picolibc", "newlib", "libc")
		path, _ := c.LibcPath("picolibc")
		cflags = append(cflags,
			"--sysroot="+path,
			"-isystem", filepath.Join(picolibcDir, "include"),
			"-isystem", filepath.Join(picolibcDir, "tinystdio"),
		)
	case "musl":
		root := goenv.Get("TINYGOROOT")
		path, _ := c.LibcPath("musl")
		arch := MuslArchitecture(c.Triple())
		cflags = append(cflags,
			"-nostdlibinc",
			"-isystem", filepath.Join(path, "include"),
			"-isystem", filepath.Join(root, "lib", "musl", "arch", arch),
			"-isystem", filepath.Join(root, "lib", "musl", "include"),
		)
	case "wasi-libc":
		root := goenv.Get("TINYGOROOT")
		cflags = append(cflags, "--sysroot="+root+"/lib/wasi-libc/sysroot")
	case "mingw-w64":
		root := goenv.Get("TINYGOROOT")
		path, _ := c.LibcPath("mingw-w64")
		cflags = append(cflags,
			"--sysroot="+path,
			"-isystem", filepath.Join(root, "lib", "mingw-w64", "mingw-w64-headers", "crt"),
			"-isystem", filepath.Join(root, "lib", "mingw-w64", "mingw-w64-headers", "defaults", "include"),
			"-D_UCRT",
		)
	case "":
		// No libc specified, nothing to add.
	default:
		// Incorrect configuration. This could be handled in a better way, but
		// usually this will be found by developers (not by TinyGo users).
		panic("unknown libc: " + c.Target.Libc)
	}
	// Always emit debug information. It is optionally stripped at link time.
	cflags = append(cflags, "-g")
	// Use the same optimization level as TinyGo.
	cflags = append(cflags, "-O"+c.Options.Opt)
	// Set the LLVM target triple.
	cflags = append(cflags, "--target="+c.Triple())
	// Set the -mcpu (or similar) flag.
	if c.Target.CPU != "" {
		if c.GOARCH() == "amd64" || c.GOARCH() == "386" {
			// x86 prefers the -march flag (-mcpu is deprecated there).
			cflags = append(cflags, "-march="+c.Target.CPU)
		} else if strings.HasPrefix(c.Triple(), "avr") {
			// AVR MCUs use -mmcu instead of -mcpu.
			cflags = append(cflags, "-mmcu="+c.Target.CPU)
		} else {
			// The rest just uses -mcpu.
			cflags = append(cflags, "-mcpu="+c.Target.CPU)
		}
	}
	return cflags
}

// LDFlags returns the flags to pass to the linker. A few more flags are needed
// (like the one for the compiler runtime), but this represents the majority of
// the flags.
func (c *Config) LDFlags() []string {
	root := goenv.Get("TINYGOROOT")
	// Merge and adjust LDFlags.
	var ldflags []string
	for _, flag := range c.Target.LDFlags {
		ldflags = append(ldflags, strings.ReplaceAll(flag, "{root}", root))
	}
	ldflags = append(ldflags, "-L", root)
	if c.Target.LinkerScript != "" {
		ldflags = append(ldflags, "-T", c.Target.LinkerScript)
	}
	return ldflags
}

// ExtraFiles returns the list of extra files to be built and linked with the
// executable. This can include extra C and assembly files.
func (c *Config) ExtraFiles() []string {
	return c.Target.ExtraFiles
}

// DumpSSA returns whether to dump Go SSA while compiling (-dumpssa flag). Only
// enable this for debugging.
func (c *Config) DumpSSA() bool {
	return c.Options.DumpSSA
}

// VerifyIR returns whether to run extra checks on the IR. This is normally
// disabled but enabled during testing.
func (c *Config) VerifyIR() bool {
	return c.Options.VerifyIR
}

// Debug returns whether debug (DWARF) information should be retained by the
// linker. By default, debug information is retained but it can be removed with
// the -no-debug flag.
func (c *Config) Debug() bool {
	return c.Options.Debug
}

// BinaryFormat returns an appropriate binary format, based on the file
// extension and the configured binary format in the target JSON file.
func (c *Config) BinaryFormat(ext string) string {
	switch ext {
	case ".bin", ".gba", ".nro":
		// The simplest format possible: dump everything in a raw binary file.
		if c.Target.BinaryFormat != "" {
			return c.Target.BinaryFormat
		}
		return "bin"
	case ".hex":
		// Similar to bin, but includes the start address and is thus usually a
		// better format.
		return "hex"
	case ".uf2":
		// Special purpose firmware format, mainly used on Adafruit boards.
		// More information:
		// https://github.com/Microsoft/uf2
		return "uf2"
	case ".zip":
		if c.Target.BinaryFormat != "" {
			return c.Target.BinaryFormat
		}
		return "zip"
	default:
		// Use the ELF format for unrecognized file formats.
		return "elf"
	}
}

// Programmer returns the flash method and OpenOCD interface name given a
// particular configuration. It may either be all configured in the target JSON
// file or be modified using the -programmmer command-line option.
func (c *Config) Programmer() (method, openocdInterface string) {
	switch c.Options.Programmer {
	case "":
		// No configuration supplied.
		return c.Target.FlashMethod, c.Target.OpenOCDInterface
	case "openocd", "msd", "command":
		// The -programmer flag only specifies the flash method.
		return c.Options.Programmer, c.Target.OpenOCDInterface
	case "bmp":
		// The -programmer flag only specifies the flash method.
		return c.Options.Programmer, ""
	default:
		// The -programmer flag specifies something else, assume it specifies
		// the OpenOCD interface name.
		return "openocd", c.Options.Programmer
	}
}

// OpenOCDConfiguration returns a list of command line arguments to OpenOCD.
// This list of command-line arguments is based on the various OpenOCD-related
// flags in the target specification.
func (c *Config) OpenOCDConfiguration() (args []string, err error) {
	_, openocdInterface := c.Programmer()
	if openocdInterface == "" {
		return nil, errors.New("OpenOCD programmer not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(openocdInterface) {
		return nil, fmt.Errorf("OpenOCD programmer has an invalid name: %#v", openocdInterface)
	}
	if c.Target.OpenOCDTarget == "" {
		return nil, errors.New("OpenOCD chip not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(c.Target.OpenOCDTarget) {
		return nil, fmt.Errorf("OpenOCD target has an invalid name: %#v", c.Target.OpenOCDTarget)
	}
	if c.Target.OpenOCDTransport != "" && c.Target.OpenOCDTransport != "swd" {
		return nil, fmt.Errorf("unknown OpenOCD transport: %#v", c.Target.OpenOCDTransport)
	}
	args = []string{"-f", "interface/" + openocdInterface + ".cfg"}
	for _, cmd := range c.Target.OpenOCDCommands {
		args = append(args, "-c", cmd)
	}
	if c.Target.OpenOCDTransport != "" {
		args = append(args, "-c", "transport select "+c.Target.OpenOCDTransport)
	}
	args = append(args, "-f", "target/"+c.Target.OpenOCDTarget+".cfg")
	return args, nil
}

// CodeModel returns the code model used on this platform.
func (c *Config) CodeModel() string {
	if c.Target.CodeModel != "" {
		return c.Target.CodeModel
	}

	return "default"
}

// RelocationModel returns the relocation model in use on this platform. Valid
// values are "static", "pic", "dynamicnopic".
func (c *Config) RelocationModel() string {
	if c.Target.RelocationModel != "" {
		return c.Target.RelocationModel
	}

	return "static"
}

// WasmAbi returns the WASM ABI which is specified in the target JSON file, and
// the value is overridden by `-wasm-abi` flag if it is provided
func (c *Config) WasmAbi() string {
	if c.Options.WasmAbi != "" {
		return c.Options.WasmAbi
	}
	return c.Target.WasmAbi
}

// Emulator returns the emulator target config
func (c *Config) Emulator() []string {
	var emulator []string
	for _, s := range c.Target.Emulator {
		emulator = append(emulator, strings.ReplaceAll(s, "{root}", goenv.Get("TINYGOROOT")))
	}
	return emulator
}

type TestConfig struct {
	CompileTestBinary bool
	// TODO: Filter the test functions to run, include verbose flag, etc
}
