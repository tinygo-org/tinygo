package reflect_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

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
