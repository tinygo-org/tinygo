target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str = constant [6 x i8] c"foobar"

declare { i8*, i64, i64 } @runtime.stringToBytes(i8*, i64)

declare void @printSlice(i8* nocapture readonly, i64, i64)

declare void @writeToSlice(i8* nocapture, i64, i64)

; Test that runtime.stringToBytes can be fully optimized away.
define void @testReadOnly() {
entry:
  %0 = call fastcc { i8*, i64, i64 } @runtime.stringToBytes(i8* getelementptr inbounds ([6 x i8], [6 x i8]* @str, i32 0, i32 0), i64 6)
  %1 = extractvalue { i8*, i64, i64 } %0, 0
  %2 = extractvalue { i8*, i64, i64 } %0, 1
  %3 = extractvalue { i8*, i64, i64 } %0, 2
  call fastcc void @printSlice(i8* %1, i64 %2, i64 %3)
  ret void
}

; Test that even though the slice is written to, some values can be propagated.
define void @testReadWrite() {
entry:
  %0 = call fastcc { i8*, i64, i64 } @runtime.stringToBytes(i8* getelementptr inbounds ([6 x i8], [6 x i8]* @str, i32 0, i32 0), i64 6)
  %1 = extractvalue { i8*, i64, i64 } %0, 0
  %2 = extractvalue { i8*, i64, i64 } %0, 1
  %3 = extractvalue { i8*, i64, i64 } %0, 2
  call fastcc void @writeToSlice(i8* %1, i64 %2, i64 %3)
  ret void
}
