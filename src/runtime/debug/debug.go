// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package debug is a very partially implemented package to allow compilation.
package debug

import (
	"errors"
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

// modinfo holds the serialized module build information for the running binary,
// in the same textual format produced by BuildInfo.String() (minus the leading
// "go\t..." line). It is empty unless the TinyGo builder embedded it (see
// builder.Build, which fills runtime/debug.modinfo from `go list -json`), or it
// was set explicitly via -ldflags="-X runtime/debug.modinfo=...".
//
// TinyGo controls both the writer and this reader, so the value is stored bare:
// it carries none of the delimiters cmd/go wraps around runtime.modinfo.
//
// What TinyGo writes is deliberately a subset. There are no "=>" replacement
// lines and no checksums, and none of the build settings the go toolchain
// records (-tags, GOOS, GOARCH, CGO_ENABLED and so on) beyond the vcs.* ones.
var modinfo string

// ReadBuildInfo returns the build information embedded
// in the running binary. The information is available only
// in binaries built with module support.
//
// TinyGo populates GoVersion always, and Path/Main/Deps/Settings when the
// builder (or -ldflags -X) embedded module info; see the modinfo var.
func ReadBuildInfo() (info *BuildInfo, ok bool) {
	goVersion := runtime.Compiler + runtime.Version()
	data := modinfo
	if data == "" {
		// No module info embedded; still report the toolchain version so
		// callers that only want GoVersion keep working.
		return &BuildInfo{GoVersion: goVersion}, true
	}
	bi, err := ParseBuildInfo(data)
	if err != nil {
		return &BuildInfo{GoVersion: goVersion}, true
	}
	// GoVersion is stored separately from the module string (as in upstream Go).
	bi.GoVersion = goVersion
	return bi, true
}

// parseError reports which line of the module string failed to parse.
//
// Upstream spells this fmt.Errorf("...: line %d: %w", ...). Doing the same
// here would link fmt into every binary that calls ReadBuildInfo, which on a
// small target costs more flash than the feature is worth, so the wrapping is
// written out by hand. The message is identical and Unwrap keeps errors.Is
// and errors.As working as they do upstream.
type parseError struct {
	line int
	err  error
}

func (e *parseError) Error() string {
	return "could not parse Go build info: line " + strconv.Itoa(e.line) + ": " + e.err.Error()
}

func (e *parseError) Unwrap() error { return e.err }

// ParseBuildInfo parses the string returned by BuildInfo.String (excluding the
// leading "go" line) back into a BuildInfo. It is the reverse of that method
// and is ported from the standard library's runtime/debug.
func ParseBuildInfo(data string) (bi *BuildInfo, err error) {
	lineNum := 1
	defer func() {
		if err != nil {
			err = &parseError{line: lineNum, err: err}
		}
	}()

	const (
		pathLine  = "path\t"
		modLine   = "mod\t"
		depLine   = "dep\t"
		repLine   = "=>\t"
		buildLine = "build\t"
		newline   = "\n"
		tab       = "\t"
	)

	readModuleLine := func(elem []string) (Module, error) {
		if len(elem) != 2 && len(elem) != 3 {
			return Module{}, errors.New("expected 2 or 3 columns; got " + strconv.Itoa(len(elem)))
		}
		version := elem[1]
		sum := ""
		if len(elem) == 3 {
			sum = elem[2]
		}
		return Module{
			Path:    elem[0],
			Version: version,
			Sum:     sum,
		}, nil
	}

	bi = new(BuildInfo)
	var (
		last *Module
		line string
		ok   bool
	)
	// Reverse of BuildInfo.String(), except for go version.
	for len(data) > 0 {
		line, data, ok = strings.Cut(data, newline)
		if !ok {
			break
		}
		switch {
		case strings.HasPrefix(line, pathLine):
			elem := line[len(pathLine):]
			bi.Path = elem
		case strings.HasPrefix(line, modLine):
			elem := strings.Split(line[len(modLine):], tab)
			last = &bi.Main
			*last, err = readModuleLine(elem)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, depLine):
			elem := strings.Split(line[len(depLine):], tab)
			last = new(Module)
			bi.Deps = append(bi.Deps, last)
			*last, err = readModuleLine(elem)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, repLine):
			elem := strings.Split(line[len(repLine):], tab)
			if len(elem) != 3 {
				return nil, errors.New("expected 3 columns for replacement; got " + strconv.Itoa(len(elem)))
			}
			if last == nil {
				return nil, errors.New("replacement with no module on previous line")
			}
			last.Replace = &Module{
				Path:    elem[0],
				Version: elem[1],
				Sum:     elem[2],
			}
			last = nil
		case strings.HasPrefix(line, buildLine):
			kv := line[len(buildLine):]
			if len(kv) < 1 {
				return nil, errors.New("build line missing '='")
			}

			var key, rawValue string
			switch kv[0] {
			case '=':
				return nil, errors.New("build line with missing key")

			case '`', '"':
				rawKey, err := strconv.QuotedPrefix(kv)
				if err != nil {
					return nil, errors.New("invalid quoted key in build line")
				}
				if len(kv) == len(rawKey) {
					return nil, errors.New("build line missing '=' after quoted key")
				}
				if c := kv[len(rawKey)]; c != '=' {
					// %q on a byte formats a single-quoted rune, not a string.
					return nil, errors.New("unexpected character after quoted key: " + strconv.QuoteRune(rune(c)))
				}
				key, _ = strconv.Unquote(rawKey)
				rawValue = kv[len(rawKey)+1:]

			default:
				var ok bool
				key, rawValue, ok = strings.Cut(kv, "=")
				if !ok {
					return nil, errors.New("build line missing '=' after key")
				}
				if quoteKey(key) {
					return nil, errors.New("unquoted key " + strconv.Quote(key) + " must be quoted")
				}
			}

			var value string
			if len(rawValue) > 0 {
				switch rawValue[0] {
				case '`', '"':
					var err error
					value, err = strconv.Unquote(rawValue)
					if err != nil {
						return nil, errors.New("invalid quoted value in build line")
					}

				default:
					value = rawValue
					if quoteValue(value) {
						return nil, errors.New("unquoted value " + strconv.Quote(value) + " must be quoted")
					}
				}
			}

			bi.Settings = append(bi.Settings, BuildSetting{Key: key, Value: value})
		}
		lineNum++
	}
	return bi, nil
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
		buf.WriteString("go\t")
		buf.WriteString(bi.GoVersion)
		buf.WriteByte('\n')
	}
	if bi.Path != "" {
		buf.WriteString("path\t")
		buf.WriteString(bi.Path)
		buf.WriteByte('\n')
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
		buf.WriteString("build\t")
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(value)
		buf.WriteByte('\n')
	}

	return buf.String()
}
