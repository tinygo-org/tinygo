//go:build none

// This file is included in the build on systems that support the Boehm GC,
// despite the //go:build line above.

#include <stdint.h>

typedef void (* GC_push_other_roots_proc)(void);
void GC_set_push_other_roots(GC_push_other_roots_proc);

typedef void(* GC_warn_proc)(const char *msg, uintptr_t arg);
void GC_set_warn_proc(GC_warn_proc p);

void tinygo_runtime_bdwgc_callback(void);

static void callback(void) {
    tinygo_runtime_bdwgc_callback();
}

static void warn_proc(const char *msg, uintptr_t arg) {
}

void tinygo_runtime_bdwgc_init(void) {
    GC_set_push_other_roots(callback);
#if defined(__wasm__)
    // There are a lot of warnings on WebAssembly in the form:
    //
    //     GC Warning: Repeated allocation of very large block (appr. size 68 KiB):
    //         May lead to memory leak and poor performance
    //
    // The usual advice is to use something like GC_malloc_ignore_off_page but
    // unfortunately for most allocations that's not allowed: Go allocations can
    // legitimately hold pointers further than one page in the allocation. So
    // instead we just disable the warning.
    GC_set_warn_proc(warn_proc);
#endif
}
