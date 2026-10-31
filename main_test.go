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

func TestBuildAndRunBuildErrorReturnsNil(t *testing.T) {
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  &compileopts.TargetSpec{Libc: "unknown"},
	}
	result, err := buildAndRun("", config, io.Discard, nil, nil, 0, func(cmd *exec.Cmd, result *builder.BuildResult) error {
		t.Fatal("run callback called after a build error")
		return nil
	})
	if err == nil || err.Error() != "unknown libc: unknown" {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result on build error, got %+v", result)
	}
}

func TestFlashRejectsUnsupportedFirmwareExtension(t *testing.T) {
	err := Flash("testdata/print.go", "", "firmware.bin", &compileopts.Options{Target: "pico"})
	want := "target pico does not support firmware output extension .bin; use .elf, .hex, or .uf2"
	if err == nil || err.Error() != want {
		t.Fatalf("Flash() error = %v, want %q", err, want)
	}
}

func TestFlashAcceptsSupportedFirmwareExtension(t *testing.T) {
	err := Flash("github.com/tinygo-org/tinygo/testdata/does-not-exist-output-format", "", "firmware.uf2", &compileopts.Options{Target: "pico", Opt: "z"})
	if err == nil {
		t.Fatal("Flash() succeeded for a missing package")
	}
	if strings.Contains(err.Error(), "does not support firmware output extension") || strings.Contains(err.Error(), "output format") {
		t.Fatalf("Flash() rejected supported UF2 output: %v", err)
	}
}

func TestBuildOutputDirectoryWithFirmwareSuffix(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "output.bin")
	if err := os.Mkdir(outputDir, 0700); err != nil {
		t.Fatal(err)
	}
	config, err := builder.NewConfig(&compileopts.Options{Target: "pico", Opt: "z"})
	if err != nil {
		t.Fatal(err)
	}
	err = Build("does/not/exist", outputDir, config)
	if err == nil || strings.Contains(err.Error(), "does not support firmware output extension") {
		t.Fatalf("Build() error = %v, want a package-loading error rather than an extension error", err)
	}
}

var supportedLinuxArches = map[string]string{
	"AMD64Linux": "linux/amd64",
	"X86Linux":   "linux/386",
	"ARMLinux":   "linux/arm/6",
	"ARM64Linux": "linux/arm64",
	"MIPSLinux":  "linux/mips/hardfloat",
}

var sema = make(chan struct{}, runtime.NumCPU())

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
	_, err = buildAndRun(pkgName, config, stdout, cmdArgs, environmentVars, 2*time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
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
