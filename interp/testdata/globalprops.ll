target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.target = global i8 0
@main.bytes = hidden global [8 x i8] zeroinitializer, section ".data.bytes", align 8

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Store a pointer into a byte array, so the global is replaced with one of
  ; the raw memory layout.
  store ptr @main.target, ptr @main.bytes
  ret void
}
