target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = local_unnamed_addr global i1 true
@main.v2 = local_unnamed_addr global i1 false

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
