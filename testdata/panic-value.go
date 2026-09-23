package main

// point does not implement error or fmt.Stringer, so printitf falls through
// to its default case, which must not allocate on the panic path.
type point struct {
	x, y int
}

func main() {
	defer func() {
		println("done")
	}()
	panic(point{1, 2})
}
