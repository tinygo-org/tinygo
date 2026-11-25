target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@store.value = global i32 1
@store.pointer = global ptr null

declare void @use(ptr) local_unnamed_addr

declare void @consume(i32) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @use(ptr @store.pointer)
  store ptr @store.value, ptr @store.pointer, align 8
  %ptr = load ptr, ptr @store.pointer, align 8
  %value = load i32, ptr %ptr, align 4
  call void @consume(i32 %value)
  store i32 2, ptr @store.value, align 4
  ret void
}
