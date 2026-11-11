package main

// Test that aggregates passed by pointer keep value semantics.

type key [33]byte

type pair [2]key

type record struct {
	id   int
	key  key
	name string
	sum  [20]int32
}

func mk(b byte) (k key) {
	for i := range k {
		k[i] = b + byte(i)
	}
	return k
}

func mutate(k key) byte {
	k[0] = 99
	return k[0]
}

func mutateRecord(r record) record {
	r.key[1] = 77
	r.sum[19] = -1
	r.name += "!"
	return r
}

func sorted(a, b key) pair {
	if a[0] < b[0] {
		return pair{a, b}
	}
	return pair{b, a}
}

func choose(cond bool, a, b key) key {
	x := a
	if cond {
		x = b
	}
	return x
}

func swap(a, b key) (key, key) {
	return b, a
}

func deferred(k key) (out key) {
	defer func(v key) {
		out = v
	}(k)
	k[0] = 55
	return k
}

type keyer interface {
	apply(k key, n int) key
}

type adder byte

func (a adder) apply(k key, n int) key {
	k[n] += byte(a)
	return k
}

type big struct {
	values [20]int32
}

func (b big) apply(k key, n int) key {
	b.values[n] = int32(k[n])
	k[n] = byte(b.values[n] + 1)
	return k
}

type exported struct {
	add byte
}

//export aggregateExportedApply
func (e exported) apply(k key, n int) key {
	k[n] += e.add
	return k
}

func viaInterface(i keyer, k key) key {
	return i.apply(k, 2)
}

func main() {
	a, b := mk(1), mk(9)

	println("mutate:", mutate(a), a[0])

	r := record{id: 1, key: a, name: "r"}
	r2 := mutateRecord(r)
	println("record:", r.key[1], r2.key[1], r.sum[19], r2.sum[19], r.name, r2.name)

	p := sorted(b, a)
	println("sorted:", p[0][0], p[1][0], p[0][32], p[1][32])

	println("choose:", choose(false, a, b)[3], choose(true, a, b)[3])

	x, y := swap(a, b)
	x, y = y, x
	x, y = swap(x, y)
	println("swap:", x[0], y[0])

	println("deferred:", deferred(a)[0], a[0])

	k := a
	f := func() key {
		k[4] = 44
		return k
	}
	c := k
	println("closure:", f()[4], k[4], c[4])

	ch := make(chan key, 1)
	s := a
	ch <- s
	s[5] = 88
	println("channel:", (<-ch)[5], s[5])

	done := make(chan pair)
	go func(q pair) {
		q[0][6] = 66
		done <- q
	}(p)
	q := <-done
	println("goroutine:", q[0][6], p[0][6])

	m := map[key]pair{a: p}
	got := m[a]
	got[1][7] = 33
	println("map:", got[1][7], m[a][1][7], len(m))

	var boxed interface{} = a
	u := boxed.(key)
	u[8] = 22
	println("interface:", u[8], boxed.(key)[8], boxed == interface{}(a))

	arr := []key{a, b}
	for _, v := range arr {
		v[9] = 11
	}
	println("range:", arr[0][9], arr[1][9])

	for _, i := range []keyer{adder(5), big{}, exported{add: 7}} {
		out := viaInterface(i, a)
		println("method:", out[2], a[2])
	}

	apply := keyer(exported{add: 1}).apply
	println("method value:", apply(a, 3)[3], a[3])
}
