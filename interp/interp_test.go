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

func TestInterpBootHookConstantCallback(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	i64Type := ctx.Int64Type()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	roPtrGlobal := llvm.AddGlobal(mod, ptrType, "board.roPtr")
	roPtrGlobal.SetInitializer(llvm.ConstNull(ptrType))
	uartGlobal := llvm.AddGlobal(mod, i32Type, "board.uart")
	uartGlobal.SetInitializer(llvm.ConstInt(i32Type, 42, false))
	writtenGlobal := llvm.AddGlobal(mod, i32Type, "board.written")
	writtenGlobal.SetInitializer(llvm.ConstNull(i32Type))
	observedGlobal := llvm.AddGlobal(mod, i32Type, "board.observed")
	observedGlobal.SetInitializer(llvm.ConstNull(i32Type))

	cbType := llvm.FunctionType(ctx.VoidType(), nil, false)
	cbFn := llvm.AddFunction(mod, "board.callback", cbType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(cbFn, "entry"))
	roLoaded := builder.CreateLoad(ptrType, roPtrGlobal, "ro")
	mmioStore := builder.CreateStore(llvm.ConstInt(i32Type, 1, false), roLoaded)
	mmioStore.SetVolatile(true)
	builder.CreateStore(llvm.ConstInt(i32Type, 9, false), writtenGlobal)
	builder.CreateRetVoid()

	descType := ctx.StructType([]llvm.Type{i64Type}, false)
	descGlobal := llvm.AddGlobal(mod, descType, "board.descriptor")
	descGlobal.SetGlobalConstant(true)
	descGlobal.SetInitializer(ctx.ConstStruct([]llvm.Value{
		llvm.ConstPtrToInt(cbFn, i64Type),
	}, false))

	sinkType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	sinkFn := llvm.AddFunction(mod, "board.sink", sinkType)

	hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit0", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	builder.CreateCall(sinkType, sinkFn, []llvm.Value{descGlobal}, "")
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	builder.CreateStore(uartGlobal, roPtrGlobal)
	loaded := builder.CreateLoad(i32Type, writtenGlobal, "written")
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
	ir := mod.NamedFunction("board.init").String()
	if strings.Contains(ir, "@board.roPtr") {
		t.Errorf("board.init did not fold store to read-only @board.roPtr:\n%s", ir)
	}
	if !strings.Contains(ir, "load i32, ptr @board.written") {
		t.Errorf("board.init folded away runtime load of @board.written:\n%s", ir)
	}
}

func TestInterpBootHookWriteThroughLoadedPtr(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	objType := ctx.StructType([]llvm.Type{i32Type, i32Type}, false)

	pkgGlobal := llvm.AddGlobal(mod, ptrType, "board.pkgGlobal")
	pkgGlobal.SetInitializer(llvm.ConstNull(ptrType))
	objGlobal := llvm.AddGlobal(mod, objType, "board.obj")
	objGlobal.SetInitializer(llvm.ConstNull(objType))
	observedGlobal := llvm.AddGlobal(mod, ptrType, "board.observed")
	observedGlobal.SetInitializer(llvm.ConstNull(ptrType))

	hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit1", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	loadedPtr := builder.CreateLoad(ptrType, pkgGlobal, "p")
	fieldPtr := builder.CreateInBoundsGEP(
		objType,
		loadedPtr,
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), llvm.ConstInt(i32Type, 1, false)},
		"field",
	)
	builder.CreateStore(llvm.ConstInt(i32Type, 99, false), fieldPtr)
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	builder.CreateStore(objGlobal, pkgGlobal)
	readPtr := builder.CreateLoad(ptrType, pkgGlobal, "p")
	builder.CreateStore(readPtr, observedGlobal)
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

	ir := mod.NamedFunction("board.init").String()
	if !strings.Contains(ir, "store ptr @board.obj, ptr @board.pkgGlobal") {
		t.Errorf("board.init lost runtime store to @board.pkgGlobal:\n%s", ir)
	}
	if !strings.Contains(ir, "load ptr, ptr @board.pkgGlobal") {
		t.Errorf("board.init folded away runtime load of @board.pkgGlobal:\n%s", ir)
	}
}

