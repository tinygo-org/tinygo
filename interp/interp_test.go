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

func TestInterpRunFuncBootHooks(t *testing.T) {
	t.Parallel()

	for _, hookName := range []string{"runtime.boardInit1", "runtime.socInit0"} {
		t.Run(hookName, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("board")
			defer mod.Dispose()
			mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
			builder := ctx.NewBuilder()
			defer builder.Dispose()

			i32Type := ctx.Int32Type()
			ptrType := llvm.PointerType(ctx.Int8Type(), 0)
			// RunFunc protects globals within the hook-defining package module.
			// Cross-package mutations are deferred to whole-program interp.Run.
			serialGlobal := llvm.AddGlobal(mod, ptrType, "board.serial")
			serialGlobal.SetInitializer(llvm.ConstNull(ptrType))
			roPtrGlobal := llvm.AddGlobal(mod, ptrType, "board.roPtr")
			roPtrGlobal.SetInitializer(llvm.ConstNull(ptrType))
			uartGlobal := llvm.AddGlobal(mod, i32Type, "board.uart")
			uartGlobal.SetInitializer(llvm.ConstInt(i32Type, 42, false))
			observedGlobal := llvm.AddGlobal(mod, ptrType, "board.observed")
			observedGlobal.SetInitializer(llvm.ConstNull(ptrType))

			hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
			hookFn := llvm.AddFunction(mod, hookName, hookType)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
			builder.CreateStore(uartGlobal, serialGlobal)
			roLoaded := builder.CreateLoad(ptrType, roPtrGlobal, "ro")
			mmioStore := builder.CreateStore(llvm.ConstInt(i32Type, 1, false), roLoaded)
			mmioStore.SetVolatile(true)
			builder.CreateRetVoid()

			initFn := llvm.AddFunction(mod, "board.init", hookType)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
			builder.CreateStore(uartGlobal, roPtrGlobal)
			loaded := builder.CreateLoad(ptrType, serialGlobal, "serial")
			builder.CreateStore(loaded, observedGlobal)
			builder.CreateRetVoid()

			if err := RunFunc(
				initFn,
				time.Second,
				DefaultMaxInterpBlockEntries,
				false,
			); err != nil {
				t.Fatal(err)
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}

			if roInit := roPtrGlobal.Initializer(); roInit != uartGlobal {
				t.Errorf("board.roPtr initializer = %v, want @board.uart", roInit)
			}
			updatedInit := mod.NamedFunction("board.init")
			if updatedInit.IsNil() {
				t.Fatal("missing board.init")
			}
			ir := updatedInit.String()
			if strings.Contains(ir, "@board.roPtr") {
				t.Errorf("board.init did not fold store to read-only @board.roPtr:\n%s", ir)
			}
			if !strings.Contains(ir, "load ptr, ptr @board.serial") {
				t.Errorf("board.init folded away runtime load of @board.serial:\n%s", ir)
			}
			if !strings.Contains(ir, "store ptr %serial, ptr @board.observed") {
				t.Errorf("board.init lost runtime store to @board.observed:\n%s", ir)
			}
		})
	}
}

func TestInterpBootHookICmp(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	arrType := llvm.ArrayType(i32Type, 2)
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	roTable := llvm.AddGlobal(mod, arrType, "board.roTable")
	roTable.SetInitializer(llvm.ConstNull(arrType))
	roDirect := llvm.AddGlobal(mod, i32Type, "board.roDirect")
	roDirect.SetInitializer(llvm.ConstNull(i32Type))
	serialGlobal := llvm.AddGlobal(mod, i32Type, "board.serial")
	serialGlobal.SetInitializer(llvm.ConstNull(i32Type))
	observedGlobal := llvm.AddGlobal(mod, i32Type, "board.observed")
	observedGlobal.SetInitializer(llvm.ConstNull(i32Type))

	hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit0", hookType)
	hookEntry := ctx.AddBasicBlock(hookFn, "entry")
	builder.SetInsertPointAtEnd(hookEntry)

	// Cover both instruction-GEP icmp (at non-constant index) and direct-global icmp.
	idxAlloca := builder.CreateAlloca(i32Type, "idx")
	builder.CreateStore(llvm.ConstInt(i32Type, 1, false), idxAlloca)
	dynIdx := builder.CreateLoad(i32Type, idxAlloca, "dynIdx")
	dynGEP := builder.CreateInBoundsGEP(
		arrType,
		roTable,
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), dynIdx},
		"dynElem",
	)
	_ = builder.CreateICmp(llvm.IntEQ, dynGEP, llvm.ConstNull(ptrType), "gepIsNull")
	_ = builder.CreateLoad(i32Type, dynGEP, "elemVal")
	_ = builder.CreateICmp(llvm.IntEQ, roDirect, llvm.ConstNull(ptrType), "directIsNull")
	_ = builder.CreateLoad(i32Type, roDirect, "directVal")
	builder.CreateStore(llvm.ConstInt(i32Type, 7, false), serialGlobal)
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	wantTable := llvm.ConstArray(
		i32Type,
		[]llvm.Value{llvm.ConstInt(i32Type, 11, false), llvm.ConstInt(i32Type, 22, false)},
	)
	builder.CreateStore(wantTable, roTable)
	builder.CreateStore(llvm.ConstInt(i32Type, 33, false), roDirect)
	loaded := builder.CreateLoad(i32Type, serialGlobal, "serial")
	builder.CreateStore(loaded, observedGlobal)
	builder.CreateRetVoid()

	if err := RunFunc(
		initFn,
		time.Second,
		DefaultMaxInterpBlockEntries,
		false,
	); err != nil {
		t.Fatal(err)
	}
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	if got := roTable.Initializer(); got != wantTable {
		t.Errorf("board.roTable initializer = %v, want %v", got, wantTable)
	}
	if got := roDirect.Initializer(); got != llvm.ConstInt(i32Type, 33, false) {
		t.Errorf("board.roDirect initializer = %v, want 33", got)
	}
	ir := mod.NamedFunction("board.init").String()
	if strings.Contains(ir, "@board.roTable") || strings.Contains(ir, "@board.roDirect") {
		t.Errorf("board.init did not fold stores to read-only globals:\n%s", ir)
	}
	if !strings.Contains(ir, "load i32, ptr @board.serial") {
		t.Errorf("board.init folded away runtime load of @board.serial:\n%s", ir)
	}
}
