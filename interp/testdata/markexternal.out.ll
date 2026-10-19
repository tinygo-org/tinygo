target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 1
@main.g = global ptr @main.x
@main.y = local_unnamed_addr global i64 0

declare void @externalCall(ptr) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @externalCall(ptr @main.g)
  %x = load i64, ptr @main.x, align 8
  store i64 %x, ptr @main.y, align 8
  ret void
}
