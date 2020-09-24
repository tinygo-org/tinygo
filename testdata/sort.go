package main

import "sort"

// sort.Slice implicitly uses reflect.Swapper

func strings() {
	data := []string{"aaaa", "cccc", "bbb", "fff", "ggg"}
	sort.Slice(data, func(i, j int) bool {
		return data[i] > data[j]
	})
	println("strings")
	for _, d := range data {
		println(d)
	}
}

func int64s() {
	sd := []int64{1, 6, 3, 2, 1923, 123, -123, -29, 3, 0, 1}
	sort.Slice(sd, func(i, j int) bool {
		return sd[i] > sd[j]
	})
	println("int64s")
	for _, d := range sd {
		println(d)
	}

	ud := []uint64{1, 6, 3, 2, 1923, 123, 29, 3, 0, 1}
	sort.Slice(ud, func(i, j int) bool {
		return ud[i] > ud[j]
	})
	println("uint64s")
	for _, d := range ud {
		println(d)
	}
}

func int32s() {
	sd := []int32{1, 6, 3, 2, 1923, 123, -123, -29, 3, 0, 1}
	sort.Slice(sd, func(i, j int) bool {
		return sd[i] > sd[j]
	})
	println("int32s")
	for _, d := range sd {
		println(d)
	}

	ud := []uint32{1, 6, 3, 2, 1923, 123, 29, 3, 0, 1}
	sort.Slice(ud, func(i, j int) bool {
		return ud[i] > ud[j]
	})
	println("uint32s")
	for _, d := range ud {
		println(d)
	}
}

func int16s() {
	sd := []int16{1, 6, 3, 2, 1923, 123, -123, -29, 3, 0, 1}
	sort.Slice(sd, func(i, j int) bool {
		return sd[i] > sd[j]
	})
	println("int16s")
	for _, d := range sd {
		println(d)
	}

	ud := []uint16{1, 6, 3, 2, 1923, 123, 29, 3, 0, 1}
	sort.Slice(ud, func(i, j int) bool {
		return ud[i] > ud[j]
	})
	println("uint16s")
	for _, d := range ud {
		println(d)
	}
}

func int8s() {
	sd := []int8{1, 6, 3, 2, 123, -123, -29, 3, 0, 1}
	sort.Slice(sd, func(i, j int) bool {
		return sd[i] > sd[j]
	})
	println("int8s")
	for _, d := range sd {
		println(d)
	}

	ud := []uint8{1, 6, 3, 2, 123, 29, 3, 0, 1}
	sort.Slice(ud, func(i, j int) bool {
		return ud[i] > ud[j]
	})
	println("uint16s")
	for _, d := range ud {
		println(d)
	}
}

func ints() {
	sd := []int{1, 6, 3, 2, 123, -123, -29, 3, 0, 1}
	sort.Slice(sd, func(i, j int) bool {
		return sd[i] > sd[j]
	})
	println("ints")
	for _, d := range sd {
		println(d)
	}

	ud := []uint{1, 6, 3, 2, 123, 29, 3, 0, 1}
	sort.Slice(ud, func(i, j int) bool {
		return ud[i] > ud[j]
	})
	println("uints")
	for _, d := range ud {
		println(d)
	}
}

func structs() {
	type s struct {
		name string
		a    uint64
		b    uint32
		c    uint16
		d    int
		e    *struct {
			aa uint16
			bb int
		}
	}

	data := []s{
		{
			name: "struct 1",
			d:    100,
			e: &struct {
				aa uint16
				bb int
			}{aa: 123, bb: -10},
		},
		{
			name: "struct 2",
			d:    1,
			e: &struct {
				aa uint16
				bb int
			}{aa: 15, bb: 10},
		},
		{
			name: "struct 3",
			d:    10,
			e: &struct {
				aa uint16
				bb int
			}{aa: 31, bb: -1030},
		},
		{
			name: "struct 4",
			e: &struct {
				aa uint16
				bb int
			}{},
		},
	}
	sort.Slice(data, func(i, j int) bool {
		di := data[i]
		dj := data[j]
		return di.d*di.e.bb > dj.d*dj.e.bb
	})
	println("structs")
	for _, d := range data {
		println(d.name)
	}
}

func main() {
	strings()
	int64s()
	int32s()
	int16s()
	int8s()
	ints()
	structs()
}
