target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7em-unknown-unknown-eabi"

@ext = external global [16 x i8]
@main.v = global i64 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

; The i64 at offset 4 of an external global is only 4 byte aligned, and runs at runtime.
define internal void @main.init() unnamed_addr {
entry:
  %p = getelementptr inbounds i8, ptr @ext, i32 4
  %v = load i64, ptr %p, align 4
  store i64 %v, ptr @main.v, align 8
  %w = add i64 %v, 1
  store i64 %w, ptr %p, align 4
  ret void
}
