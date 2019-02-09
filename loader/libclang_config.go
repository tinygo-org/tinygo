// +build !byollvm

package loader

/*
#cgo linux  CFLAGS: -I/usr/lib/llvm-7/include
#cgo darwin CFLAGS: -I/usr/local/opt/llvm/include
#cgo linux  LDFLAGS: -L/usr/lib/llvm-7/lib -lclang
#cgo darwin LDFLAGS: -L/usr/local/opt/llvm/lib -lclang -lffi
*/
import "C"
