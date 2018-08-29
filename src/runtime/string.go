package runtime

// This file implements functions related to Go strings.

type _string struct {
	length lenType
	ptr    *uint8
}

func stringEqual(x, y string) bool {
	if len(x) != len(y) {
		return false
	}
	for i := 0; i < len(x); i++ {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
