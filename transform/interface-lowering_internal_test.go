package transform

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeInterfacesIndirectABIRoots(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("indirect-abi-roots")
	defer mod.Dispose()
	builder := ctx.NewBuilder()
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
		name string
		path string
		want bool
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
			if got := p.usesReflectMethods(); got != tc.want {
				t.Errorf("usesReflectMethods() = %v, want %v", got, tc.want)
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
