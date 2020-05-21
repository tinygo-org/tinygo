package bytealg

const (
	// Index can search any valid length of string.

	MaxLen        = int(-1) >> 31
	MaxBruteForce = MaxLen
)

// Compare two byte slices.
// Returns -1 if the first differing byte is lower in a, or 1 if the first differing byte is greater in b.
// If the byte slices are equal, returns 0.
// If the lengths are different and there are no differing bytes, compares based on length.
func Compare(a, b []byte) int {
	// Compare for differing bytes.
	for i := 0; i < len(a) && i < len(b); i++ {
		switch {
		case a[0] < b[0]:
			return -1
		case a[0] > b[0]:
			return 1
		}
	}

	// Compare lengths.
	switch {
	case len(a) > len(b):
		return 1
	case len(a) < len(b):
		return -1
	default:
		return 0
	}
}

// Count the number of instances of a byte in a slice.
func Count(b []byte, c byte) int {
	// Use a simple implementation, as there is no intrinsic that does this like we want.
	n := 0
	for _, v := range b {
		if v == c {
			n++
		}
	}
	return n
}

// Count the number of instances of a byte in a string.
func CountString(s string, c byte) int {
	// Use a simple implementation, as there is no intrinsic that does this like we want.
	// Currently, the compiler does not generate zero-copy byte-string conversions, so this needs to be seperate from Count.
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// Cutover is not reachable in TinyGo, but must exist as it is referenced.
func Cutover(n int) int {
	// Setting MaxLen and MaxBruteForce should force a different path to be taken.
	// This should never be called.
	panic("cutover is unreachable")
}

// Equal checks if two byte slices are equal.
// It is equivalent to bytes.Equal.
func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// Index finds the base index of the first instance of the byte sequence b in a.
// If a does not contain b, this returns -1.
func Index(a, b []byte) int {
	for i := 0; i <= len(a)-len(b); i++ {
		if Equal(a[i:i+len(b)], b) {
			return i
		}
	}
	return -1
}

// Index finds the index of the first instance of the specified byte in the slice.
// If the byte is not found, this returns -1.
func IndexByte(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}

// Index finds the index of the first instance of the specified byte in the string.
// If the byte is not found, this returns -1.
func IndexByteString(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// Index finds the base index of the first instance of a substring in a string.
// If the substring is not found, this returns -1.
func IndexString(str, sub string) int {
	for i := 0; i <= len(str)-len(sub); i++ {
		if str[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
