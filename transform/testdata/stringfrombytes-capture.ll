declare { ptr, i64 } @runtime.stringFromBytes(ptr nocapture readonly, i64, i64)

define ptr @lastByte(ptr readonly captures(ret: address, read_provenance) %src, i64 %len) memory(none) {
  %index = sub i64 %len, 1
  %result = getelementptr i8, ptr %src, i64 %index
  ret ptr %result
}

define i8 @returnCapture() {
  %src = alloca i8
  store i8 65, ptr %src
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %last = call ptr @lastByte(ptr %p, i64 1)
  store i8 66, ptr %src
  %result = load i8, ptr %last
  ret i8 %result
}
