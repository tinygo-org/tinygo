target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

; One byte per rune in two planes, like the width table in go-runewidth.
@lut = global [2 x [1114112 x i8]] zeroinitializer

define void @runtime.initAll() unnamed_addr {
entry:
  call void @lut.init(ptr undef)
  ret void
}

define internal void @lut.init(ptr %context) unnamed_addr {
entry:
  store i8 1, ptr @lut
  %last = getelementptr [2 x [1114112 x i8]], ptr @lut, i32 0, i32 0, i32 767
  store i8 2, ptr %last
  ret void
}
