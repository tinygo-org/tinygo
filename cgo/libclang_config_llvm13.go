// +build !byollvm
// +build llvm13

package cgo

/*
#cgo linux   CFLAGS:  -I/usr/lib/llvm-13/include
#cgo darwin  CFLAGS:  -I/usr/local/opt/llvm@13/include
#cgo freebsd CFLAGS:  -I/usr/local/llvm13/include
#cgo linux   LDFLAGS: -L/usr/lib/llvm-13/lib -lclang
#cgo darwin  LDFLAGS: -L/usr/local/opt/llvm@13/lib -lclang -lffi
#cgo freebsd LDFLAGS: -L/usr/local/llvm13/lib -lclang
*/
import "C"
