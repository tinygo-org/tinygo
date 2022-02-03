//go:build darwin
// +build darwin

package runtime

const GOOS = "darwin"

const (
	// See https://github.com/golang/go/blob/master/src/syscall/zerrors_darwin_amd64.go
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = 0x1000 // MAP_ANON
)

// Source: https://opensource.apple.com/source/Libc/Libc-1439.100.3/include/time.h.auto.html
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)
