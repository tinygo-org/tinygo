package main

import (
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"
)

var rngState uint64 = 0x9e3779b97f4a7c15

func rnd() uint64 {
	rngState ^= rngState << 13
	rngState ^= rngState >> 7
	rngState ^= rngState << 17
	return rngState
}

func check(name string, ok bool) {
	if !ok {
		println("FAIL", name)
		panic(name)
	}
}

// ref builds a string without using string(b).
func ref(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		sb.WriteByte(c)
	}
	return sb.String()
}

//go:noinline
func lookup(m map[string]int, b []byte) (int, bool) {
	v, ok := m[string(b)]
	return v, ok
}

//go:noinline
func lookupDirect(m map[string]int, b []byte) int {
	return m[string(b)]
}

//go:noinline
func runeCount(b []byte) int {
	return utf8.RuneCountInString(string(b))
}

//go:noinline
func mutate(b []byte) int {
	for i := range b {
		b[i] ^= 0x55
	}
	return len(b)
}

//go:noinline
func lenAndLookup(m map[string]int, s string, n int) int {
	return m[s] + n
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rnd())
	}
	return b
}

// Differential test against a lookup that cannot use the optimization.
func testMapDifferential() {
	m := map[string]int{}
	var keys [][]byte
	for i := 0; i < 100; i++ {
		k := randBytes(int(rnd() % 40))
		keys = append(keys, k)
		m[ref(k)] = i + 1
	}
	b := make([]byte, 0, 64)
	for i := 0; i < 5000; i++ {
		b = b[:0]
		if rnd()%2 == 0 {
			b = append(b, keys[rnd()%uint64(len(keys))]...)
			if rnd()%3 == 0 && len(b) > 0 {
				b[rnd()%uint64(len(b))] ^= 1
			}
		} else {
			b = append(b, randBytes(int(rnd()%40))...)
		}
		want, wantOK := m[ref(b)]
		got, gotOK := lookup(m, b)
		check("mapDifferential", got == want && gotOK == wantOK)
		check("mapDifferentialDirect", lookupDirect(m, b) == want)
	}
	println("ok mapDifferential")
}

func testRuneCountDifferential() {
	for i := 0; i < 5000; i++ {
		b := randBytes(int(rnd() % 100))
		if i%3 == 0 {
			// Bias toward multibyte sequences.
			for j := range b {
				if b[j]&0x80 == 0 && rnd()%2 == 0 {
					b[j] |= 0xc0
				}
			}
		}
		check("runeCount", runeCount(b) == utf8.RuneCount(b))
	}
	println("ok runeCountDifferential")
}

// A write after the call must not affect the result of an earlier call, and
// the lookup must see the current bytes each time.
func testMutationBetweenCalls() {
	m := map[string]int{"abcd": 1, "abcd" + "\x00": 2}
	b := []byte("abcd")
	for i := 0; i < 1000; i++ {
		v, ok := lookup(m, b)
		if i%2 == 0 {
			check("mutBetween even", ok && v == 1)
		} else {
			check("mutBetween odd", !ok)
		}
		mutate(b)
	}
	println("ok mutationBetweenCalls")
}

// The conversion is evaluated before later arguments, which mutate the bytes.
func testArgumentOrder() {
	m := map[string]int{"hello": 7}
	b := []byte("hello")
	got := lenAndLookup(m, string(b), mutate(b))
	check("argumentOrder", got == 7+5)
	check("argumentOrder after", string(b) != "hello")

	b = []byte("hello")
	got = m[string(b)] + mutate(b)
	check("argumentOrder map", got == 7+5)
	println("ok argumentOrder")
}

// The string outlives the call, so it must stay a snapshot.
func testStringOutlivesCall() {
	m := map[string]int{"abc": 1}
	b := []byte("abc")
	s := string(b)
	v1 := m[s]
	n1 := utf8.RuneCountInString(s)
	b[0] = 'z'
	check("outlives lookup", m[s] == v1 && v1 == 1)
	check("outlives count", utf8.RuneCountInString(s) == n1)
	check("outlives value", s == "abc")

	// Used in a call and also read later.
	b = []byte("xyz")
	t := string(b)
	n := runeCountString(t)
	b[1] = 'q'
	check("outlives later read", t == "xyz" && n == 3)
	println("ok stringOutlivesCall")
}

//go:noinline
func runeCountString(s string) int { return utf8.RuneCountInString(s) }

var sink string