func TestInterpBootHookRejectsIndirectAndInterfaceCalls(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		iface   bool
		wantErr string
	}{
		{name: "indirect call", iface: false, wantErr: "indirect call is not allowed in boot hook"},
		{name: "interface call", iface: true, wantErr: "interface call is not allowed in boot hook"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("board")
			defer mod.Dispose()
			mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
			builder := ctx.NewBuilder()
			defer builder.Dispose()

			ptrType := llvm.PointerType(ctx.Int8Type(), 0)
			hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
			cbType := llvm.FunctionType(ctx.VoidType(), nil, false)

			hookFn := llvm.AddFunction(mod, "runtime.boardInit1", hookType)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
			if tc.iface {
				ifaceFn := llvm.AddFunction(mod, "interface:{Init:func()}", cbType)
				builder.CreateCall(cbType, ifaceFn, nil, "")
			} else {
				builder.CreateCall(cbType, hookFn.Param(0), nil, "")
			}
			builder.CreateRetVoid()

			initFn := llvm.AddFunction(mod, "board.init", hookType)
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
			builder.CreateRetVoid()

			err := RunFunc(
				initFn,
				time.Second,
				DefaultMaxInterpBlockEntries,
				false,
			)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got error %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestInterpBootHookVectorConstant(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	vecType := llvm.VectorType(i32Type, 2)
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	vecGlobal := llvm.AddGlobal(mod, vecType, "board.vec")
	vecGlobal.SetGlobalConstant(true)
	vecGlobal.SetInitializer(llvm.ConstVector([]llvm.Value{
		llvm.ConstInt(i32Type, 1, false),
		llvm.ConstInt(i32Type, 2, false),
	}, false))

	sinkType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	sinkFn := llvm.AddFunction(mod, "board.sink", sinkType)

	hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit0", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	builder.CreateCall(sinkType, sinkFn, []llvm.Value{vecGlobal}, "")
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	builder.CreateRetVoid()

	if err := RunFunc(
		initFn,
		time.Second,
		DefaultMaxInterpBlockEntries,
		false,
	); err != nil {
		t.Fatal(err)
	}
}

func TestInterpBootHookVolatileGEPStore(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	arrType := llvm.ArrayType(i32Type, 4)
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	bufGlobal := llvm.AddGlobal(mod, arrType, "board.buf")
	bufGlobal.SetInitializer(llvm.ConstNull(arrType))
	observedGlobal := llvm.AddGlobal(mod, i32Type, "board.observed")
	observedGlobal.SetInitializer(llvm.ConstNull(i32Type))

	hookType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit1", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	idxAlloca := builder.CreateAlloca(i32Type, "idx")
	builder.CreateStore(llvm.ConstInt(i32Type, 1, false), idxAlloca)
	dynIdx := builder.CreateLoad(i32Type, idxAlloca, "dynIdx")
	elemPtr := builder.CreateInBoundsGEP(
		arrType,
		bufGlobal,
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), dynIdx},
		"elem",
	)
	volStore := builder.CreateStore(llvm.ConstInt(i32Type, 1, false), elemPtr)
	volStore.SetVolatile(true)
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	loaded := builder.CreateLoad(i32Type, bufGlobal, "buf0")
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

	ir := mod.NamedFunction("board.init").String()
	if !strings.Contains(ir, "load i32, ptr @board.buf") {
		t.Errorf("board.init folded away runtime load of @board.buf:\n%s", ir)
	}
}

