// +build !byollvm

package cgo

/*
#cgo linux  CFLAGS: -I/usr/lib/llvm-9/include
#cgo darwin CFLAGS: -I/usr/local/opt/llvm@9/include
#cgo linux  LDFLAGS: -L/usr/lib/llvm-9/lib -lclang
#cgo darwin LDFLAGS: -L/usr/local/opt/llvm@9/lib -lclang -lffi
*/
import "C"
