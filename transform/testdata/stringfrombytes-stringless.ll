target datalayout = "e-m:e-p:32:32-i64:64-v128:64-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime._string = type { ptr, i32 }

declare %runtime._string @runtime.stringFromBytes(ptr, i32, i32, ptr)

declare i1 @runtime.stringLess(ptr, i32, ptr, i32, ptr)

declare void @runtime.trackPointer(ptr, ptr, ptr)

declare void @useString(ptr, i32)

define i1 @main.bytesLess(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr %stackalloc, ptr undef)
  %2 = extractvalue %runtime._string %0, 1
  %3 = call %runtime._string @runtime.stringFromBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %4 = extractvalue %runtime._string %3, 0
  call void @runtime.trackPointer(ptr %4, ptr %stackalloc, ptr undef)
  %5 = extractvalue %runtime._string %3, 1
  %6 = call i1 @runtime.stringLess(ptr %1, i32 %2, ptr %4, i32 %5, ptr undef)
  ret i1 %6
}

define i1 @main.bytesLessString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %2 = extractvalue %runtime._string %0, 1
  %3 = call i1 @runtime.stringLess(ptr %1, i32 %2, ptr %s.data, i32 %s.len, ptr undef)
  ret i1 %3
}

define i1 @main.stringLessBytes(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %2 = extractvalue %runtime._string %0, 1
  %3 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %1, i32 %2, ptr undef)
  ret i1 %3
}

define i1 @main.keepStringConversion(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %2 = extractvalue %runtime._string %0, 1
  call void @useString(ptr %1, i32 %2)
  %3 = call %runtime._string @runtime.stringFromBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %4 = extractvalue %runtime._string %3, 0
  %5 = extractvalue %runtime._string %3, 1
  %6 = call i1 @runtime.stringLess(ptr %1, i32 %2, ptr %4, i32 %5, ptr undef)
  ret i1 %6
}
