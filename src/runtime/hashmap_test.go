package runtime_test

import (
	"maps"
	"math"
	"strconv"
	"testing"
)

func checkClone[K comparable, V comparable](t *testing.T, m map[K]V) {
	t.Helper()
	want := make(map[K]V, len(m))
	for k, v := range m {
		want[k] = v
	}
	c := maps.Clone(m)
	if len(c) != len(want) {
		t.Fatalf("len(clone) = %d, want %d", len(c), len(want))
	}
	for k, v := range want {
		cv, ok := c[k]
		if !ok {
			t.Fatalf("clone missing key %v", k)
		}
		if cv != v {
			t.Fatalf("clone[%v] = %v, want %v", k, cv, v)
		}
	}
	for k := range c {
		delete(c, k)
	}
	if len(c) != 0 {
		t.Fatalf("len(clone) = %d after deleting all keys, want 0", len(c))
	}
	if len(m) != len(want) {
		t.Fatalf("len(original) = %d after deleting from clone, want %d", len(m), len(want))
	}
	for k, v := range want {
		mv, ok := m[k]
		if !ok {
			t.Fatalf("original missing key %v after deleting from clone", k)
		}
		if mv != v {
			t.Fatalf("original[%v] = %v after deleting from clone, want %v", k, mv, v)
		}
	}
}

func TestMapCloneSizes(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, 9, 63, 64, 65, 1000, 10000} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			m := make(map[int]int)
			for i := range n {
				m[i] = i * 3
			}
			checkClone(t, m)
		})
	}
}

func TestMapCloneAfterDelete(t *testing.T) {
	for _, n := range []int{8, 9, 64, 1000, 10000} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			m := make(map[int]int)
			for i := range n {
				m[i] = i
			}
			for i := 0; i < n; i += 3 {
				delete(m, i)
			}
			checkClone(t, m)
		})
	}
}

func TestMapCloneSizeHint(t *testing.T) {
	// A large size hint with few entries leaves most buckets empty.
	m := make(map[int]int, 4096)
	for i := range 5 {
		m[i*1000] = i
	}
	checkClone(t, m)
}

func TestMapCloneStrings(t *testing.T) {
	m := make(map[string]string)
	for i := range 5000 {
		m["key"+strconv.Itoa(i)] = strconv.Itoa(i)
	}
	checkClone(t, m)
}

func TestMapCloneInterfaceKeys(t *testing.T) {
	m := make(map[any]int)
	for i := range 500 {
		m[i] = i
		m[strconv.Itoa(i)] = i
	}
	checkClone(t, m)
}

// bigValue is larger than hashmapMaxKeySize and hashmapMaxValueSize so keys
// and values are stored out of line.
type bigValue struct {
	a [40]int64
}

func TestMapCloneIndirect(t *testing.T) {
	m := make(map[bigValue]bigValue)
	for i := range 300 {
		var k, v bigValue
		k.a[0], k.a[39] = int64(i), int64(-i)
		v.a[5] = int64(i * 7)
		m[k] = v
	}
	checkClone(t, m)

	c := maps.Clone(m)
	for k, v := range c {
		v.a[5] = -1
		c[k] = v
	}
	if len(m) != 300 {
		t.Fatalf("len(original) = %d after writing to clone, want 300", len(m))
	}
	for i := range 300 {
		var k, want bigValue
		k.a[0], k.a[39] = int64(i), int64(-i)
		want.a[5] = int64(i * 7)
		v, ok := m[k]
		if !ok {
			t.Fatalf("original missing key %d after writing to clone", i)
		}
		if v != want {
			t.Fatalf("original[%d] = %v after writing to clone, want %v", i, v, want)
		}
	}
}

func TestMapCloneNaN(t *testing.T) {
	m := make(map[float64]int)
	for i := range 20 {
		m[math.NaN()] = i
		m[float64(i)] = i
	}
	c := maps.Clone(m)
	if len(c) != len(m) {
		t.Fatalf("len(clone) = %d, want %d", len(c), len(m))
	}
	nan := 0
	for k := range c {
		if k != k {
			nan++
		}
	}
	if nan != 20 {
		t.Fatalf("clone has %d NaN keys, want 20", nan)
	}
}

func TestMapCloneNil(t *testing.T) {
	var m map[int]int
	if c := maps.Clone(m); c != nil {
		t.Fatalf("maps.Clone(nil) = %v, want nil", c)
	}
}

func TestMapCloneGrow(t *testing.T) {
	// Inserting into a clone must grow it independently of the original.
	m := make(map[int]int)
	for i := range 100 {
		m[i] = i
	}
	c := maps.Clone(m)
	for i := 100; i < 10000; i++ {
		c[i] = i
	}
	if len(c) != 10000 {
		t.Fatalf("len(clone) = %d after growing clone, want 10000", len(c))
	}
	for i := range 10000 {
		v, ok := c[i]
		if !ok {
			t.Fatalf("clone missing key %d after growing clone", i)
		}
		if v != i {
			t.Fatalf("clone[%d] = %d after growing clone, want %d", i, v, i)
		}
	}
	if len(m) != 100 {
		t.Fatalf("len(original) = %d after growing clone, want 100", len(m))
	}
	for i := range 100 {
		v, ok := m[i]
		if !ok {
			t.Fatalf("original missing key %d after growing clone", i)
		}
		if v != i {
			t.Fatalf("original[%d] = %d after growing clone, want %d", i, v, i)
		}
	}
	checkClone(t, c)
}

var clonedMap map[int]int

func TestMapCloneAllocs(t *testing.T) {
	m := map[int]int{1: 1, 2: 2, 3: 3, 4: 4}
	allocs := testing.AllocsPerRun(100, func() {
		clonedMap = maps.Clone(m)
	})
	if allocs != 2 {
		t.Errorf("got %v allocations for maps.Clone, want 2", allocs)
	}
}
