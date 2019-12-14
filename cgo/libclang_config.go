// +build !byollvm

package cgo

/*
#cgo linux  CFLAGS: -I/usr/lib/llvm-9/include
#cgo darwin CFLAGS: -I/usr/local/opt/llvm@9/include
#cgo freebsd CFLAGS: -I/usr/local/llvm90/include
#cgo linux  LDFLAGS: -L/usr/lib/llvm-9/lib -lclang
#cgo darwin LDFLAGS: -L/usr/local/opt/llvm@9/lib -lclang -lffi
#cgo freebsd LDFLAGS: -L/usr/local/llvm90/lib -lclang
*/
import "C"
