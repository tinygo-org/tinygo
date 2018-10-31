package main

func init() {
	println("init")
}

func main() {
	println("main")
	println("v1:", v1)
	println("v2:", v2.x, v2.y)
	println("v3:", len(v3), cap(v3), v3[0], v3[3])
	println("v4:", len(v4), v4 == nil)
	println("v5:", len(v5), v5 == nil)
	println("v6:", v6)
}

type (
	t2 struct {
		x int
		y int
	}
)

var (
	v1 = 3
	v2 = t2{2, 5}
	v3 = []int{2, 3, 5, 7}
	v4 map[string]int
	v5 = map[string]int{}
	v6 = float64(v1) < 2.6
)
