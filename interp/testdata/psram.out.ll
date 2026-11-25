target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.table = internal global [4 x i8] zeroinitializer, section ".psram"
@main.normal = internal global [4 x i8] c"\02\00\00\00"

declare void @externalUse(ptr) unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  store i8 1, ptr @main.table, align 1
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @externalUse(ptr @main.table)
  call void @externalUse(ptr @main.normal)
  ret void
}
