// +build !byollvm
// +build llvm12

package cgo

/*
#cgo linux   CFLAGS:  -I/usr/lib/llvm-12/include
#cgo darwin  CFLAGS:  -I/usr/local/opt/llvm@12/include
#cgo freebsd CFLAGS:  -I/usr/local/llvm12/include
#cgo linux   LDFLAGS: -L/usr/lib/llvm-12/lib -lclang
#cgo darwin  LDFLAGS: -L/usr/local/opt/llvm@12/lib -lclang -lffi
#cgo freebsd LDFLAGS: -L/usr/local/llvm12/lib -lclang
*/
import "C"
