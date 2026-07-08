target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.zeroSized = local_unnamed_addr global {} zeroinitializer
@main.v = local_unnamed_addr global i64 0

define void @runtime.initAll() unnamed_addr {
entry:
  %val = load i64, ptr @main.zeroSized, align 8
  store i64 %val, ptr @main.v, align 8
  ret void
}
