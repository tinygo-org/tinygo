declare { ptr, i64 } @runtime.stringFromBytes(ptr nocapture readonly, i64, i64)

define i8 @readByte(ptr nocapture readonly %src) memory(argmem: read) {
  %result = load i8, ptr %src
  ret i8 %result
}

define i8 @selectedSource(i1 %second) {
  %a = alloca i8
  %b = alloca i8
  store i8 65, ptr %a
  store i8 67, ptr %b
  %src = select i1 %second, ptr %b, ptr %a
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  store i8 66, ptr %a
  %result = call i8 @readByte(ptr %p)
  ret i8 %result
}

define i8 @readAndWrite(ptr nocapture readonly %src, ptr nocapture %out) #0 {
  %result = load i8, ptr %src
  store i8 66, ptr %out
  ret i8 %result
}

define i8 @writableArgument() {
  %src = alloca i8
  store i8 65, ptr %src
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %first = call i8 @readAndWrite(ptr %p, ptr %src)
  %result = call i8 @readAndWrite(ptr %p, ptr %src)
  ret i8 %result
}

define i8 @readOnlyControl() {
  %src = alloca i8
  store i8 65, ptr %src
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %src, i64 1, i64 1)
  %p = extractvalue { ptr, i64 } %s, 0
  %result = call i8 @readByte(ptr %p)
  ret i8 %result
}

attributes #0 = { "tinygo-writes-args-only" }
