package mapbench

import "testing"

type compositeKey struct {
	S string
	N int32
}

var intSink int

func BenchmarkMapStringShortGet(b *testing.B) {
	m := make(map[string]int, 100)
	for i := 0; i < 100; i++ {
		m[string(rune('A'+i%26))+string(rune('a'+i/26))] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intSink += m["Qa"]
	}
}

func BenchmarkMapStringLongGet(b *testing.B) {
	m := make(map[string]int, 100)
	for i := 0; i < 100; i++ {
		s := "this-is-a-longer-key-for-testing-"
		for j := 0; j < 3; j++ {
			s += string(rune('A' + (i+j)%26))
		}
		m[s] = i
	}
	key := "this-is-a-longer-key-for-testing-ABC"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intSink += m[key]
	}
}

func BenchmarkMapCompositeGet(b *testing.B) {
	m := make(map[compositeKey]int, 100)
	for i := 0; i < 100; i++ {
		m[compositeKey{S: string(rune('A'+i%26)) + string(rune('a'+i/26)), N: int32(i)}] = i
	}
	key := compositeKey{S: "Qa", N: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intSink += m[key]
	}
}

func BenchmarkMapIntGet(b *testing.B) {
	m := make(map[int]int, 100)
	for i := 0; i < 100; i++ {
		m[i*7] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intSink += m[42]
	}
}

func BenchmarkMapCompositeSet(b *testing.B) {
	m := make(map[compositeKey]int, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[compositeKey{S: "key", N: int32(i)}] = i
	}
}

type bigKey [256]byte

func BenchmarkMapBigKeyGet(b *testing.B) {
	m := make(map[bigKey]int, 100)
	for i := 0; i < 100; i++ {
		var k bigKey
		k[0] = byte(i)
		m[k] = i
	}
	var k bigKey
	k[0] = 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intSink += m[k]
	}
}

func BenchmarkMapBigKeySet(b *testing.B) {
	m := make(map[bigKey]int, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var k bigKey
		k[0] = byte(i)
		k[1] = byte(i >> 8)
		k[2] = byte(i >> 16)
		k[3] = byte(i >> 24)
		m[k] = i
	}
}
