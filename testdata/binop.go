package main

func main() {
	// string equality
	println(a == "a")
	println(a == "b")
	println(a != "a")
	println(a != "b")

	// struct equality
	println(s1 == Struct1{3, true})
	println(s1 == Struct1{4, true})
	println(s1 == Struct1{3, false})
	println(s1 == Struct1{4, false})
	println(s1 != Struct1{3, true})
	println(s1 != Struct1{4, true})
	println(s1 != Struct1{3, false})
	println(s1 != Struct1{4, false})

	// blank fields in structs
	println(s2 == Struct2{"foo", 0.0, 5})
	println(s2 == Struct2{"foo", 0.0, 7})
	println(s2 == Struct2{"foo", 1.0, 5})
	println(s2 == Struct2{"foo", 1.0, 7})
}

var x = true
var y = false

var a = "a"
var s1 = Struct1{3, true}
var s2 = Struct2{"foo", 0.0, 5}

type Struct1 struct {
	i int
	b bool
}

type Struct2 struct {
	s string
	_ float64
	i int
}
