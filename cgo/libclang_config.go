// +build !byollvm
// +build !llvm12

package cgo

/*
#cgo linux   CFLAGS:  -I/usr/lib/llvm-11/include
#cgo darwin  CFLAGS:  -I/usr/local/opt/llvm@11/include
#cgo freebsd CFLAGS:  -I/usr/local/llvm11/include
#cgo linux   LDFLAGS: -L/usr/lib/llvm-11/lib -lclang
#cgo darwin  LDFLAGS: -L/usr/local/opt/llvm@11/lib -lclang -lffi
#cgo freebsd LDFLAGS: -L/usr/local/llvm11/lib -lclang
*/
import "C"
