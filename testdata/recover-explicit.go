package main

func main() {
	catch()
	println("done")
}

func catch() {
	defer func() {
		println("recovered:", recover() == "panic")
	}()
	call()
	println("unreachable after call")
}

func call() {
	panic("panic")
}
