target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

declare nonnull i8* @runtime.alloc(i32)

; Test allocating a single int (i32) that should be allocated on the stack.
define void @testInt() {
  %1 = call i8* @runtime.alloc(i32 4)
  %2 = bitcast i8* %1 to i32*
  store i32 5, i32* %2
  ret void
}

; Test allocating an array of 3 i16 values that should be allocated on the
; stack.
define i16 @testArray() {
  %1 = call i8* @runtime.alloc(i32 6)
  %2 = bitcast i8* %1 to i16*
  %3 = getelementptr i16, i16* %2, i32 1
  store i16 5, i16* %3
  %4 = getelementptr i16, i16* %2, i32 2
  %5 = load i16, i16* %4
  ret i16 %5
}

; Call a function that will let the pointer escape, so the heap-to-stack
; transform shouldn't be applied.
define void @testEscapingCall() {
  %1 = call i8* @runtime.alloc(i32 4)
  %2 = bitcast i8* %1 to i32*
  %3 = call i32* @escapeIntPtr(i32* %2)
  ret void
}

define void @testEscapingCall2() {
  %1 = call i8* @runtime.alloc(i32 4)
  %2 = bitcast i8* %1 to i32*
  %3 = call i32* @escapeIntPtrSometimes(i32* %2, i32* %2)
  ret void
}

; Call a function that doesn't let the pointer escape.
define void @testNonEscapingCall() {
  %1 = call i8* @runtime.alloc(i32 4)
  %2 = bitcast i8* %1 to i32*
  %3 = call i32* @noescapeIntPtr(i32* %2)
  ret void
}

; Return the allocated value, which lets it escape.
define i32* @testEscapingReturn() {
  %1 = call i8* @runtime.alloc(i32 4)
  %2 = bitcast i8* %1 to i32*
  ret i32* %2
}

declare i32* @escapeIntPtr(i32*)

declare i32* @noescapeIntPtr(i32* nocapture)

declare i32* @escapeIntPtrSometimes(i32* nocapture, i32*)
