package reflect_test

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"testing"
)

//go:noinline
func captureFunctionPointer(value int) func() int {
	return func() int { return value }
}

func TestClosureCodePointer(t *testing.T) {
	first := reflect.ValueOf(captureFunctionPointer(1))
	second := reflect.ValueOf(captureFunctionPointer(2))
	if first.Pointer() == 0 || first.Pointer() != second.Pointer() {
		t.Fatal("closures from the same expression did not share a code pointer")
	}
}

func callAllocationCases() []struct {
	name string
	fn   reflect.Value
	args []reflect.Value
	max  float64
} {
	type pair [2]*int
	type large [1100]byte
	x := 42
	return []struct {
		name string
		fn   reflect.Value
		args []reflect.Value
		max  float64
	}{
		{"empty", reflect.ValueOf(func() {}), nil, 0},
		{"scalar", reflect.ValueOf(func(x int) int { return x }), []reflect.Value{reflect.ValueOf(x)}, 4},
		{"pointer", reflect.ValueOf(func(x *int) *int { return x }), []reflect.Value{reflect.ValueOf(&x)}, 4},
		{"aggregate", reflect.ValueOf(func(x pair) pair { return x }), []reflect.Value{reflect.ValueOf(pair{&x, &x})}, 5},
		{"indirect", reflect.ValueOf(func(x large) large { return x }), []reflect.Value{reflect.ValueOf(large{})}, 5},
		{"interface", reflect.ValueOf(func(x any) any { return x }), []reflect.Value{reflect.ValueOf(&x)}, 6},
	}
}

var callAllocationResults []reflect.Value

func TestValueCallAllocations(t *testing.T) {
	for _, test := range callAllocationCases() {
		t.Run(test.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(100, func() {
				callAllocationResults = test.fn.Call(test.args)
			}); got > test.max {
				t.Fatalf("Call allocated %v times, want at most %v", got, test.max)
			}
		})
	}
	value := reflect.ValueOf(func(values ...int) int { return len(values) })
	args := []reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)}
	if got := testing.AllocsPerRun(100, func() {
		callAllocationResults = value.Call(args)
	}); got > 7 {
		t.Fatalf("variadic Call allocated %v times, want at most 7", got)
	}
	args = []reflect.Value{reflect.ValueOf([]int{1, 2})}
	if got := testing.AllocsPerRun(100, func() {
		callAllocationResults = value.CallSlice(args)
	}); got > 4 {
		t.Fatalf("CallSlice allocated %v times, want at most 4", got)
	}
}

func BenchmarkValueCall(b *testing.B) {
	for _, test := range callAllocationCases() {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				callAllocationResults = test.fn.Call(test.args)
			}
		})
	}
}

func TestValueCallStorageLifetime(t *testing.T) {
	type payload struct {
		Pointers [2]*int
		Text     string
	}
	type empty struct{}
	x, y := new(int), new(int)
	*x, *y = 17, 23
	input := payload{[2]*int{x, y}, "before"}
	value := reflect.ValueOf(func(arg payload, _ empty) (byte, payload, empty, *int) {
		runtime.GC()
		arg.Text = "after"
		if *arg.Pointers[0] != 17 || *arg.Pointers[1] != 23 {
			t.Fatal("Call lost its argument pointers")
		}
		return 42, arg, empty{}, arg.Pointers[0]
	})
	args := []reflect.Value{reflect.ValueOf(&input).Elem(), reflect.ValueOf(empty{})}
	out := value.Call(args)
	input.Text = "changed"
	args = nil
	runtime.GC()
	if out[0].Uint() != 42 || out[1].Interface().(payload).Text != "after" ||
		out[2].Type() != reflect.TypeOf(empty{}) || *out[3].Interface().(*int) != 17 {
		t.Fatal("Call results did not retain their values")
	}
	for _, result := range out {
		if result.CanAddr() || result.CanSet() {
			t.Fatal("Call returned addressable storage")
		}
	}
	next := value.Call([]reflect.Value{reflect.ValueOf(input), reflect.ValueOf(empty{})})
	if next[0].Uint() != 42 || out[1].Interface().(payload).Text != "after" {
		t.Fatal("Call reused a previous result buffer")
	}
}

