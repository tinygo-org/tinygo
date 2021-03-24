package main

import (
	"encoding/json"
)

func main() {
	println("int:", encode(3))
	println("float64:", encode(3.14))
	println("string:", encode("foo"))
	println("slice of strings:", encode([]string{"foo", "bar"}))
}

func encode(itf interface{}) string {
	buf, err := json.Marshal(itf)
	if err != nil {
		panic("failed to JSON encode: " + err.Error())
	}
	return string(buf)
}
