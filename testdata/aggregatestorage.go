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

func compareArray(src, other *[4096]byte) bool {
	value := *src
	same := value == *other
	src[0] = 42
	return same
}

func passStorage(value storageValue) byte {
	return value.data[0]
}

func allZero(data *[4096]byte) bool {
	for _, value := range data {
		if value != 0 {
			return false
		}
	}
	return true
}

func main() {
	var src, dst storageValue
	src.data[0] = 7
	src.data[len(src.data)-1] = 9
	src.ptr = &src.data[0]
	copyStorage(&dst, &src)
	println("copy:", dst.data[0], dst.data[len(dst.data)-1], dst.ptr == src.ptr, passStorage(dst))

	resetStorage(&dst)
	println("reset:", allZero(&dst.data), dst.ptr == nil, src.data[0], src.data[len(src.data)-1])

	resetArray(&src.data)
	println("reset array:", allZero(&src.data))

	small := [4]byte{1, 2, 3, 4}
	resetSmallArray(&small)
	println("reset small array:", small == [4]byte{})

	var a, b [4096]byte
	a[1], b[1] = 5, 5
	println("compare:", compareArray(&a, &b), a[0])
}