func TestValueCall(t *testing.T) {
	offset := 3
	fn := func(a int, b string) (int, string) {
		return a + offset, b + "!"
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{
		reflect.ValueOf(4),
		reflect.ValueOf("go"),
	})
	if got := int(out[0].Int()); got != 7 {
		t.Fatalf("first result = %d, want 7", got)
	}
	if got := out[1].String(); got != "go!" {
		t.Fatalf("second result = %q, want %q", got, "go!")
	}
	if out[0].CanAddr() || out[0].CanSet() {
		t.Fatal("Call result is addressable or settable")
	}
}

func TestValueCallNamedFunction(t *testing.T) {
	type named func(int) int
	var fn named = func(value int) int {
		return value + 1
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(41)})
	if got := int(out[0].Int()); got != 42 {
		t.Fatalf("result = %d, want 42", got)
	}
}

func TestValueCallAggregate(t *testing.T) {
	type pair struct {
		Number int
		Text   string
	}
	fn := func(value pair) pair {
		value.Number++
		value.Text += "!"
		return value
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(pair{41, "go"})})
	if got := out[0].Interface().(pair); got != (pair{42, "go!"}) {
		t.Fatalf("result = %#v, want %#v", got, pair{42, "go!"})
	}
}

func TestCallInlineValues(t *testing.T) {
	const number = uint16(0x1234)
	pair := [2]byte{0x56, 0x78}
	fn := reflect.ValueOf(func(n uint16, p [2]byte) (uint16, [2]byte) {
		return n + 1, [2]byte{p[1], p[0]}
	})
	out := fn.Call([]reflect.Value{reflect.ValueOf(number), reflect.ValueOf(pair)})
	if got := out[0].Interface().(uint16); got != number+1 {
		t.Errorf("Call returned %#x, want %#x", got, number+1)
	}
	if got := out[1].Interface().([2]byte); got != ([2]byte{pair[1], pair[0]}) {
		t.Errorf("Call returned %#v, want reversed input", got)
	}

	generated := reflect.MakeFunc(fn.Type(), func(in []reflect.Value) []reflect.Value {
		if n, p := in[0].Interface().(uint16), in[1].Interface().([2]byte); n != number || p != pair {
			t.Errorf("MakeFunc callback received %#x, %#v", n, p)
		}
		return in
	}).Interface().(func(uint16, [2]byte) (uint16, [2]byte))
	if n, p := generated(number, pair); n != number || p != pair {
		t.Errorf("MakeFunc returned %#x, %#v", n, p)
	}
}

func TestValueCallZeroSized(t *testing.T) {
	type empty struct{}
	fn := func(empty) empty {
		return empty{}
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(empty{})})
	if len(out) != 1 || out[0].Type() != reflect.TypeOf(empty{}) {
		t.Fatalf("Call returned %#v, want one empty result", out)
	}
}

func TestValueCallMixedABI(t *testing.T) {
	type pair [2]uintptr
	fn := func(b byte, c int, d byte, e pair, f byte, g float32, h byte) (byte, int, byte, pair, byte, float32, byte) {
		return b, c, d, e, f, g, h
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{
		reflect.ValueOf(byte(10)),
		reflect.ValueOf(20),
		reflect.ValueOf(byte(30)),
		reflect.ValueOf(pair{40, 50}),
		reflect.ValueOf(byte(60)),
		reflect.ValueOf(float32(70)),
		reflect.ValueOf(byte(80)),
	})
	if len(out) != 7 {
		t.Fatalf("Call returned %d values, want 7", len(out))
	}
	if byte(out[0].Uint()) != 10 ||
		int(out[1].Int()) != 20 ||
		byte(out[2].Uint()) != 30 ||
		out[3].Interface().(pair) != (pair{40, 50}) ||
		byte(out[4].Uint()) != 60 ||
		float32(out[5].Float()) != 70 ||
		byte(out[6].Uint()) != 80 {
		t.Fatalf("Call returned unexpected mixed results")
	}
}

func TestValueCallInterface(t *testing.T) {
	fn := func(value any) any {
		return struct{ Value any }{value}
	}
	out := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(42)})
	got := out[0].Interface().(struct{ Value any })
	if got.Value != 42 {
		t.Fatalf("result = %#v, want 42", got.Value)
	}
}

