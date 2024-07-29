package main

import "time"

func main() {
	thing := &Thing{"foo"}
	println("thing:", thing.String())
	thing.Print()
	printItf(5)
	printItf(byte('x'))
	printItf("foo")
	printItf(Foo(18))
	printItf(*thing)
	printItf(thing)
	printItf(Stringer(thing))
	printItf(struct{ n int }{})
	printItf(struct {
		n int `foo:"bar"`
	}{})
	printItf(Number(3))
	array := Array([4]uint32{1, 7, 11, 13})
	printItf(array)
	printItf(ArrayStruct{3, array})
	printItf(SmallPair{3, 5})
	printItf(nil)
	s := Stringer(thing)
	println("Stringer.String():", s.String())
	var itf interface{} = s
	println("Stringer.(*Thing).String():", itf.(Stringer).String())
	if s, ok := s.(interface{ String() string }); ok {
		println("s has String() method:", s.String())
	}

	println("nested switch:", nestedSwitch('v', 3))

	// Try putting a linked list in an interface:
	// https://github.com/tinygo-org/tinygo/issues/309
	itf = linkedList{}

	// Test bugfix for assertion on named empty interface:
	// https://github.com/tinygo-org/tinygo/issues/453
	_, _ = itf.(Empty)

	var v Byter = FooByte(3)
	println("Byte(): ", v.Byte())

	var n int
	var f float32
	var interfaceEqualTests = []struct {
		equal bool
		lhs   interface{}
		rhs   interface{}
	}{
		{true, true, true},
		{true, int(1), int(1)},
		{true, int8(1), int8(1)},
		{true, int16(1), int16(1)},
		{true, int32(1), int32(1)},
		{true, int64(1), int64(1)},
		{true, uint(1), uint(1)},
		{false, uint(1), uint(2)},
		{true, uint8(1), uint8(1)},
		{true, uint16(1), uint16(1)},
		{true, uint32(1), uint32(1)},
		{true, uint64(1), uint64(1)},
		{true, uintptr(1), uintptr(1)},
		{true, float32(1.1), float32(1.1)},
		{true, float64(1.1), float64(1.1)},
		{true, complex(100, 8), complex(100, 8)},
		{false, complex(100, 8), complex(101, 8)},
		{false, complex(100, 8), complex(100, 9)},
		{true, complex64(8), complex64(8)},
		{true, complex128(8), complex128(8)},
		{true, "string", "string"},
		{false, "string", "stringx"},
		{true, [2]int16{-5, 201}, [2]int16{-5, 201}},
		{false, [2]int16{-5, 201}, [2]int16{-5, 202}},
		{false, [2]int16{-5, 201}, [2]int16{5, 201}},
		{true, &n, &n},
		{false, &n, new(int)},
		{false, new(int), new(int)},
		{false, &n, &f},
		{true, struct {
			a int
			b int
		}{3, 5}, struct {
			a int
			b int
		}{3, 5}},
		{false, struct {
			a int
			b int
		}{3, 5}, struct {
			a int
			b int
		}{3, 6}},
		{true, named1(), named1()},
		{true, named2(), named2()},
		{false, named1(), named2()},
		{false, named2(), named3()},
		{true, namedptr1(), namedptr1()},
		{false, namedptr1(), namedptr2()},
	}
	for i, tc := range interfaceEqualTests {
		if (tc.lhs == tc.rhs) != tc.equal {
			println("test", i, "of interfaceEqualTests failed")
		}
	}

	// test interface blocking
	blockDynamic(NonBlocker{})
	println("non-blocking call on sometimes-blocking interface")
	blockDynamic(SleepBlocker(time.Millisecond))
	println("slept 1ms")
	blockStatic(SleepBlocker(time.Millisecond))
	println("slept 1ms")

	// check that pointer-to-pointer type switches work
	ptrptrswitch()

	// check that type asserts to interfaces with no methods work
	emptyintfcrash()
}

