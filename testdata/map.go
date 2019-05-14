package main

var testmap1 = map[string]int{"data": 3}
var testmap2 = map[string]int{
	"one":    1,
	"two":    2,
	"three":  3,
	"four":   4,
	"five":   5,
	"six":    6,
	"seven":  7,
	"eight":  8,
	"nine":   9,
	"ten":    10,
	"eleven": 11,
	"twelve": 12,
}

type ArrayKey [4]byte

var testMapArrayKey = map[ArrayKey]int{
	ArrayKey([4]byte{1, 2, 3, 4}): 1234,
	ArrayKey([4]byte{4, 3, 2, 1}): 4321,
}
var testmapIntInt = map[int]int{1: 1, 2: 4, 3: 9}

func main() {
	m := map[string]int{"answer": 42, "foo": 3}
	readMap(m, "answer")
	readMap(testmap1, "data")
	readMap(testmap2, "three")
	readMap(testmap2, "ten")
	delete(testmap2, "six")
	readMap(testmap2, "seven")
	lookup(testmap2, "eight")
	lookup(testmap2, "nokey")

	// operations on nil maps
	var nilmap map[string]int
	println(m == nil, m != nil, len(m))
	println(nilmap == nil, nilmap != nil, len(nilmap))
	println(testmapIntInt[2])
	testmapIntInt[2] = 42
	println(testmapIntInt[2])

	arrKey := ArrayKey([4]byte{4, 3, 2, 1})
	println(testMapArrayKey[arrKey])
	testMapArrayKey[arrKey] = 5555
	println(testMapArrayKey[arrKey])

	// test preallocated map
	squares := make(map[int]int, 200)
	testBigMap(squares, 100)
	println("tested preallocated map")

	// test growing maps
	squares = make(map[int]int, 0)
	testBigMap(squares, 10)
	squares = make(map[int]int, 20)
	testBigMap(squares, 40)
	println("tested growing of a map")
}

func readMap(m map[string]int, key string) {
	println("map length:", len(m))
	println("map read:", key, "=", m[key])
	for k, v := range m {
		println(" ", k, "=", v)
	}
}

func lookup(m map[string]int, key string) {
	value, ok := m[key]
	println("lookup with comma-ok:", key, value, ok)
}

func testBigMap(squares map[int]int, n int) {
	for i := 0; i < n; i++ {
		if len(squares) != i {
			println("unexpected length:", len(squares), "at i =", i)
		}
		squares[i] = i*i
		for j := 0; j <= i; j++ {
			if v, ok := squares[j]; !ok || v != j*j {
				if !ok {
					println("key not found in squares map:", j)
				} else {
					println("unexpected value read back from squares map:", j, v)
				}
				return
			}
		}
	}
}