func TestValueCallInterfaceConversion(t *testing.T) {
	type reader interface {
		Read([]byte) (int, error)
	}
	fn := func(value reader) reader {
		return value
	}
	var value reader
	out := reflect.ValueOf(fn).Call([]reflect.Value{
		reflect.ValueOf(&value).Elem(),
	})
	if len(out) != 1 || !out[0].IsNil() || out[0].Type() != reflect.TypeOf((*reader)(nil)).Elem() {
		t.Fatalf("Call returned %#v, want a nil reader", out)
	}
}

func TestValueCallVariadic(t *testing.T) {
	fn := func(prefix string, values ...int) int {
		total := len(prefix)
		for _, value := range values {
			total += value
		}
		return total
	}
	value := reflect.ValueOf(fn)
	out := value.Call([]reflect.Value{
		reflect.ValueOf("go"),
		reflect.ValueOf(3),
		reflect.ValueOf(4),
	})
	if got := int(out[0].Int()); got != 9 {
		t.Fatalf("Call result = %d, want 9", got)
	}
	out = value.CallSlice([]reflect.Value{
		reflect.ValueOf("go"),
		reflect.ValueOf([]int{3, 4}),
	})
	if got := int(out[0].Int()); got != 9 {
		t.Fatalf("CallSlice result = %d, want 9", got)
	}
}

func TestValueCallImportedFunction(t *testing.T) {
	out := reflect.ValueOf(fmt.Sprintf).Call([]reflect.Value{
		reflect.ValueOf("%s %d"),
		reflect.ValueOf("value"),
		reflect.ValueOf(42),
	})
	if got := out[0].String(); got != "value 42" {
		t.Fatalf("Sprintf result = %q, want %q", got, "value 42")
	}
}

func TestValueCallIndirectAggregate(t *testing.T) {
	type large [1100]byte
	fn := func(value large) large {
		value[0]++
		value[len(value)-1]++
		return value
	}
	var input large
	input[0] = 1
	input[len(input)-1] = 2
	out := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(input)})
	got := out[0].Interface().(large)
	if got[0] != 2 || got[len(got)-1] != 3 {
		t.Fatalf("result endpoints = %d, %d; want 2, 3", got[0], got[len(got)-1])
	}
}

func TestValueCallPanics(t *testing.T) {
	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		fn()
	}

	mustPanic("non-function", func() {
		reflect.ValueOf(1).Call(nil)
	})
	mustPanic("nil function", func() {
		var fn func()
		reflect.ValueOf(fn).Call(nil)
	})
	hidden := reflect.ValueOf(struct {
		fn       func()
		variadic func(...int)
	}{func() {}, func(...int) {}})
	mustPanic("unexported function", func() {
		hidden.Field(0).Call(nil)
	})
	mustPanic("unexported variadic function", func() {
		hidden.Field(1).CallSlice([]reflect.Value{reflect.ValueOf([]int{})})
	})
	type nested struct{ Function func() }
	indirect := reflect.ValueOf(struct{ hidden nested }{nested{func() {}}})
	mustPanic("function through unexported field", func() {
		indirect.Field(0).Field(0).Call(nil)
	})
	mustPanic("wrong argument count", func() {
		reflect.ValueOf(func(int) {}).Call(nil)
	})
	mustPanic("wrong argument type", func() {
		reflect.ValueOf(func(int) {}).Call([]reflect.Value{reflect.ValueOf("x")})
	})
	mustPanic("CallSlice on non-variadic function", func() {
		reflect.ValueOf(func([]int) {}).CallSlice([]reflect.Value{reflect.ValueOf([]int{})})
	})
	mustPanic("called function", func() {
		reflect.ValueOf(func() { panic("called") }).Call(nil)
	})
}

