package main

func main() {

	ch := make(chan bool, 1)
	println("1")
	go func() {
		println("2")
		ch <- true
		println("3")
	}()
	println("4")
	v := <-ch
	println(v)

}