func printItf(val interface{}) {
	switch val := val.(type) {
	case Unmatched:
		panic("matched the unmatchable")
	case Doubler:
		println("is Doubler:", val.Double())
	case Tuple:
		println("is Tuple:", val.Nth(0), val.Nth(1), val.Nth(2), val.Nth(3))
		val.Print()
	case int:
		println("is int:", val)
	case byte:
		println("is byte:", val)
	case string:
		println("is string:", val)
	case Thing:
		println("is Thing:", val.String())
	case *Thing:
		println("is *Thing:", val.String())
	case struct{ i int }:
		println("is struct{i int}")
	case struct{ n int }:
		println("is struct{n int}")
	case struct {
		n int `foo:"bar"`
	}:
		println("is struct{n int `foo:\"bar\"`}")
	case Foo:
		println("is Foo:", val)
	case nil:
		println("is nil")
	default:
		println("is ?")
	}
}

var (
	// Test for type assert support in the interp package.
	globalThing interface{} = Foo(3)
	_                       = globalThing.(Foo)
)

func nestedSwitch(verb rune, arg interface{}) bool {
	switch verb {
	case 'v', 's':
		switch arg.(type) {
		case int:
			return true
		}
	}
	return false
}

func blockDynamic(blocker DynamicBlocker) {
	blocker.Block()
}

func blockStatic(blocker StaticBlocker) {
	blocker.Sleep()
}

type Thing struct {
	name string
}

func (t Thing) String() string {
	return t.name
}

func (t Thing) Print() {
	println("Thing.Print:", t.name)
}

type Stringer interface {
	String() string
}

type Foo int

type Number int

func (n Number) Double() int {
	return int(n) * 2
}

type Doubler interface {
	Double() int
}

type Tuple interface {
	Nth(int) uint32
	Print()
}

type Array [4]uint32

func (a Array) Nth(n int) uint32 {
	return a[n]
}

func (a Array) Print() {
	println("Array len:", len(a))
}

type ArrayStruct struct {
	n int
	a Array
}

func (a ArrayStruct) Nth(n int) uint32 {
	return a.a[n]
}

func (a ArrayStruct) Print() {
	println("ArrayStruct.Print:", len(a.a), a.n)
}

type SmallPair struct {
	a byte
	b byte
}

func (p SmallPair) Nth(n int) uint32 {
	return uint32(int(p.a)*n + int(p.b)*n)
}

func (p SmallPair) Print() {
	println("SmallPair.Print:", p.a, p.b)
}

// There is no type that matches this method.
type Unmatched interface {
	NeverImplementedMethod()
}

type linkedList struct {
	addr *linkedList
}

type DynamicBlocker interface {
	Block()
}

type NonBlocker struct{}

func (b NonBlocker) Block() {}

type SleepBlocker time.Duration

func (s SleepBlocker) Block() {
	time.Sleep(time.Duration(s))
}

func (s SleepBlocker) Sleep() {
	s.Block()
}

type StaticBlocker interface {
	Sleep()
}

type Empty interface{}

type FooByte int

func (f FooByte) Byte() byte { return byte(f) }

type Byter interface {
	Byte() uint8
}

// Make sure that named types inside functions do not alias with any other named
// functions.

type named int

func named1() any {
	return named(0)
}

func named2() any {
	type named int
	return named(0)
}

func named3() any {
	type named int
	return named(0)
}

func namedptr1() interface{} {
	type Test int
	return (*Test)(nil)
}

func namedptr2() interface{} {
	type Test byte
	return (*Test)(nil)
}

func ptrptrswitch() {
	identify(0)
	identify(new(int))
	identify(new(*int))
}

func identify(itf any) {
	switch itf.(type) {
	case int:
		println("type is int")
	case *int:
		println("type is *int")
	case **int:
		println("type is **int")
	default:
		println("other type??")
	}
}

func emptyintfcrash() {
	if x, ok := any(5).(any); ok {
		println("x is", x.(int))
	}
}
