// +build !byollvm

package loader

/*
#cgo CFLAGS: -I/usr/lib/llvm-7/include
#cgo LDFLAGS: -L/usr/lib/llvm-7/lib -lclang
*/
import "C"
