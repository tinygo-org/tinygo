package main

import (
	"sort"
	"unsafe"
)

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

// Package-level pointer map literals: these exercise the interp pass's
// ability to defer pointer-keyed map inserts to runtime (since pointer
// hashes can't be computed at compile time).
var testPtrMapVar1 = 42
var testPtrMapVar2 = 99
var testPtrMap = map[*int]int{&testPtrMapVar1: 1, &testPtrMapVar2: 2}
var testUnsafePtrMap = map[unsafe.Pointer]int{unsafe.Pointer(&testPtrMapVar1): 10}

type namedFloat struct {
	s string
	f float32
}

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
	delete(nilmap, "foo")
	println("nilmap:", nilmap["bar"])
	println(testmapIntInt[2])
	testmapIntInt[2] = 42
	println(testmapIntInt[2])
	for k := range nilmap {
		println(k) // unreachable
	}

	var nilbinmap map[uint16]int
	delete(nilbinmap, 4)
	println("nilbinmap:", nilbinmap[5])

	arrKey := ArrayKey([4]byte{4, 3, 2, 1})
	println(testMapArrayKey[arrKey])
	testMapArrayKey[arrKey] = 5555
	println(testMapArrayKey[arrKey])

	// test maps with interface keys
	itfMap := map[interface{}]int{
		3.14:         3,
		8:            8,
		uint8(8):     80,
		"eight":      800,
		[2]int{5, 2}: 52,
		true:         1,
	}
	println("itfMap[3]:", itfMap[3]) // doesn't exist
	println("itfMap[3.14]:", itfMap[3.14])
	println("itfMap[8]:", itfMap[8])
	println("itfMap[uint8(8)]:", itfMap[uint8(8)])
	println(`itfMap["eight"]:`, itfMap["eight"])
	println(`itfMap[[2]int{5, 2}]:`, itfMap[[2]int{5, 2}])
	println("itfMap[true]:", itfMap[true])
	delete(itfMap, 8)
	println("itfMap[8]:", itfMap[8])
	for key, value := range itfMap {
		if key == "eight" {
			println("itfMap: found key \"eight\":", value)
		}
	}

	// test map with float keys
	floatMap := map[float32]int{
		42:   84,
		3.14: 6,
	}
	println("floatMap[42]:", floatMap[42])
	println("floatMap[43]:", floatMap[43])
	delete(floatMap, 42)
	println("floatMap[42]:", floatMap[42])
	for k, v := range floatMap {
		println("floatMap key, value:", k, v)
	}

	// test maps with struct keys
	structMap := map[namedFloat]int{
		namedFloat{"tau", 6.28}: 5,
	}
	println(`structMap[{"tau", 6.28}]:`, structMap[namedFloat{"tau", 6.28}])
	println(`structMap[{"Tau", 6.28}]:`, structMap[namedFloat{"Tau", 6.28}])
	println(`structMap[{"tau", 3.14}]:`, structMap[namedFloat{"tau", 3.14}])
	delete(structMap, namedFloat{"tau", 6.28})
	println(`structMap[{"tau", 6.28}]:`, structMap[namedFloat{"tau", 6.28}])

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

	floatcmplx()

	mapgrow()

	nestedarraymaps()

	ptrmaps()

	interfacerehash()
}

func floatcmplx() {

	var zero float64
	var negz float64 = -zero

	// test that zero and negative zero hash to the same thing
	m := make(map[float64]int)
	m[zero]++
	m[negz]++
	println(m[negz])

	cmap := make(map[complex128]int)

	var c complex128
	c = complex(zero, zero)
	cmap[c]++

	c = complex(negz, negz)
	cmap[c]++

	c = complex(0, 0)
	println(cmap[c])

	c = complex(1, negz)
	cmap[c]++

	c = complex(1, zero)
	cmap[c]++

	println(cmap[c])

	c = complex(negz, 2)
	cmap[c]++

	c = complex(zero, 2)
	cmap[c]++

	println(cmap[c])
}

