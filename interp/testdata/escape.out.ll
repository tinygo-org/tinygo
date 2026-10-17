target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@store.value = global i32 1
@store.pointer = global ptr null
@insert.value = global i32 1
@insert.count = local_unnamed_addr global i32 0
@insert.pair = local_unnamed_addr global { ptr, i32 } zeroinitializer

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
  %count = load atomic i32, ptr @insert.count seq_cst, align 4
  %pair.insertvalue0 = insertvalue { ptr, i32 } { ptr @insert.value, i32 0 }, i32 %count, 1
  store { ptr, i32 } %pair.insertvalue0, ptr @insert.pair, align 8
  %ptr1 = load ptr, ptr @insert.pair, align 8
  %value2 = load i32, ptr %ptr1, align 4
  call void @consume(i32 %value2)
  store i32 2, ptr @insert.value, align 4
  ret void
}