func TestInterpBootHookSkipsRuntimeAllocAndPanicBodies(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i32Type := ctx.Int32Type()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	voidType := ctx.VoidType()
	cbType := llvm.FunctionType(voidType, nil, false)

	// Simulate runtime.alloc body reaching an indirect call (e.g. finalizer).
	allocFnType := llvm.FunctionType(ptrType, []llvm.Type{i32Type, ptrType, ptrType}, false)
	allocFn := llvm.AddFunction(mod, "runtime.alloc", allocFnType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(allocFn, "entry"))
	builder.CreateCall(cbType, allocFn.Param(1), nil, "")
	builder.CreateRet(llvm.ConstNull(ptrType))

	// Simulate runtime.slicePanic and runtime._panic calling an unlowered interface method.
	ifaceFn := llvm.AddFunction(mod, "interface:{Error:func()}", cbType)
	for _, panicName := range []string{"runtime.slicePanic", "runtime._panic"} {
		panicFn := llvm.AddFunction(mod, panicName, cbType)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(panicFn, "entry"))
		builder.CreateCall(cbType, ifaceFn, nil, "")
		builder.CreateRetVoid()
	}

	hookType := llvm.FunctionType(voidType, []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit1", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	builder.CreateCall(
		allocFnType,
		allocFn,
		[]llvm.Value{
			llvm.ConstInt(i32Type, 8, false),
			llvm.ConstNull(ptrType),
			llvm.ConstNull(ptrType),
		},
		"",
	)
	builder.CreateCall(cbType, mod.NamedFunction("runtime.slicePanic"), nil, "")
	builder.CreateCall(cbType, mod.NamedFunction("runtime._panic"), nil, "")
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	builder.CreateRetVoid()

	if err := RunFunc(
		initFn,
		time.Second,
		DefaultMaxInterpBlockEntries,
		false,
	); err != nil {
		t.Fatal(err)
	}
}

