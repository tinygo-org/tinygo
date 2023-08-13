//go:build none

// This file is included in the build on systems that support the Boehm GC.

typedef void (* GC_push_other_roots_proc)(void);
void GC_set_push_other_roots(GC_push_other_roots_proc);

void tinygo_runtime_bdwgc_callback(void);

static void callback(void) {
    tinygo_runtime_bdwgc_callback();
}

void tinygo_runtime_bdwgc_init(void) {
    GC_set_push_other_roots(callback);
}
