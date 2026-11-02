package main

var value int

func init() {
	defer func() {
		if recover() != nil {
			value = -1
		}
	}()
	initialize()
}

func initialize() {
	value = 42
}

func main() {
	println(value)
}
