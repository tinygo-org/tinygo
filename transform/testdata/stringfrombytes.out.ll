target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare { ptr, i64 } @runtime.stringFromBytes(ptr readonly captures(none), i64, i64)

declare void @write()

; Function Attrs: memory(argmem: read)
declare i64 @count(ptr readonly captures(none), i64) #0

declare i64 @countWrites(ptr readonly captures(none), i64)

; Function Attrs: memory(argmem: read)
declare i64 @countCapture(ptr readonly, i64) #0

define i64 @testReadOnly(ptr %p, i64 %len, i64 %cap) {
entry:
  %n = call i64 @count(ptr %p, i64 %len)
  ret i64 %n
}

define i64 @testCalleeWrites(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @countWrites(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

define i64 @testCalleeCaptures(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @countCapture(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

define i64 @testWriteBetween(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  store i8 1, ptr %p, align 1
  %n = call i64 @count(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

define void @testStored(ptr %p, i64 %len, i64 %cap, ptr %out) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  store ptr %s.ptr, ptr %out, align 8
  ret void
}

define i64 @testCalleeLocalWrites(ptr %p, i64 %len, i64 %cap) {
entry:
  %n = call i64 @countLocalImpl(ptr %p, i64 %len)
  ret i64 %n
}

define i64 @countLocalImpl(ptr readonly captures(none) %s, i64 %len) {
entry:
  %it = alloca i64, align 8
  store i64 0, ptr %it, align 8
  %c = load i8, ptr %s, align 1
  %z = zext i8 %c to i64
  ret i64 %z
}

declare i1 @runtime.hashmapStringGet(ptr readonly captures(none), ptr readonly captures(none), i64, ptr captures(none), i64, ptr) #1

define i1 @testMapLookup(ptr %m, ptr %p, i64 %len, i64 %cap) {
entry:
  %slot = alloca i64, align 8
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %p, i64 %len, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

define i1 @testMapLookupNonLocalSlot(ptr %m, ptr %p, i64 %len, i64 %cap, ptr %out) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %s.ptr, i64 %s.len, ptr %out, i64 8, ptr undef)
  ret i1 %ok
}

attributes #0 = { memory(argmem: read) }
attributes #1 = { "tinygo-writes-args-only" }
