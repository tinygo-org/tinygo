target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.target = global i8 0
@main.bytes = hidden local_unnamed_addr global ptr @main.target, section ".data.bytes", align 8

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
