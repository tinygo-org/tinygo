package main

import "reflect"

type checker = func(any) bool

func siblingA() (any, checker) {
	type Foo struct{ V int }
	return Foo{V: 1}, func(x any) bool { _, ok := x.(Foo); return ok }
}

func siblingB() (any, checker) {
	type Foo struct{ V int }
	return Foo{V: 2}, func(x any) bool { _, ok := x.(Foo); return ok }
}

func nestedScopes() (any, any, any, checker, checker, checker) {
	var av, bv, cv any
	var ac, bc, cc checker
	{
		type Bar struct{ X int }
		av = Bar{X: 10}
		ac = func(x any) bool { _, ok := x.(Bar); return ok }
	}
	{
		type Bar struct{ X int }
		bv = Bar{X: 20}
		bc = func(x any) bool { _, ok := x.(Bar); return ok }
	}
	{
		type Bar struct{ X int }
		cv = Bar{X: 30}
		cc = func(x any) bool { _, ok := x.(Bar); return ok }
	}
	return av, bv, cv, ac, bc, cc
}

func genericWithLocals[T any]() (any, any, checker, checker) {
	type UsesT struct{ V T }
	type NoT struct{ V int }
	var z T
	return UsesT{V: z}, NoT{V: 42},
		func(x any) bool { _, ok := x.(UsesT); return ok },
		func(x any) bool { _, ok := x.(NoT); return ok }
}

func siblingClosures() (any, any, checker, checker) {
	var av, bv any
	var ac, bc checker
	func() {
		type C struct{ V int }
		av = C{V: 1}
		ac = func(x any) bool { _, ok := x.(C); return ok }
	}()
	func() {
		type C struct{ V int }
		bv = C{V: 2}
		bc = func(x any) bool { _, ok := x.(C); return ok }
	}()
	return av, bv, ac, bc
}

func closureInGenericUsesT[T any]() (any, checker) {
	var v any
	var c checker
	func() {
		type X struct{ V T }
		var z T
		v = X{V: z}
		c = func(x any) bool { _, ok := x.(X); return ok }
	}()
	return v, c
}

func closureInGenericNoT[T any]() (any, checker) {
	var v any
	var c checker
	func() {
		type X struct{ V int }
		v = X{V: 7}
		c = func(x any) bool { _, ok := x.(X); return ok }
	}()
	return v, c
}

func siblingClosuresInGeneric[T any]() (any, any, checker, checker) {
	var av, bv any
	var ac, bc checker
	func() {
		type C struct{ V T }
		var z T
		av = C{V: z}
		ac = func(x any) bool { _, ok := x.(C); return ok }
	}()
	func() {
		type C struct{ V T }
		var z T
		bv = C{V: z}
		bc = func(x any) bool { _, ok := x.(C); return ok }
	}()
	return av, bv, ac, bc
}

func doublyNestedInGeneric[T any]() (any, checker) {
	var v any
	var c checker
	func() {
		func() {
			type Y struct{ V T }
			var z T
			v = Y{V: z}
			c = func(x any) bool { _, ok := x.(Y); return ok }
		}()
	}()
	return v, c
}

func expect(name string, ok bool) {
	if ok {
		println("ok:", name)
	} else {
		println("BUG:", name)
	}
}

// issue5180Copy1 and issue5180CopyIgnoreNilMembers are the original
// repro from https://github.com/tinygo-org/tinygo/issues/5180. Each
// declares its own local Foo with a different field type, then uses
// reflect.New to construct a value via the runtime type and asserts it
// back. Without distinct local-type names the two Foos collide and the
// second assertion panics.
func issue5180Copy1() bool {
	type Foo struct{ A int }
	f1 := &Foo{}
	dst := reflect.New(reflect.TypeOf(f1).Elem()).Interface()
	_, ok := dst.(*Foo)
	return ok
}

func issue5180CopyIgnoreNilMembers() (ok bool) {
	type Foo struct{ A *int }
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	f1 := &Foo{}
	dst := reflect.New(reflect.TypeOf(f1).Elem()).Interface()
	_, ok = dst.(*Foo)
	return ok
}

