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
  %0 = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %1 = extractvalue { ptr, i64, i64 } %0, 0
  %2 = extractvalue { ptr, i64, i64 } %0, 1
  %3 = extractvalue { ptr, i64, i64 } %0, 2
  %4 = call fastcc i1 @runtime.stringEqual(ptr %1, i64 %2, ptr %1, i64 %2)
  %5 = call fastcc { ptr, i64, i64 } @runtime.stringFromBytes(ptr %1, i64 %2, i64 %3)
  %6 = extractvalue { ptr, i64, i64 } %5, 0
  call fastcc void @maybeSideEffect()
  %7 = call fastcc i1 @runtime.stringEqual(ptr %6, i64 %2, ptr %6, i64 %2)
  %8 = call fastcc { ptr, i64, i64 } @runtime.stringFromBytes(ptr %1, i64 %2, i64 %3)
  %9 = extractvalue { ptr, i64, i64 } %8, 0
  %10 = call fastcc i1 @runtime.stringEqual(ptr %9, i64 %2, ptr %9, i64 %2)
  call fastcc void @readString(ptr %9, i64 %2)
  ret void
}
