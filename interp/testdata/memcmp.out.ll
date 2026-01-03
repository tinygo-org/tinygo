target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@cmpLen0 = local_unnamed_addr global i32 0
@cmp12 = local_unnamed_addr global i32 -1
@cmp21 = local_unnamed_addr global i32 1
@cmp11 = local_unnamed_addr global i32 0
@cmp22 = local_unnamed_addr global i32 0
@cmp12Partial = local_unnamed_addr global i32 0
@cmp21Partial = local_unnamed_addr global i32 0

define void @runtime.initAll() unnamed_addr {
  ret void
}
