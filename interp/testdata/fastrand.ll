target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@runtime.xorshift32State = global i32 1
@runtime.xorshift64State = global i64 1
@foo.rand32 = global i32 0
@foo.rand64 = global i64 0

declare i32 @runtime.fastrand(ptr)

declare i64 @runtime.fastrand64(ptr)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @external.init(ptr undef)
  call void @foo.init(ptr undef)
  ret void
}

; This init is reverted, which marks the RNG state as externally modified.
define internal void @external.init(ptr %context) unnamed_addr {
  %val = load i32, ptr @runtime.xorshift32State
  store i32 %val, ptr @runtime.xorshift32State
  unreachable
}

; The RNG state is unknown here, but the values must still be constant.
define internal void @foo.init(ptr %context) unnamed_addr {
  %a = call i32 @runtime.fastrand(ptr undef)
  store i32 %a, ptr @foo.rand32
  %b = call i64 @runtime.fastrand64(ptr undef)
  store i64 %b, ptr @foo.rand64
  ret void
}
