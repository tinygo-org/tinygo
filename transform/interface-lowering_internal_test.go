package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

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
			p := lowerInterfacesPass{mod: mod}
			if got := p.usesReflectMethods(); got != tc.want {
				t.Errorf("usesReflectMethods() = %v, want %v", got, tc.want)
			}
			if _, got := p.reflectMethodNames(); got != tc.want {
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
		want   []bool
	}{
		{"unused", "", false, false, false, false, []bool{false, false, false}},
		{"invalid name", "", true, false, false, false, []bool{false, false, false}},
		{"constant name", "Keep", true, false, false, false, []bool{true, false, false}},
		{"multiple names", "Keep Other", true, false, false, false, []bool{true, true, false}},
		{"dynamic name", "", true, true, false, false, []bool{true, true, false}},
		{"unexported dispatch", "Keep", true, false, true, false, []bool{true, true, true}},
		{"nested reflection", "Keep", true, false, false, true, []bool{true, true, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("reflect-method-dce")
			defer mod.Dispose()
			builder := ctx.NewBuilder()
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
			builder.SetInsertPointAtEnd(ctx.AddBasicBlock(dead, "entry"))
			builder.CreateRetVoid()
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
			var used bool
			p.methodFuncNames, used = p.reflectMethodNames()
			if used != tc.marked {
				t.Errorf("reachable reflection = %v, want %v", used, tc.marked)
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

			p := lowerInterfacesPass{mod: mod}
			if got := p.usesReflectMakeFunc(); got != tc.want {
				t.Errorf("usesReflectMakeFunc() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPruneUnusedReflectMakeFunc(t *testing.T) {
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

	if err := pruneDeadCodeBeforeInterfaceLowering(mod); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"reflect/makefunc:func:{}{}",
		"internal/reflectlite.makeFuncCall",
	} {
		if !mod.NamedFunction(name).IsNil() {
			t.Errorf("%s was not removed", name)
		}
	}
	const typeName = "reflect/types.type:named:unusedMakeFuncType"
	if !mod.NamedGlobal(typeName).IsNil() {
		t.Errorf("%s was not removed", typeName)
	}
}
