package main

// Test maps with keys and values larger than 128 bytes, which triggers
// indirect storage in the bucket (pointers instead of inline data).
//
// This is a separate file from map.go because the compiler generates many
// large stack temporaries for map operations on [256]byte keys, which
// overflows the goroutine stack on AVR (384 bytes). AVR skips this test.

type BigKey [256]byte
type BigValue [256]byte

func main() {
	// Large key, small value.
	m1 := make(map[BigKey]int)
	var k1 BigKey
	k1[0] = 1
	k1[255] = 42
	m1[k1] = 100

	var k1same BigKey
	k1same[0] = 1
	k1same[255] = 42

	var k1diff BigKey
	k1diff[0] = 2

	println("bigkey get:", m1[k1])
	println("bigkey get same:", m1[k1same])
	println("bigkey get diff:", m1[k1diff])

	// Overwrite.
	m1[k1] = 200
	println("bigkey overwrite:", m1[k1])

	// Small key, large value.
	m2 := make(map[int]BigValue)
	var v BigValue
	v[0] = 7
	v[255] = 99
	m2[1] = v
	got := m2[1]
	println("bigval get:", got[0], got[255])

	// Both large.
	m3 := make(map[BigKey]BigValue)
	m3[k1] = v
	got3 := m3[k1]
	println("bigboth get:", got3[0], got3[255])

	// Delete.
	delete(m3, k1)
	got3 = m3[k1]
	println("bigboth deleted:", got3[0])

	// Iteration.
	m1[k1diff] = 300
	count := 0
	for range m1 {
		count++
	}
	println("bigkey len:", len(m1), "iterated:", count)
}
