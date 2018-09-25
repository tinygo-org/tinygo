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

type s9 struct {
}

func test1(s s1) {
	println("test1")
}

func test2(s s2) {
	println("test2")
}

func test3(s s3) {
	println("test3")
}

func test4(s s4) {
	println("test4")
}

func test5(s s5) {
	println("test5")
}

func test6(s s6) {
	println("test6")
}

func test7(s s7) {
	println("test7")
}

func test8(s s8) {
	println("test8")
}

func test9(s s9) {
	println("test9")
}

func main() {
	test1(s1{})
	test2(s2{})
	test3(s3{})
	test4(s4{})
	test5(s5{})
	test6(s6{})
	test7(s7{})
	test8(s8{})
	test9(s9{})
}
