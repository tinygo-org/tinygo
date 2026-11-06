target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.s = global { [7 x i8], ptr } { [7 x i8] zeroinitializer, ptr @main.x }
@main.n = global i16 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
  ; Load one padding byte and the first byte of the pointer.
  %addr = getelementptr i8, ptr @main.s, i64 7
  %n = load i16, ptr %addr
  store i16 %n, ptr @main.n
  ret void
}
