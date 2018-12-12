package main

import "reflect"

type myint int

func main() {
	println("matching types")
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(int(5)))
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(uint(5)))
	println(reflect.TypeOf(myint(3)) == reflect.TypeOf(int(5)))
}
