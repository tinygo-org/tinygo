package main

// TODO: add .ll test files to check the output

type s0 struct {
}

type s1 struct {
	a byte
}

type s2 struct {
	a byte
	b byte
}

type s3 struct {
	a byte
	b byte
	c byte
}

// should not be expanded
type s4 struct {
	a byte
	b byte
	c byte
	d byte
}

type s5 struct {
	a struct {
		aa byte
		ab byte
	}
	b byte
}

type s6 struct {
	a string
	b byte
}

type s7 struct {
	a interface{}
	b byte
}

// should not be expanded
type s8 struct {
	a []byte // 3 elements
	b byte   // 1 element
}

func test0(s s0) {
	println("test0")
}

func test1(s s1) {
	println("test1", s.a)
}

func test2(s s2) {
	println("test2", s.a, s.b)
}

func test3(s s3) {
	println("test3", s.a, s.b, s.c)
}

func test4(s s4) {
	println("test4", s.a, s.b, s.c, s.d)
}

func test5(s s5) {
	println("test5", s.a.aa, s.a.ab, s.b)
}

func test6(s s6) {
	println("test6", s.a, len(s.a), s.b)
}

func test7(s s7) {
	println("test7", s.a, s.b)
}

func test8(s s8) {
	println("test8", len(s.a), cap(s.a), s.a[0], s.a[1], s.b)
}

func main() {
	test0(s0{})
	test1(s1{1})
	test2(s2{1, 2})
	test3(s3{1, 2, 3})
	test4(s4{1, 2, 3, 4})
	test5(s5{a: struct {
		aa byte
		ab byte
	}{1, 2}, b: 3})
	test6(s6{"foo", 5})
	test7(s7{a: nil, b: 8})
	test8(s8{[]byte{12, 13, 14}[:2], 6})
}
