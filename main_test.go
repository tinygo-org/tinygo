package main

// This file tests the compiler by running Go files in testdata/*.go and
// comparing their output with the expected output in testdata/*.txt.

import (
	"bufio"
	"bytes"
	"debug/dwarf"
	"debug/elf"
	"debug/pe"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aykevl/go-wasm"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/diagnostics"
	"github.com/tinygo-org/tinygo/goenv"
)

const TESTDATA = "testdata"

var testTarget = flag.String("target", "", "override test target")

var supportedLinuxArches = map[string]string{
	"AMD64Linux": "linux/amd64",
	"X86Linux":   "linux/386",
	"ARMLinux":   "linux/arm/6",
	"ARM64Linux": "linux/arm64",
	"MIPSLinux":  "linux/mips/hardfloat",
	"WASIp1":     "wasip1/wasm",
}

var sema = make(chan struct{}, runtime.NumCPU())

func TestBoehmFinalizerInterior(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		target, scheduler string
	}{
		{"wasm", "asyncify"},
		{"wasm", "none"},
		{"", "tasks"},
		{"", "threads"},
		{"", "none"},
	} {
		if test.target == "" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			continue
		}
		t.Run(test.target+"/"+test.scheduler, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget(test.target, sema)
			options.GC = "boehm"
			options.Scheduler = test.scheduler
			options.Tags = append(options.Tags, "runtime_asserts")
			options.GlobalValues = map[string]map[string]string{
				"main": {"finalizerTestMode": "interior"},
			}
			runTest("finalizer.go", options, t, nil, nil)
		})
	}
}

func TestBoehmFinalizerInteriorInvalid(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("host finalizers need Linux or Darwin")
	}
	options := optionsFromTarget("", sema)
	options.GC = "boehm"
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"interior-pointer", "interior-large"} {
		t.Run(name, func(t *testing.T) {
			output := &bytes.Buffer{}
			_, err := buildAndRun("testdata/finalizer.go", config, output, []string{name}, nil, time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
				cmd.Stdout = nil
				cmd.Stderr = nil
				data, err := cmd.CombinedOutput()
				output.Write(data)
				return err
			})
			if err == nil || !strings.Contains(output.String(), "pointer not at beginning of allocated block") {
				t.Fatalf("interior registration result: %v\n%s", err, output.String())
			}
		})
	}
}

func TestFinalizerRunnerRegistration(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		target, gc, scheduler string
	}{
		{"wasm", "boehm", "asyncify"},
		{"wasm", "conservative", "asyncify"},
		{"", "boehm", "tasks"},
		{"", "boehm", "threads"},
	} {
		if test.target == "" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			continue
		}
		t.Run(test.target+"/"+test.gc+"/"+test.scheduler, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget(test.target, sema)
			options.GC = test.gc
			options.Scheduler = test.scheduler
			options.GlobalValues = map[string]map[string]string{
				"main": {"finalizerIdleMode": "registration"},
			}
			runTest("finalizeridle.go", options, t, nil, nil)
		})
	}
}

func TestParseGoLinkFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want []string
		err  string
	}{
		{name: "empty"},
		{name: "windows-path", in: `-extldflags=-LC:\x\lib`, want: []string{`-LC:\x\lib`}},
		{
			name: "quoted-windows-paths",
			in:   `-extldflags '-LC:\x\lib "-LC:\Program Files\x"'`,
			want: []string{`-LC:\x\lib`, `-LC:\Program Files\x`},
		},
		{
			name: "literal-quotes-and-backslashes",
			in:   `-extldflags '-DNAME="value" a\ b ""'`,
			want: []string{`-DNAME="value"`, `a\`, "b", ""},
		},
		{
			name: "unterminated-ldflags",
			in:   `-extldflags 'abc`,
			err:  "unterminated ' string",
		},
		{
			name: "unterminated-extldflags",
			in:   `-extldflags "'abc"`,
			err:  "could not parse -extldflags: unterminated ' string",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, got, err := parseGoLinkFlag(tc.in)
			if tc.err != "" {
				if err == nil || err.Error() != tc.err {
					t.Fatalf("parseGoLinkFlag(%q) error = %v, want %q", tc.in, err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("external flags = %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("global-values", func(t *testing.T) {
		got, _, err := parseGoLinkFlag(`-X main.version=1.0 -X 'example.com/app.name=hello world' -X 'main.path=C:\x\lib'`)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]map[string]string{
			"main":            {"version": "1.0", "path": `C:\x\lib`},
			"example.com/app": {"name": "hello world"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("global values = %v, want %v", got, want)
		}
	})
}

func TestTrimPath(t *testing.T) {
	t.Setenv("CGO_CFLAGS", "-iquoteinclude -includestdint.h -imacros relative.h")
	root := t.TempDir()
	var binaries [][]byte
	options := optionsFromTarget(*testTarget, sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a", "b"} {
		dir := filepath.Join(root, name)
		if err := os.CopyFS(dir, os.DirFS(filepath.Join(TESTDATA, "trimpath"))); err != nil {
			t.Fatal(err)
		}

		outpath := filepath.Join(root, name+".out")
		cache := filepath.Join(root, "cache-"+name)
		cmd := trimPathCommand(t, dir, "build", "-trimpath", "-p=4", "-o", outpath, ".")
		cmd.Env = append(cmd.Env, "XDG_CACHE_HOME="+cache, "HOME="+cache, "LocalAppData="+cache)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build failed: %v\n%s", err, output)
		}
		binary, err := os.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		for _, localPath := range []string{root, goenv.Get("GOROOT"), goenv.Get("TINYGOROOT"), goenv.Get("GOCACHE")} {
			if bytes.Contains(binary, []byte(localPath)) || bytes.Contains(binary, []byte(filepath.ToSlash(localPath))) {
				t.Errorf("trimmed binary contains local path %q", localPath)
			}
		}
		binaries = append(binaries, binary)
	}

	if !bytes.Equal(binaries[0], binaries[1]) {
		t.Error("trimmed binaries built in different directories are not identical")
	}
	if config.GOOS() != "darwin" {
		files := trimPathDWARFFiles(t, filepath.Join(root, "a.out"), config.GOOS())
		headerPath := "/_/example.com/dependency@v1.2.3/include/shared.h"
		if config.GOOS() == "windows" {
			headerPath = "//_/_/example.com/dependency@v1.2.3/include/shared.h"
		}
		for _, want := range []string{
			"example.com/trimpath/main.go",
			"example.com/dependency@v1.2.3/subpackage/dependency.go",
			headerPath,
		} {
			if !slices.Contains(files, want) {
				t.Errorf("missing DWARF path %q in %v", want, files)
			}
		}
		for _, file := range files {
			if strings.Contains(file, filepath.ToSlash(root)) || strings.Count(file, "example.com/") > 1 {
				t.Errorf("invalid DWARF path %q", file)
			}
		}
	}

	dir := filepath.Join(root, "a")
	for _, trim := range []bool{false, true} {
		outpath := filepath.Join(root, "nodebug.out")
		args := []string{"build", "-no-debug", "-o", outpath}
		if trim {
			args = append(args, "-trimpath")
		}
		cmd := trimPathCommand(t, dir, append(args, ".")...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build failed: %v\n%s", err, output)
		}
		binary, err := os.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		containsPath := bytes.Contains(binary, []byte(dir)) || bytes.Contains(binary, []byte(filepath.ToSlash(dir)))
		if containsPath == trim {
			t.Errorf("trimpath=%v: contains local __FILE__ path=%v", trim, containsPath)
		}
	}
}

func TestTrimPathStackSizes(t *testing.T) {
	dir := t.TempDir()
	source := `package main
func main() {
	done := make(chan int)
	go func() { done <- 42 }()
	println(<-done)
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, trim := range []bool{false, true} {
		t.Run(fmt.Sprintf("trimpath=%v", trim), func(t *testing.T) {
			outpath := filepath.Join(dir, "main.elf")
			args := []string{"build", "-target=pico2", "-o", outpath}
			if trim {
				args = append(args, "-trimpath")
			}
			cmd := trimPathCommand(t, dir, append(args, "main.go")...)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build failed: %v\n%s", err, output)
			}
			file, err := elf.Open(outpath)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if section := file.Section(".tinygo_stacksizes"); section == nil || section.Size == 0 {
				t.Fatal("missing automatic stack-size data")
			}
			binary, err := os.ReadFile(outpath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(binary, []byte("task_stack.go")) {
				t.Fatal("missing stack-size source file")
			}
			sourceDir := filepath.Join(goenv.Get("TINYGOROOT"), "src")
			containsPath := bytes.Contains(binary, []byte(sourceDir)) || bytes.Contains(binary, []byte(filepath.ToSlash(sourceDir)))
			if containsPath == trim {
				t.Errorf("trimpath=%v: contains local runtime source directory=%v", trim, containsPath)
			}
		})
	}
}

func trimPathCommand(t *testing.T, dir string, args ...string) *exec.Cmd {
	t.Helper()
	options := optionsFromTarget(*testTarget, sema)
	if options.Target != "" {
		args = slices.Insert(args, 1, "-target="+options.Target)
	}
	cmd := exec.Command(os.Args[0], append([]string{"test-main"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"TINYGOROOT="+goenv.Get("TINYGOROOT"),
		"GOPATH="+goenv.Get("GOPATH"),
		"GOOS="+options.GOOS,
		"GOARCH="+options.GOARCH,
		"GOARM="+options.GOARM,
	)
	return cmd
}

func trimPathDWARFFiles(t *testing.T, filename, goos string) []string {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var data *dwarf.Data
	switch goos {
	case "windows":
		file, parseErr := pe.NewFile(f)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		data, err = file.DWARF()
	case "wasip1", "wasip2", "js":
		file, parseErr := wasm.Parse(f)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		data, err = file.DWARF()
	default:
		file, parseErr := elf.NewFile(f)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		data, err = file.DWARF()
	}
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	reader := data.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			t.Fatal(err)
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			continue
		}
		lines, err := data.LineReader(entry)
		if err != nil {
			t.Fatal(err)
		}
		if lines == nil {
			continue
		}
		var line dwarf.LineEntry
		for {
			if err := lines.Next(&line); err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				break
			}
		}
		for _, file := range lines.Files() {
			if file != nil {
				files = append(files, strings.ReplaceAll(file.Name, "\\", "/"))
			}
		}
	}
	return files
}

func TestTrimPathTestPackages(t *testing.T) {
	options := optionsFromTarget(*testTarget, sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	if config.GOOS() == "darwin" {
		t.Skip("Mach-O executables keep DWARF in separate object files")
	}
	for _, externalOnly := range []bool{false, true} {
		dir := t.TempDir()
		files := map[string]string{
			"go.mod":           "module example.com/pkg_test\n\ngo 1.23\n",
			"external_test.go": "package pkg_test\nimport \"testing\"\nfunc TestPath(t *testing.T) { t.Log(\"path\") }\n",
		}
		if !externalOnly {
			files["pkg.go"] = "package pkg\nvar Value = 1\n"
		}
		for name, data := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		outpath := filepath.Join(dir, "test.out")
		cmd := trimPathCommand(t, dir, "test", "-trimpath", "-c", "-o", outpath, ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("test build failed: %v\n%s", err, output)
		}
		got := trimPathDWARFFiles(t, outpath, config.GOOS())
		if !slices.Contains(got, "example.com/pkg_test/external_test.go") {
			t.Errorf("externalOnly=%v: incorrect test paths: %v", externalOnly, got)
		}
		if !slices.Contains(got, "_testmain.go") {
			t.Errorf("externalOnly=%v: missing generated test main path", externalOnly)
		}
		for _, file := range got {
			if filepath.IsAbs(filepath.FromSlash(file)) && !strings.HasPrefix(file, "/_/") && !strings.HasPrefix(file, "//_/_/") {
				t.Errorf("externalOnly=%v: untrimmed test path %q", externalOnly, file)
			}
		}
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()

	tests := []string{
		"alias.go",
		"atomic.go",
		"binop.go",
		"buildinfo.go",
		"calls.go",
		"cgo/",
		"channel.go",
		"coreassign.go",
		"embed/",
		"finalizer.go",
		"finalizerbits.go",
		"finalizeridle.go",
		"finalizerinvariants.go",
		"finalizerlarge.go",
		"float.go",
		"gc.go",
		"generics.go",
		"go1.21.go",
		"go1.22/",
		"go1.23/",
		"goroutines.go",
		"init.go",
		"init_multi.go",
		"interface-blank.go",
		"interface.go",
		"json.go",
		"localtypes/",
		"localtypes.go",
		"map.go",
		"map_bigkey.go",
		"math.go",
		"oldgo/",
		"print.go",
		"reflect.go",
		"signal.go",
		"signalnotify.go",
		"slice.go",
		"sort.go",
		"stdlib.go",
		"string.go",
		"structs.go",
		"testing.go",
		"timers.go",
		"typealias.go",
		"zeroalloc.go",
	}

	if *testTarget != "" {
		// This makes it possible to run one specific test (instead of all),
		// which is especially useful to quickly check whether some changes
		// affect a particular target architecture.
		options := optionsFromTarget(*testTarget, sema)
		runPlatTests(options, tests, t)
		if *testTarget == "wasip1" {
			t.Run("cgo-realloc", func(t *testing.T) {
				runTest("cgo-realloc/", options, t, nil, nil)
			})
		}
		return
	}

	t.Run("Host", func(t *testing.T) {
		t.Parallel()
		hostOptions := optionsFromTarget("", sema)
		runPlatTests(hostOptions, tests, t)
		t.Run("testing.go-verbose", func(t *testing.T) {
			t.Parallel()
			runTest("testing-verbose.go", hostOptions, t, nil, nil)
		})

		// scheduler.threads needs threadID, which exists only on Linux and Darwin.
		// scheduler.none does not link on Windows.
		switch runtime.GOOS {
		case "darwin", "linux":
			for _, scheduler := range []string{"threads", "none"} {
				t.Run("finalizerinvariants.go-gc-conservative-scheduler-"+scheduler, func(t *testing.T) {
					t.Parallel()
					options := compileopts.Options(hostOptions)
					options.GC = "conservative"
					options.Scheduler = scheduler
					options.Tags = append(append([]string(nil), hostOptions.Tags...), "runtime_asserts")
					runTest("finalizerinvariants.go", options, t, nil, nil)
				})
			}
		}
	})

	t.Run("boehm-finalizers", func(t *testing.T) {
		t.Parallel()
		options := optionsFromTarget("wasm", sema)
		options.GC = "boehm"
		t.Run("finalizerlarge.go-graph", func(t *testing.T) {
			t.Parallel()
			testOptions := compileopts.Options(options)
			testOptions.GlobalValues = map[string]map[string]string{
				"main": {"finalizerLargeMode": "graph"},
			}
			runTest("finalizerlarge.go", testOptions, t, nil, nil)
		})
		for _, name := range []string{
			"finalizer.go",
			"finalizerbits.go",
			"finalizeridle.go",
			"finalizerlarge.go",
			"finalizerinvariants.go",
		} {
			name := name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				testOptions := compileopts.Options(options)
				if name == "finalizerlarge.go" {
					testOptions.GlobalValues = map[string]map[string]string{
						"main": {"finalizerLargeMode": "runner-payload"},
					}
				}
				runTest(name, testOptions, t, nil, nil)
			})
		}
	})

	// Test a few build options.
	t.Run("build-options", func(t *testing.T) {
		t.Parallel()

		// Test with few optimizations enabled (no inlining, etc).
		t.Run("opt=1", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "1"
			runTestWithConfig("stdlib.go", t, opts, nil, nil)
		})

		// Test with only the bare minimum of optimizations enabled.
		// TODO: fix this for stdlib.go, which currently fails.
		t.Run("opt=0", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "0"
			runTestWithConfig("print.go", t, opts, nil, nil)
		})

		t.Run("opt=0-gc=boehm", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.Opt = "0"
			opts.GC = "boehm"
			runTestWithConfig("gc-boehm-opt0.go", t, opts, nil, nil)
		})

		// Regression test: at -opt=0 the compiler does not always remove a
		// local escaping through a pointer cast, so printitf used to
		// allocate on the panic path. printitf is //go:noheap, so a
		// regression here makes the build fail with a linker error.
		t.Run("opt=0-printitf-cortex-m-qemu", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.Opt = "0"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/panic-value.go", t.TempDir()+"/panic-value", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		// Regression test: at -opt=0 the compiler does not always remove a
		// local escaping through Queue.Push, so runGC used to allocate while
		// marking the runqueue. runGC is //go:noheap, so a regression here
		// makes the build fail with a linker error instead of silently
		// hanging like it used to.
		t.Run("opt=0-gc-cortex-m-qemu", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.Opt = "0"
			emuCheck(t, opts)
			runTestWithConfig("gc.go", t, opts, nil, nil)
		})

		t.Run("gc=none-runtime-panic", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("cortex-m-qemu", sema)
			opts.GC = "none"
			opts.Scheduler = "none"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/trivialpanic.go", t.TempDir()+"/trivialpanic", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		t.Run("scheduler=none-uefi", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("uefi-amd64", sema)
			opts.Scheduler = "none"
			config, err := builder.NewConfig(&opts)
			if err != nil {
				t.Fatal(err)
			}
			err = Build("testdata/trivialpanic.go", t.TempDir()+"/trivialpanic", config)
			if err != nil {
				w := &bytes.Buffer{}
				diagnostics.CreateDiagnostics(err).WriteTo(w, "")
				t.Fatal(w.String())
			}
		})

		t.Run("ldflags", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.GlobalValues = map[string]map[string]string{
				"main": {
					"someGlobal": "foobar",
				},
			}
			runTestWithConfig("ldflags.go", t, opts, nil, nil)
		})

		// Same as ldflags, but the global has a default value in source.
		// -ldflags -X must override it, matching standard Go behaviour.
		t.Run("ldflags-initialized", func(t *testing.T) {
			t.Parallel()
			opts := optionsFromTarget("", sema)
			opts.GlobalValues = map[string]map[string]string{
				"main": {
					"someGlobal": "foobar",
				},
			}
			runTestWithConfig("ldflags-initialized.go", t, opts, nil, nil)
		})
	})

	// Short mode skips the targets that every host can test. Linux CI tests
	// them without -short.
	if !testing.Short() {
		t.Run("EmulatedCortexM3", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("cortex-m-qemu", sema), tests, t)
		})

		t.Run("EmulatedRISCV", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("riscv-qemu", sema), tests, t)
		})

		t.Run("AVR", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("simavr", sema), tests, t)
		})

		t.Run("WebAssembly", func(t *testing.T) {
			t.Parallel()

			runPlatTests(optionsFromTarget("wasm", sema), tests, t)
			runGCLivenessTest(optionsFromTarget("wasm", sema), t)
			// Test with -gc=boehm.
			t.Run("gc.go-boehm", func(t *testing.T) {
				t.Parallel()
				optionsBoehm := optionsFromTarget("wasm", sema)
				optionsBoehm.GC = "boehm"
				runTest("gc.go", optionsBoehm, t, nil, nil)
				runTest("gc-boehm.go", optionsBoehm, t, nil, nil)
			})
		})
		t.Run("WASIp1", func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget("wasip1", sema)
			runPlatTests(options, tests, t)
			runGCLivenessTest(options, t)
			t.Run("cgo-realloc", func(t *testing.T) {
				runTest("cgo-realloc/", options, t, nil, nil)
			})

			// Test with -gc=boehm.
			t.Run("gc.go-boehm", func(t *testing.T) {
				t.Parallel()
				optionsBoehm := optionsFromTarget("wasip1", sema)
				optionsBoehm.GC = "boehm"
				runTest("gc.go", optionsBoehm, t, nil, nil)
				runTest("gc-boehm.go", optionsBoehm, t, nil, nil)
			})
		})
		t.Run("WASIp2", func(t *testing.T) {
			t.Parallel()
			runPlatTests(optionsFromTarget("wasip2", sema), tests, t)
			runGCLivenessTest(optionsFromTarget("wasip2", sema), t)
		})
	}

	// Linux CI also tests the other Linux architectures without -short, so
	// short mode skips them. Windows CI runs only in short mode, so
	// Windows386 continues to run.
	if runtime.GOOS == "linux" && !testing.Short() {
		for name, osArch := range supportedLinuxArches {
			options := optionsFromOSARCH(osArch, sema)
			if options.GOARCH != runtime.GOARCH { // Native architecture already run above.
				t.Run(name, func(t *testing.T) {
					runPlatTests(options, tests, t)
				})
			}
		}
		t.Run("MIPS little-endian", func(t *testing.T) {
			// Run a single test for GOARCH=mipsle to see whether it works at
			// all. It is already mostly tested because GOARCH=mips and
			// GOARCH=mipsle are so similar, but it's good to have an extra test
			// to be sure.
			t.Parallel()
			options := optionsFromOSARCH("linux/mipsle/softfloat", sema)
			emuCheck(t, options)
			runTest("cgo/", options, t, nil, nil)
		})
	} else if runtime.GOOS == "windows" {
		if runtime.GOARCH != "386" {
			t.Run("Windows386", func(t *testing.T) {
				t.Parallel()
				runPlatTests(optionsFromOSARCH("windows/386", sema), tests, t)
			})
		}
	}
}

func TestESP32QEMU(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("esp32-qemu", sema)
	emuCheck(t, options)
	machines, err := exec.Command("qemu-system-xtensa", "-machine", "help").Output()
	if err != nil {
		t.Fatal("failed to list qemu-system-xtensa machines:", err)
	}
	if !regexp.MustCompile(`(?m)^esp32\s`).Match(machines) {
		t.Skip("qemu-system-xtensa does not support the ESP32 machine")
	}
	for _, name := range []string{"print.go", "recover.go"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTest(name, options, t, nil, nil)
		})
	}
}

func runPlatTests(options compileopts.Options, tests []string, t *testing.T) {
	emuCheck(t, options)

	spec, err := compileopts.LoadTarget(&options)
	if err != nil {
		t.Fatal("failed to load target spec:", err)
	}

	// FIXME: this should really be:
	// isWebAssembly := strings.HasPrefix(spec.Triple, "wasm")
	isWASI := strings.HasPrefix(options.Target, "wasi")
	isWebAssembly := isWASI || strings.HasPrefix(options.Target, "wasm") || (options.Target == "" && strings.HasPrefix(options.GOARCH, "wasm"))
	isBaremetal := options.Target == "simavr" || options.Target == "cortex-m-qemu" || options.Target == "riscv-qemu"
	_, goMinor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal("could not get version:", goMinor)
	}

	for _, name := range tests {
		if name == "goroutines.go" && (spec.Scheduler == "threads" || spec.Scheduler == "cores") {
			// This test intentionally checks concurrent scheduling by comparing
			// output order, so only run it with non-threaded schedulers.
			continue
		}
		if name == "coreassign.go" && spec.Scheduler != "cores" {
			// Core assignment needs the cores scheduler.
			continue
		}
		if options.GOOS == "linux" && (options.GOARCH == "arm" || options.GOARCH == "386") {
			switch name {
			case "timers.go":
				// Timer tests do not work because syscall.seek is implemented
				// as Assembly in mainline Go and causes linker failure
				continue
			}
		}
		if options.GOOS == "linux" && (options.GOARCH == "mips" || options.GOARCH == "mipsle") {
			if name == "atomic.go" || name == "timers.go" {
				// 64-bit atomic operations aren't currently supported on MIPS.
				continue
			}
		}
		if options.GOOS == "linux" && options.GOARCH == "mips" {
			if name == "cgo/" {
				// CGo isn't supported yet on big-endian systems (needs updates
				// to bitfield access methods).
				continue
			}
		}
		if options.Target == "cortex-m-qemu" && goMinor >= 27 && name == "json.go" {
			// Go 1.27 jsonv2 exceeds the LM3S6965's 256KiB flash. json.go
			// is still covered by larger targets such as riscv-qemu.
			continue
		}
		if options.Target == "simavr" {
			// Not all tests are currently supported on AVR.
			// Skip the ones that aren't.
			switch name {
			case "reflect.go":
				// Reflect tests do not run correctly, probably because of the
				// limited amount of memory.
				continue

			case "map_bigkey.go":
				// Compiler generates many large stack temporaries for [256]byte
				// map keys, overflowing the goroutine stack (384 bytes).
				continue

			case "gc.go":
				// Does not pass due to high mark false positive rate.
				continue

			case "buildinfo.go", "json.go", "stdlib.go", "testing.go":
				// Too big for AVR. Doesn't fit in flash/RAM.
				continue

			case "cgo/":
				// CGo function pointers don't work on AVR (needs LLVM 16 and
				// some compiler changes).
				continue

			case "timers.go":
				// Crashes starting with Go 1.23.
				// Bug: https://github.com/llvm/llvm-project/issues/104032
				continue

			default:
			}
		}
		if options.Target == "wasip2" {
			switch name {
			case "cgo/":
				// waisp2 use our own libc; cgo tests fail
				continue
			}
		}
		if options.GOOS == "windows" {
			switch name {
			case "signal.go", "signalnotify.go":
				// os/signal does not link on Windows.
				continue
			}
		}
		if isWebAssembly || isBaremetal {
			switch name {
			case "signal.go":
				// Signals only work on POSIX-like systems.
				continue
			}
		}
		if options.Target != "wasm" {
			switch name {
			case "finalizer.go", "finalizerbits.go", "finalizeridle.go", "finalizerlarge.go":
				// These tests require deterministic finalization on target wasm.
				// finalizerinvariants.go covers other block GC targets.
				continue
			}
		}
		if options.Target == "simavr" {
			switch name {
			case "finalizerinvariants.go":
				// Skip because runtime.GC does not return. See the gc.go exclusion above.
				continue
			}
		}
		name := name // redefine to avoid race condition
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testOptions := compileopts.Options(options)
			switch name {
			case "finalizerinvariants.go", "finalizerlarge.go", "interface.go", "interface-blank.go":
				testOptions.Tags = append(append([]string(nil), options.Tags...), "runtime_asserts")
			}
			if testOptions.Target == "simavr" && (name == "math.go" || name == "string.go") {
				// These tests exceed simavr's default 384-byte goroutine stack.
				testOptions.StackSize = 512
			}
			runTest(name, testOptions, t, nil, nil)
		})
	}
	if !strings.HasPrefix(spec.Emulator, "simavr ") {
		t.Run("env.go", func(t *testing.T) {
			t.Parallel()
			runTest("env.go", options, t, []string{"first", "second"}, []string{"ENV1=VALUE1", "ENV2=VALUE2"})
		})
	}
	if isWebAssembly {
		t.Run("alias.go-scheduler-none", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.Scheduler = "none"
			runTest("alias.go", options, t, nil, nil)
		})
	}
	if options.Target == "" || isWASI {
		t.Run("filesystem.go", func(t *testing.T) {
			t.Parallel()
			runTest("filesystem.go", options, t, nil, nil)
		})
	}
	if options.Target == "" || options.Target == "wasm" || isWASI {
		t.Run("rand.go", func(t *testing.T) {
			t.Parallel()
			runTest("rand.go", options, t, nil, nil)
		})
	}
	t.Run("recover.go", func(t *testing.T) {
		t.Parallel()
		runTest("recover.go", options, t, nil, nil)
	})
	if isWebAssembly {
		t.Run("recover-explicit.go", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.Scheduler = "none"
			options.PanicUnwind = "explicit"
			runTest("recover-explicit.go", options, t, nil, nil)
		})
		t.Run("calls.go-unwind-none", func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options(options)
			options.PanicUnwind = "none"
			runTest("calls.go", options, t, nil, nil)
		})
	}
}

func emuCheck(t *testing.T, options compileopts.Options) {
	// Check if the emulator is installed.
	spec, err := compileopts.LoadTarget(&options)
	if err != nil {
		t.Fatal("failed to load target spec:", err)
	}
	if spec.Emulator != "" {
		emulatorCommand, _, _ := strings.Cut(spec.Emulator, " ")
		_, err := exec.LookPath(emulatorCommand)
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				t.Skipf("emulator not installed: %q", emulatorCommand)
			}

			t.Errorf("searching for emulator: %v", err)
			return
		}
	}
}

func optionsFromTarget(target string, sema chan struct{}) compileopts.Options {
	separators := strings.Count(target, "/")
	if (separators == 1 || separators == 2) && !strings.HasSuffix(target, ".json") {
		return optionsFromOSARCH(target, sema)
	}
	return compileopts.Options{
		// GOOS/GOARCH are only used if target == ""
		GOOS:          goenv.Get("GOOS"),
		GOARCH:        goenv.Get("GOARCH"),
		GOARM:         goenv.Get("GOARM"),
		GOMIPS:        goenv.Get("GOMIPS"),
		Target:        target,
		Semaphore:     sema,
		InterpTimeout: 180 * time.Second,
		Debug:         true,
		VerifyIR:      true,
		Opt:           "z",
	}
}

// optionsFromOSARCH returns a set of options based on the "osarch" string. This
// string is in the form of "os/arch/subarch", with the subarch only sometimes
// being necessary. Examples are "darwin/amd64" or "linux/arm/7".
func optionsFromOSARCH(osarch string, sema chan struct{}) compileopts.Options {
	parts := strings.Split(osarch, "/")
	options := compileopts.Options{
		GOOS:          parts[0],
		GOARCH:        parts[1],
		Semaphore:     sema,
		InterpTimeout: 180 * time.Second,
		Debug:         true,
		VerifyIR:      true,
		Opt:           "z",
	}
	switch options.GOARCH {
	case "arm":
		options.GOARM = parts[2]
	case "mips", "mipsle":
		options.GOMIPS = parts[2]
	}
	return options
}

// runGCLivenessTest runs testdata/gc-liveness-repro.go. Only targets with
// stack objects can lose a GC root the way it reproduces, so it is registered
// per wasm target instead of being part of the shared test list.
func runGCLivenessTest(options compileopts.Options, t *testing.T) {
	t.Run("gc-liveness-repro.go", func(t *testing.T) {
		t.Parallel()
		runTest("gc-liveness-repro.go", options, t, nil, nil)
	})
}

func runTest(name string, options compileopts.Options, t *testing.T, cmdArgs, environmentVars []string) {
	t.Helper()
	runTestWithConfig(name, t, options, cmdArgs, environmentVars)
}

func runTestWithConfig(name string, t *testing.T, options compileopts.Options, cmdArgs, environmentVars []string) {
	t.Helper()
	// Get the expected output for this test.
	// Note: not using filepath.Join as it strips the path separator at the end
	// of the path.
	path := TESTDATA + "/" + name
	// Get the expected output for this test.
	expectedOutputPath := path[:len(path)-3] + ".txt"
	pkgName := "./" + path
	if path[len(path)-1] == '/' {
		expectedOutputPath = path + "out.txt"
		options.Directory = path
		pkgName = "."
	}
	isWebAssembly := strings.HasPrefix(options.Target, "wasi") ||
		strings.HasPrefix(options.Target, "wasm") ||
		strings.HasPrefix(options.GOARCH, "wasm")
	if name == "testing.go" && isWebAssembly {
		expectedOutputPath = TESTDATA + "/testing-wasm.txt"
	}

	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	// Build the test binary.
	stdout := &bytes.Buffer{}
	_, err = buildAndRun(pkgName, config, stdout, cmdArgs, environmentVars, 2*time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
		if config.EmulatorName() == "simavr" {
			// simavr before v1.8 wrote firmware output to stderr and loader logs
			// to stdout, but PR #490 swapped these streams:
			// https://github.com/buserror/simavr/pull/490
			cmd.Stdout = stdout
		}
		return cmd.Run()
	})
	if err != nil {
		w := &bytes.Buffer{}
		diagnostics.CreateDiagnostics(err).WriteTo(w, "")
		for line := range strings.SplitSeq(strings.TrimRight(w.String(), "\n"), "\n") {
			t.Log(line)
		}
		if stdout.Len() != 0 {
			t.Logf("output:\n%s", stdout.String())
		}
		t.Fail()
		return
	}

	actual := stdout.Bytes()
	if config.EmulatorName() == "simavr" {
		actual = cleanSimAVRTestOutput(actual)
	}
	if config.EmulatorName() == "qemu-system-xtensa" {
		actual = cleanESP32QEMUOutput(actual)
	}
	if name == "testing.go" || name == "testing-verbose.go" {
		// Strip actual time.
		re := regexp.MustCompile(`\([0-9]\.[0-9][0-9]s\)`)
		actual = re.ReplaceAllLiteral(actual, []byte{'(', '0', '.', '0', '0', 's', ')'})
	}

	// Check whether the command ran successfully.
	if err != nil {
		t.Error("failed to run:", err)
	}
	checkOutput(t, expectedOutputPath, actual)

	if t.Failed() {
		r := bufio.NewReader(bytes.NewReader(actual))
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			t.Log("stdout:", line[:len(line)-1])
		}
		t.Fail()
	}
}

func cleanESP32QEMUOutput(output []byte) []byte {
	entryLine := bytes.Index(output, []byte("\nentry "))
	if entryLine < 0 {
		return output
	}
	entryLineEnd := bytes.IndexByte(output[entryLine+1:], '\n')
	if entryLineEnd < 0 {
		return output
	}
	return output[entryLine+1+entryLineEnd+1:]
}

func cleanSimAVRTestOutput(output []byte) []byte {
	output = bytes.ReplaceAll(output, []byte{0x1b, '[', '3', '2', 'm'}, nil)
	output = bytes.ReplaceAll(output, []byte{0x1b, '[', '0', 'm'}, nil)
	output = bytes.ReplaceAll(output, []byte{'.', '.', '\n'}, []byte{'\n'})
	output = bytes.ReplaceAll(output, []byte{'\n', '.', '\n'}, []byte{'\n', '\n'})

	var cleaned []byte
	for _, line := range bytes.SplitAfter(output, []byte{'\n'}) {
		trimmedLine := bytes.TrimRight(line, "\r\n")
		if simavrLoadTextLogPattern.Match(trimmedLine) || simavrLoadBytesLogPattern.Match(trimmedLine) {
			continue
		}
		cleaned = append(cleaned, line...)
	}
	return cleaned
}

var (
	simavrLoadTextLogPattern  = regexp.MustCompile(`^Loaded [0-9]+ \.[A-Za-z0-9_]+( at address 0x[0-9a-fA-F]+)?$`)
	simavrLoadBytesLogPattern = regexp.MustCompile(`^Loaded [0-9]+ bytes of [A-Za-z]+ data at (0x)?[0-9a-fA-F]+$`)
)

// Check whether the output of a test equals the expected output.
func checkOutput(t *testing.T, filename string, actual []byte) {
	t.Helper()
	expectedOutput, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal("could not read output file:", err)
	}
	checkOutputData(t, expectedOutput, actual)
}

func checkOutputData(t *testing.T, expectedOutput, actual []byte) {
	t.Helper()
	actual = bytes.ReplaceAll(actual, []byte("\r\n"), []byte("\n"))

	if !bytes.Equal(actual, expectedOutput) {
		t.Errorf("output did not match (expected %d bytes, got %d bytes):", len(expectedOutput), len(actual))
		t.Error(string(Diff("expected", expectedOutput, "actual", actual)))
	}
}

func TestTest(t *testing.T) {
	t.Parallel()

	type targ struct {
		name string
		opts compileopts.Options
	}
	targs := []targ{
		// Host
		{"Host", optionsFromTarget("", sema)},
	}
	if !testing.Short() {
		if runtime.GOOS == "linux" {
			for name, osArch := range supportedLinuxArches {
				options := optionsFromOSARCH(osArch, sema)
				if options.GOARCH != runtime.GOARCH { // Native architecture already run above.
					targs = append(targs, targ{name, options})
				}
			}
		}

		targs = append(targs,
			// QEMU microcontrollers
			targ{"EmulatedCortexM3", optionsFromTarget("cortex-m-qemu", sema)},
			targ{"EmulatedRISCV", optionsFromTarget("riscv-qemu", sema)},

			// Node/Wasmtime
			targ{"WASM", optionsFromTarget("wasm", sema)},
			targ{"WASI", optionsFromTarget("wasip1", sema)},
		)
	}
	for _, targ := range targs {
		t.Run(targ.name, func(t *testing.T) {
			t.Parallel()

			emuCheck(t, targ.opts)

			t.Run("StringComparisons", func(t *testing.T) {
				t.Parallel()

				var wg sync.WaitGroup
				defer wg.Wait()

				out := ioLogger(t, &wg)
				defer out.Close()

				opts := targ.opts
				passed, err := Test("github.com/tinygo-org/tinygo/tests/stringcompare", false, out, out, &opts, "")
				if err != nil {
					t.Errorf("test error: %v", err)
				}
				if !passed {
					t.Error("test failed")
				}
			})

			t.Run("Pass", func(t *testing.T) {
				t.Parallel()

				// Test a package which builds and passes normally.

				var wg sync.WaitGroup
				defer wg.Wait()

				out := ioLogger(t, &wg)
				defer out.Close()

				opts := targ.opts
				passed, err := Test("github.com/tinygo-org/tinygo/tests/testing/pass", false, out, out, &opts, "")
				if err != nil {
					t.Errorf("test error: %v", err)
				}
				if !passed {
					t.Error("test failed")
				}
			})

			t.Run("Fail", func(t *testing.T) {
				t.Parallel()

				// Test a package which builds fine but fails.

				var wg sync.WaitGroup
				defer wg.Wait()

				out := ioLogger(t, &wg)
				defer out.Close()

				opts := targ.opts
				passed, err := Test("github.com/tinygo-org/tinygo/tests/testing/fail", false, out, out, &opts, "")
				if err != nil {
					t.Errorf("test error: %v", err)
				}
				if passed {
					t.Error("test passed")
				}
			})

			if targ.name != "Host" {
				// Emulated tests are somewhat slow, and these do not need to be run across every platform.
				return
			}

			t.Run("Nothing", func(t *testing.T) {
				t.Parallel()

				// Test a package with no test files.

				var wg sync.WaitGroup
				defer wg.Wait()

				out := ioLogger(t, &wg)
				defer out.Close()

				var output bytes.Buffer
				opts := targ.opts
				passed, err := Test("github.com/tinygo-org/tinygo/tests/testing/nothing", false, io.MultiWriter(&output, out), out, &opts, "")
				if err != nil {
					t.Errorf("test error: %v", err)
				}
				if !passed {
					t.Error("test failed")
				}
				if !strings.Contains(output.String(), "[no test files]") {
					t.Error("missing [no test files] in output")
				}
			})

			t.Run("BuildErr", func(t *testing.T) {
				t.Parallel()

				// Test a package which fails to build.

				var wg sync.WaitGroup
				defer wg.Wait()

				out := ioLogger(t, &wg)
				defer out.Close()

				opts := targ.opts
				passed, err := Test("github.com/tinygo-org/tinygo/tests/testing/builderr", false, out, out, &opts, "")
				if err == nil {
					t.Error("test did not error")
				}
				if passed {
					t.Error("test passed")
				}
			})

			t.Run("NoTestsInvalid", func(t *testing.T) {
				t.Parallel()

				for _, test := range []struct {
					name   string
					source string
					want   string
				}{
					{"undefined", "package probe\nfunc check() { undefinedName() }\n", "undefinedName"},
					{"syntax", "package probe\nfunc check() {\n", "expected"},
				} {
					t.Run(test.name, func(t *testing.T) {
						dir := t.TempDir()
						for name, content := range map[string]string{
							"go.mod":   "module example.com/notests\n\ngo 1.25.0\n",
							"probe.go": test.source,
						} {
							if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
								t.Fatal(err)
							}
						}
						opts := targ.opts
						opts.Directory = dir
						var output bytes.Buffer
						passed, err := Test(".", false, &output, &output, &opts, "")
						if err == nil || !strings.Contains(err.Error(), test.want) {
							t.Fatalf("error = %v, want diagnostic containing %q", err, test.want)
						}
						if passed {
							t.Error("invalid package passed")
						}
						if strings.Contains(output.String(), "[no test files]") {
							t.Error("invalid package reported as having no test files")
						}
					})
				}
			})
		})
	}
}

func ioLogger(t *testing.T, wg *sync.WaitGroup) io.WriteCloser {
	r, w := io.Pipe()
	wg.Go(func() {
		defer r.Close()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			t.Log(scanner.Text())
		}
	})

	return w
}

func TestGetListOfPackages(t *testing.T) {
	opts := optionsFromTarget("", sema)
	tests := []struct {
		pkgs          []string
		expectedPkgs  []string
		expectesError bool
	}{
		{
			pkgs: []string{"./tests/testing/recurse/..."},
			expectedPkgs: []string{
				"github.com/tinygo-org/tinygo/tests/testing/recurse",
				"github.com/tinygo-org/tinygo/tests/testing/recurse/subdir",
			},
		},
		{
			pkgs: []string{"./tests/testing/pass"},
			expectedPkgs: []string{
				"github.com/tinygo-org/tinygo/tests/testing/pass",
			},
		},
		{
			pkgs: []string{"fmt", "./tests/testing/pass"},
			expectedPkgs: []string{
				"fmt",
				"github.com/tinygo-org/tinygo/tests/testing/pass",
			},
		},
		{
			pkgs:          []string{"./tests/testing"},
			expectesError: true,
		},
	}

	for _, test := range tests {
		packages, err := getListOfPackages(test.pkgs, &opts)
		if err != nil && !test.expectesError {
			t.Errorf("unexpected error: %v", err)
		} else if err == nil && test.expectesError {
			t.Error("expected error, but got none")
		}

		var actualPkgs []string
		for _, pkg := range packages {
			actualPkgs = append(actualPkgs, pkg.ImportPath)
			if pkg.Standard != (pkg.ImportPath == "fmt") {
				t.Errorf("package %s: unexpected Standard value %v", pkg.ImportPath, pkg.Standard)
			}
		}
		if !reflect.DeepEqual(test.expectedPkgs, actualPkgs) {
			t.Errorf("expected two slices to be equal, expected %v got %v", test.expectedPkgs, actualPkgs)
		}
	}
}

// This TestMain is necessary because TinyGo may also be invoked to run certain
// LLVM tools in a separate process. Not capturing these invocations would lead
// to recursive tests.
func TestMain(m *testing.M) {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "clang", "ld.lld", "wasm-ld":
			// Invoke a specific tool.
			err := builder.RunTool(os.Args[1], os.Args[2:]...)
			if err != nil {
				// The tool should have printed an error message already.
				// Don't print another error message here.
				os.Exit(1)
			}
			os.Exit(0)
		case "test-main":
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			main()
			os.Exit(0)
		}
	}

	// make test sets these to build TinyGo against the host LLVM.
	for _, name := range []string{"CGO_CPPFLAGS", "CGO_CFLAGS", "CGO_CXXFLAGS"} {
		os.Unsetenv(name)
	}

	// Run normal tests.
	os.Exit(m.Run())
}

func TestInitAllDWARF(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("test reads ELF DWARF")
	}

	options := optionsFromTarget("", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build("testdata/init.go", "", t.TempDir(), config)
	if err != nil {
		t.Fatal("failed to build binary:", err)
	}
	f, err := elf.Open(result.Binary)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := f.DWARF()
	if err != nil {
		t.Fatal(err)
	}
	r := d.Reader()
	for {
		entry, err := r.Next()
		if err != nil {
			t.Fatal(err)
		}
		if entry == nil {
			t.Fatal("no DWARF subprogram for runtime.initAll")
		}
		if entry.Tag == dwarf.TagSubprogram && entry.Val(dwarf.AttrName) == "runtime.initAll" {
			return
		}
	}
}
