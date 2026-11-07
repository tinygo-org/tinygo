package main

type storageValue struct {
	data [4096]byte
	ptr  *byte
}

func resetStorage(value *storageValue) {
	*value = storageValue{}
}

func resetArray(value *[4096]byte) {
	*value = [4096]byte{}
}

func resetSmallArray(value *[4]byte) {
	*value = [4]byte{}
}

func copyStorage(dst, src *storageValue) {
	value := *src
	src.data[0] = 42
	*dst = value
}

func passStorage(value storageValue) byte {
	return value.data[0]
}

func main() {
	var src, dst storageValue
	src.data[0] = 7
	src.data[len(src.data)-1] = 9
	src.ptr = &src.data[0]
	copyStorage(&dst, &src)
	if dst.data[0] != 7 || dst.data[len(dst.data)-1] != 9 || dst.ptr != src.ptr || passStorage(dst) != 7 {
		panic("aggregate copy changed")
	}
	resetStorage(&dst)
	for _, value := range dst.data {
		if value != 0 {
			panic("aggregate reset failed")
		}
	}
	if dst.ptr != nil || src.data[0] != 42 || src.data[len(src.data)-1] != 9 {
		panic("aggregate reset changed adjacent storage")
	}
	resetArray(&src.data)
	for _, value := range src.data {
		if value != 0 {
			panic("array reset failed")
		}
	}
	small := [4]byte{1, 2, 3, 4}
	resetSmallArray(&small)
	if small != [4]byte{} {
		panic("small array reset failed")
	}
}
