target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@store.value = global i32 1
@store.pointer = global ptr null
@insert.value = global i32 1
@insert.count = global i32 0
@insert.pair = global { ptr, i32 } zeroinitializer

declare void @use(ptr)

declare void @consume(i32)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @store.init(ptr undef)
  call void @insert.init(ptr undef)
  ret void
}

; The runtime store lets runtime code read store.value, so the last store must
; also run at runtime.
define internal void @store.init(ptr %context) unnamed_addr {
entry:
  call void @use(ptr @store.pointer)
  store ptr @store.value, ptr @store.pointer
  %ptr = load ptr, ptr @store.pointer
  %value = load i32, ptr %ptr
  call void @consume(i32 %value)
  store i32 2, ptr @store.value
  ret void
}

; The pair is made at runtime, so runtime code can read insert.value through it.
define internal void @insert.init(ptr %context) unnamed_addr {
entry:
  %count = load atomic i32, ptr @insert.count seq_cst, align 4
  %pair = insertvalue { ptr, i32 } { ptr @insert.value, i32 0 }, i32 %count, 1
  store { ptr, i32 } %pair, ptr @insert.pair
  %ptr = load ptr, ptr @insert.pair
  %value = load i32, ptr %ptr
  call void @consume(i32 %value)
  store i32 2, ptr @insert.value
  ret void
}
