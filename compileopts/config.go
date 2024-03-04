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

	"github.com/google/shlex"
	"github.com/tinygo-org/tinygo/goenv"
)

// Config keeps all configuration affecting the build in a single struct.
type Config struct {
	Options        *Options
	Target         *TargetSpec
	GoMinorVersion int
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

// ABI returns the -mabi= flag for this target (like -mabi=lp64). A zero-length
// string is returned if the target doesn't specify an ABI.
func (c *Config) ABI() string {
	return c.Target.ABI
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
	tags := append(c.Target.BuildTags, []string{
		"tinygo",                                     // that's the compiler
		"purego",                                     // to get various crypto packages to work
		"math_big_pure_go",                           // to get math/big to work
		"gc." + c.GC(), "scheduler." + c.Scheduler(), // used inside the runtime package
		"serial." + c.Serial()}...) // used inside the machine package
	for i := 1; i <= c.GoMinorVersion; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	tags = append(tags, c.Options.Tags...)
	return tags
}

// GC returns the garbage collection strategy in use on this platform. Valid
// values are "none", "leaking", "conservative" and "precise".
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
	case "conservative", "custom", "precise":
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
// "asyncify" and "tasks".
func (c *Config) Scheduler() string {
	if c.Options.Scheduler != "" {
		return c.Options.Scheduler
	}
	if c.Target.Scheduler != "" {
		return c.Target.Scheduler
	}
	// Fall back to none.
	return "none"
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
func (c *Config) OptLevel() (level string, speedLevel, sizeLevel int) {
	switch c.Options.Opt {
	case "none", "0":
		return "O0", 0, 0
	case "1":
		return "O1", 1, 0
	case "2":
		return "O2", 2, 0
	case "s":
		return "Os", 2, 1
	case "z":
		return "Oz", 2, 2 // default
	default:
		// This is not shown to the user: valid choices are already checked as
		// part of Options.Verify(). It is here as a sanity check.
		panic("unknown optimization level: -opt=" + c.Options.Opt)
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

// StackSize returns the default stack size to be used for goroutines, if the
// stack size could not be determined automatically at compile time.
func (c *Config) StackSize() uint64 {
	if c.Options.StackSize != 0 {
		return c.Options.StackSize
	}
	return c.Target.DefaultStackSize
}

// MaxStackAlloc returns the size of the maximum allocation to put on the stack vs heap.
func (c *Config) MaxStackAlloc() uint64 {
	if c.StackSize() > 32*1024 {
		return 1024
	}

	return 256
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
	archname := c.Triple()
	if c.CPU() != "" {
		archname += "-" + c.CPU()
	}
	if c.ABI() != "" {
		archname += "-" + c.ABI()
	}

	// Try to load a precompiled library.
	precompiledDir := filepath.Join(goenv.Get("TINYGOROOT"), "pkg", archname, name)
	if _, err := os.Stat(precompiledDir); err == nil {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return precompiledDir, true
	}

	// No precompiled library found. Determine the path name that will be used
	// in the build cache.
	return filepath.Join(goenv.Get("GOCACHE"), name+"-"+archname), false
}

// DefaultBinaryExtension returns the default extension for binaries, such as
// .exe, .wasm, or no extension (depending on the target).
func (c *Config) DefaultBinaryExtension() string {
	parts := strings.Split(c.Triple(), "-")
	if parts[0] == "wasm32" {
		// WebAssembly files always have the .wasm file extension.
		return ".wasm"
	}
	if len(parts) >= 3 && parts[2] == "windows" {
		// Windows uses .exe.
		return ".exe"
	}
	if len(parts) >= 3 && parts[2] == "unknown" {
		// There appears to be a convention to use the .elf file extension for
		// ELF files intended for microcontrollers. I'm not aware of the origin
		// of this, it's just something that is used by many projects.
		// I think it's a good tradition, so let's keep it.
		return ".elf"
	}
	// Linux, MacOS, etc, don't use a file extension. Use it as a fallback.
	return ""
}

// CFlags returns the flags to pass to the C compiler. This is necessary for CGo
// preprocessing.
func (c *Config) CFlags(libclang bool) []string {
	var cflags []string
	for _, flag := range c.Target.CFlags {
		cflags = append(cflags, strings.ReplaceAll(flag, "{root}", goenv.Get("TINYGOROOT")))
	}
	resourceDir := goenv.ClangResourceDir(libclang)
	if resourceDir != "" {
		// The resource directory contains the built-in clang headers like
		// stdbool.h, stdint.h, float.h, etc.
		// It is left empty if we're using an external compiler (that already
		// knows these headers).
		cflags = append(cflags,
			"-resource-dir="+resourceDir,
		)
	}
	switch c.Target.Libc {
	case "darwin-libSystem":
		root := goenv.Get("TINYGOROOT")
		cflags = append(cflags,
			"-nostdlibinc",
			"-isystem", filepath.Join(root, "lib/macos-minimal-sdk/src/usr/include"),
		)
	case "picolibc":
		root := goenv.Get("TINYGOROOT")
		picolibcDir := filepath.Join(root, "lib", "picolibc", "newlib", "libc")
		path, _ := c.LibcPath("picolibc")
		cflags = append(cflags,
			"-nostdlibinc",
			"-isystem", filepath.Join(path, "include"),
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
	case "wasmbuiltins":
		// nothing to add (library is purely for builtins)
	case "mingw-w64":
		root := goenv.Get("TINYGOROOT")
		path, _ := c.LibcPath("mingw-w64")
		cflags = append(cflags,
			"-nostdlibinc",
			"-isystem", filepath.Join(path, "include"),
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
	cflags = append(cflags, "-gdwarf-4")
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
	// Set the -mabi flag, if needed.
	if c.ABI() != "" {
		cflags = append(cflags, "-mabi="+c.ABI())
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
// linker. By default, debug information is retained, but it can be removed
// with the -no-debug flag.
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
	case ".img":
		// Image file. Only defined for the ESP32 at the moment, where it is a
		// full (runnable) image that can be used in the Espressif QEMU fork.
		if c.Target.BinaryFormat != "" {
			return c.Target.BinaryFormat + "-img"
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
	if !regexp.MustCompile(`^[\p{L}0-9_-]+$`).MatchString(openocdInterface) {
		return nil, fmt.Errorf("OpenOCD programmer has an invalid name: %#v", openocdInterface)
	}
	if c.Target.OpenOCDTarget == "" {
		return nil, errors.New("OpenOCD chip not set")
	}
	if !regexp.MustCompile(`^[\p{L}0-9_-]+$`).MatchString(c.Target.OpenOCDTarget) {
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
		transport := c.Target.OpenOCDTransport
		if transport == "swd" {
			switch openocdInterface {
			case "stlink-dap":
				transport = "dapdirect_swd"
			}
		}
		args = append(args, "-c", "transport select "+transport)
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

// EmulatorName is a shorthand to get the command for this emulator, something
// like qemu-system-arm or simavr.
func (c *Config) EmulatorName() string {
	parts := strings.SplitN(c.Target.Emulator, " ", 2)
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

// EmulatorFormat returns the binary format for the emulator and the associated
// file extension. An empty string means to pass directly whatever the linker
// produces directly without conversion (usually ELF format).
func (c *Config) EmulatorFormat() (format, fileExt string) {
	switch {
	case strings.Contains(c.Target.Emulator, "{img}"):
		return "img", ".img"
	default:
		return "", ""
	}
}

// Emulator returns a ready-to-run command to run the given binary in an
// emulator. Give it the format (returned by EmulatorFormat()) and the path to
// the compiled binary.
func (c *Config) Emulator(format, binary string) ([]string, error) {
	parts, err := shlex.Split(c.Target.Emulator)
	if err != nil {
		return nil, fmt.Errorf("could not parse emulator command: %w", err)
	}
	var emulator []string
	for _, s := range parts {
		s = strings.ReplaceAll(s, "{root}", goenv.Get("TINYGOROOT"))
		// Allow replacement of what's usually /tmp except notably Windows.
		s = strings.ReplaceAll(s, "{tmpDir}", os.TempDir())
		s = strings.ReplaceAll(s, "{"+format+"}", binary)
		emulator = append(emulator, s)
	}
	return emulator, nil
}

type TestConfig struct {
	CompileTestBinary bool
	CompileOnly       bool
	Verbose           bool
	Short             bool
	RunRegexp         string
	SkipRegexp        string
	Count             *int
	BenchRegexp       string
	BenchTime         string
	BenchMem          bool
	Shuffle           string
}
