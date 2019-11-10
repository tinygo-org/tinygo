// +build amd64 arm,go1.13 arm64 ppc64le ppc64 s390x

package runtime

// This file implements the string counting functions used by the strings
// package, for example. It must be reimplemented here as a replacement for the
// Go stdlib asm implementations, but only when the asm implementations are used
// (this varies by Go version).
// Track this file for updates:
// https://github.com/golang/go/blob/master/src/internal/bytealg/count_native.go

// countString copies the implementation from
// https://github.com/golang/go/blob/67f181bfd84dfd5942fe9a29d8a20c9ce5eb2fea/src/internal/bytealg/count_generic.go#L1
//go:linkname countString internal/bytealg.CountString
func countString(s string, c byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}
