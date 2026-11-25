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
