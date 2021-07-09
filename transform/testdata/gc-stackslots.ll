target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

%runtime.stackChainObject = type { %runtime.stackChainObject*, i32 }

@runtime.stackChainStart = external global %runtime.stackChainObject*
@someGlobal = global i8 3

declare void @runtime.trackPointer(i8* nocapture readonly)

declare noalias nonnull i8* @runtime.alloc(i32, i8*)

; Generic function that returns a pointer (that must be tracked).
define i8* @getPointer() {
    ret i8* @someGlobal
}

define i8* @needsStackSlots() {
  ; Tracked pointer. Although, in this case the value is immediately returned
  ; so tracking it is not really necessary.
  %ptr = call i8* @runtime.alloc(i32 4, i8* null)
  call void @runtime.trackPointer(i8* %ptr)
  ; Restoring the stack pointer can happen at this position, before the return.
  ; This avoids issues with tail calls.
  call void @someArbitraryFunction()
  %val = load i8, i8* @someGlobal
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
  %unused = call i8* @runtime.alloc(i32 4, i8* null)
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

define i8* @fibNext(i8* %x, i8* %y) {
  %x.val = load i8, i8* %x
  %y.val = load i8, i8* %y
  %out.val = add i8 %x.val, %y.val
  %out.alloc = call i8* @runtime.alloc(i32 1, i8* null)
  call void @runtime.trackPointer(i8* %out.alloc)
  store i8 %out.val, i8* %out.alloc
  ret i8* %out.alloc
}

define i8* @allocLoop() {
entry:
  %entry.x = call i8* @runtime.alloc(i32 1, i8* null)
  call void @runtime.trackPointer(i8* %entry.x)
  %entry.y = call i8* @runtime.alloc(i32 1, i8* null)
  call void @runtime.trackPointer(i8* %entry.y)
  store i8 1, i8* %entry.y
  br label %loop

loop:
  %prev.y = phi i8* [ %entry.y, %entry ], [ %prev.x, %loop ]
  %prev.x = phi i8* [ %entry.x, %entry ], [ %next.x, %loop ]
  call void @runtime.trackPointer(i8* %prev.x)
  call void @runtime.trackPointer(i8* %prev.y)
  %next.x = call i8* @fibNext(i8* %prev.x, i8* %prev.y)
  call void @runtime.trackPointer(i8* %next.x)
  %next.x.val = load i8, i8* %next.x
  %loop.done = icmp ult i8 40, %next.x.val
  br i1 %loop.done, label %end, label %loop

end:
  ret i8* %next.x
}

declare [32 x i8]* @arrayAlloc()

define void @testGEPBitcast() {
  %arr = call [32 x i8]* @arrayAlloc()
  %arr.bitcast = getelementptr [32 x i8], [32 x i8]* %arr, i32 0, i32 0
  call void @runtime.trackPointer(i8* %arr.bitcast)
  %other = call i8* @runtime.alloc(i32 1, i8* null)
  call void @runtime.trackPointer(i8* %other)
  ret void
}

define void @someArbitraryFunction() {
  ret void
}
