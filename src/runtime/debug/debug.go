// Package debug is a dummy package that is not yet implemented.
package debug

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// SetMaxStack sets the maximum amount of memory that can be used by a single
// goroutine stack.
//
// Not implemented.
func SetMaxStack(n int) int {
	return n
}

// PrintStack prints to standard error the stack trace returned by runtime.Stack.
//
// Not implemented.
func PrintStack() {}

// Stack returns a formatted stack trace of the goroutine that calls it.
//
// Not implemented.
func Stack() []byte {
	return nil
}

// ReadBuildInfo returns the build information embedded
// in the running binary. The information is available only
// in binaries built with module support.
//
// Not implemented.
func ReadBuildInfo() (info *BuildInfo, ok bool) {
	return &BuildInfo{GoVersion: runtime.Compiler + runtime.Version()}, true
}

// BuildInfo represents the build information read from
// the running binary.
type BuildInfo struct {
	GoVersion string    // version of the Go toolchain that built the binary, e.g. "go1.19.2"
	Path      string    // The main package path
	Main      Module    // The module containing the main package
	Deps      []*Module // Module dependencies
	Settings  []BuildSetting
}

type BuildSetting struct {
	// Key and Value describe the build setting.
	// Key must not contain an equals sign, space, tab, or newline.
	// Value must not contain newlines ('\n').
	Key, Value string
}

// Module represents a module.
type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

// Not implemented.
func SetGCPercent(n int) int {
	return n
}

// Start of stolen from big go. TODO: import/reuse without copy pasta.

// quoteKey reports whether key is required to be quoted.
func quoteKey(key string) bool {
	return len(key) == 0 || strings.ContainsAny(key, "= \t\r\n\"`")
}

// quoteValue reports whether value is required to be quoted.
func quoteValue(value string) bool {
	return strings.ContainsAny(value, " \t\r\n\"`")
}

func (bi *BuildInfo) String() string {
	buf := new(strings.Builder)
	if bi.GoVersion != "" {
		fmt.Fprintf(buf, "go\t%s\n", bi.GoVersion)
	}
	if bi.Path != "" {
		fmt.Fprintf(buf, "path\t%s\n", bi.Path)
	}
	var formatMod func(string, Module)
	formatMod = func(word string, m Module) {
		buf.WriteString(word)
		buf.WriteByte('\t')
		buf.WriteString(m.Path)
		buf.WriteByte('\t')
		buf.WriteString(m.Version)
		if m.Replace == nil {
			buf.WriteByte('\t')
			buf.WriteString(m.Sum)
		} else {
			buf.WriteByte('\n')
			formatMod("=>", *m.Replace)
		}
		buf.WriteByte('\n')
	}
	if bi.Main != (Module{}) {
		formatMod("mod", bi.Main)
	}
	for _, dep := range bi.Deps {
		formatMod("dep", *dep)
	}
	for _, s := range bi.Settings {
		key := s.Key
		if quoteKey(key) {
			key = strconv.Quote(key)
		}
		value := s.Value
		if quoteValue(value) {
			value = strconv.Quote(value)
		}
		fmt.Fprintf(buf, "build\t%s=%s\n", key, value)
	}

	return buf.String()
}
