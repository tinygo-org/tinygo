package main

func copy64(dst, src []uint64) int {
	return copy(dst, src)
}

func arrCopy(dst, src *[4]uint32) {
	copy(dst[:], src[:])
}
