target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@uminResult = local_unnamed_addr global i32 12
@sminResult = local_unnamed_addr global i32 -1
@umaxResult = local_unnamed_addr global i32 -1
@smaxResult = local_unnamed_addr global i32 12

define void @runtime.initAll() local_unnamed_addr {
  ret void
}