func TestInterpBootHookLoadedPtrThroughHelperCall(t *testing.T) {
	t.Parallel()

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("board")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:64:64-i64:64-n8:16:32:64")
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	i8Type := ctx.Int8Type()
	i32Type := ctx.Int32Type()
	ptrType := llvm.PointerType(i8Type, 0)
	voidType := ctx.VoidType()
	pairType := ctx.StructType([]llvm.Type{ptrType, i8Type}, false)
	objType := ctx.StructType([]llvm.Type{i32Type, i32Type}, false)

	mmioPtrGlobal := llvm.AddGlobal(mod, ptrType, "board.mmioPtr")
	mmioPtrGlobal.SetInitializer(llvm.ConstNull(ptrType))
	mutPtrGlobal := llvm.AddGlobal(mod, ptrType, "board.mutPtr")
	mutPtrGlobal.SetInitializer(llvm.ConstNull(ptrType))
	localMutPtrGlobal := llvm.AddGlobal(mod, ptrType, "board.localMutPtr")
	localMutPtrGlobal.SetInitializer(llvm.ConstNull(ptrType))

	uartGlobal := llvm.AddGlobal(mod, objType, "board.uart")
	uartGlobal.SetInitializer(llvm.ConstNull(objType))
	mutObjGlobal := llvm.AddGlobal(mod, objType, "board.mutObj")
	mutObjGlobal.SetInitializer(llvm.ConstNull(objType))
	localMutObjGlobal := llvm.AddGlobal(mod, objType, "board.localMutObj")
	localMutObjGlobal.SetInitializer(llvm.ConstNull(objType))
	observedGlobal := llvm.AddGlobal(mod, ptrType, "board.observed")
	observedGlobal.SetInitializer(llvm.ConstNull(ptrType))

	trackType := llvm.FunctionType(voidType, []llvm.Type{ptrType, ptrType}, false)
	trackFn := llvm.AddFunction(mod, "runtime.trackPointer", trackType)

	// Helper stores a pointer into a local alloca {ptr, i8} and returns it by value.
	wrapType := llvm.FunctionType(pairType, []llvm.Type{ptrType}, false)
	wrapFn := llvm.AddFunction(mod, "board.wrapMMIO", wrapType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(wrapFn, "entry"))
	pairAlloca := builder.CreateAlloca(pairType, "pair")
	builder.CreateStore(llvm.ConstNull(pairType), pairAlloca)
	f0 := builder.CreateInBoundsGEP(
		pairType,
		pairAlloca,
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), llvm.ConstInt(i32Type, 0, false)},
		"f0",
	)
	builder.CreateStore(wrapFn.Param(0), f0)
	loadedPair := builder.CreateLoad(pairType, pairAlloca, "retPair")
	builder.CreateRet(loadedPair)

	// Helper extracts the pointer from {ptr, i8} and performs a volatile store.
	writeMMIOType := llvm.FunctionType(voidType, []llvm.Type{pairType}, false)
	writeMMIOFn := llvm.AddFunction(mod, "board.writeMMIO", writeMMIOType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(writeMMIOFn, "entry"))
	extracted := builder.CreateExtractValue(writeMMIOFn.Param(0), 0, "reg")
	volStore := builder.CreateStore(llvm.ConstInt(i32Type, 1, false), extracted)
	volStore.SetVolatile(true)
	builder.CreateRetVoid()

	// Helper performs a non-volatile store through its pointer parameter.
	mutateType := llvm.FunctionType(voidType, []llvm.Type{ptrType}, false)
	mutateFn := llvm.AddFunction(mod, "board.mutateObj", mutateType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(mutateFn, "entry"))
	fieldPtr := builder.CreateInBoundsGEP(
		objType,
		mutateFn.Param(0),
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), llvm.ConstInt(i32Type, 1, false)},
		"field",
	)
	builder.CreateStore(llvm.ConstInt(i32Type, 7, false), fieldPtr)
	builder.CreateRetVoid()

	hookType := llvm.FunctionType(voidType, []llvm.Type{ptrType}, false)
	hookFn := llvm.AddFunction(mod, "runtime.boardInit1", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(hookFn, "entry"))
	mmioLoaded := builder.CreateLoad(ptrType, mmioPtrGlobal, "mmio")
	builder.CreateCall(trackType, trackFn, []llvm.Value{mmioLoaded, llvm.ConstNull(ptrType)}, "")
	wrapped := builder.CreateCall(wrapType, wrapFn, []llvm.Value{mmioLoaded}, "wrapped")
	builder.CreateCall(writeMMIOType, writeMMIOFn, []llvm.Value{wrapped}, "")

	mutLoaded := builder.CreateLoad(ptrType, mutPtrGlobal, "mut")
	builder.CreateCall(mutateType, mutateFn, []llvm.Value{mutLoaded}, "")

	localMutLoaded := builder.CreateLoad(ptrType, localMutPtrGlobal, "localMut")
	localSlot := builder.CreateAlloca(ptrType, "slot")
	builder.CreateStore(localMutLoaded, localSlot)
	reloaded := builder.CreateLoad(ptrType, localSlot, "reloaded")
	localField := builder.CreateInBoundsGEP(
		objType,
		reloaded,
		[]llvm.Value{llvm.ConstInt(i32Type, 0, false), llvm.ConstInt(i32Type, 1, false)},
		"localField",
	)
	builder.CreateStore(llvm.ConstInt(i32Type, 9, false), localField)
	builder.CreateRetVoid()

	initFn := llvm.AddFunction(mod, "board.init", hookType)
	builder.SetInsertPointAtEnd(ctx.AddBasicBlock(initFn, "entry"))
	builder.CreateStore(uartGlobal, mmioPtrGlobal)
	builder.CreateStore(mutObjGlobal, mutPtrGlobal)
	builder.CreateStore(localMutObjGlobal, localMutPtrGlobal)
	r1 := builder.CreateLoad(ptrType, mutPtrGlobal, "r1")
	builder.CreateStore(r1, observedGlobal)
	r2 := builder.CreateLoad(ptrType, localMutPtrGlobal, "r2")
	builder.CreateStore(r2, observedGlobal)
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

	if got := mmioPtrGlobal.Initializer(); got != uartGlobal {
		t.Errorf("board.mmioPtr initializer = %v, want @board.uart", got)
	}
	ir := mod.NamedFunction("board.init").String()
	if strings.Contains(ir, "@board.mmioPtr") {
		t.Errorf("board.init did not fold store to read-only @board.mmioPtr:\n%s", ir)
	}
	if !strings.Contains(ir, "load ptr, ptr @board.mutPtr") {
		t.Errorf("board.init folded away runtime load of @board.mutPtr:\n%s", ir)
	}
	if !strings.Contains(ir, "load ptr, ptr @board.localMutPtr") {
		t.Errorf("board.init folded away runtime load of @board.localMutPtr:\n%s", ir)
	}
}
