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
	println("v7:", cap(v7), string(v7))
	println("v8:", v8)
	println("v9:", len(v9), v9[0], v9[1], v9[2])

	println(uint8SliceSrc[0])
	println(uint8SliceDst[0])
	println(intSliceSrc[0])
	println(intSliceDst[0])
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
	v7 = []byte("foo")
	v8 string
	v9 []int

	uint8SliceSrc = []uint8{3, 100}
	uint8SliceDst []uint8
	intSliceSrc   = []int16{5, 123, 1024}
	intSliceDst   []int16

	someList    *linkedList
	someBigList *bigLinkedList
)

type linkedList struct {
	prev *linkedList
	next *linkedList
	v    int // arbitrary value (don't care)
}

func init() {
	someList = &linkedList{
		v: -1,
	}
	for i := 0; i < 3; i++ {
		prev := someList
		someList = &linkedList{
			v:    i,
			prev: prev,
		}
		prev.next = someList
	}
}

type bigLinkedList struct {
	prev *bigLinkedList
	next *bigLinkedList
	v    int
	buf  [100]*int
}

func init() {
	// Create a circular reference.
	someBigList = &bigLinkedList{
		v: -1,
	}
	for i := 0; i < 3; i++ {
		prev := someBigList
		someBigList = &bigLinkedList{
			v:    i,
			prev: prev,
		}
		prev.next = someBigList
	}
}

func init() {
	uint8SliceDst = make([]uint8, len(uint8SliceSrc))
	copy(uint8SliceDst, uint8SliceSrc)

	intSliceDst = make([]int16, len(intSliceSrc))
	copy(intSliceDst, intSliceSrc)

	v8 = sliceString("foobarbaz", 3, 8)
	v9 = sliceSlice([]int{0, 1, 2, 3, 4, 5}, 2, 5)
}

func sliceString(s string, start, end int) string {
	return s[start:end]
}

func sliceSlice(s []int, start, end int) []int {
	return s[start:end]
}

type outside struct{}

func init() {
	_, _ = any(0).(interface{ DoesNotExist() })
	_, _ = any("").(interface{ DoesNotExist() })
	_, _ = any(outside{}).(interface{ DoesNotExist() })

	type inside struct{}
	_, _ = any(inside{}).(interface{ DoesNotExist() })
}
