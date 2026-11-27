package transform_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestInterfaceLowering(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, passes string }{
		{"interface", "globaldce"},
		{"interface-offsets", "function(instcombine),globaldce"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testTransform(t, "testdata/"+tc.name, func(mod llvm.Module) {
				if err := transform.LowerInterfaces(mod, defaultTestConfig); err != nil {
					t.Fatal(err)
				}
				po := llvm.NewPassBuilderOptions()
				defer po.Dispose()
				if err := mod.RunPasses(tc.passes, llvm.TargetMachine{}, po); err != nil {
					t.Fatal(err)
				}
			})
		})
	}
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
@"reflect/types.type:struct:metadata" = linkonce_odr constant { ptr, i8, i16, { i32, [1 x ptr], [1 x ptr], [1 x ptr], [1 x i32] } } { ptr @"Metadata$methodset", i8 25, i16 -32768, { i32, [1 x ptr], [1 x ptr], [1 x ptr], [1 x i32] } { i32 1, [1 x ptr] [ptr @"reflect/methods.Unused(float64)"], [1 x ptr] zeroinitializer, [1 x ptr] [ptr @"reflect/types.type:func:{basic:float64}{}"], [1 x i32] zeroinitializer } }, align 4

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
		// The method set of an interface without methods is a zeroinitializer.
		name: "boxed with empty interface type",
		ir: `
@"reflect/types.type:interface:{}" = linkonce_odr constant { i8, ptr, { i32, [0 x ptr], [0 x ptr], [0 x ptr] } } { i8 20, ptr null, { i32, [0 x ptr], [0 x ptr], [0 x ptr] } zeroinitializer }, align 4

define ptr @box() {
  ret ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Number", i32 0, i32 1)
}
`,
		dispatch: true,
	}, {
		// Struct field names share the type prefix, but are i8 arrays.
		name: "boxed with struct field name",
		ir: `
@"reflect/types.type:struct:{Method:basic:int}.Method" = linkonce_odr unnamed_addr constant [9 x i8] c"\04\00Method\00", align 1

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

// TestInterfaceLoweringTypeIDSwitch checks that invoke thunks with many
// implementers switch on the type ID, and still compare types without one.
func TestInterfaceLoweringTypeIDSwitch(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		numTypes int // plus Legacy
		layout   string
		switched bool
	}{
		{"below threshold", 3, "", false},
		{"above threshold", 5, "", true},
		{"more than 255 types", 260, "", true},
		// Like AVR, where the compiler emits no typeID byte.
		{"no type ID field", 5, "-i16:8", false},
	} {
		var ir strings.Builder
		ir.WriteString(`
target datalayout = "e-m:e-p:32:32` + tc.layout + `-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 2, ptr null }, align 4
@"reflect/methods.Double() int" = linkonce_odr constant i8 0

declare void @runtime.nilPanic(ptr)
declare i32 @"Doubler.Double$invoke"(ptr %receiver, ptr %typecode, ptr %context) #0

define i32 @callDouble(ptr %typecode, ptr %value) {
  %ret = call i32 @"Doubler.Double$invoke"(ptr %value, ptr %typecode, ptr undef)
  ret i32 %ret
}

declare i1 @"Doubler$typeassert"(ptr %typecode) #1

define i1 @isDoubler(ptr %typecode) {
  %ok = call i1 @"Doubler$typeassert"(ptr %typecode)
  ret i1 %ok
}

; Descriptor without numMethods, so it has no type ID byte.
@"Legacy$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Double() int"], { ptr } { ptr @"(Legacy).Double$invoke" } }
@"reflect/types.type:named:Legacy" = linkonce_odr constant { ptr, i8, ptr, ptr } { ptr @"Legacy$methodset", i8 34, ptr null, ptr @"reflect/types.type:basic:int" }, align 4
define i32 @"(Legacy).Double$invoke"(ptr %receiver, ptr %context) {
  ret i32 0
}
@boxLegacy = global ptr getelementptr inbounds ({ ptr, i8, ptr, ptr }, ptr @"reflect/types.type:named:Legacy", i32 0, i32 1)

attributes #0 = { "tinygo-invoke"="reflect/methods.Double() int" "tinygo-methods"="reflect/methods.Double() int" }
attributes #1 = { "tinygo-methods"="reflect/methods.Double() int" }
`)
		// The compiler emits a typeID byte after the kind byte, except on AVR.
		descType, descID := "{ ptr, i8, i8, i16, ptr, ptr, ptr }", "i8 0, "
		if tc.layout != "" {
			descType, descID = "{ ptr, i8, i16, ptr, ptr, ptr }", ""
		}
		for i := range tc.numTypes {
			fmt.Fprintf(&ir, `
