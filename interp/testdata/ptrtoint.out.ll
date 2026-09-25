target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = local_unnamed_addr global i32 10
@main.v2 = local_unnamed_addr global i64 65535
@main.v3 = local_unnamed_addr global i8 0
@main.global = global i8 0
@main.v4 = local_unnamed_addr global i8 2
@main.v5 = local_unnamed_addr global i8 0
@main.v6 = local_unnamed_addr global { i8, i8 } { i8 ptrtoint (ptr @main.global to i8), i8 7 }
@main.v7 = local_unnamed_addr global i8 0

define void @runtime.initAll() unnamed_addr {
entry:
  store i8 ptrtoint (ptr @main.global to i8), ptr @main.v3, align 1
  store i8 ptrtoint (ptr @main.global to i8), ptr @main.v5, align 1
  %v7 = load i8, ptr getelementptr inbounds (i8, ptr @main.v6, i64 1), align 1
  store i8 %v7, ptr @main.v7, align 1
  ret void
}
