package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/interp"
	"tinygo.org/x/go-llvm"
)

func TestCGoHeaderCompileIsPathIndependent(t *testing.T) {
	var outputs [][]byte
	for _, dirName := range []string{"first", "second"} {
		dir := filepath.Join(t.TempDir(), dirName)
		if err := os.Mkdir(dir, 0o777); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(dir, "snippet.c")
		output := filepath.Join(dir, "snippet.bc")
		if err := os.WriteFile(source, []byte("const char *sourceName = __FILE__;\n"), 0o666); err != nil {
			t.Fatal(err)
		}
		flags := cgoHeaderCompileArgs(source, output, []string{
			"-gdwarf-4",
			"--target=x86_64-unknown-linux-gnu",
		})
		if err := runCCompiler(flags...); err != nil {
			t.Fatal(err)
		}

		ctx := llvm.NewContext()
		mod := ctx.NewModule("package")
		headerMod, err := ctx.ParseBitcodeFile(output)
		if err != nil {
			t.Fatal(err)
		}
		if err := llvm.LinkModules(mod, headerMod); err != nil {
			t.Fatal(err)
		}
		buf := llvm.WriteBitcodeToMemoryBuffer(mod)
		outputs = append(outputs, bytes.Clone(buf.Bytes()))
		buf.Dispose()
		mod.Dispose()
		ctx.Dispose()
	}
	if !bytes.Equal(outputs[0], outputs[1]) {
		t.Fatal("CGo header bitcode depends on its temporary source path")
	}
}

func TestCGoHeaderCacheDependencies(t *testing.T) {
	dir, config := cgoCacheTestConfig(t)
	source := filepath.Join(dir, "main.go")
	writeCGoCacheFile(t, source, `package main
/*
#ifndef _FORTIFY_SOURCE
#include "hidden.h"
#endif
static int value(void) {
#ifdef _FORTIFY_SOURCE
	return 0;
#else
	return HIDDEN;
#endif
}
*/
import "C"
func main() { println(C.value()) }
`)
	compilations := 0
	config.Options.PrintCommands = func(command string, args ...string) {
		if command == "clang" && slices.Contains(args, "-emit-llvm") && !slices.Contains(args, "-M") {
			compilations++
		}
	}
	previous := 0
	for i, value := range []string{"1", "2", "2"} {
		writeCGoCacheFile(t, filepath.Join(dir, "hidden.h"), "#define HIDDEN "+value+"\n")
		checkCGoCacheBuild(t, dir, source, value, config)
		if i == 1 && compilations == previous {
			t.Fatal("editing a compiler-only header reused cached bitcode")
		}
		if i == 2 && compilations != previous {
			t.Fatal("unchanged CGo inputs were recompiled")
		}
		previous = compilations
	}
}

func TestCGoHeaderCacheParseAndCompile(t *testing.T) {
	dir, config := cgoCacheTestConfig(t)
	source := filepath.Join(dir, "main.go")
	header := filepath.Join(dir, "value.h")
	writeCGoCacheFile(t, source, fmt.Sprintf(`package main
/*
#include "%s"
static int value(void) { return NUMBER; }
*/
import "C"
func main() { println(C.NUMBER, C.value()) }
`, filepath.ToSlash(header)))
	writeCGoCacheFile(t, header, "#define NUMBER 1\n")
	changed := false
	config.Options.PrintCommands = func(command string, args ...string) {
		if !changed && command == "clang" && slices.Contains(args, "-M") && slices.Contains(args, "-emit-llvm") {
			changed = true
			if err := os.WriteFile(header, []byte("#define NUMBER 2\n"), 0o600); err != nil {
				t.Error(err)
			}
		}
	}
	for _, want := range []string{"1 2", "2 2"} {
		checkCGoCacheBuild(t, dir, source, want, config)
	}
}

func TestCGoHeaderCacheParserOnlyLookup(t *testing.T) {
	dir, config := cgoCacheTestConfig(t)
	source := filepath.Join(dir, "main.go")
	writeCGoCacheFile(t, source, `package main
	/*
	#ifdef _FORTIFY_SOURCE
	#if __has_include("optional.h")
	#define NUMBER 2
	#endif
	#endif
	#ifndef NUMBER
	#define NUMBER 1
	#endif
	*/
	import "C"
	func main() { println(C.NUMBER) }
	`)
	for i, want := range []string{"1", "2", "1"} {
		header := filepath.Join(dir, "optional.h")
		if i == 1 {
			writeCGoCacheFile(t, header, "/* optional */\n")
		} else if i == 2 {
			if err := os.Remove(header); err != nil {
				t.Fatal(err)
			}
		}
		checkCGoCacheBuild(t, dir, source, want, config)
	}
}

func TestCGoHeaderCacheRejectsChangedCompileInputs(t *testing.T) {
	dir, config := cgoCacheTestConfig(t)
	source := filepath.Join(dir, "main.go")
	header := filepath.Join(dir, "hidden.h")
	writeCGoCacheFile(t, source, `package main
	/*
	#ifdef _FORTIFY_SOURCE
	static int value(void) { return 0; }
	#else
	#include "hidden.h"
	static int value(void) { return NUMBER; }
	#endif
	*/
	import "C"
	func main() { println(C.value()) }
	`)
	writeCGoCacheFile(t, header, "#define NUMBER 1\n")
	changed := false
	config.Options.PrintCommands = func(command string, args ...string) {
		if command == "clang" && slices.Contains(args, "-emit-llvm") && !slices.Contains(args, "-M") && !changed {
			changed = true
			if err := os.WriteFile(header, []byte("#define NUMBER 2\n"), 0o600); err != nil {
				t.Error(err)
			}
		}
	}
	_, err := Build(source, filepath.Join(dir, "probe"), t.TempDir(), config)
	if !changed || err == nil || !strings.Contains(err.Error(), "CGo inputs changed during compilation") {
		t.Fatalf("build with changed inputs: err = %v, edit performed = %v", err, changed)
	}
	writeCGoCacheFile(t, header, "#define NUMBER 1\n")
	checkCGoCacheBuild(t, dir, source, "1", config)
}

func cgoCacheTestConfig(t *testing.T) (string, *compileopts.Config) {
	t.Helper()
	cache := t.TempDir()
	for _, name := range []string{"HOME", "XDG_CACHE_HOME", "LocalAppData"} {
		t.Setenv(name, cache)
	}
	config, err := NewConfig(&compileopts.Options{
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		Opt: "z", TrimPath: true, Semaphore: make(chan struct{}, 1),
		InterpTimeout: time.Minute, InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
	})
	if err != nil {
		t.Fatal(err)
	}
	return t.TempDir(), config
}

func writeCGoCacheFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func checkCGoCacheBuild(t *testing.T, dir, source, want string, config *compileopts.Config) {
	t.Helper()
	result, err := Build(source, filepath.Join(dir, "probe"), t.TempDir(), config)
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(result.Executable).CombinedOutput()
	if got := string(bytes.TrimSpace(output)); err != nil || got != want {
		t.Fatalf("CGo result = %q, err = %v, want %q", got, err, want)
	}
}
