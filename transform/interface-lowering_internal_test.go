package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

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
