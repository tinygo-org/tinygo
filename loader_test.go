package main

import (
	"bytes"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
)

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
		})
	}
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

func TestWASIWorkingDirectory(t *testing.T) {
	opts := optionsFromTarget("wasip1", sema)
	emuCheck(t, opts)
	config, err := builder.NewConfig(&opts)
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build(TESTDATA+"/env.go", "wasm", t.TempDir(), config)
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	preopen := "--dir=" + wd + "::/workspace"
	for _, tc := range []struct {
		name string
		cwd  string
		args []string
	}{
		{"NoPreopens", "", nil},
		{"FirstPreopen", "/workspace", []string{preopen, "--env=EXPECT_FILE=testdata/filesystem.txt"}},
		{"PWD", "/workspace/testdata", []string{preopen, "--env=PWD=/workspace/testdata", "--env=EXPECT_FILE=filesystem.txt"}},
		{"NormalizedPWD", "/workspace/testdata", []string{preopen, "--env=PWD=/workspace/../workspace/testdata", "--env=EXPECT_FILE=filesystem.txt"}},
		{"TrailingSlash", "/workspace/testdata/", []string{preopen, "--env=PWD=/workspace/testdata/", "--env=EXPECT_FILE=filesystem.txt"}},
		{"UnmappedPWD", "/no-access", []string{"--env=PWD=/no-access", "--env=EXPECT_FILE_FAILURE=missing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"run", "--env=ENV1=VALUE1", "--env=ENV2=VALUE2", "--env=EXPECT_CWD=" + tc.cwd}
			args = append(args, tc.args...)
			args = append(args, result.Binary, "first", "second")
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			output, err := exec.CommandContext(ctx, "wasmtime", args...).CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run: %v\n%s", err, output)
			}
			checkOutput(t, TESTDATA+"/env.txt", output)
		})
	}
}

func TestWASIp2WorkingDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-target tests disabled")
	}
	opts := optionsFromTarget("wasip2", sema)
	opts.Tags = append(opts.Tags, "runtime_asserts")
	emuCheck(t, opts)
	config, err := builder.NewConfig(&opts)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	result, err := builder.Build("testdata/wasip2_cwd.go", "wasm", dir, config)
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		pwd      string
		cwd      string
		unmapped bool
	}{
		{"NestedPreopen", "/workspace/testdata", "/workspace/testdata", false},
		{"NormalizedPWD", "/workspace/../workspace/testdata/", "/workspace/testdata/", false},
		{"UnmappedPWD", "/no-access", "/no-access", true},
		{"UnmappedNormalizedPWD", "/no-access/../no-access/", "/no-access/", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			output, err := exec.CommandContext(ctx, "wasmtime", "run", "--dir="+wd+"::/workspace",
				"--dir="+filepath.Join(wd, "testdata")+"::/workspace/test",
				"--env=PWD="+tc.pwd, "--env=EXPECT_CWD="+tc.cwd,
				"--env=EXPECT_UNMAPPED_CWD="+fmt.Sprint(tc.unmapped),
				result.Binary).CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run: %v\n%s", err, output)
			}
		})
	}
}

