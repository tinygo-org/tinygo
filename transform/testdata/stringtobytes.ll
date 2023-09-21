target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str = constant [6 x i8] c"foobar"

declare { ptr, i64, i64 } @runtime.stringToBytes(ptr, i64)

declare void @printSlice(ptr nocapture readonly, i64, i64)

declare void @writeToSlice(ptr nocapture, i64, i64)

; Test that runtime.stringToBytes can be fully optimized away.
define void @testReadOnly() {
entry:
  %0 = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %1 = extractvalue { ptr, i64, i64 } %0, 0
  %2 = extractvalue { ptr, i64, i64 } %0, 1
  %3 = extractvalue { ptr, i64, i64 } %0, 2
  call fastcc void @printSlice(ptr %1, i64 %2, i64 %3)
  ret void
}

; Test that even though the slice is written to, some values can be propagated.
define void @testReadWrite() {
entry:
  %0 = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %1 = extractvalue { ptr, i64, i64 } %0, 0
  %2 = extractvalue { ptr, i64, i64 } %0, 1
  %3 = extractvalue { ptr, i64, i64 } %0, 2
  call fastcc void @writeToSlice(ptr %1, i64 %2, i64 %3)
  ret void
}

; Test that pointer values are never propagated if there is even a single write
; to the pointer value (but len/cap values still can be).
define void @testReadSome() {
entry:
  %s = call fastcc { ptr, i64, i64 } @runtime.stringToBytes(ptr @str, i64 6)
  %s.ptr = extractvalue { ptr, i64, i64 } %s, 0
  %s.len = extractvalue { ptr, i64, i64 } %s, 1
  %s.cap = extractvalue { ptr, i64, i64 } %s, 2
  call fastcc void @writeToSlice(ptr %s.ptr, i64 %s.len, i64 %s.cap)
  %s.ptr2 = extractvalue { ptr, i64, i64 } %s, 0
  %s.len2 = extractvalue { ptr, i64, i64 } %s, 1
  %s.cap2 = extractvalue { ptr, i64, i64 } %s, 2
  call fastcc void @printSlice(ptr %s.ptr2, i64 %s.len2, i64 %s.cap2)
  ret void
}
