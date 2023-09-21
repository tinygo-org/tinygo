target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str = constant [6 x i8] c"foobar"

declare { ptr, i64, i64 } @runtime.stringToBytes(ptr, i64)

declare void @printSlice(ptr nocapture readonly, i64, i64)

declare void @writeToSlice(ptr nocapture, i64, i64)

define void @testReadOnly() {
entry:
  call fastcc void @printSlice(ptr @str, i64 6, i64 6)
  ret void
}

define void @testReadWrite() {
entry:
  %0 = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %1 = extractvalue { ptr, i64, i64 } %0, 0
  call fastcc void @writeToSlice(ptr %1, i64 6, i64 6)
  ret void
}

define void @testReadSome() {
entry:
  %s = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %s.ptr = extractvalue { ptr, i64, i64 } %s, 0
  call fastcc void @writeToSlice(ptr %s.ptr, i64 6, i64 6)
  %s.ptr2 = extractvalue { ptr, i64, i64 } %s, 0
  call fastcc void @printSlice(ptr %s.ptr2, i64 6, i64 6)
  ret void
}
