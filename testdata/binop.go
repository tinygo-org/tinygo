package main

func main() {
	println("string equality")
	println(a == "a")
	println(a == "b")
	println(a != "a")
	println(a != "b")
	println("string inequality")
	println(a > "b")
	println("b" > a)
	println("b" > "b")
	println(a <= "b")
	println("b" <= a)
	println("b" <= "b")
	println(a > "b")
	println("b" > a)
	println("b" > "b")
	println(a <= "b")
	println("b" <= a)
	println("b" <= "b")
	println(a < "aa")
	println("aa" < a)
	println("ab" < "aa")
	println("aa" < "ab")

	h := "hello"
	println("h < h", h < h)
	println("h <= h", h <= h)
	println("h == h", h == h)
	println("h >= h", h >= h)
	println("h > h", h > h)

	println("array equality")
	println(a1 == [2]int{1, 2})
	println(a1 != [2]int{1, 2})
	println(a1 == [2]int{1, 3})
	println(a1 == [2]int{2, 2})
	println(a1 == [2]int{2, 1})
	println(a1 != [2]int{2, 1})

	println("struct equality")
	println(s1 == Struct1{3, true})
	println(s1 == Struct1{4, true})
	println(s1 == Struct1{3, false})
	println(s1 == Struct1{4, false})
	println(s1 != Struct1{3, true})
	println(s1 != Struct1{4, true})
	println(s1 != Struct1{3, false})
	println(s1 != Struct1{4, false})

	println("blank fields in structs")
	println(s2 == Struct2{"foo", 0.0, 5})
	println(s2 == Struct2{"foo", 0.0, 7})
	println(s2 == Struct2{"foo", 1.0, 5})
	println(s2 == Struct2{"foo", 1.0, 7})

	println("complex numbers")
	println(c64 == 3+2i)
	println(c64 == 4+2i)
	println(c64 == 3+3i)
	println(c64 != 3+2i)
	println(c64 != 4+2i)
	println(c64 != 3+3i)
	println(c128 == 3+2i)
	println(c128 == 4+2i)
	println(c128 == 3+3i)
	println(c128 != 3+2i)
	println(c128 != 4+2i)
	println(c128 != 3+3i)

	println("shifts")
	println(shlSimple == 4)
	println(shlOverflow == 0)
	println(shrSimple == 1)
	println(shrOverflow == 0)
	println(ashrNeg == -1)
	println(ashrOverflow == 0)
	println(ashrNegOverflow == -1)

	// fix for a bug in constant numbers
	println("constant number")
	x := uint32(5)
	println(uint32(x) / (20e0 / 1))

	// check for signed integer overflow
	println("-2147483648 / -1:", sdiv32(-2147483648, -1))
	println("-2147483648 % -1:", srem32(-2147483648, -1))
}

var x = true
var y = false

var a = "a"
var s1 = Struct1{3, true}
var s2 = Struct2{"foo", 0.0, 5}

var a1 = [2]int{1, 2}

var c64 = 3 + 2i
var c128 = 4 + 3i

type Int int

type Struct1 struct {
	i Int
	b bool
}

type Struct2 struct {
	s string
	_ float64
	i int
}

func shl(x uint, y uint) uint {
	return x << y
}

func shr(x uint, y uint) uint {
	return x >> y
}

func ashr(x int, y uint) int {
	return x >> y
}

func sdiv32(x, y int32) int32 {
	return x / y
}

func srem32(x, y int32) int32 {
	return x % y
}

var shlSimple = shl(2, 1)
var shlOverflow = shl(2, 1000)
var shrSimple = shr(2, 1)
var shrOverflow = shr(2, 1000000)
var ashrNeg = ashr(-1, 1)
var ashrOverflow = ashr(1, 1000000)
var ashrNegOverflow = ashr(-1, 1000000)
