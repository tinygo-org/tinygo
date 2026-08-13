target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.value = global i32 1
@main.result = local_unnamed_addr global i32 0

declare void @externalAggregate({ ptr }) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @externalAggregate({ ptr } { ptr @main.value })
  %value = load i32, ptr @main.value, align 4
  store i32 %value, ptr @main.result, align 4
  ret void
}