// Stored keys and captured strings must be copies.
func testCapture() {
	m := map[string]int{}
	b := []byte("key1")
	m[string(b)] = 1
	b[3] = '2'
	m[string(b)] = 2
	b[3] = '3'
	check("capture set", len(m) == 2 && m["key1"] == 1 && m["key2"] == 2)

	b = []byte("global")
	sink = string(b)
	b[0] = 'G'
	check("capture global", sink == "global")

	// Callees that return a substring of their argument.
	b = []byte("  padded  ")
	trimmed := strings.TrimSpace(string(b))
	mutate(b)
	check("capture trim", trimmed == "padded")

	b = []byte("lower")
	low := strings.ToLower(string(b))
	mutate(b)
	check("capture tolower", low == "lower")

	b = []byte("a,b,c")
	parts := strings.Split(string(b), ",")
	mutate(b)
	check("capture split", len(parts) == 3 && parts[0] == "a" && parts[2] == "c")

	b = []byte("one two three")
	fields := strings.Fields(string(b))
	mutate(b)
	check("capture fields", len(fields) == 3 && fields[1] == "two")

	// strconv errors keep the input string.
	b = []byte("12x")
	_, err := strconv.Atoi(string(b))
	mutate(b)
	check("capture atoi", err != nil && strings.Contains(err.Error(), "12x"))
	println("ok capture")
}

// The result slot and the key are the same memory.
func testResultAliasesKey() {
	m := map[string][4]byte{
		"\x01\x02\x03\x04": {'a', 'b', 'c', 'd'},
		"abcd":             {'w', 'x', 'y', 'z'},
	}
	arr := [4]byte{1, 2, 3, 4}
	arr = m[string(arr[:])]
	check("aliasing step1", string(arr[:]) == "abcd")
	arr = m[string(arr[:])]
	check("aliasing step2", string(arr[:]) == "wxyz")
	arr = m[string(arr[:])]
	check("aliasing miss", arr == [4]byte{})

	type holder struct {
		key [4]byte
		val [4]byte
	}
	h := holder{key: [4]byte{1, 2, 3, 4}}
	h.val = m[string(h.key[:])]
	check("aliasing struct", string(h.val[:]) == "abcd" && h.key == [4]byte{1, 2, 3, 4})
	println("ok resultAliasesKey")
}

// Subslices, overlapping views and different capacities.
func testSubslices() {
	m := map[string]int{"bc": 1, "abc": 2, "": 3, "c": 4}
	buf := []byte("abcabc")
	for i := 0; i <= len(buf); i++ {
		for j := i; j <= len(buf); j++ {
			want := m[ref(buf[i:j])]
			check("subslice", lookupDirect(m, buf[i:j]) == want)
		}
	}
	var nilSlice []byte
	check("nil slice", lookupDirect(m, nilSlice) == 3)
	check("empty slice", lookupDirect(m, buf[:0]) == 3)
	check("nil runecount", runeCount(nilSlice) == 0)
	println("ok subslices")
}

func testLargeKeys() {
	big := make([]byte, 1<<12)
	for i := range big {
		big[i] = byte(i * 7)
	}
	m := map[string]int{ref(big): 99}
	check("large hit", lookupDirect(m, big) == 99)
	big[len(big)-1] ^= 1
	check("large miss", lookupDirect(m, big) == 0)
	big[len(big)-1] ^= 1
	check("large hit again", lookupDirect(m, big) == 99)
	check("large runecount", runeCount(big) == utf8.RuneCount(big))
	println("ok largeKeys")
}

// Each goroutine owns its buffer and mutates it between lookups.
func testGoroutines() {
	m := map[string]int{}
	for i := 0; i < 64; i++ {
		m[string([]byte{byte('a' + i%26), byte('0' + i/26)})] = i + 1
	}
	const workers = 4
	done := make(chan bool, workers)
	for w := 0; w < workers; w++ {
		go func(w int) {
			b := make([]byte, 2)
			ok := true
			for i := 0; i < 2000; i++ {
				idx := (i*7 + w) % 64
				b[0] = byte('a' + idx%26)
				b[1] = byte('0' + idx/26)
				if lookupDirect(m, b) != idx+1 {
					ok = false
				}
				if i%50 == 0 {
					// Yield so other workers interleave.
					done2 := make(chan bool, 1)
					done2 <- true
					<-done2
				}
				b[1] = 'Z'
				if lookupDirect(m, b) != 0 {
					ok = false
				}
			}
			done <- ok
		}(w)
	}
	for w := 0; w < workers; w++ {
		check("goroutine", <-done)
	}
	println("ok goroutines")
}

// Cases below must keep the copy. Each wrapper is noinline so the IR check in
// the transform tests can find the conversion by function name.

var alias []byte
var gate = make(chan bool)
var started = make(chan bool)

//go:noinline
func keepBlockingCallee(s string) int {
	started <- true
	<-gate
	return utf8.RuneCountInString(s)
}

//go:noinline
func keepMutatesViaGlobal(s string) int {
	alias[0] = 'a'
	return int(s[0])
}

//go:noinline
func keepMutatesViaParam(s string, p []byte) int {
	p[0] = 'a'
	return int(s[0])
}

//go:noinline
func keepMutatesViaFunc(s string, f func()) int {
	f()
	return int(s[0])
}

type mutator interface{ mut() }
type mutImpl struct{ b []byte }

