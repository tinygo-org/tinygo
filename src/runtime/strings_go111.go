// +build !go1.12

package runtime

import "internal/bytealg"

// The following functions provide compatibility with Go 1.11.
// See the following:
// https://github.com/tinygo-org/tinygo/issues/351
// https://github.com/golang/go/commit/ad4a58e31501bce5de2aad90a620eaecdc1eecb8

//go:linkname indexByte strings.IndexByte
func indexByte(s string, c byte) int {
	return bytealg.IndexByteString(s, c)
}

//go:linkname bytesEqual bytes.Equal
func bytesEqual(a, b []byte) bool {
	return bytealg.Equal(a, b)
}
