target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 1
@main.g = global ptr null
@main.y = global i64 0

declare void @externalCall(ptr)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; Only the interp state of main.g points to main.x, not its LLVM initializer.
  store ptr @main.x, ptr @main.g
  ; The external call can change main.x through main.g.
  call void @externalCall(ptr @main.g)
  %x = load i64, ptr @main.x
  store i64 %x, ptr @main.y
  ret void
}
