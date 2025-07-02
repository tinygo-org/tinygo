target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@runtime.zeroSizedAlloc = internal global i8 0, align 1

declare nonnull ptr @runtime.alloc(i32, ptr)

; Test allocating a single int (i32) that should be allocated on the stack.
define void @testInt() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  store i32 5, ptr %alloc
  ret void
}

; Test allocating an array of 3 i16 values that should be allocated on the
; stack.
define i16 @testArray() {
  %alloc = call align 2 ptr @runtime.alloc(i32 6, ptr null)
  %alloc.1 = getelementptr i16, ptr %alloc, i32 1
  store i16 5, ptr %alloc.1
  %alloc.2 = getelementptr i16, ptr %alloc, i32 2
  %val = load i16, ptr %alloc.2
  ret i16 %val
}

; Test allocating objects with an unknown alignment.
define void @testUnknownAlign() {
  %alloc32 = call ptr @runtime.alloc(i32 32, ptr null)
  store i8 5, ptr %alloc32
  %alloc24 = call ptr @runtime.alloc(i32 24, ptr null)
  store i16 5, ptr %alloc24
  %alloc12 = call ptr @runtime.alloc(i32 12, ptr null)
  store i16 5, ptr %alloc12
  %alloc6 = call ptr @runtime.alloc(i32 6, ptr null)
  store i16 5, ptr %alloc6
  %alloc3 = call ptr @runtime.alloc(i32 3, ptr null)
  store i16 5, ptr %alloc3
  ret void
}

; Call a function that will let the pointer escape, so the heap-to-stack
; transform shouldn't be applied.
define void @testEscapingCall() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  %val = call ptr @escapeIntPtr(ptr %alloc)
  ret void
}

define void @testEscapingCall2() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  %val = call ptr @escapeIntPtrSometimes(ptr %alloc, ptr %alloc)
  ret void
}

; Call a function that doesn't let the pointer escape.
define void @testNonEscapingCall() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  %val = call ptr @noescapeIntPtr(ptr %alloc)
  ret void
}

; Return the allocated value, which lets it escape.
define ptr @testEscapingReturn() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  ret ptr %alloc
}

; Do a non-escaping allocation in a loop.
define void @testNonEscapingLoop() {
entry:
  br label %loop
loop:
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  %ptr = call ptr @noescapeIntPtr(ptr %alloc)
  %result = icmp eq ptr null, %ptr
  br i1 %result, label %loop, label %end
end:
  ret void
}

; Test a zero-sized allocation.
define void @testZeroSizedAlloc() {
  %alloc = call align 1 ptr @runtime.alloc(i32 0, ptr null)
  %ptr = call ptr @noescapeIntPtr(ptr %alloc)
  ret void
}

declare ptr @escapeIntPtr(ptr)

declare ptr @noescapeIntPtr(ptr nocapture)

declare ptr @escapeIntPtrSometimes(ptr nocapture, ptr)
