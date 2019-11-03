// Package compileopts contains the configuration for a single to-be-built
// binary.
package compileopts

import (
	"fmt"
	"strconv"
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

// Triple returns the LLVM target triple, like armv6m-none-eabi.
func (c *Config) Triple() string {
	return c.Target.Triple
}

// CPU returns the LLVM CPU name, like atmega328p or arm7tdmi. It may return an
// empty string if the CPU name is not known.
func (c *Config) CPU() string {
	return c.Target.CPU
}

// Features returns a list of features this CPU supports. For example, for a
// RISC-V processor, that could be ["+a", "+c", "+m"]. For many targets, an
// empty list will be returned.
func (c *Config) Features() []string {
	return c.Target.Features
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

// BuildTags returns the complete list of build tags used during this build.
func (c *Config) BuildTags() []string {
	tags := append(c.Target.BuildTags, []string{"tinygo", "gc." + c.GC(), "scheduler." + c.Scheduler()}...)
	for i := 1; i <= c.GoMinorVersion; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	if extraTags := strings.Fields(c.Options.Tags); len(extraTags) != 0 {
		tags = append(tags, extraTags...)
	}
	return tags
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

// Scheduler returns the scheduler implementation. Valid values are "coroutines"
// and "tasks".
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

// PanicStrategy returns the panic strategy selected for this target. Valid
// values are "print" (print the panic value, then exit) or "trap" (issue a trap
// instruction).
func (c *Config) PanicStrategy() string {
	return c.Options.PanicStrategy
}

// CFlags returns the flags to pass to the C compiler. This is necessary for CGo
// preprocessing.
func (c *Config) CFlags() []string {
	cflags := append([]string{}, c.Options.CFlags...)
	for _, flag := range c.Target.CFlags {
		cflags = append(cflags, strings.Replace(flag, "{root}", goenv.Get("TINYGOROOT"), -1))
	}
	return cflags
}

// LDFlags returns the flags to pass to the linker. A few more flags are needed
// (like the one for the compiler runtime), but this represents the majority of
// the flags.
func (c *Config) LDFlags() []string {
	root := goenv.Get("TINYGOROOT")
	// Merge and adjust LDFlags.
	ldflags := append([]string{}, c.Options.LDFlags...)
	for _, flag := range c.Target.LDFlags {
		ldflags = append(ldflags, strings.Replace(flag, "{root}", root, -1))
	}
	ldflags = append(ldflags, "-L", root)
	if c.Target.GOARCH == "wasm" {
		// Round heap size to next multiple of 65536 (the WebAssembly page
		// size).
		heapSize := (c.Options.HeapSize + (65536 - 1)) &^ (65536 - 1)
		ldflags = append(ldflags, "--initial-memory="+strconv.FormatInt(heapSize, 10))
	}
	return ldflags
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

// Debug returns whether to add debug symbols to the IR, for debugging with GDB
// and similar.
func (c *Config) Debug() bool {
	return c.Options.Debug
}

type TestConfig struct {
	CompileTestBinary bool
	// TODO: Filter the test functions to run, include verbose flag, etc
}