@"T%[1]d$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Double() int"], { ptr } { ptr @"(T%[1]d).Double$invoke" } }
@"reflect/types.type:named:T%[1]d" = linkonce_odr constant %[2]s { ptr @"T%[1]d$methodset", i8 34, %[3]si16 1, ptr null, ptr @"reflect/types.type:basic:int", ptr null }, align 4
define i32 @"(T%[1]d).Double$invoke"(ptr %%receiver, ptr %%context) {
  ret i32 %[1]d
}
@box%[1]d = global ptr getelementptr inbounds (%[2]s, ptr @"reflect/types.type:named:T%[1]d", i32 0, i32 1)
`, i, descType, descID)
		}

		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			path := filepath.Join(t.TempDir(), "in.ll")
			if err := os.WriteFile(path, []byte(ir.String()), 0o666); err != nil {
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
			if got := strings.Contains(thunk, "switch i8"); got != tc.switched {
				t.Errorf("thunk uses switch: got %v, want %v\n%s", got, tc.switched, thunk)
			}
			if !strings.Contains(thunk, `icmp eq ptr %actualType, @"reflect/types.type:named:Legacy"`) {
				t.Errorf("thunk does not compare the type without an ID byte:\n%s", thunk)
			}
			for i := range tc.numTypes {
				if !strings.Contains(thunk, fmt.Sprintf(`@"(T%d).Double$invoke"`, i)) {
					t.Errorf("thunk does not call (T%d).Double:\n%s", i, thunk)
				}
			}
			if !strings.Contains(thunk, `@"(Legacy).Double$invoke"`) {
				t.Errorf("thunk does not call (Legacy).Double:\n%s", thunk)
			}
			if tc.switched {
				checkNilBeforeTypeID(t, thunk, "nilpanic")
				if !strings.Contains(thunk, "nilpanic:") || !strings.Contains(thunk, "@runtime.nilPanic(") {
					t.Errorf("thunk does not call runtime.nilPanic:\n%s", thunk)
				}
			}

			assert := mod.NamedFunction("Doubler$typeassert").String()
			for _, want := range []string{"switch i8", "%istagged", "%meta"} {
				if got := strings.Contains(assert, want); got != tc.switched {
					t.Errorf("type assert contains %q: got %v, want %v\n%s", want, got, tc.switched, assert)
				}
			}
			if !strings.Contains(assert, `icmp eq ptr %actualType, @"reflect/types.type:named:Legacy"`) {
				t.Errorf("type assert does not compare the type without an ID byte:\n%s", assert)
			}
			if tc.switched {
				checkNilBeforeTypeID(t, assert, "false")
			}

			seen := make(map[uint64]bool)
			for i := range tc.numTypes {
				name := fmt.Sprintf("reflect/types.type:named:T%d", i)
				var id uint64
				if field := mod.NamedGlobal(name).Initializer().AggregateElement(1); field.Type() == ctx.Int8Type() {
					id = field.ZExtValue()
				}
				if id != 0 {
					if !tc.switched || seen[id] {
						t.Errorf("T%d: bad type ID %d", i, id)
					}
					seen[id] = true
					continue
				}
				// Types without an ID must still be compared.
				cmp := `icmp eq ptr %actualType, @"` + name + `"`
				if tc.switched && (!strings.Contains(thunk, cmp) || !strings.Contains(assert, cmp)) {
					t.Errorf("T%d has no type ID and is not compared", i)
				}
			}
			if wantIDs := min(tc.numTypes, 255); tc.switched && len(seen) != wantIDs {
				t.Errorf("got %d type IDs, want %d", len(seen), wantIDs)
			}
		})
	}
}

// checkNilBeforeTypeID checks that a nil typecode branches to failLabel before
// the type ID load, which would read address 1.
func checkNilBeforeTypeID(t *testing.T, fn, failLabel string) {
	t.Helper()
	isNil := strings.Index(fn, "%isnil = icmp eq ptr %actualType, null")
	branch := strings.Index(fn, "label %"+failLabel+", label %typeid")
	load := strings.Index(fn, "load i8, ptr %typeid.ptr")
	if isNil < 0 || branch < isNil || load < branch {
		t.Errorf("nil typecode does not branch to %s before the type ID load:\n%s", failLabel, fn)
	}
}
