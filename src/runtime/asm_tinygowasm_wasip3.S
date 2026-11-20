// With the wasip3 libcall thread context the stack pointer is not a wasm
// global. It is read through an import.
.global  tinygo_getCurrentStackPointer
.hidden  tinygo_getCurrentStackPointer
.type    tinygo_getCurrentStackPointer,@function
tinygo_getCurrentStackPointer: // func getCurrentStackPointer() uintptr
    .functype tinygo_getCurrentStackPointer() -> (i32)
    .functype __wasm_get_stack_pointer() -> (i32)
    call __wasm_get_stack_pointer
    return
    end_function

// The host calls cabi_realloc on the small stack that wasi-libc provides for
// the task hook. Run the allocation on a stack of its own instead.
.functype cabi_realloc_go(i32, i32, i32, i32) -> (i32)
.functype __wasm_set_stack_pointer(i32) -> ()
.functype __wasm_get_stack_pointer() -> (i32)
.global  cabi_realloc
.type    cabi_realloc,@function
.export_name cabi_realloc, cabi_realloc
cabi_realloc: // func cabi_realloc(ptr, oldsize, align, newsize unsafe.Pointer) unsafe.Pointer
    .functype cabi_realloc(i32, i32, i32, i32) -> (i32)
    .local i32
    call __wasm_get_stack_pointer
    local.set 4
    i32.const cabi_realloc_stack+16384
    call __wasm_set_stack_pointer
    local.get 0
    local.get 1
    local.get 2
    local.get 3
    call cabi_realloc_go
    local.get 4
    call __wasm_set_stack_pointer
    end_function

.section .bss.cabi_realloc_stack,"",@
.p2align 4
.type cabi_realloc_stack,@object
cabi_realloc_stack:
    .skip 16384
    .size cabi_realloc_stack, 16384
