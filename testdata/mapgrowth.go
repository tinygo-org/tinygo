package main

func main() {

	m := make(map[int]int)

	const (
		Delete = 500
		N      = Delete * 2
	)

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

	seen := make([]bool, 500)

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
