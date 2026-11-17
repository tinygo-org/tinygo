target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.target = global i8 0
@main.bytes = hidden thread_local(initialexec) local_unnamed_addr addrspace(1) global ptr @main.target, section ".data.bytes", align 8

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
