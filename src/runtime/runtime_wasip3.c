//go:build none

// Replacement for the parts of wasi-libc's crt1 that are needed by the
// wasip3 libcall thread context support. See
// lib/wasi-libc/libc-bottom-half/crt/wasip3_symbol_references.h.

#include <stddef.h>
#include <stdint.h>
#include <wasi/version.h>
#include <wasi/wasip3_tls.h>

extern void __wasm_task_hook(uint32_t);
extern void cabi_realloc(void);
__attribute__((used)) static void *__wasm_task_hook_ref = __wasm_task_hook;
__attribute__((used)) static void *cabi_realloc_ref = cabi_realloc;

static size_t tls_size_and_align(size_t *align) {
  *align = __builtin_wasm_tls_align();
  return __builtin_wasm_tls_size();
}

void __wasm_init_tls(void *base);

__attribute__((visibility("default")))
struct __wasilibc_library_tls_info __wasm_library_tls_info = {
  .tls_size_and_align = tls_size_and_align,
  .init_tls = __wasm_init_tls,
};

__asm__(
".globl      __wasm_get_stack_pointer\n"
".type       __wasm_get_stack_pointer,@function\n"
".functype __wasm_get_stack_pointer () -> (i32)\n"
".import_module __wasm_get_stack_pointer, \"env\"\n"
".import_name __wasm_get_stack_pointer, \"__wasm_get_stack_pointer\"\n"

".globl      __wasm_set_stack_pointer\n"
".type       __wasm_set_stack_pointer,@function\n"
".functype __wasm_set_stack_pointer (i32) -> ()\n"
".import_module __wasm_set_stack_pointer, \"env\"\n"
".import_name __wasm_set_stack_pointer, \"__wasm_set_stack_pointer\"\n"

".globl      __wasm_get_tls_base\n"
".type       __wasm_get_tls_base,@function\n"
".functype __wasm_get_tls_base () -> (i32)\n"
".import_module __wasm_get_tls_base, \"env\"\n"
".import_name __wasm_get_tls_base, \"__wasm_get_tls_base\"\n"

".globl      __wasm_set_tls_base\n"
".type       __wasm_set_tls_base,@function\n"
".functype __wasm_set_tls_base (i32) -> ()\n"
".import_module __wasm_set_tls_base, \"env\"\n"
".import_name __wasm_set_tls_base, \"__wasm_set_tls_base\"\n"
);
