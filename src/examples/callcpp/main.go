package main

/*
int add(int a, int b);
*/
import "C"
import "fmt"

func main() {
	a := C.add(C.int(1), C.int(2))
	fmt.Println(a)
}
