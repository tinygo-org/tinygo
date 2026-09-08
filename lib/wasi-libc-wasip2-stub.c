// wasi-libc's libc-bottom-half/sources/wasip2.c references
// __component_type_object_force_link_wasip2 to force the linker to retain
// the vendored wasip2_component_type.o object, which encodes WIT type
// metadata for wasi-libc's own crt1 "wasi:cli/run" export.
//
// TinyGo doesn't use wasi-libc's crt1 (it provides its own entry points in
// src/runtime/runtime_wasmentry.go) or that metadata (its own
// component-embedding step in builder/build.go generates the wasi:cli/run
// export separately), so this stub satisfies the linker without pulling in
// the real object, which we don't otherwise build.
void __component_type_object_force_link_wasip2(void) {}
