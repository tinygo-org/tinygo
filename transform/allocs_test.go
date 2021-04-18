package transform_test

import (
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/loader"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", func(mod llvm.Module) {
		transform.OptimizeAllocs(mod, nil, nil)
	})
}

type allocsTestOutput struct {
	filename string
	line     int
	msg      string
}

func (out allocsTestOutput) String() string {
	return out.filename + ":" + strconv.Itoa(out.line) + ": " + out.msg
}

// Test with a Go file as input (for more accurate tests).
func TestAllocs2(t *testing.T) {
	t.Parallel()

	target, err := compileopts.LoadTarget("i686--linux")
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  target,
	}
	compilerConfig := &compiler.Config{
		Triple:             config.Triple(),
		GOOS:               config.GOOS(),
		GOARCH:             config.GOARCH(),
		CodeModel:          config.CodeModel(),
		RelocationModel:    config.RelocationModel(),
		Scheduler:          config.Scheduler(),
		FuncImplementation: config.FuncImplementation(),
		AutomaticStackSize: config.AutomaticStackSize(),
		Debug:              true,
	}
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, []string{"./testdata/allocs2.go"}, config.ClangHeaders, types.Config{
		Sizes: compiler.Sizes(machine),
	})
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	err = lprogram.Parse()
	if err != nil {
		t.Fatal("could not parse", err)
	}

	// Compile AST to IR.
	program := lprogram.LoadSSA()
	pkg := lprogram.MainPkg()
	mod, errs := compiler.CompilePackage("allocs2.go", pkg, program.Package(pkg.Pkg), machine, compilerConfig, false)
	if errs != nil {
		for _, err := range errs {
			t.Error(err)
		}
		return
	}

	// Run functionattrs pass, which is necessary for escape analysis.
	pm := llvm.NewPassManager()
	defer pm.Dispose()
	pm.AddInstructionCombiningPass()
	pm.AddFunctionAttrsPass()
	pm.Run(mod)

	// Run heap to stack transform.
	var testOutputs []allocsTestOutput
	transform.OptimizeAllocs(mod, regexp.MustCompile("."), func(pos token.Position, msg string) {
		testOutputs = append(testOutputs, allocsTestOutput{
			filename: filepath.Base(pos.Filename),
			line:     pos.Line,
			msg:      msg,
		})
	})
	sort.Slice(testOutputs, func(i, j int) bool {
		return testOutputs[i].line < testOutputs[j].line
	})
	testOutput := ""
	for _, out := range testOutputs {
		testOutput += out.String() + "\n"
	}

	// Load expected test output (the OUT: lines).
	testInput, err := ioutil.ReadFile("./testdata/allocs2.go")
	if err != nil {
		t.Fatal("could not read test input:", err)
	}
	var expectedTestOutput string
	for i, line := range strings.Split(strings.ReplaceAll(string(testInput), "\r\n", "\n"), "\n") {
		if idx := strings.Index(line, " // OUT: "); idx > 0 {
			msg := line[idx+len(" // OUT: "):]
			expectedTestOutput += "allocs2.go:" + strconv.Itoa(i+1) + ": " + msg + "\n"
		}
	}

	if testOutput != expectedTestOutput {
		t.Errorf("output does not match expected output:\n%s", testOutput)
	}
}
