package transform

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

func TestRetainedTypecodePrefix16Bit(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("typecode")
	defer mod.Dispose()
	mod.SetDataLayout("e-p:16:8-i16:8")
	td := llvm.NewTargetData(mod.DataLayout())
	defer td.Dispose()
	i8, i16 := ctx.Int8Type(), ctx.Int16Type()
	ptr := llvm.PointerType(i8, 0)
	field := llvm.ConstInt(i8, 34, false)
	init := ctx.ConstStruct([]llvm.Value{llvm.ConstNull(ptr), field}, false)
	old := llvm.AddGlobal(mod, init.Type(), "descriptor")
	old.SetInitializer(init)
	raw := llvm.AddGlobal(mod, ptr, "raw")
	raw.SetInitializer(old)

	p := lowerInterfacesPass{mod: mod, ctx: ctx, targetData: td, uintptrType: i16}
	kind := p.replaceTypecode(&typeInfo{typecode: old}, []llvm.Value{field})
	storage := mod.NamedGlobal("descriptor")
	if storage.Alignment() < 4 || kind != llvmutil.ConstPtrAdd(i8, storage, llvm.ConstInt(i16, 4, false), false) {
		t.Fatalf("kind pointer is not tag-aligned: %v in %v", kind, storage)
	}
	if raw.Initializer() != llvmutil.ConstPtrAdd(i8, storage, llvm.ConstInt(i16, 2, false), false) {
		t.Fatalf("original base is not two bytes before the kind: %v", raw)
	}
}

func TestOptimizeInterfacesIndirectABIRoots(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("indirect-abi-roots")
	defer mod.Dispose()
	builder := mod.NewBuilder()
	defer builder.Dispose()

	for _, name := range []string{"dead", "retained"} {
		fn := llvm.AddFunction(mod, name, llvm.FunctionType(ctx.VoidType(), nil, false))
		fn.SetLinkage(llvm.InternalLinkage)
		builder.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
		builder.CreateRetVoid()
		llvmutil.AppendToGlobal(mod, indirectABIRoots, fn)
		llvmutil.AppendToGlobal(mod, "llvm.used", fn)
		if name == "retained" {
			llvmutil.AppendToGlobal(mod, "llvm.used", fn)
		}
	}

	config := &compileopts.Config{Options: &compileopts.Options{}}
	if err := optimizeInterfaces(mod, config); err != nil {
		t.Fatal(err)
	}
	if !mod.NamedFunction("dead").IsNil() {
		t.Fatal("function with only a temporary ABI root survived")
	}
	retained := mod.NamedFunction("retained")
	roots := mod.NamedGlobal(indirectABIRoots)
	if retained.IsNil() || roots.IsNil() || roots.Initializer().Operand(0) != retained {
		t.Fatal("temporary ABI root for the independently retained function was lost")
	}
}

