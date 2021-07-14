target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalCall(i64)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @foo.init(i8* undef, i8* undef)
  call void @main.init(i8* undef, i8* undef)
  ret void
}

define internal void @foo.init(i8* %context, i8* %parentHandle) unnamed_addr {
  unreachable ; this triggers a revert of @foo.init.
}

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  call void @externalCall(i64 3)
  ret void
}