func main() {
	expect("issue5180 TestCopy1", issue5180Copy1())
	expect("issue5180 TestCopyIgnoreNilMembers", issue5180CopyIgnoreNilMembers())

	// Two siblings with same-named local types are distinct.
	aV, aC := siblingA()
	bV, bC := siblingB()
	expect("siblingA.Foo accepts own", aC(aV))
	expect("siblingB.Foo accepts own", bC(bV))
	expect("siblingA.Foo rejects siblingB.Foo", !aC(bV))
	expect("siblingB.Foo rejects siblingA.Foo", !bC(aV))

	// Three sibling-scope locals in one function are mutually distinct.
	n1V, n2V, n3V, n1C, n2C, n3C := nestedScopes()
	expect("nestedScopes.Bar#1 accepts own", n1C(n1V))
	expect("nestedScopes.Bar#2 accepts own", n2C(n2V))
	expect("nestedScopes.Bar#3 accepts own", n3C(n3V))
	expect("nestedScopes.Bar#1 rejects #2", !n1C(n2V))
	expect("nestedScopes.Bar#2 rejects #3", !n2C(n3V))
	expect("nestedScopes.Bar#1 rejects #3", !n1C(n3V))

	// Generic instantiations are distinct from each other but match
	// themselves across calls.
	iU, iN, iUC, iNC := genericWithLocals[int]()
	sU, sN, sUC, sNC := genericWithLocals[string]()
	iU2, iN2, _, _ := genericWithLocals[int]()
	expect("genericWithLocals[int].UsesT accepts own", iUC(iU))
	expect("genericWithLocals[int].NoT accepts own", iNC(iN))
	expect("genericWithLocals[string].UsesT accepts own", sUC(sU))
	expect("genericWithLocals[string].NoT accepts own", sNC(sN))
	expect("genericWithLocals[int].UsesT rejects [string].UsesT", !iUC(sU))
	expect("genericWithLocals[int].NoT rejects [string].NoT", !iNC(sN))
	expect("genericWithLocals[string].UsesT rejects [int].UsesT", !sUC(iU))
	expect("genericWithLocals[string].NoT rejects [int].NoT", !sNC(iN))
	expect("genericWithLocals[int].UsesT matches across calls", iUC(iU2))
	expect("genericWithLocals[int].NoT matches across calls", iNC(iN2))

	// Sibling closures in a non-generic function.
	scA, scB, scAC, scBC := siblingClosures()
	expect("siblingClosures.C#1 accepts own", scAC(scA))
	expect("siblingClosures.C#2 accepts own", scBC(scB))
	expect("siblingClosures.C#1 rejects C#2", !scAC(scB))
	expect("siblingClosures.C#2 rejects C#1", !scBC(scA))

	// Closure inside generic function, type uses T.
	cgi, cgiC := closureInGenericUsesT[int]()
	cgs, cgsC := closureInGenericUsesT[string]()
	expect("closureInGenericUsesT[int].X accepts own", cgiC(cgi))
	expect("closureInGenericUsesT[string].X accepts own", cgsC(cgs))
	expect("closureInGenericUsesT[int].X rejects [string].X", !cgiC(cgs))
	expect("closureInGenericUsesT[string].X rejects [int].X", !cgsC(cgi))

	// Closure inside generic function, type does not use T.
	cni, cniC := closureInGenericNoT[int]()
	cns, cnsC := closureInGenericNoT[string]()
	expect("closureInGenericNoT[int].X accepts own", cniC(cni))
	expect("closureInGenericNoT[string].X accepts own", cnsC(cns))
	expect("closureInGenericNoT[int].X rejects [string].X", !cniC(cns))
	expect("closureInGenericNoT[string].X rejects [int].X", !cnsC(cni))

	// Sibling closures inside a generic function.
	sgIA, sgIB, sgIAC, sgIBC := siblingClosuresInGeneric[int]()
	sgSA, _, sgSAC, _ := siblingClosuresInGeneric[string]()
	expect("siblingClosuresInGeneric[int].C#1 accepts own", sgIAC(sgIA))
	expect("siblingClosuresInGeneric[int].C#2 accepts own", sgIBC(sgIB))
	expect("siblingClosuresInGeneric[int].C#1 rejects C#2", !sgIAC(sgIB))
	expect("siblingClosuresInGeneric[int].C#1 rejects [string].C#1", !sgIAC(sgSA))
	expect("siblingClosuresInGeneric[string].C#1 rejects [int].C#1", !sgSAC(sgIA))

	// Doubly-nested closure inside a generic function.
	dni, dniC := doublyNestedInGeneric[int]()
	dns, dnsC := doublyNestedInGeneric[string]()
	expect("doublyNestedInGeneric[int].Y accepts own", dniC(dni))
	expect("doublyNestedInGeneric[string].Y accepts own", dnsC(dns))
	expect("doublyNestedInGeneric[int].Y rejects [string].Y", !dniC(dns))
	expect("doublyNestedInGeneric[string].Y rejects [int].Y", !dnsC(dni))
}
