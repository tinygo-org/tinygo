// Interface type asserts with a type ID switch, on types whose descriptor byte 1
// is not a type ID.
package main

import (
	"errors"
	"reflect"
	"strconv"
)

type I interface{ M() int }
type Stringer interface{ String() string }
type J interface{ N() int }

type A int
type B struct{ x int }
type C []int
type D string
type E struct{ a, b int }
type F int
type G func(int) int
type H struct{}

func (a A) M() int  { return 1 }
func (b *B) M() int { return 2 }
func (c C) M() int  { return 3 }
func (d D) M() int  { return 4 }
func (e E) M() int  { return 5 }
func (f *F) M() int { return 6 }
func (g G) M() int  { return 7 }

func (a A) N() int { return 11 }
func (d D) N() int { return 14 }
func (h H) N() int { return 18 }
func (e E) N() int { return 15 }
func (c C) N() int { return 13 }

func isI(x any) bool { _, ok := x.(I); return ok }
func isJ(x any) bool { _, ok := x.(J); return ok }

func kind(x any) string {
	switch v := x.(type) {
	case I:
		return "I" + strconv.Itoa(v.M())
	case J:
		return "J" + strconv.Itoa(v.N())
	case Stringer:
		return "Stringer"
	case error:
		return "error"
	}
	return "none"
}

func main() {
	b := &B{1}
	pb := &b
	ppb := &pb
	pppb := &ppb
	f := F(1)
	vals := []struct {
		name string
		v    any
	}{
		{"nil", nil},
		{"A", A(1)}, {"*B", b}, {"C", C{1}}, {"D", D("x")}, {"E", E{}}, {"*F", &f}, {"G", G(func(int) int { return 0 })},
		{"*A", new(A)}, {"B", B{}}, {"F", f}, {"H", H{}}, {"*H", &H{}},
		{"**B", pb}, {"***B", ppb}, {"****B", pppb},
		{"int", 1}, {"*int", new(int)}, {"string", "s"}, {"[]int", []int{}},
		{"func()", func() {}},
		{"func(int)", func(int) {}},
		{"func(int,int)", func(int, int) {}},
		{"func(3)", func(int, int, int) {}},
		{"func(4)", func(int, int, int, int) {}},
		{"func(5)", func(int, int, int, int, int) {}},
		{"func(6)", func(int, int, int, int, int, int) {}},
		{"func(7)", func(int, int, int, int, int, int, int) {}},
		{"func(int)int", func(int) int { return 0 }},
		{"error", errors.New("e")},
		{"StructOf", reflect.New(reflect.StructOf([]reflect.StructField{{Name: "X", Type: reflect.TypeOf(0)}})).Elem().Interface()},
		{"FuncOf", reflect.Zero(reflect.FuncOf([]reflect.Type{reflect.TypeOf(A(0)), reflect.TypeOf(B{})}, nil, false)).Interface()},
		{"SliceOf", reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(E{})), 0, 0).Interface()},
		{"MapOf", reflect.MakeMap(reflect.MapOf(reflect.TypeOf(D("")), reflect.TypeOf(E{}))).Interface()},
		{"ArrayOf", reflect.New(reflect.ArrayOf(3, reflect.TypeOf(H{}))).Elem().Interface()},
		{"ChanOf", reflect.MakeChan(reflect.ChanOf(reflect.BothDir, reflect.TypeOf(F(0))), 0).Interface()},
		{"PointerTo(A)", reflect.New(reflect.TypeOf(A(0))).Interface()},
		{"PointerTo(B)", reflect.New(reflect.TypeOf(B{})).Interface()},
	}
	for _, v := range vals {
		println(v.name, isI(v.v), isJ(v.v), kind(v.v))
	}
}
