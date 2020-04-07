// +build !go1.12

package runtime

import "internal/bytealg"

// indexByte provides compatibility with Go 1.11.
// See the following:
// https://github.com/tinygo-org/tinygo/issues/351
// https://github.com/golang/go/commit/ad4a58e31501bce5de2aad90a620eaecdc1eecb8
//go:linkname indexByte strings.IndexByte
func indexByte(s string, c byte) int {
	return bytealg.IndexByteString(s, c)
}
