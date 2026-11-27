target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@main.i8 = global [4 x i8] c"abcd"
@main.i16 = global [3 x i16] [i16 258, i16 772, i16 1286]
@main.i32 = global [2 x i32] [i32 16909060, i32 84281096]
@main.i64 = global [2 x i64] [i64 72623859790382856, i64 1]
@main.f32 = global [2 x float] [float 1.0, float 2.0]
@main.f64 = global [2 x double] [double 1.0, double 2.0]

@main.b16 = global i8 0
@main.b32 = global i8 0
@main.b64 = global i8 0
@main.bf32 = global i8 0
@main.bf64 = global i8 0
@main.w16 = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

; Read single bytes from constant data arrays, then write single bytes so the
; arrays are converted back to LLVM constants.
define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %p16 = getelementptr inbounds i8, ptr @main.i16, i32 1
  %b16 = load i8, ptr %p16
  store i8 %b16, ptr @main.b16
  %p32 = getelementptr inbounds i8, ptr @main.i32, i32 5
  %b32 = load i8, ptr %p32
  store i8 %b32, ptr @main.b32
  %p64 = getelementptr inbounds i8, ptr @main.i64, i32 2
  %b64 = load i8, ptr %p64
  store i8 %b64, ptr @main.b64
  %pf32 = getelementptr inbounds i8, ptr @main.f32, i32 7
  %bf32 = load i8, ptr %pf32
  store i8 %bf32, ptr @main.bf32
  %pf64 = getelementptr inbounds i8, ptr @main.f64, i32 15
  %bf64 = load i8, ptr %pf64
  store i8 %bf64, ptr @main.bf64
  %w16 = load i32, ptr @main.i16
  store i32 %w16, ptr @main.w16

  store i8 122, ptr @main.i8
  store i8 9, ptr %p16
  store i8 9, ptr %p32
  store i8 9, ptr %p64
  %wf32 = getelementptr inbounds i8, ptr @main.f32, i32 5
  store i8 0, ptr %wf32
  %wf64 = getelementptr inbounds i8, ptr @main.f64, i32 9
  store i8 0, ptr %wf64
  ret void
}
