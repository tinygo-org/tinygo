// wasi-libc's libc-bottom-half/sources/wasip3.c references
// __component_type_object_force_link_wasip3 to force the linker to retain
// the vendored wasip3_component_type.o object, which encodes WIT type
// metadata for wasi-libc's own crt1 "wasi:cli/run" export.
//
// TinyGo doesn't use wasi-libc's crt1 or that metadata, so this stub satisfies
// the linker without pulling in the real object. See also
// lib/wasi-libc-wasip2-stub.c.
void __component_type_object_force_link_wasip3(void) {}
