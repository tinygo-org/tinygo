//go:build tinygo

package stackbuffer

import (
	"strconv"
	"strings"
	"testing"
)

var (
	sinkInt       int
	globalString  string
	globalStrings []string
	globalAny     any
	globalMap     = map[string]int{}
)

// clobber overwrites the stack below the caller.
//
//go:noinline
func clobber(n int) int {
	var pad [256]byte
	for i := range pad {
		pad[i] = 0xAA
	}
	if n > 0 {
		return clobber(n-1) + int(pad[n&127])
	}
	return int(pad[0])
}

func bytesOfLen(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + i%26)
	}
	return b
}

func sumOf(b []byte) int {
	n := 0
	for _, c := range b {
		n += int(c)
	}
	return n
}

//go:noinline
func readString(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func returnString(b []byte) string { return string(b) }

// The bytes change before the string is used, so the copy is needed. It uses
// the stack up to 32 bytes and the heap above that.
func TestAllocations(t *testing.T) {
	for _, n := range []int{0, 1, 5, 31, 32, 33, 100} {
		b := bytesOfLen(n)
		orig := bytesOfLen(n)
		want := sumOf(orig)
		wantAllocs := 0.0
		if n > 32 {
			wantAllocs = 1
		}
		var got int
		allocs := testing.AllocsPerRun(50, func() {
			s := string(b)
			for i := range b {
				b[i] = 'x'
			}
			got = readString(s)
			copy(b, orig)
		})
		if got != want {
			t.Errorf("length %d: got %d, want %d", n, got, want)
		}
		if allocs != wantAllocs {
			t.Errorf("length %d: got %v allocations, want %v", n, allocs, wantAllocs)
		}
	}
}

func TestStdlibCallsDoNotAllocate(t *testing.T) {
	number := []byte("123456789")
	text := []byte("hello, world!")
	var n int
	var f float64
	var ok bool
	var err error
	allocs := testing.AllocsPerRun(100, func() {
		n, err = strconv.Atoi(string(number))
		f, _ = strconv.ParseFloat(string(number), 64)
		ok = strings.Contains(string(text), "world")
		sinkInt = n + int(f)
	})
	if allocs != 0 {
		t.Errorf("got %v allocations, want 0", allocs)
	}
	if n != 123456789 || f != 123456789 || !ok || err != nil {
		t.Errorf("got %d %v %v %v", n, f, ok, err)
	}
}

// The error of Atoi keeps a copy of the string.
func TestErrorKeepsString(t *testing.T) {
	b := []byte("12x4")
	_, err := strconv.Atoi(string(b))
	b[0] = 'Z'
	clobber(8)
	if got, want := err.Error(), `strconv.Atoi: parsing "12x4": invalid syntax`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringIsACopy(t *testing.T) {
	b := []byte("hello")
	s := string(b)
	b[0] = 'J'
	if s != "hello" {
		t.Errorf("string changed with the bytes: %q", s)
	}
	sinkInt = readString(s)

	c := []byte("world")
	got := readAndChange(string(c), c)
	if got != "world" {
		t.Errorf("string changed with the bytes: %q", got)
	}
}

//go:noinline
func readAndChange(s string, b []byte) string {
	for i := range b {
		b[i] = 'X'
	}
	return strings.Clone(s)
}

//go:noinline
func storeInGlobal(b []byte) { globalString = string(b) }

//go:noinline
func storeInInterface(b []byte) { globalAny = string(b) }

//go:noinline
func appendToKept(b []byte) { globalStrings = append(globalStrings, string(b)) }

//go:noinline
func putInMap(b []byte) { globalMap[string(b)] = len(b) }

//go:noinline
func captureInClosure(b []byte) func() string {
	s := string(b)
	return func() string { return s }
}

//go:noinline
func sendOnChannel(b []byte, ch chan string) { ch <- string(b) }

//go:noinline
func storeInStruct(b []byte, p *struct{ name string }) { p.name = string(b) }

// A string that escapes must be a heap copy that survives the stack.
func TestEscapingStrings(t *testing.T) {
	b := []byte("escape-through-here")

	storeInGlobal(b)
	r := returnString(b)
	storeInInterface(b)
	appendToKept(b)
	putInMap(b)
	f := captureInClosure(b)
	ch := make(chan string, 1)
	sendOnChannel(b, ch)
	var st struct{ name string }
	storeInStruct(b, &st)

	for i := range b {
		b[i] = '#'
	}
	clobber(8)

	const want = "escape-through-here"
	if globalString != want {
		t.Errorf("global: %q", globalString)
	}
	if r != want {
		t.Errorf("returned: %q", r)
	}
	if globalAny.(string) != want {
		t.Errorf("interface: %q", globalAny)
	}
	if got := globalStrings[len(globalStrings)-1]; got != want {
		t.Errorf("slice: %q", got)
	}
	if _, ok := globalMap[want]; !ok {
		t.Errorf("map key is missing")
	}
	if got := f(); got != want {
		t.Errorf("closure: %q", got)
	}
	if got := <-ch; got != want {
		t.Errorf("channel: %q", got)
	}
	if st.name != want {
		t.Errorf("struct: %q", st.name)
	}
}

// Strings made in a loop and kept must not share a buffer.
func TestStringsKeptFromLoop(t *testing.T) {
	var kept [8]string
	allocs := testing.AllocsPerRun(20, func() {
		b := make([]byte, 5)
		for i := range kept {
			for j := range b {
				b[j] = byte('a' + i)
			}
			kept[i] = string(b)
		}
	})
	if allocs < 8 {
		t.Errorf("got %v allocations, want at least 8", allocs)
	}
	clobber(8)
	for i, s := range kept {
		if want := strings.Repeat(string(rune('a'+i)), 5); s != want {
			t.Errorf("string %d is %q, want %q", i, s, want)
		}
	}
}

// The string of the previous iteration is compared with the next one.
func TestPreviousStringInLoop(t *testing.T) {
	b := make([]byte, 6)
	prev := ""
	same := 0
	for i := 0; i < 6; i++ {
		for j := range b {
			b[j] = byte('a' + i/2)
		}
		cur := string(b)
		if cur == prev {
			same++
		}
		prev = cur
	}
	clobber(8)
	if same != 3 {
		t.Errorf("got %d equal neighbours, want 3", same)
	}
	if prev != "cccccc" {
		t.Errorf("got %q, want cccccc", prev)
	}
}

func TestLastStringAfterLoop(t *testing.T) {
	var last string
	b := make([]byte, 4)
	for i := 0; i < 5; i++ {
		for j := range b {
			b[j] = byte('A' + i)
		}
		last = string(b)
	}
	b[0] = '!'
	clobber(8)
	if last != "EEEE" {
		t.Errorf("got %q, want EEEE", last)
	}
}

//go:noinline
func concatenate(b []byte) string { return string(b) + "-suffix" }

func TestConcatenatedString(t *testing.T) {
	b := []byte("abc")
	s := concatenate(b)
	b[0] = 'X'
	clobber(8)
	if s != "abc-suffix" {
		t.Errorf("got %q", s)
	}
}

// Lengths on both sides of the buffer size, with the contents checked.
func TestLengths(t *testing.T) {
	for n := 0; n <= 70; n++ {
		b := bytesOfLen(n)
		s := returnString(b)
		for i := range b {
			b[i] = '?'
		}
		clobber(4)
		if want := string(bytesOfLen(n)); s != want {
			t.Fatalf("length %d: got %q, want %q", n, s, want)
		}
		if got := readString(string(bytesOfLen(n))); got != sumOf(bytesOfLen(n)) {
			t.Fatalf("length %d: read %d", n, got)
		}
	}
}
