target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@intToPtrResult = local_unnamed_addr global i8 2

define void @runtime.initAll() local_unnamed_addr {
  ret void
}
