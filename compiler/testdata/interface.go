// This file tests interface types and interface builtins.

package main

// Test interface construction.

func simpleType() interface{} {
	return 0
}

func pointerType() interface{} {
	// Pointers have an element type, in this case int.
	var v *int
	return v
}

func interfaceType() interface{} {
	// Interfaces can exist in interfaces, but only indirectly (through
	// pointers).
	var v *error
	return v
}

func anonymousInterfaceType() interface{} {
	var v *interface {
		String() string
	}
	return v
}

// Test interface builtins.

func isInt(itf interface{}) bool {
	_, ok := itf.(int)
	return ok
}

func isError(itf interface{}) bool {
	// Interface assert on (builtin) named interface type.
	_, ok := itf.(error)
	return ok
}

func isStringer(itf interface{}) bool {
	// Interface assert on anonymous interface type.
	_, ok := itf.(interface {
		String() string
	})
	return ok
}

type fooInterface interface {
	String() string
	foo(int) byte
}

func callFooMethod(itf fooInterface) uint8 {
	return itf.foo(3)
}

func callErrorMethod(itf error) string {
	return itf.Error()
}

func namedFoo() {
	type Foo struct {
		A int
	}
	f1 := &Foo{}
	fcopy := copyOf(f1)
	f2 := fcopy.(*Foo)
	println(f2.A)
}

func namedFoo2Nested() {
	type Foo struct {
		A *int
	}
	f1 := &Foo{}
	fcopy := copyOf(f1)
	f2 := fcopy.(*Foo)
	println(f2.A == nil)

	if f2.A == nil {
		type Foo struct {
			A *byte
		}
		nestedf1 := &Foo{}
		fcopy := copyOf(nestedf1)
		nestedf2 := fcopy.(*Foo)
		println(nestedf2.A == nil)
	}
}

func copyOf(src interface{}) (dst interface{}) {
	return src
}