func TestMakeFuncAllocations(t *testing.T) {
	for _, test := range callAllocationCases() {
		t.Run(test.name, func(t *testing.T) {
			value := reflect.MakeFunc(test.fn.Type(), func(in []reflect.Value) []reflect.Value { return in })
			max := test.max + 1
			if test.name == "empty" {
				max = 0
			}
			if got := testing.AllocsPerRun(100, func() {
				callAllocationResults = value.Call(test.args)
			}); got > max {
				t.Fatalf("MakeFunc Call allocated %v times, want at most %v", got, max)
			}
		})
	}
	value := reflect.MakeFunc(reflect.TypeOf((func())(nil)), func(in []reflect.Value) []reflect.Value { return in })
	if got := testing.AllocsPerRun(100, value.Interface().(func())); got != 0 {
		t.Fatalf("MakeFunc empty function allocated %v times, want 0", got)
	}
	value = reflect.MakeFunc(reflect.TypeOf((func(int) int)(nil)), func(in []reflect.Value) []reflect.Value { return in })
	fn := value.Interface().(func(int) int)
	if got := testing.AllocsPerRun(100, func() {
		if fn(42) != 42 {
			t.Fatal("MakeFunc returned the wrong value")
		}
	}); got > 1 {
		t.Fatalf("MakeFunc scalar function allocated %v times, want at most 1", got)
	}
}

func BenchmarkMakeFuncCall(b *testing.B) {
	for _, test := range callAllocationCases() {
		b.Run(test.name, func(b *testing.B) {
			value := reflect.MakeFunc(test.fn.Type(), func(in []reflect.Value) []reflect.Value { return in })
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				callAllocationResults = value.Call(test.args)
			}
		})
	}
}

func TestMakeFuncRetainedArguments(t *testing.T) {
	t.Run("direct aggregate", func(t *testing.T) {
		type pair [2]uintptr
		var saved reflect.Value
		fn := reflect.MakeFunc(reflect.TypeFor[func(pair) pair](), func(in []reflect.Value) []reflect.Value {
			if !saved.IsValid() {
				saved = in[0]
			}
			return in
		}).Interface().(func(pair) pair)
		fn(pair{1, 2})
		fn(pair{3, 4})
		if got := saved.Interface().(pair); got != (pair{1, 2}) {
			t.Fatalf("retained argument = %v, want [1 2]", got)
		}
	})

	type payload struct {
		Pointer *int
		Padding [1100]byte
	}
	var saved []reflect.Value
	value := reflect.MakeFunc(reflect.TypeOf((func(payload) payload)(nil)), func(in []reflect.Value) []reflect.Value {
		saved = append(saved, in[0])
		runtime.GC()
		return in
	})
	fn := value.Interface().(func(payload) payload)
	input := payload{Pointer: new(int)}
	*input.Pointer = 17
	input.Padding[0] = 1
	first := fn(input)
	input.Padding[0] = 2
	input.Pointer = new(int)
	*input.Pointer = 23
	second := fn(input)
	input.Padding[0] = 3
	third := value.Call([]reflect.Value{reflect.ValueOf(&input).Elem()})[0]
	input.Padding[0] = 4
	input.Pointer = nil
	runtime.GC()
	for i, want := range []payload{first, second, third.Interface().(payload)} {
		got := saved[i].Interface().(payload)
		if got != want || got.Padding[0] != byte(i+1) || *got.Pointer != []int{17, 23, 23}[i] {
			t.Fatalf("retained argument %d was changed", i)
		}
		if saved[i].CanAddr() || saved[i].CanSet() {
			t.Fatalf("retained argument %d is addressable", i)
		}
	}
}

func TestMakeFuncABI(t *testing.T) {
	type pair [2]uintptr
	typ := reflect.TypeOf(func(byte, int, byte, pair, byte, float32, byte) (byte, int, byte, pair, byte, float32, byte) {
		return 0, 0, 0, pair{}, 0, 0, 0
	})
	value := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		return in
	})
	fn := value.Interface().(func(byte, int, byte, pair, byte, float32, byte) (byte, int, byte, pair, byte, float32, byte))
	b, c, d, e, f, g, h := fn(10, 20, 30, pair{40, 50}, 60, 70, 80)
	if b != 10 || c != 20 || d != 30 || e != (pair{40, 50}) || f != 60 || g != 70 || h != 80 {
		t.Fatalf("MakeFunc returned %d, %d, %d, %v, %d, %g, %d", b, c, d, e, f, g, h)
	}
}

