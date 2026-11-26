//go:build wasip3

package syscall

import "unsafe"

// Types from wasi-libc's wasi/__generated_wasip3.h.
type wasip3Preopen struct {
	descriptor int32
	namePtr    *byte
	nameLen    uintptr
}

type wasip3PreopenList struct {
	ptr *wasip3Preopen
	len uintptr
}

// void filesystem_preopens_get_directories(filesystem_preopens_list_tuple2_own_descriptor_string_t *ret);
//
//export filesystem_preopens_get_directories
func filesystem_preopens_get_directories(ret *wasip3PreopenList)

// void filesystem_descriptor_drop_own(filesystem_own_descriptor_t handle);
//
//export filesystem_descriptor_drop_own
func filesystem_descriptor_drop_own(handle int32)

// This frees the names but does not drop the descriptors.
//
// void filesystem_preopens_list_tuple2_own_descriptor_string_free(filesystem_preopens_list_tuple2_own_descriptor_string_t *ptr);
//
//export filesystem_preopens_list_tuple2_own_descriptor_string_free
func filesystem_preopens_list_tuple2_own_descriptor_string_free(ptr *wasip3PreopenList)

func firstPreopenDirectory() string {
	var list wasip3PreopenList
	filesystem_preopens_get_directories(&list)
	dirs := unsafe.Slice(list.ptr, list.len)
	var name string
	if len(dirs) != 0 {
		name = string(unsafe.Slice(dirs[0].namePtr, dirs[0].nameLen))
	}
	for _, dir := range dirs {
		filesystem_descriptor_drop_own(dir.descriptor)
	}
	filesystem_preopens_list_tuple2_own_descriptor_string_free(&list)
	return name
}
