//go:build linux
// +build linux

package runtime

const GOOS = "linux"

const (
	// See https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/mman-common.h
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = 0x20
)

// Source: https://github.com/torvalds/linux/blob/master/include/uapi/linux/time.h
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)
