package rand

import "errors"

func init() {
	Reader = &reader{}
}

type reader struct {
}

var errRandom = errors.New("failed to obtain random data from rand_s")

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	var randomByte uint32
	for i := range b {
		// Call rand_s every four bytes because it's a C int (always 32-bit in
		// Windows).
		if i%4 == 0 {
			errCode := libc_rand_s(&randomByte)
			if errCode != 0 {
				// According to the documentation, it can return an error.
				return n, errRandom
			}
		} else {
			randomByte >>= 8
		}
		b[i] = byte(randomByte)
	}

	return len(b), nil
}

// Cryptographically secure random number generator.
// https://docs.microsoft.com/en-us/cpp/c-runtime-library/reference/rand-s?view=msvc-170
// errno_t rand_s(unsigned int* randomValue);
//
//export rand_s
func libc_rand_s(randomValue *uint32) int32
