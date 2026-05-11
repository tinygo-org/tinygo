target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@overlap.buf = local_unnamed_addr global [4 x i8] c"\01\01\02\09"
@alias.src = local_unnamed_addr global [4 x i8] c"\05\06\07\08"
@alias.dst = local_unnamed_addr global [2 x i8] c"\09\07"
@reload.buf = local_unnamed_addr global [4 x i8] c"c\02\03\09"
@reload.out = local_unnamed_addr global [2 x i8] c"\01\02"

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
