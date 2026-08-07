target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@allocFunction = constant ptr @runtime.alloc

declare ptr @runtime.alloc(i32, ptr)

declare void @use(ptr)

define void @passAllocator() {
entry:
  call void @use(ptr @runtime.alloc)
  ret void
}

define ptr @allocate() {
entry:
  %allocation = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 5 to ptr))
  ret ptr %allocation
}