func (m mutImpl) mut() { m.b[0] = 'a' }

//go:noinline
func keepMutatesViaIface(s string, m mutator) int {
	m.mut()
	return int(s[0])
}

//go:noinline
func keepPtrAsUintptr(s string) uintptr {
	return uintptr(unsafe.Pointer(unsafe.StringData(s)))
}

//go:noinline
func keepReturnsSubstring(s string) string { return s[1:] }

//go:noinline
func keepGoStmt(b []byte, out chan int) {
	go func(s string) { out <- int(s[0]) }(string(b))
	b[0] = 'z'
}

//go:noinline
func keepDefer(b []byte) (r int) {
	defer func(s string) { r = int(s[0]) }(string(b))
	b[0] = 'z'
	return 0
}

//go:noinline
func keepInlineStoreBetween(b []byte) int {
	return lenAndLookup(map[string]int{"abc": 1}, string(b), func() int { b[0] = 'z'; return 0 }())
}

//go:noinline
func keepStoreViaAlias(b []byte, q []byte) int {
	return lenAndLookup(map[string]int{"abc": 1}, string(b), func() int { q[0] = 'z'; return 0 }())
}

//go:noinline
func keepBuilder(b []byte) string {
	var sb strings.Builder
	sb.WriteString(string(b))
	b[0] = 'z'
	return sb.String()
}

//go:noinline
func keepAppend(dst []byte, b []byte) []byte {
	return append(dst, string(b)...)
}

//go:noinline
func keepIndirectCallee(s string, f func(string) int) int { return f(s) }

//go:noinline
func keepAtomicWritesSource(s string, p *int32) int {
	atomic.AddInt32(p, 0x01010101)
	return int(s[0])
}

//go:noinline
func keepAtomicCASSource(s string, p *int32) int {
	atomic.CompareAndSwapInt32(p, *p, 0x7a7a7a7a)
	return int(s[0])
}

//go:noinline
func convertThenCall(b []byte, f func(string) int) int { return f(string(b)) }

func testMustKeepCopy() {
	// The callee is only known at run time and mutates the bytes.
	w := []byte("xbc")
	check("keep indirect", convertThenCall(w, func(s string) int { w[0] = 'a'; return int(s[0]) }) == 'x')

	// The callee writes the source with atomics.
	var word int32 = 0x78787878
	wb := unsafe.Slice((*byte)(unsafe.Pointer(&word)), 4)
	check("keep atomic add", keepAtomicWritesSource(string(wb), &word) == 'x')
	word = 0x78787878
	check("keep atomic cas", keepAtomicCASSource(string(wb), &word) == 'x')

	// Blocking callee: another goroutine mutates the bytes while it waits.
	b := []byte("\xe6\x97\xa5")
	res := make(chan int)
	go func() { res <- keepBlockingCallee(string(b)) }()
	<-started
	copy(b, "abc")
	gate <- true
	check("keep blocking", <-res == 1)

	b = []byte("xbc")
	alias = b
	check("keep global", keepMutatesViaGlobal(string(b)) == 'x')
	b = []byte("xbc")
	check("keep param", keepMutatesViaParam(string(b), b) == 'x')
	b = []byte("xbc")
	check("keep func", keepMutatesViaFunc(string(b), func() { b[0] = 'a' }) == 'x')
	b = []byte("xbc")
	check("keep iface", keepMutatesViaIface(string(b), mutImpl{b}) == 'x')

	// A copy must have its own address.
	b = []byte("xbc")
	check("keep addr", keepPtrAsUintptr(string(b)) != uintptr(unsafe.Pointer(&b[0])))
	b = []byte("xbc")
	sub := keepReturnsSubstring(string(b))
	b[1] = 'z'
	check("keep substring", sub == "bc")

	b = []byte("xbc")
	out := make(chan int, 1)
	keepGoStmt(b, out)
	check("keep go", <-out == 'x')
	b = []byte("xbc")
	check("keep defer", keepDefer(b) == 'x')

	b = []byte("abc")
	check("keep inline store", keepInlineStoreBetween(b) == 1)
	b = []byte("abc")
	check("keep store alias", keepStoreViaAlias(b, b) == 1)
	b = []byte("abc")
	check("keep builder", keepBuilder(b) == "abc")
	b = []byte("abc")
	check("keep append", string(keepAppend(nil, b)) == "abc")
	// The destination of the append overlaps the source.
	buf := make([]byte, 3, 16)
	copy(buf, "abc")
	buf = keepAppend(buf[:0], buf)
	check("keep append overlap", string(buf) == "abc")
	println("ok mustKeepCopy")
}

func main() {
	testMapDifferential()
	testRuneCountDifferential()
	testMutationBetweenCalls()
	testArgumentOrder()
	testStringOutlivesCall()
	testCapture()
	testResultAliasesKey()
	testSubslices()
	testLargeKeys()
	testGoroutines()
	testMustKeepCopy()
}
