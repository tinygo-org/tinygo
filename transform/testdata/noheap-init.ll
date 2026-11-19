target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@sink = global i32 0
@escaped = global ptr null

declare ptr @runtime.alloc(i32, ptr, ptr)
declare noalias nonnull ptr @runtime.alloc_noheap(i32, ptr, ptr) #0
declare void @runtime.free(ptr)
declare void @runtime.nilPanic()
declare i32 @runtime.gcGlobalRootCount()
declare ptr @runtime.gcGlobalRoot(i32)
declare i32 @runtime.gcGlobalRootSize(i32)

define void @runtime.hwInit0() {
entry:
  %p = call align 4 ptr @runtime.alloc_noheap(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef)
  store i32 7, ptr %p, align 4
  %v = load i32, ptr %p, align 4
  store i32 %v, ptr @sink, align 4
  ret void
}

define void @runtime.hwInit0Escaping() {
entry:
  %p = call align 4 ptr @runtime.alloc_noheap(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef)
  store ptr %p, ptr @escaped, align 4
  ret void
}

define void @main.main() {
entry:
  call void @runtime.hwInit0()
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" }
