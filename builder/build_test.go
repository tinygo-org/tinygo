package builder

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("LocalAppData", cache)
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	if err := os.WriteFile(source, []byte(`package main
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
`), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(&compileopts.Options{
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		Opt: "z", TrimPath: true, Semaphore: make(chan struct{}, 1),
		InterpTimeout: time.Minute, InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
	})
	if err != nil {
		t.Fatal(err)
	}
	compilations := 0
	config.Options.PrintCommands = func(command string, args ...string) {
		if command == "clang" && slices.Contains(args, "-emit-llvm") && !slices.Contains(args, "-M") {
			compilations++
		}
	}
	previous := 0
	for i, value := range []string{"1", "2", "2"} {
		if err := os.WriteFile(filepath.Join(dir, "hidden.h"), []byte("#define HIDDEN "+value+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		result, err := Build(source, filepath.Join(dir, "probe"), t.TempDir(), config)
		if err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(result.Executable).CombinedOutput()
		if err != nil {
			t.Fatalf("running CGo program: %v\n%s", err, output)
		}
		if got := string(bytes.TrimSpace(output)); got != value {
			t.Fatalf("CGo result = %q, want %q", got, value)
		}
		if i == 1 && compilations == previous {
			t.Fatal("editing a compiler-only header reused cached bitcode")
		}
		if i == 2 && compilations != previous {
			t.Fatal("unchanged CGo inputs were recompiled")
		}
		previous = compilations
	}
}
