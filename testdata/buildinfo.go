package main

// Tests runtime/debug.ReadBuildInfo and the parser behind it.
//
// The output has to be the same on every machine, so nothing here prints a
// version, a module path or a checksum that the build happens to have. What is
// printed is either derived (does the version exist, does it name the
// compiler) or comes from the fixed string parsed below.

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// A module string in the format BuildInfo.String writes, minus the leading go
// line — which is what ParseBuildInfo takes. Covers the four line kinds: path,
// mod, dep and build, plus a replaced dependency.
const sample = "path\texample.com/prog\n" +
	"mod\texample.com/prog\t(devel)\t\n" +
	"dep\texample.com/a\tv1.2.3\th1:aaa=\n" +
	"dep\texample.com/b\tv0.1.0\t\n" +
	"=>\texample.com/b-fork\tv0.2.0\th1:bbb=\n" +
	"build\t-tags=sample\n" +
	"build\tCGO_ENABLED=0\n"

func main() {
	readBuildInfo()
	parse()
	roundTrip()
	malformed()
}

func readBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	// ReadBuildInfo always succeeds: with no module information embedded it
	// still reports the toolchain version, so callers that only want
	// GoVersion keep working.
	println("read ok:", ok)
	println("info not nil:", info != nil)
	if info == nil {
		return
	}
	println("go version set:", info.GoVersion != "")
	println("names the compiler:", strings.HasPrefix(info.GoVersion, runtime.Compiler))
}

func parse() {
	bi, err := debug.ParseBuildInfo(sample)
	if err != nil {
		println("parse error:", err.Error())
		return
	}
	println("path:", bi.Path)
	println("main path:", bi.Main.Path)
	println("main version:", bi.Main.Version)
	println("deps:", len(bi.Deps))
	for _, d := range bi.Deps {
		println("dep:", d.Path, d.Version, d.Sum)
		if d.Replace != nil {
			println("  replaced by:", d.Replace.Path, d.Replace.Version, d.Replace.Sum)
		}
	}
	println("settings:", len(bi.Settings))
	for _, s := range bi.Settings {
		println("setting:", s.Key, s.Value)
	}
}

func roundTrip() {
	bi, err := debug.ParseBuildInfo(sample)
	if err != nil {
		println("round trip parse error:", err.Error())
		return
	}
	// String writes a leading go line that ParseBuildInfo does not read, so it
	// is dropped before parsing again. Everything else must survive.
	out := bi.String()
	if i := strings.Index(out, "\n"); i >= 0 && strings.HasPrefix(out, "go\t") {
		out = out[i+1:]
	}
	again, err := debug.ParseBuildInfo(out)
	if err != nil {
		println("round trip reparse error:", err.Error())
		return
	}
	println("round trip path:", again.Path == bi.Path)
	println("round trip main:", again.Main == bi.Main)
	println("round trip deps:", len(again.Deps) == len(bi.Deps))
	println("round trip settings:", len(again.Settings) == len(bi.Settings))
}

func malformed() {
	// A line whose prefix is not one of the known kinds is skipped rather than
	// rejected, which is what upstream does and what keeps an older parser
	// reading a newer module string.
	for _, in := range []string{
		"path\n",              // "path" without the tab is not the path line
		"build\n",             // likewise
		"future\tsomething\n", // a line kind this version does not know
	} {
		_, err := debug.ParseBuildInfo(in)
		println("unknown line skipped:", err == nil)
	}

	// A line that is one of the known kinds but the wrong shape is an error,
	// because that is a module string this parser has misread rather than one
	// it does not recognize. ReadBuildInfo falls back to the toolchain version
	// when that happens.
	for _, in := range []string{
		"mod\texample.com/prog\n",            // a module needs 2 or 3 columns
		"dep\texample.com/a\n",               // likewise
		"=>\texample.com/x\tv1.0.0\th1:x=\n", // a replacement with nothing to replace
		"mod\ta\tb\tc\td\n",                  // too many columns
	} {
		_, err := debug.ParseBuildInfo(in)
		println("malformed rejected:", err != nil)
	}
}