func TestWasmtimePreopens(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"None", nil, false},
		{"Root", []string{"--dir=/"}, false},
		{"Custom", []string{"--dir=/host::/custom"}, false},
		{"Tmp", []string{"--dir=/tmp"}, true},
		{"AliasedTmp", []string{"--dir=/host::/tmp"}, true},
		{"SeparateFlag", []string{"--dir", "/host::/tmp"}, true},
		{"TrailingSlash", []string{"--dir=/host::/tmp/"}, true},
		{"UnnormalizedGuest", []string{"--dir=/host::/other/../tmp"}, false},
		{"IncompleteFlag", []string{"--dir"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasWasmtimePreopen(tc.args, "/tmp"); got != tc.want {
				t.Fatalf("tmp preopen = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveDirectory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	type directory struct {
		name   string
		target string
	}
	dirs := []directory{{target, target}}
	longTarget := filepath.Join(root, strings.Repeat("x", 240))
	if err := os.Mkdir(longTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	dirs = append(dirs, directory{longTarget, longTarget})
	link := filepath.Join(root, "link")
	if runtime.GOOS == "windows" {
		if output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
			t.Fatalf("creating junction: %v\n%s", err, output)
		}
	} else if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	dirs = append(dirs, directory{link, target})
	for _, tc := range dirs {
		resolved, err := resolveDirectory(tc.name)
		if err != nil {
			t.Fatalf("resolve %q: %v", tc.name, err)
		}
		parent, err := os.Stat(filepath.Dir(resolved))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.Stat(root)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(parent, want) || filepath.Base(resolved) != filepath.Base(tc.target) {
			t.Fatalf("resolved %q to %q, want %q", tc.name, resolved, tc.target)
		}
	}
	if _, err := resolveDirectory(filepath.Join(root, "missing")); !os.IsNotExist(err) {
		t.Fatalf("missing directory: %v", err)
	}
}

func TestWasmtimeWorkingDirectory(t *testing.T) {
	type testCase struct {
		name string
		wd   string
		tmp  string
		test bool
		dirs []string
		pwd  string
	}
	tests := []testCase{
		{"Run", "/workspace/package", "/host/tmp", false, []string{"--dir=.", "--dir=/workspace/package", "--dir=/host/tmp::/tmp"}, "/workspace/package"},
		{"Test", "/workspace/package", "/host/tmp", true, []string{"--dir=/"}, "/workspace/package"},
		{"Root", "/", "/host/tmp", true, []string{"--dir=/"}, "/"},
		{"Backslash", `/workspace/a\b`, "/host/tmp", true, []string{"--dir=/"}, `/workspace/a\b`},
	}
	if runtime.GOOS == "windows" {
		tests = []testCase{
			{"Run", `D:\workspace\package`, `C:\Temp`, false, []string{"--dir=.", `--dir=D:\workspace\package::/workspace/package`, `--dir=C:\Temp::/tmp`}, "/workspace/package"},
			{"Test", `D:\workspace\package`, `C:\Temp`, true, []string{`--dir=D:\::/`, `--dir=C:\Temp::/tmp`}, "/workspace/package"},
			{"Root", `D:\`, `C:\Temp`, true, []string{`--dir=D:\::/`, `--dir=C:\Temp::/tmp`}, "/"},
			{"UNC", `\\server\share\package`, `C:\Temp`, true, []string{`--dir=\\server\share\::/`, `--dir=C:\Temp::/tmp`}, "/package"},
			{"UNCRoot", `\\server\share`, `C:\Temp`, true, []string{`--dir=\\server\share\::/`, `--dir=C:\Temp::/tmp`}, "/"},
			{"ExtendedPath", `\\?\D:\workspace\package`, `C:\Temp`, true, []string{`--dir=\\?\D:\::/`, `--dir=C:\Temp::/tmp`}, "/workspace/package"},
			{"SpacesAndQuotes", `D:\workspace\package's files`, `C:\Temp files`, true, []string{`--dir=D:\::/`, `--dir=C:\Temp files::/tmp`}, "/workspace/package's files"},
		}
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dirs, pwd := wasmtimePreopens(tc.wd, tc.tmp, tc.test, nil)
			if !slices.Equal(dirs, tc.dirs) || pwd != tc.pwd {
				t.Fatalf("preopens = %q, PWD = %q; want %q, %q", dirs, pwd, tc.dirs, tc.pwd)
			}
			custom := []string{"--dir=/custom::/tmp"}
			dirs, pwd = wasmtimePreopens(tc.wd, tc.tmp, tc.test, custom)
			want := tc.dirs
			if !tc.test || runtime.GOOS == "windows" {
				want = want[:len(want)-1]
			}
			if !slices.Equal(dirs, want) || pwd != tc.pwd {
				t.Fatalf("explicit tmp preopen: preopens = %q, PWD = %q; want %q, %q", dirs, pwd, want, tc.pwd)
			}
		})
	}
}

func TestWASITestPreopens(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-target tests disabled")
	}
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "spaces and 'quotes'")
	pkgDir := filepath.Join(dir, "package")
	if err := os.MkdirAll(filepath.Join(pkgDir, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	pkgNameForTest := pkgDir
	if runtime.GOOS != "windows" {
		pkgNameForTest = filepath.Join(dir, "alias")
		if err := os.Symlink(pkgDir, pkgNameForTest); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range map[string]string{
		"go.mod":              "module example.com/wasi-test-root\n\ngo 1.23\n",
		"parent.txt":          "fixture\n",
		"package/fixture.txt": "fixture\n",
		"package/access_test.go": `package access

import (
	"os"
	"testing"
)

var initialCWD, initialCWDError = os.Getwd()

func TestFileAccess(t *testing.T) {
	if initialCWDError != nil {
		t.Fatal(initialCWDError)
	}
	if initialCWD != os.Getenv("PWD") {
		t.Fatalf("initial cwd = %q, want %q", initialCWD, os.Getenv("PWD"))
	}
	for _, name := range []string{"fixture.txt", "../parent.txt", "/custom/parent.txt"} {
		if data, err := os.ReadFile(name); err != nil || string(data) != "fixture\n" {
			t.Fatalf("read %q: %q, %v", name, data, err)
		}
	}
	if err := os.Chdir("child"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile("../../parent.txt"); err != nil || string(data) != "fixture\n" {
		t.Fatalf("read after chdir: %q, %v", data, err)
	}
	tmp, err := os.CreateTemp("", "wasi-root")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
}
`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"wasip1", "wasip2"} {
		for _, tc := range []struct {
			name        string
			compileTest bool
			tmpMount    bool
		}{
			{"Run", false, false},
			{"RunWithTmpMount", false, true},
			{"Test", true, false},
		} {
			pkgName := "testdata/stdlib.go"
			if tc.compileTest {
				pkgName = pkgNameForTest
			}
			t.Run(target+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				opts := optionsFromTarget(target, sema)
				opts.TestConfig.CompileTestBinary = tc.compileTest
				if tc.compileTest {
					opts.Directory = pkgDir
				}
				emuCheck(t, opts)
				config, err := builder.NewConfig(&opts)
				if err != nil {
					t.Fatal(err)
				}
				custom := "--dir=" + dir + "::/custom"
				mounts := "'" + strings.ReplaceAll(custom, `'`, `'\''`) + "'"
				customTmp := "--dir=" + dir + "::/tmp"
				if tc.tmpMount {
					mounts += " '" + strings.ReplaceAll(customTmp, `'`, `'\''`) + "'"
				}
				config.Target.Emulator = strings.Replace(config.Target.Emulator, " {}", " "+mounts+" {}", 1)
				output := &bytes.Buffer{}
				_, err = buildAndRun(pkgName, config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, _ builder.BuildResult) error {
					var dirs []string
					for _, arg := range cmd.Args {
						if strings.HasPrefix(arg, "--dir=") {
							dirs = append(dirs, arg)
						}
					}
					guestWD := wd
					mount := wd
					if runtime.GOOS == "windows" {
						volume := filepath.VolumeName(wd)
						guestWD = filepath.ToSlash(wd[len(volume):])
						mount += "::" + guestWD
					}
					want := []string{"--dir=.", "--dir=" + mount}
					if tc.compileTest {
						resolved, err := resolveDirectory(pkgDir)
						if err != nil {
							t.Fatal(err)
						}
						guestWD = filepath.ToSlash(resolved[len(filepath.VolumeName(resolved)):])
						root := "/"
						if runtime.GOOS == "windows" {
							root = filepath.VolumeName(resolved) + `\::/`
						}
						want = []string{"--dir=" + root}
					}
					if (!tc.compileTest || runtime.GOOS == "windows") && !tc.tmpMount {
						want = append(want, "--dir="+os.TempDir()+"::/tmp")
					}
					want = append(want, custom)
					if tc.tmpMount {
						want = append(want, customTmp)
					}
					if !slices.Equal(dirs, want) {
						t.Errorf("preopens = %v, want %v", dirs, want)
					}
					if !slices.Contains(cmd.Args, "--env=PWD="+guestWD) {
						t.Errorf("missing guest PWD %q in %v", guestWD, cmd.Args)
					}
					return cmd.Run()
				})
				if err != nil {
					t.Fatalf("failed to run: %v\n%s", err, output)
				}
			})
		}
	}
}
