target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalAggregate({ ptr })

@main.value = global i32 1
@main.result = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; The pointer is hidden inside an aggregate argument.
  %arg = insertvalue { ptr } undef, ptr @main.value, 0

  ; This call runs at runtime and may modify @main.value.
  call void @externalAggregate({ ptr } %arg)

  ; Therefore this load must also remain at runtime.
  %value = load i32, ptr @main.value
  store i32 %value, ptr @main.result
  ret void
}