func TestMakeFuncCall(t *testing.T) {
	typ := reflect.TypeOf(func(int) int { return 0 })
	value := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		if in[0].CanAddr() {
			t.Error("MakeFunc argument is addressable")
		}
		return []reflect.Value{reflect.ValueOf(int(in[0].Int()) + 1)}
	})
	out := value.Call([]reflect.Value{reflect.ValueOf(41)})
	if got := int(out[0].Int()); got != 42 {
		t.Fatalf("Call returned %d, want 42", got)
	}
}

func TestMakeFuncAssignableResult(t *testing.T) {
	typ := reflect.TypeOf(func() error { return nil })
	value := reflect.MakeFunc(typ, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(io.EOF)}
	})
	if got := value.Interface().(func() error)(); got != io.EOF {
		t.Fatalf("MakeFunc returned %v, want %v", got, io.EOF)
	}
}

func TestMakeFuncZeroSized(t *testing.T) {
	type empty struct{}
	typ := reflect.TypeOf(func(empty) empty { return empty{} })
	value := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		return in
	})
	if got := value.Interface().(func(empty) empty)(empty{}); got != (empty{}) {
		t.Fatalf("MakeFunc returned %#v", got)
	}
}

func TestMakeFuncIndirectAggregate(t *testing.T) {
	type large [1100]byte
	typ := reflect.TypeOf(func(large) large { return large{} })
	value := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		result := in[0].Interface().(large)
		result[0]++
		result[len(result)-1]++
		return []reflect.Value{reflect.ValueOf(result)}
	})
	var input large
	input[0] = 1
	input[len(input)-1] = 2
	got := value.Interface().(func(large) large)(input)
	if got[0] != 2 || got[len(got)-1] != 3 {
		t.Fatalf("result endpoints = %d, %d; want 2, 3", got[0], got[len(got)-1])
	}
}

func TestMakeFuncDynamicType(t *testing.T) {
	type input string
	type output float64
	typ := reflect.FuncOf(
		[]reflect.Type{reflect.TypeOf(input(""))},
		[]reflect.Type{reflect.TypeOf(output(0))},
		false,
	)
	value := reflect.MakeFunc(typ, func(in []reflect.Value) []reflect.Value {
		if got := in[0].Interface().(input); got != "gopher" {
			t.Fatalf("argument = %q, want %q", got, "gopher")
		}
		return []reflect.Value{reflect.ValueOf(output(3.14))}
	})
	out := value.Call([]reflect.Value{reflect.ValueOf(input("gopher"))})
	if got := out[0].Interface().(output); got != 3.14 {
		t.Fatalf("Call returned %v, want 3.14", got)
	}
}

func TestMakeFuncPanics(t *testing.T) {
	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		fn()
	}

	mustPanic("non-function type", func() {
		reflect.MakeFunc(reflect.TypeOf(0), func([]reflect.Value) []reflect.Value { return nil })
	})
	mustPanic("wrong result count", func() {
		value := reflect.MakeFunc(reflect.TypeOf(func() int { return 0 }), func([]reflect.Value) []reflect.Value {
			return nil
		})
		value.Call(nil)
	})
	mustPanic("zero result", func() {
		value := reflect.MakeFunc(reflect.TypeOf(func() int { return 0 }), func([]reflect.Value) []reflect.Value {
			return []reflect.Value{{}}
		})
		value.Call(nil)
	})
	mustPanic("wrong result type", func() {
		value := reflect.MakeFunc(reflect.TypeOf(func() int { return 0 }), func([]reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.ValueOf("bad")}
		})
		value.Call(nil)
	})
}

type methodValueTarget struct {
	Base int
}

func (v methodValueTarget) Add(value int) int {
	return v.Base + value
}

func (v *methodValueTarget) Set(value int) {
	v.Base = value
}

func (v methodValueTarget) Sum(values ...int) int {
	total := v.Base
	for _, value := range values {
		total += value
	}
	return total
}

type embeddedMethodTarget struct {
	methodValueTarget
}

func TestTypeMethodFunc(t *testing.T) {
	typ := reflect.TypeOf(methodValueTarget{})
	method, ok := typ.MethodByName("Add")
	if !ok {
		t.Fatal("Add method was not found")
	}

	out := method.Func.Call([]reflect.Value{
		reflect.ValueOf(methodValueTarget{Base: 40}),
		reflect.ValueOf(2),
	})
	if got := int(out[0].Int()); got != 42 {
		t.Fatalf("Method.Func returned %d, want 42", got)
	}
}

