target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.stackChainObject = type { %runtime.stackChainObject*, i32 }

@runtime.stackChainStart = external global %runtime.stackChainObject*
@someGlobal = global i8 3

declare void @runtime.trackPointer(i8* nocapture readonly)

declare noalias nonnull i8* @runtime.alloc(i32)

; Generic function that returns a pointer (that must be tracked).
define i8* @getPointer() {
    ret i8* @someGlobal
}

define i8* @needsStackSlots() {
  ; Tracked pointer. Although, in this case the value is immediately returned
  ; so tracking it is not really necessary.
  %ptr = call i8* @runtime.alloc(i32 4)
  call void @runtime.trackPointer(i8* %ptr)
  ret i8* %ptr
}

; Check some edge cases of pointer tracking.
define i8* @needsStackSlots2() {
  ; Only one stack slot should be created for this (but at the moment, one is
  ; created for each call to runtime.trackPointer).
  %ptr1 = call i8* @getPointer()
  call void @runtime.trackPointer(i8* %ptr1)
  call void @runtime.trackPointer(i8* %ptr1)
  call void @runtime.trackPointer(i8* %ptr1)

  ; Create a pointer that does not need to be tracked (but is tracked).
  %ptr2 = getelementptr i8, i8* @someGlobal, i32 0
  call void @runtime.trackPointer(i8* %ptr2)

  ; Here is finally the point where an allocation happens.
  %unused = call i8* @runtime.alloc(i32 4)
  call void @runtime.trackPointer(i8* %unused)

  ret i8* %ptr1
}

; Return a pointer from a caller. Because it doesn't allocate, no stack objects
; need to be created.
define i8* @noAllocatingFunction() {
  %ptr = call i8* @getPointer()
  call void @runtime.trackPointer(i8* %ptr)
  ret i8* %ptr
}
