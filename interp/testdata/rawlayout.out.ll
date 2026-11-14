target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.pointerFree = local_unnamed_addr global ptr @"main$alloc"
@main.misaligned = local_unnamed_addr global ptr @"main$alloc.1"
@main.self = local_unnamed_addr global ptr @"main$alloc.2"
@main.straddle = local_unnamed_addr global ptr @"main$alloc.3"
@main.aligned = local_unnamed_addr global ptr @"main$alloc.4"
@"main$alloc" = internal global <{ [8 x i8], ptr }> <{ [8 x i8] zeroinitializer, ptr @main.x }>, align 8
@"main$alloc.1" = internal global <{ [4 x i8], ptr, [4 x i8] }> <{ [4 x i8] zeroinitializer, ptr @main.x, [4 x i8] zeroinitializer }>, align 8
@"main$alloc.2" = internal global <{ [8 x i8], ptr }> <{ [8 x i8] zeroinitializer, ptr @"main$alloc.2" }>, align 8
@"main$alloc.3" = internal global <{ ptr, [4 x i8], ptr, [4 x i8] }> <{ ptr @main.x, [4 x i8] zeroinitializer, ptr @main.x, [4 x i8] zeroinitializer }>, align 8
@"main$alloc.4" = internal global { ptr, i64, i64 } { ptr null, i64 ptrtoint (ptr @main.x to i64), i64 0 }, align 8

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
