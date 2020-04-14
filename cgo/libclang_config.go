// +build !byollvm
// +build !llvm9,!llvm11

package cgo

/*
#cgo linux  CFLAGS: -I/usr/lib/llvm-10/include
#cgo darwin CFLAGS: -I/usr/local/opt/llvm@10/include
#cgo freebsd CFLAGS: -I/usr/local/llvm10/include
#cgo linux  LDFLAGS: -L/usr/lib/llvm-10/lib -lclang
#cgo darwin LDFLAGS: -L/usr/local/opt/llvm@10/lib -lclang -lffi
#cgo freebsd LDFLAGS: -L/usr/local/llvm10/lib -lclang
*/
import "C"