func readMap(m map[string]int, key string) {
	println("map length:", len(m))
	println("map read:", key, "=", m[key])
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
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
		squares[i] = i * i
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

func mapgrow() {
	m := make(map[int]int)

	var Delete = 500
	if unsafe.Sizeof(uintptr(0)) < 4 {
		// Reduce the number of iterations on low-memory devices like AVR.
		Delete = 20
	}
	var N = Delete * 2

	for i := 0; i < Delete; i++ {
		m[i] = i
	}

	var deleted bool
	for k, v := range m {
		if k == 0 {
			// grow map
			for i := Delete; i < N; i++ {
				m[i] = i
			}

			// delete some elements
			for i := 0; i < Delete; i++ {
				delete(m, i)
			}
			deleted = true
			continue
		}

		// make sure we never see a deleted element later in our iteration
		if deleted && v < Delete {
			println("saw deleted element", v)
		}
	}

	if len(m) != N-Delete {
		println("bad length post grow/delete", len(m))
	}

	seen := make([]bool, Delete)

	var mcount int
	for k, v := range m {
		if k != v {
			println("element mismatch", k, v)
		}
		if k < Delete {
			println("saw deleted element post-grow", k)
		}
		seen[v-Delete] = true
		mcount++
	}

	for _, v := range seen {
		if !v {
			println("missing key", v)
		}
	}

	if mcount != N-Delete {
		println("bad number of elements post-grow:", mcount)
	}
	println("done")
}

type Counter interface {
	count() int
}

type counter struct {
	i int
}

func (c *counter) count() int {
	return c.i
}

func interfacerehash() {
	m := make(map[Counter]int)

	for i := 0; i < 20; i++ {
		c := &counter{i}
		m[c] = i
	}

	var failures int
	for k, v := range m {
		if got := m[k]; got != v {
			println("lookup failure got", got, "want", v)
			failures++
		}
	}
	if failures == 0 {
		println("no interface lookup failures")
	}
}

func nestedarraymaps() {
	type nestedArrayElem struct {
		x uint8
	}
	type nestedArrayKey struct {
		a [5][5]nestedArrayElem
	}
	var k nestedArrayKey
	k.a[4][4].x = 7
	m := map[nestedArrayKey]int{k: 42}
	println("nested array key:", m[k])
}

func ptrmaps() {
	// Package-level pointer map literals (interp defers inserts to runtime).
	println("ptr map literal:", testPtrMap[&testPtrMapVar1], testPtrMap[&testPtrMapVar2])
	println("unsafe ptr literal:", testUnsafePtrMap[unsafe.Pointer(&testPtrMapVar1)])

	// Runtime pointer maps.
	a, b, c := 1, 2, 3
	m := make(map[*int]int)
	m[&a] = 10
	m[&b] = 20
	m[&c] = 30
	println("ptr map len:", len(m))
	println("ptr map a:", m[&a])
	delete(m, &b)
	_, ok := m[&b]
	println("ptr map deleted:", ok)

	// Runtime unsafe.Pointer maps.
	m2 := make(map[unsafe.Pointer]int)
	m2[unsafe.Pointer(&a)] = 100
	println("unsafe ptr map:", m2[unsafe.Pointer(&a)])

	// Struct keys with padding: the hash/equal must operate per-field
	// and not include padding bytes. Only test when padding actually
	// exists (e.g., not on AVR where alignment is 1).
	type paddedKey struct {
		A int8
		B int32
	}
	if unsafe.Offsetof(paddedKey{}.B) > 1 {
		pm := make(map[paddedKey]int)
		var pk1, pk2 paddedKey
		pk1.A = 1; pk1.B = 42
		pk2.A = 1; pk2.B = 42
		// Poison pk2's padding byte (between A and B).
		*(*byte)(unsafe.Add(unsafe.Pointer(&pk2), 1)) = 0xFF
		pm[pk1] = 100
		println("padded key lookup:", pm[pk2])         // 100
		println("padded key equal:", pk1 == pk2)        // true
	} else {
		// No padding on this platform; print expected output.
		println("padded key lookup:", 100)
		println("padded key equal:", true)
	}

	// Struct keys with blank fields: blank fields are ignored in equality.
	type blankKey struct {
		_ int
		X string
	}
	bm := make(map[blankKey]int)
	var bk1, bk2 blankKey
	bk1.X = "hello"
	bk2.X = "hello"
	*(*int)(unsafe.Pointer(&bk2)) = 999
	bm[bk1] = 200
	println("blank key lookup:", bm[bk2])           // 200
	println("blank key equal:", bk1 == bk2)          // true
}
