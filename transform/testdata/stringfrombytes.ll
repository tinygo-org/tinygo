target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare { ptr, i64 } @runtime.stringFromBytes(ptr nocapture readonly, i64, i64)

declare void @write()

declare i64 @count(ptr nocapture readonly, i64) memory(argmem: read)

declare i64 @countWrites(ptr nocapture readonly, i64)

declare i64 @countCapture(ptr readonly, i64) memory(argmem: read)

; The copy is removed.
define i64 @testReadOnly(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @count(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

; The callee may write memory.
define i64 @testCalleeWrites(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @countWrites(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

; The callee may keep the pointer.
define i64 @testCalleeCaptures(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @countCapture(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

; The bytes may change between the conversion and the use.
define i64 @testWriteBetween(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  store i8 1, ptr %p
  %n = call i64 @count(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

; The string is stored.
define void @testStored(ptr %p, i64 %len, i64 %cap, ptr %out) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  store ptr %s.ptr, ptr %out
  ret void
}

; The callee only writes its own stack memory.
define i64 @testCalleeLocalWrites(ptr %p, i64 %len, i64 %cap) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %n = call i64 @countLocalImpl(ptr %s.ptr, i64 %s.len)
  ret i64 %n
}

define i64 @countLocalImpl(ptr nocapture readonly %s, i64 %len) {
entry:
  %it = alloca i64
  store i64 0, ptr %it
  %c = load i8, ptr %s
  %z = zext i8 %c to i64
  ret i64 %z
}

declare i1 @runtime.hashmapStringGet(ptr readonly nocapture, ptr readonly nocapture, i64, ptr nocapture, i64, ptr) #0

attributes #0 = { "tinygo-writes-args-only" }

; A map lookup only reads the key and writes the result slot.
define i1 @testMapLookup(ptr %m, ptr %p, i64 %len, i64 %cap) {
entry:
  %slot = alloca i64
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %s.ptr, i64 %s.len, ptr %slot, i64 8, ptr undef)
  ret i1 %ok
}

; The result slot is not a local alloca, so it could alias the key bytes.
define i1 @testMapLookupNonLocalSlot(ptr %m, ptr %p, i64 %len, i64 %cap, ptr %out) {
entry:
  %s = call { ptr, i64 } @runtime.stringFromBytes(ptr %p, i64 %len, i64 %cap)
  %s.ptr = extractvalue { ptr, i64 } %s, 0
  %s.len = extractvalue { ptr, i64 } %s, 1
  %ok = call i1 @runtime.hashmapStringGet(ptr %m, ptr %s.ptr, i64 %s.len, ptr %out, i64 8, ptr undef)
  ret i1 %ok
}
