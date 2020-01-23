package runtime

//go:linkname indexBytePortable internal/bytealg.IndexByte
func indexBytePortable(s []byte, c byte) int {
	for i, b := range s {
		if b == c {
			return i
		}
	}
	return -1
}