func TestUsesReflectMethods(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		constant bool
		want     bool
	}{
		{
			name: "unused marker",
			path: "testdata/reflect-method-unused.ll",
		},
		{
			name: "used marker",
			path: "testdata/reflect-method-used.ll",
			want: true,
		},
		{
			name:     "unused constant-name marker",
			path:     "testdata/reflect-method-unused.ll",
			constant: true,
		},
		{
			name:     "used constant-name marker",
			path:     "testdata/reflect-method-used.ll",
			constant: true,
			want:     true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()

			if tc.constant {
				fn := mod.FirstFunction()
				fn.RemoveStringAttributeAtIndex(-1, "tinygo-reflect-method")
				fn.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", ""))
			}
			if got := analyzeReflection(mod, beforeInterfaceLowering).methods; got != tc.want {
				t.Errorf("usesReflectMethods() = %v, want %v", got, tc.want)
			}
			if got := analyzeReflection(mod, afterInterfaceLowering).methods; got != tc.want {
				t.Errorf("reachable reflection = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReflectMethodFunctionDCE(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		names  string
		marked bool
		all    bool
		direct bool
		nested bool
		root   string
		want   []bool
	}{
		{"unused", "", false, false, false, false, "", []bool{false, false, false}},
		{"invalid name", "", true, false, false, false, "", []bool{false, false, false}},
		{"constant name", "Keep", true, false, false, false, "", []bool{true, false, false}},
		{"multiple names", "Keep Other", true, false, false, false, "", []bool{true, true, false}},
		{"dynamic name", "", true, true, false, false, "", []bool{true, true, false}},
		{"unexported dispatch", "Keep", true, false, true, false, "", []bool{true, true, true}},
		{"nested reflection", "Keep", true, false, false, true, "", []bool{true, true, false}},
		{"temporary ABI root", "Keep", true, false, false, false, "temporary", []bool{true, false, false}},
		{"permanent ABI root", "Keep", true, false, false, false, "permanent", []bool{true, true, false}},
		{"exported ABI root", "Keep", true, false, false, false, "exported", []bool{true, true, false}},
		{"called ABI root", "Keep", true, false, false, false, "direct", []bool{true, true, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("reflect-method-dce")
			defer mod.Dispose()
			builder := mod.NewBuilder()
			defer builder.Dispose()
			uintptrType := ctx.Int64Type()
			ptrType := llvm.PointerType(ctx.Int8Type(), 0)
			p := lowerInterfacesPass{
				ctx:         ctx,
				mod:         mod,
				builder:     builder,
				uintptrType: uintptrType,
				ptrType:     ptrType,
			}

			lookup := llvm.AddFunction(mod, "lookup", llvm.FunctionType(ctx.VoidType(), nil, false))
			if tc.all {
				lookup.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method", ""))
			} else if tc.marked {
				lookup.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", tc.names))
			}
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(lookup, "entry"))
			builder.CreateRetVoid()
			dead := llvm.AddFunction(mod, "dead", lookup.GlobalValueType())
			dead.SetLinkage(llvm.InternalLinkage)
			dead.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method", ""))
			dead.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-method-names", "Other"))
			dead.AddFunctionAttr(ctx.CreateStringAttribute("tinygo-reflect-makefunc", ""))
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(dead, "entry"))
			builder.CreateRetVoid()
			if tc.root != "" {
				llvmutil.AppendToGlobal(mod, "llvm.used", dead)
				llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", dead)
				switch tc.root {
				case "permanent":
					llvmutil.AppendToGlobal(mod, "llvm.used", dead)
				case "exported":
					dead.SetLinkage(llvm.ExternalLinkage)
				case "direct":
					builder.SetInsertPointBefore(lookup.EntryBasicBlock().LastInstruction())
					builder.CreateCall(dead.GlobalValueType(), dead, nil, "")
				}
			}
			var signatures, functions []llvm.Value
			for _, name := range []string{"Keep", "Other", "Example.hidden"} {
				signature := llvm.AddGlobal(mod, ctx.Int8Type(), "reflect/types.signature:"+name+":func:{}{}")
				signature.SetInitializer(llvm.ConstNull(ctx.Int8Type()))
				signatures = append(signatures, signature)
				fn := llvm.AddFunction(mod, name, llvm.FunctionType(ctx.VoidType(), nil, false))
				fn.SetLinkage(llvm.InternalLinkage)
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
				if name == "Example.hidden" || name == "Keep" && tc.nested {
					builder.CreateCall(dead.GlobalValueType(), dead, nil, "")
				}
				builder.CreateRetVoid()
				functions = append(functions, llvm.ConstPtrToInt(fn, uintptrType))
				if tc.direct && name == "Example.hidden" {
					dispatch := llvm.AddGlobal(mod, ptrType, "interface-dispatch")
					dispatch.SetInitializer(fn)
				}
			}
			field := ctx.ConstStruct([]llvm.Value{
				llvm.ConstInt(uintptrType, 3, false),
				llvm.ConstArray(ptrType, signatures),
				llvm.ConstArray(ptrType, signatures),
				llvm.ConstArray(ptrType, signatures),
				llvm.ConstArray(uintptrType, functions),
			}, false)
			descriptor := ctx.ConstStruct([]llvm.Value{field}, false)
			global := llvm.AddGlobal(mod, descriptor.Type(), "reflect/types.type:named:target")
			global.SetInitializer(descriptor)
			restoreRoots := suspendIndirectABIRoots(mod)
			beforeAnalysis := mod.String()
			usage := analyzeReflection(mod, afterInterfaceLowering)
			p.methodFuncNames = usage.methodNames
			if mod.String() != beforeAnalysis {
				t.Error("reflection analysis changed the module")
			}
			if usage.methods != tc.marked {
				t.Errorf("reachable reflection = %v, want %v", usage.methods, tc.marked)
			}
			if got, want := usage.makeFunc, tc.nested || tc.direct || tc.root == "permanent" || tc.root == "exported" || tc.root == "direct"; got != want {
				t.Errorf("reachable MakeFunc = %v, want %v", got, want)
			}
			filtered := p.filterMethodFunctions(field)
			functionArray := builder.CreateExtractValue(filtered, 4, "")
			if fn := builder.CreateExtractValue(functionArray, 2, ""); !fn.IsNull() {
				t.Error("unexported reflected function pointer was retained")
			}
			for i := 0; i < 4; i++ {
				if got, want := builder.CreateExtractValue(filtered, i, ""), builder.CreateExtractValue(field, i, ""); got.C != want.C {
					t.Errorf("method metadata field %d changed", i)
				}
			}
			global.SetInitializer(ctx.ConstStruct([]llvm.Value{filtered}, false))
			po := llvm.NewPassBuilderOptions()
			defer po.Dispose()
			if err := mod.RunPasses("globaldce", llvm.TargetMachine{}, po); err != nil {
				t.Fatal(err)
			}
			restoreRoots()
			for i, name := range []string{"Keep", "Other", "Example.hidden"} {
				if got := !mod.NamedFunction(name).IsNil(); got != tc.want[i] {
					t.Errorf("method %s retained = %v, want %v", name, got, tc.want[i])
				}
			}
		})
	}
}

func TestUsesReflectMakeFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		root string
		want bool
	}{
		{
			name: "unused marker",
			path: "testdata/reflect-makefunc-unused.ll",
		},
		{
			name: "used marker",
			path: "testdata/reflect-makefunc-used.ll",
			want: true,
		},
		{
			name: "exported marker",
			path: "testdata/reflect-makefunc-unused.ll",
			root: "exported",
			want: true,
		},
		{
			name: "temporary ABI root",
			path: "testdata/reflect-makefunc-unused.ll",
			root: "temporary",
		},
		{
			name: "permanent ABI root",
			path: "testdata/reflect-makefunc-unused.ll",
			root: "permanent",
			want: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()

			if tc.root != "" {
				fn := mod.NamedFunction("unused")
				llvmutil.AppendToGlobal(mod, "llvm.used", fn)
				llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", fn)
				switch tc.root {
				case "exported":
					fn.SetLinkage(llvm.ExternalLinkage)
				case "permanent":
					llvmutil.AppendToGlobal(mod, "llvm.used", fn)
				}
			}
			restoreRoots := suspendIndirectABIRoots(mod)
			defer restoreRoots()
			if got := analyzeReflection(mod, afterInterfaceLowering).makeFunc; got != tc.want {
				t.Errorf("usesReflectMakeFunc() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPruneUnusedReflectMakeFunc(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		root string
		want bool
	}{
		{"unused", "", false},
		{"exported marker", "exported", true},
		{"temporary ABI root", "temporary", false},
		{"independent LLVM root", "permanent", true},
		{"interface methodset", "interface", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			buf, err := llvm.NewMemoryBufferFromFile("testdata/reflect-makefunc-unused.ll")
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()
			fn := mod.NamedFunction("unused")
			switch tc.root {
			case "exported":
				fn.SetLinkage(llvm.ExternalLinkage)
			case "temporary", "permanent":
				llvmutil.AppendToGlobal(mod, "llvm.used", fn)
				llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", fn)
				if tc.root == "permanent" {
					llvmutil.AppendToGlobal(mod, "llvm.used", fn)
				}
			case "interface":
				ptrType := llvm.PointerType(ctx.Int8Type(), 0)
				global := llvm.AddGlobal(mod, ptrType, "target$methodset")
				global.SetInitializer(llvm.ConstPointerCast(fn, ptrType))
			}

			restoreRoots := suspendIndirectABIRoots(mod)
			err = pruneDeadCodeBeforeInterfaceLowering(mod)
			restoreRoots()
			if err != nil {
				t.Fatal(err)
			}
			switch tc.root {
			case "temporary":
				if !mod.NamedFunction("unused").IsNil() {
					t.Error("dead indirect-ABI function survived")
				}
			case "permanent":
				fn := mod.NamedFunction("unused")
				if fn.IsNil() {
					t.Fatal("independently rooted indirect-ABI function was removed")
				}
				if count := llvmutil.GlobalReferenceCounts(mod, "llvm.used")[fn]; count != 2 {
					t.Errorf("llvm.used reference count = %d, want 2", count)
				}
				if count := llvmutil.GlobalReferenceCounts(mod, "tinygo.indirect-abi")[fn]; count != 1 {
					t.Errorf("temporary ABI reference count = %d, want 1", count)
				}
			}

			for _, name := range []string{
				"reflect/makefunc:func:{}{}",
				"internal/reflectlite.makeFuncCall",
			} {
				if got := !mod.NamedFunction(name).IsNil(); got != tc.want {
					t.Errorf("%s retained = %v, want %v", name, got, tc.want)
				}
			}
			const typeName = "reflect/types.type:named:unusedMakeFuncType"
			if got := !mod.NamedGlobal(typeName).IsNil(); got != tc.want {
				t.Errorf("%s retained = %v, want %v", typeName, got, tc.want)
			}
		})
	}
}