func TestInterfaceTypeMethodFunc(t *testing.T) {
	typ := reflect.TypeOf((*interface{ Add(int) int })(nil)).Elem()
	method, ok := typ.MethodByName("Add")
	if !ok {
		t.Fatal("Add method was not found")
	}
	if method.Func.IsValid() {
		t.Fatalf("interface Method.Func is %v, want zero Value", method.Func)
	}
}

func TestValueMethod(t *testing.T) {
	value := reflect.ValueOf(methodValueTarget{Base: 40}).MethodByName("Add")
	if !value.IsValid() {
		t.Fatal("Add method value was not found")
	}
	out := value.Call([]reflect.Value{reflect.ValueOf(2)})
	if got := int(out[0].Int()); got != 42 {
		t.Fatalf("method value returned %d, want 42", got)
	}
	fn := value.Interface().(func(int) int)
	if got := fn(3); got != 43 {
		t.Fatalf("method function returned %d, want 43", got)
	}
}

func TestValuePointerMethod(t *testing.T) {
	target := &methodValueTarget{}
	method := reflect.ValueOf(target).MethodByName("Set")
	method.Call([]reflect.Value{reflect.ValueOf(42)})
	if target.Base != 42 {
		t.Fatalf("Set stored %d, want 42", target.Base)
	}
}

func TestValueVariadicMethod(t *testing.T) {
	method := reflect.ValueOf(methodValueTarget{Base: 40}).MethodByName("Sum")
	out := method.Call([]reflect.Value{
		reflect.ValueOf(1),
		reflect.ValueOf(2),
	})
	if got := int(out[0].Int()); got != 43 {
		t.Fatalf("variadic method returned %d, want 43", got)
	}
	out = method.CallSlice([]reflect.Value{
		reflect.ValueOf([]int{1, 2}),
	})
	if got := int(out[0].Int()); got != 43 {
		t.Fatalf("variadic CallSlice returned %d, want 43", got)
	}
}

func TestValuePromotedMethod(t *testing.T) {
	target := embeddedMethodTarget{methodValueTarget{Base: 40}}
	out := reflect.ValueOf(target).MethodByName("Add").Call([]reflect.Value{reflect.ValueOf(2)})
	if got := int(out[0].Int()); got != 42 {
		t.Fatalf("promoted method returned %d, want 42", got)
	}
}

func TestValueInterfaceMethod(t *testing.T) {
	var target interface{ Add(int) int } = methodValueTarget{Base: 40}
	value := reflect.ValueOf(&target).Elem()
	method := value.MethodByName("Add")
	target = embeddedMethodTarget{methodValueTarget{Base: 50}}
	out := method.Call([]reflect.Value{reflect.ValueOf(2)})
	if got := int(out[0].Int()); got != 52 {
		t.Fatalf("interface method returned %d, want 52", got)
	}

	target = nil
	t.Run("nil at lookup", func(t *testing.T) {
		shouldPanic("nil interface value", func() { value.Method(0) })
	})
	t.Run("nil at call", func(t *testing.T) {
		shouldPanic("nil interface value", func() { method.Call([]reflect.Value{reflect.ValueOf(2)}) })
	})
}

func TestValueMethodByNameMissing(t *testing.T) {
	if value := reflect.ValueOf(methodValueTarget{}).MethodByName("Missing"); value.IsValid() {
		t.Fatalf("MethodByName returned %v for a missing method", value)
	}
}

type wideMethodTarget struct{}

func (wideMethodTarget) Wide(
	first, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _,
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, last int,
) int {
	return first + last
}

func TestValueWideMethod(t *testing.T) {
	value := reflect.ValueOf(wideMethodTarget{})
	args := make([]reflect.Value, 128)
	for i := range args {
		args[i] = reflect.ValueOf(i + 1)
	}
	for _, method := range []reflect.Value{value.Method(0), value.MethodByName("Wide")} {
		if got := method.Call(args)[0].Int(); got != 129 {
			t.Fatalf("wide method returned %d, want 129", got)
		}
	}
}
