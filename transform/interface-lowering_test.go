package transform_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestInterfaceLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/interface", func(mod llvm.Module) {
		err := transform.LowerInterfaces(mod, defaultTestConfig)
		if err != nil {
			t.Error(err)
		}

		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		err = mod.RunPasses("globaldce", llvm.TargetMachine{}, po)
		if err != nil {
			t.Error("failed to run passes:", err)
		}
	})
}

// TestInterfaceLoweringBoxedTypes checks which types an interface invoke
// thunk dispatches to. See https://github.com/tinygo-org/tinygo/issues/6168.
func TestInterfaceLoweringBoxedTypes(t *testing.T) {
	t.Parallel()

	const header = `
target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 2, ptr null }, align 4
@"reflect/types.type:basic:float64" = linkonce_odr constant { i8, ptr } { i8 78, ptr null }, align 4
@"reflect/types.type:func:{basic:float64}{}" = linkonce_odr constant { i8, i8, i16, ptr, [1 x ptr] } { i8 83, i8 1, i16 0, ptr null, [1 x ptr] [ptr @"reflect/types.type:basic:float64"] }, align 4
@"reflect/types.typeid:named:Number" = external constant i8
@"reflect/types.typeid:basic:float64" = external constant i8
@"reflect/methods.Double() int" = linkonce_odr constant i8 0
@"Number$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Double() int"], { ptr } { ptr @"(Number).Double$invoke" } }
@"reflect/types.type:named:Number" = linkonce_odr constant { ptr, i8, ptr, ptr } { ptr @"Number$methodset", i8 34, ptr null, ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:chan:named:Number" = linkonce_odr constant { i8, ptr, ptr } { i8 18, ptr null, ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1) }, align 4

declare void @runtime.nilPanic(ptr)
declare i1 @runtime.typeAssert(ptr, ptr, ptr)

define i32 @"(Number).Double$invoke"(ptr %receiver, ptr %context) {
  %n = ptrtoint ptr %receiver to i32
  %ret = mul i32 %n, 2
  ret i32 %ret
}

declare i32 @"Doubler.Double$invoke"(ptr %receiver, ptr %typecode, ptr %context) #0

define i32 @callDouble(ptr %typecode, ptr %value) {
  %ret = call i32 @"Doubler.Double$invoke"(ptr %value, ptr %typecode, ptr undef)
  ret i32 %ret
}

define i1 @assertNumber(ptr %typecode) {
  %ok = call i1 @runtime.typeAssert(ptr %typecode, ptr @"reflect/types.typeid:named:Number", ptr undef)
  ret i1 %ok
}

define i1 @assertFloat(ptr %typecode) {
  %ok = call i1 @runtime.typeAssert(ptr %typecode, ptr @"reflect/types.typeid:basic:float64", ptr undef)
  ret i1 %ok
}

attributes #0 = { "tinygo-invoke"="reflect/methods.Double() int" "tinygo-methods"="reflect/methods.Double() int" }
`

	const methodMetadata = `
@"reflect/methods.Unused(float64)" = linkonce_odr constant i8 0
@"Metadata$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Unused(float64)"], { ptr } { ptr @"Metadata.Unused$invoke" } }
@"reflect/types.type:struct:metadata" = linkonce_odr constant { ptr, i8, i16, { i32, [1 x ptr], [1 x ptr], [1 x ptr] } } { ptr @"Metadata$methodset", i8 25, i16 -32768, { i32, [1 x ptr], [1 x ptr], [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/methods.Unused(float64)"], [1 x ptr] zeroinitializer, [1 x ptr] [ptr @"reflect/types.type:func:{basic:float64}{}"] } }, align 4

define void @"Metadata.Unused$invoke"(ptr %receiver, double %arg, ptr %context) {
  ret void
}

define ptr @methodMetadata() #1 {
  ret ptr @"reflect/types.type:struct:metadata"
}

declare { ptr, ptr } @runtime.composeInterface(ptr, ptr)

define { ptr, ptr } @compose(ptr %typecode, ptr %value) {
  %ret = call { ptr, ptr } @runtime.composeInterface(ptr %typecode, ptr %value)
  ret { ptr, ptr } %ret
}
`

	tests := []struct {
		name        string
		ir          string
		dispatch    bool
		assertFloat bool
	}{{
		name: "boxed directly",
		ir: `
define ptr @box() {
  ret ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1)
}
`,
		dispatch: true,
	}, {
		name: "only element of boxed chan",
		ir: `
define ptr @box() {
  ret ptr @"reflect/types.type:chan:named:Number"
}
`,
		dispatch: false,
	}, {
		name: "only parameter of reflected function",
		ir: `
define ptr @box() {
  ret ptr @"reflect/types.type:func:{basic:float64}{}"
}
`,
	}, {
		name: "boxed float directly",
		ir: `
define ptr @box() {
  ret ptr @"reflect/types.type:basic:float64"
}
`,
		assertFloat: true,
	}, {
		name: "element of boxed chan with composeInterface",
		ir: `
declare { ptr, ptr } @runtime.composeInterface(ptr, ptr)

define ptr @box() {
  ret ptr @"reflect/types.type:chan:named:Number"
}

define { ptr, ptr } @compose(ptr %typecode, ptr %value) {
  %ret = call { ptr, ptr } @runtime.composeInterface(ptr %typecode, ptr %value)
  ret { ptr, ptr } %ret
}
`,
		dispatch:    true,
		assertFloat: true,
	}, {
		name: "unused method signature with composeInterface",
		ir:   methodMetadata + `attributes #1 = { nounwind }`,
	}, {
		name:        "method lookup retains signature",
		ir:          methodMetadata + `attributes #1 = { "tinygo-reflect-method" }`,
		dispatch:    true,
		assertFloat: true,
	}, {
		name: "boxed in global",
		ir: `
@iface = global { ptr, ptr } { ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1), ptr inttoptr (i32 3 to ptr) }
`,
		dispatch: true,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			path := filepath.Join(t.TempDir(), "in.ll")
			if err := os.WriteFile(path, []byte(header+tc.ir), 0o666); err != nil {
				t.Fatal(err)
			}
			buf, err := llvm.NewMemoryBufferFromFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatalf("could not load module:\n%v", err)
			}
			defer mod.Dispose()

			if err := transform.LowerInterfaces(mod, defaultTestConfig); err != nil {
				t.Fatal(err)
			}
			if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
				t.Fatal("IR verification failed")
			}

			thunk := mod.NamedFunction("Doubler.Double$invoke").String()
			got := strings.Contains(thunk, `@"(Number).Double$invoke"`)
			if got != tc.dispatch {
				t.Errorf("thunk calls (Number).Double: got %v, want %v\n%s", got, tc.dispatch, thunk)
			}
			for _, assertion := range []struct {
				name string
				want bool
			}{
				{"assertNumber", tc.dispatch},
				{"assertFloat", tc.assertFloat},
			} {
				body := mod.NamedFunction(assertion.name).String()
				if got := strings.Contains(body, "icmp eq"); got != assertion.want {
					t.Errorf("%s keeps type comparison: got %v, want %v\n%s", assertion.name, got, assertion.want, body)
				}
				if !assertion.want && !strings.Contains(body, "ret i1 false") {
					t.Errorf("%s did not fold to false:\n%s", assertion.name, body)
				}
			}
		})
	}
}
