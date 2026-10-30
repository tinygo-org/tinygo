package main

func main() {
	var s []int
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	println(len(s), cap(s) < len(s)*13/10)
	b := make([]byte, 0, 4)
	b = append(b, 1, 2, 3, 4, 5)
	println(cap(b))
}
