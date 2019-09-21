target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str = constant [6 x i8] c"foobar"

declare { i8*, i64, i64 } @runtime.stringToBytes(i8*, i64)

declare void @printSlice(i8* nocapture readonly, i64, i64)

declare void @writeToSlice(i8* nocapture, i64, i64)

define void @testReadOnly() {
entry:
  call fastcc void @printSlice(i8* getelementptr inbounds ([6 x i8], [6 x i8]* @str, i32 0, i32 0), i64 6, i64 6)
  ret void
}

define void @testReadWrite() {
entry:
  %0 = call fastcc { i8*, i64, i64 } @runtime.stringToBytes(i8* getelementptr inbounds ([6 x i8], [6 x i8]* @str, i32 0, i32 0), i64 6)
  %1 = extractvalue { i8*, i64, i64 } %0, 0
  call fastcc void @writeToSlice(i8* %1, i64 6, i64 6)
  ret void
}
