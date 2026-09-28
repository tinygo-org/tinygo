package transform_test

import (
	"fmt"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", func(mod llvm.Module) {
		transform.OptimizeAllocs(mod, nil, 256, nil)
	})
}

func TestAllocsAggregateEdges(t *testing.T) {
	t.Parallel()

	const path = "testdata/allocs-aggregate.ll"
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	ensureTestCacheFreshness(t, path)
	buf, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	transform.OptimizeAllocs(mod, nil, 256, nil)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	for name, wantHeap := range map[string]bool{
		"nestedStructReturn":  true,
		"nestedArrayReturn":   true,
		"arrayOfStructReturn": true,
		"structOfArrayReturn": true,
		"singleArrayReturn":   true,
		"aggregateStore":      true,
		"pointerStore":        true,
		"aggregateCall":       true,
		"unknownCall":         true,
		"indirectCall":        true,
		"repackReturn":        true,
		"forwardReturn":       true,
		"forwardLoad":         false,
		"gepReturn":           true,
		"gepLoad":             false,
		"aggregatePhi":        true,
		"aggregateSelect":     true,
		"recursiveReturn":     true,
		"duplicateArguments":  true,
		"nonEscapingScalar":   false,
		"nonEscapingLoad":     false,
		"nonEscapingNilCheck": false,
		"nonEscapingDiscard":  false,
		"nonEscapingCall":     false,
		"pointerFreeStruct":   false,
		"pointerFreeArray":    false,
		"emptyPointerArray":   false,
	} {
		t.Run(name, func(t *testing.T) {
			fn := mod.NamedFunction(name)
			if fn.IsNil() {
				t.Fatal("function not found")
			}
			var heap, stack int
			for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
				for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
					if !inst.IsACallInst().IsNil() && inst.CalledValue() == mod.NamedFunction("runtime.alloc") {
						heap++
					}
					if !inst.IsAAllocaInst().IsNil() {
						stack++
					}
				}
			}
			want := 0
			if wantHeap {
				want = 1
			}
			if heap != want || stack != 1-want {
				t.Errorf("got %d heap and %d stack allocations, want %d and %d:\n%s", heap, stack, want, 1-want, fn.String())
			}
		})
	}
}

func TestAllocsRuntimePhi(t *testing.T) {
	t.Parallel()

	mod := compileGoFileForTesting(t, "../testdata/calls.go")
	defer mod.Context().Dispose()
	defer mod.Dispose()
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	passes := "globalopt,ipsccp,instcombine,adce,function-attrs"
	if llvmutil.Version() >= 18 {
		passes = "globalopt,ipsccp,instcombine<no-verify-fixpoint>,adce,function-attrs"
	}
	if err := mod.RunPasses(passes, llvm.TargetMachine{}, po); err != nil {
		t.Fatal(err)
	}
	fn := mod.NamedFunction("main.phiReturnEscape")
	if fn.IsNil() {
		t.Fatal("phiReturnEscape not found")
	}
	var returnedPhi llvm.Value
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if !inst.IsAReturnInst().IsNil() && !inst.Operand(0).IsAPHINode().IsNil() {
				returnedPhi = inst.Operand(0)
			}
		}
	}
	if returnedPhi.IsNil() {
		t.Fatalf("runtime regression does not return a phi before escape analysis:\n%s", fn.String())
	}
	var extracted bool
	for i := 0; i < returnedPhi.IncomingCount(); i++ {
		if !returnedPhi.IncomingValue(i).IsAExtractValueInst().IsNil() {
			extracted = true
		}
	}
	if !extracted {
		t.Fatalf("returned phi does not merge an extracted aggregate:\n%s", fn.String())
	}

	transform.OptimizeAllocs(mod, nil, 256, nil)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	var allocations int
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if !inst.IsACallInst().IsNil() && inst.CalledValue() == mod.NamedFunction("runtime.alloc") {
				allocations++
			}
		}
	}
	if allocations != 1 {
		t.Fatalf("got %d heap allocations, want 1:\n%s", allocations, fn.String())
	}
}

// Test with a Go file as input (for more accurate tests).
func TestAllocs2(t *testing.T) {
	t.Parallel()

	const (
		basePath   = "testdata/allocs2"
		goFile     = basePath + ".go"
		goldenFile = basePath + ".out"
	)
	mod := compileGoFileForTesting(t, goFile)

	// Run functionattrs pass, which is necessary for escape analysis.
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	err := mod.RunPasses("function(instcombine),function-attrs", llvm.TargetMachine{}, po)
	if err != nil {
		t.Error("failed to run passes:", err)
	}

	// Run heap to stack transform.
	type report struct {
		pos    token.Position
		reason string
	}
	var reports []report
	transform.OptimizeAllocs(mod, regexp.MustCompile("."), 256, func(pos token.Position, reason string) {
		pos.Filename = goFile
		reports = append(reports, report{pos, reason})
	})
	sort.Slice(reports, func(i, j int) bool { return reports[i].pos.Line < reports[j].pos.Line })

	// Load expected test output (the OUT: lines).
	testInput, err := os.ReadFile("./testdata/allocs2.go")
	if err != nil {
		t.Fatal("could not read test input:", err)
	}
	var expectedTestOutput strings.Builder
	for i, line := range strings.Split(strings.ReplaceAll(string(testInput), "\r\n", "\n"), "\n") {
		const prefix = " // OUT: "
		if idx := strings.Index(line, prefix); idx > 0 {
			msg := line[idx+len(prefix):]
			fmt.Fprintf(&expectedTestOutput, "allocs2.go:%d: %s\n", i+1, msg)
		}
	}

	// Check whether the '// OUT' lines in allocs2.go match with the output we
	// got from the test.
	var actualTestOutput strings.Builder
	for _, r := range reports {
		fmt.Fprintf(&actualTestOutput, "allocs2.go:%d: %s\n", r.pos.Line, r.reason)
	}
	if actualTestOutput.String() != expectedTestOutput.String() {
		t.Errorf("expected:\n%s\nactual:\n%s", expectedTestOutput.String(), actualTestOutput.String())
	}

	// Render the cover report and diff it against its golden file.
	var got strings.Builder
	for _, r := range reports {
		if line := transform.FormatAllocCover(r.pos); line != "" {
			got.WriteString(line)
			got.WriteByte('\n')
		}
	}
	checkGolden(t, goldenFile+".cover", got.String())
}
