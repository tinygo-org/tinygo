target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime._string = type { ptr, i32 }

declare %runtime._string @runtime.stringFromBytes(ptr, i32, i32, ptr)

declare void @runtime.trackPointer(ptr, ptr, ptr)

declare void @useString(ptr, i32)

define i32 @main.stringFromBytesLen(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr %stackalloc, ptr undef)
  %2 = extractvalue %runtime._string %0, 1
  %len = extractvalue %runtime._string %0, 1
  ret i32 %len
}

define i32 @main.keepStringConversion(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %2 = extractvalue %runtime._string %0, 1
  call void @useString(ptr %1, i32 %2)
  %len = extractvalue %runtime._string %0, 1
  ret i32 %len
}
