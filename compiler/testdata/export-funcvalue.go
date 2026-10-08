package main

//export exported
func exported(a int32) int32 {
	return a
}

var exportedValue func(int32) int32

func main() {
	exportedValue = exported
	println(exportedValue(1))
}
