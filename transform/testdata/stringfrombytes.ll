target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str = constant [6 x i8] c"foobar"

declare { ptr, i64, i64 } @runtime.stringToBytes(ptr, i64)

declare { ptr, i64 } @runtime.stringFromBytes(ptr, i64, i64)

declare i1 @runtime.stringEqual(ptr nocapture, i64, ptr nocapture, i64)

declare void @maybeSideEffect()

declare void @readString(ptr nocapture, i64)

define void @testReadOnly() {
entry:
  ; Build byte slice.
  %0 = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %1 = extractvalue { ptr, i64, i64 } %0, 0
  %2 = extractvalue { ptr, i64, i64 } %0, 1
  %3 = extractvalue { ptr, i64, i64 } %0, 2

  ; Test that a side-effect free string equality check can optimize the stringFromBytes
  ; call away.
  %4 = call fastcc { ptr, i64, i64 } @runtime.stringFromBytes(ptr %1, i64 %2, i64 %3)
  %5 = extractvalue { ptr, i64, i64 } %4, 0
  %6 = extractvalue { ptr, i64, i64 } %4, 1
  call fastcc i1 @runtime.stringEqual(ptr %5, i64 %6, ptr %5, i64 %6)

  ; Compare it again, but with an intermittent side-effect that blocks the optimization.
  %9 = call fastcc { ptr, i64, i64 } @runtime.stringFromBytes(ptr %1, i64 %2, i64 %3)
  %10 = extractvalue { ptr, i64, i64 } %9, 0
  %11 = extractvalue { ptr, i64, i64 } %9, 1
  ; Function call may write to the slice storage.
  call fastcc void @maybeSideEffect()
  call fastcc i1 @runtime.stringEqual(ptr %10, i64 %11, ptr %10, i64 %11)

  ; Reading the string after comparing should also defeat the optimization.
  %13 = call fastcc { ptr, i64, i64 } @runtime.stringFromBytes(ptr %1, i64 %2, i64 %3)
  %14 = extractvalue { ptr, i64, i64 } %13, 0
  %15 = extractvalue { ptr, i64, i64 } %13, 1
  call fastcc i1 @runtime.stringEqual(ptr %14, i64 %15, ptr %14, i64 %15)
  call fastcc void @readString(ptr %14, i64 %15)
  ret void
}

