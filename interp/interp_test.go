package interp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/internal/ir"
	"tinygo.org/x/go-llvm"
)

func TestInterp(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.out.ll")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no test cases found in testdata")
	}
	for _, match := range matches {
		name := strings.TrimSuffix(filepath.Base(match), ".out.ll")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTest(t, "testdata/"+name)
		})
	}
}

func TestInterpFunctionUsageAttributes(t *testing.T) {
	for _, wholeProgram := range []bool{false, true} {
		name := "package"
		if wholeProgram {
			name = "program"
		}
		t.Run(name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("test")
			defer mod.Dispose()
			mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
			builder := ctx.NewBuilder()
			defer builder.Dispose()
			typ := llvm.FunctionType(ctx.VoidType(), nil, false)
			helper := llvm.AddFunction(mod, "helper", typ)
			helper.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Alpha Beta"))
			helper.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-structof", ""))
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(helper, "entry"))
			builder.CreateRetVoid()
			dead := llvm.AddFunction(mod, "dead", typ)
			dead.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Unused"))
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(dead, "entry"))
			builder.CreateRetVoid()
			init := llvm.AddFunction(mod, "example.init", typ)
			init.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method", ""))
			init.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Zeta Alpha"))
			init.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-makefunc", ""))
			init.AddFunctionAttr(ctx.CreateStringAttribute("unrelated", ""))
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(init, "entry"))
			builder.CreateCall(typ, helper, nil, "")
			builder.CreateRetVoid()
			resultName := init.Name()
			if wholeProgram {
				initAll := llvm.AddFunction(mod, "runtime.initAll", typ)
				initAll.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Gamma"))
				other := llvm.AddFunction(mod, "other.init", typ)
				other.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Delta Alpha"))
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(other, "entry"))
				builder.CreateRetVoid()
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initAll, "entry"))
				builder.CreateCall(typ, init, nil, "")
				builder.CreateCall(typ, other, nil, "")
				builder.CreateRetVoid()
				resultName = initAll.Name()
				if err := Run(mod, time.Second, DefaultMaxInterpBlockEntries, false); err != nil {
					t.Fatal(err)
				}
			} else if err := RunFunc(init, time.Second, DefaultMaxInterpBlockEntries, false); err != nil {
				t.Fatal(err)
			}
			result := mod.NamedFunction(resultName)
			methodNames := "Alpha Beta Zeta"
			if wholeProgram {
				methodNames = "Alpha Beta Delta Gamma Zeta"
			}
			for kind, want := range map[string]string{
				"tinygo-reflect-method":       "",
				"tinygo-reflect-method-names": methodNames,
				"tinygo-reflect-makefunc":     "",
				"tinygo-reflect-structof":     "",
			} {
				attr := result.GetStringAttributeAtIndex(-1, kind)
				if attr.IsNil() {
					t.Errorf("lost %s", kind)
				} else if got := attr.GetStringValue(); got != want {
					t.Errorf("%s = %q, want %q", kind, got, want)
				}
			}
			if !result.GetStringAttributeAtIndex(-1, "unrelated").IsNil() {
				t.Error("copied an unrelated attribute")
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func runTest(t *testing.T, pathPrefix string) {
	// Read the input IR.
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile(pathPrefix + ".ll")
	if err != nil {
		t.Fatalf("could not read file %s: %v", pathPrefix+".ll", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}
	defer mod.Dispose()

	// Perform the transform.
	err = Run(mod, 10*time.Minute, DefaultMaxInterpBlockEntries, false)
	if err != nil {
		if err, match := err.(*Error); match {
			println(err.Error())
			if len(err.Inst) != 0 {
				println(err.Inst)
			}
			if len(err.Traceback) > 0 {
				println("\ntraceback:")
				for _, line := range err.Traceback {
					println(line.Pos.String() + ":")
					println(line.Inst)
				}
			}
		}
		t.Fatal(err)
	}

	// To be sure, verify that the module is still valid.
	if llvm.VerifyModule(mod, llvm.PrintMessageAction) != nil {
		t.FailNow()
	}

	// Run some cleanup passes to get easy-to-read outputs.
	to := llvm.NewPassBuilderOptions()
	defer to.Dispose()
	mod.RunPasses("globalopt,dse,adce", llvm.TargetMachine{}, to)

	// Read the expected output IR.
	out, err := os.ReadFile(pathPrefix + ".out.ll")
	if err != nil {
		t.Fatalf("could not read output file %s: %v", pathPrefix+".out.ll", err)
	}

	// See whether the transform output matches with the expected output IR.
	expected := string(out)
	actual := mod.String()
	if diff := ir.Diff(expected, actual); diff != "" {
		t.Errorf("output does not match expected output:\n%s", diff)
	}
}

// TestInterpLargeArray checks that a large global written by an init function
// is serialized in linear time. It used to take about half an hour.
func TestInterpLargeArray(t *testing.T) {
	t.Parallel()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/largearray.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	start := time.Now()
	if err := Run(mod, 10*time.Minute, DefaultMaxInterpBlockEntries, false); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 30*time.Second {
		t.Errorf("interp took %v", d)
	}

	init := mod.NamedGlobal("lut").Initializer()
	plane := init.Operand(0)
	if !plane.IsConstantString() {
		t.Fatal("first plane is not a byte array")
	}
	got := plane.ConstGetAsString()
	if len(got) != 0x110000 || got[0] != 1 || got[767] != 2 || strings.Trim(got[1:767]+got[768:], "\x00") != "" {
		t.Error("first plane holds the wrong bytes")
	}
	if !init.Operand(1).IsNull() {
		t.Error("second plane is not zero")
	}
}

// TestInterpPartialPointer checks that a global which holds part of a pointer
// is reported as an error instead of being written out as zero.
func TestInterpPartialPointer(t *testing.T) {
	t.Parallel()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/partialpointer.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	if err := Run(mod, 10*time.Minute, DefaultMaxInterpBlockEntries, false); err != errUnsupportedInst {
		t.Errorf("expected %v, got %v", errUnsupportedInst, err)
	}
}

// TestInterpRollbackReason checks that the error that makes interp roll back
// names its cause and the function being interpreted.
func TestInterpRollbackReason(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "rollback.ll")
	ir := `target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

define i64 @fill(i64 %n) {
entry:
  br label %loop
loop:
  %i = phi i64 [ 0, %entry ], [ %next, %loop ]
  %next = add i64 %i, 1
  %done = icmp eq i64 %next, %n
  br i1 %done, label %exit, label %loop
exit:
  ret i64 %next
}
`
	if err := os.WriteFile(path, []byte(ir), 0o666); err != nil {
		t.Fatal(err)
	}
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	r := newRunner(mod, time.Minute, 10, false)
	defer r.dispose()
	_, _, runErr := r.run(r.getFunction(mod.NamedFunction("fill")), []value{literalValue{uint64(100)}}, nil, nil, 0)
	if runErr == nil {
		t.Fatal("expected the loop limit to stop interpretation")
	}
	if !isRecoverableError(runErr.Err) {
		t.Errorf("expected a recoverable error, got %v", runErr.Err)
	}
	if want := ": interp: loop ran too many iterations (in fill)"; !strings.HasSuffix(runErr.Error(), want) {
		t.Errorf("expected an error ending in %q, got %q", want, runErr.Error())
	}
}
