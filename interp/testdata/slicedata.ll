target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

; Reproduction of https://github.com/tinygo-org/tinygo/issues/4214.
; Dereferencing the pointer returned by unsafe.SliceData on a zero-capacity
; slice produces a load that is out of bounds of a zero-sized object. The
; interp must defer this load to runtime instead of crashing the compiler.

@main.zeroSized = global {} zeroinitializer
@main.v = global i64 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %val = load i64, ptr @main.zeroSized
  store i64 %val, ptr @main.v
  ret void
}
