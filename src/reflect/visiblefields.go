package reflect

import (
	"internal/reflectlite"
	"unsafe"
)

func VisibleFields(t Type) []StructField {
	fields := reflectlite.VisibleFields(toRawType(t))
	return *(*[]StructField)(unsafe.Pointer(&fields))
}
