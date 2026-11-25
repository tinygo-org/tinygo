package reflect

import (
	"internal/reflectlite"
	"unsafe"
)

func MakeFunc(typ Type, fn func(args []Value) (results []Value)) Value {
	header := (*struct {
		Context unsafe.Pointer
		Code    unsafe.Pointer
	})(unsafe.Pointer(&fn))
	return Value{reflectlite.MakeFunc(toRawType(typ), header.Context, header.Code)}
}
