target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@runtime.stackChainStart = external global ptr
@someGlobal = global i8 3
@ptrGlobal = global ptr null

declare void @runtime.trackPointer(ptr nocapture readonly)

declare noalias nonnull ptr @runtime.alloc(i32, ptr)

; Generic function that returns a pointer (that must be tracked).
define ptr @getPointer() {
    ret ptr @someGlobal
}

define ptr @needsStackSlots() {
  ; Tracked pointer. Although, in this case the value is immediately returned
  ; so tracking it is not really necessary.
  %ptr = call ptr @runtime.alloc(i32 4, ptr null)
  call void @runtime.trackPointer(ptr %ptr)
  call void @someArbitraryFunction()
  %val = load i8, ptr @someGlobal
  ret ptr %ptr
}

; Check some edge cases of pointer tracking.
define ptr @needsStackSlots2() {
  ; Only one stack slot should be created for this (but at the moment, one is
  ; created for each call to runtime.trackPointer).
  %ptr1 = call ptr @getPointer()
  call void @runtime.trackPointer(ptr %ptr1)
  call void @runtime.trackPointer(ptr %ptr1)
  call void @runtime.trackPointer(ptr %ptr1)

  ; Create a pointer that does not need to be tracked (but is tracked).
  %ptr2 = getelementptr i8, ptr @someGlobal, i32 0
  call void @runtime.trackPointer(ptr %ptr2)

  ; Here is finally the point where an allocation happens.
  %unused = call ptr @runtime.alloc(i32 4, ptr null)
  call void @runtime.trackPointer(ptr %unused)

  ret ptr %ptr1
}

; Return a pointer from a caller. Because it doesn't allocate, no stack objects
; need to be created.
define ptr @noAllocatingFunction() {
  %ptr = call ptr @getPointer()
  call void @runtime.trackPointer(ptr %ptr)
  ret ptr %ptr
}

define ptr @fibNext(ptr %x, ptr %y) {
  %x.val = load i8, ptr %x
  %y.val = load i8, ptr %y
  %out.val = add i8 %x.val, %y.val
  %out.alloc = call ptr @runtime.alloc(i32 1, ptr null)
  call void @runtime.trackPointer(ptr %out.alloc)
  store i8 %out.val, ptr %out.alloc
  ret ptr %out.alloc
}

define ptr @allocLoop() {
entry:
  %entry.x = call ptr @runtime.alloc(i32 1, ptr null)
  call void @runtime.trackPointer(ptr %entry.x)
  %entry.y = call ptr @runtime.alloc(i32 1, ptr null)
  call void @runtime.trackPointer(ptr %entry.y)
  store i8 1, ptr %entry.y
  br label %loop

loop:
  %prev.y = phi ptr [ %entry.y, %entry ], [ %prev.x, %loop ]
  %prev.x = phi ptr [ %entry.x, %entry ], [ %next.x, %loop ]
  call void @runtime.trackPointer(ptr %prev.x)
  call void @runtime.trackPointer(ptr %prev.y)
  %next.x = call ptr @fibNext(ptr %prev.x, ptr %prev.y)
  call void @runtime.trackPointer(ptr %next.x)
  %next.x.val = load i8, ptr %next.x
  %loop.done = icmp ult i8 40, %next.x.val
  br i1 %loop.done, label %end, label %loop

end:
  ret ptr %next.x
}

declare ptr @arrayAlloc()

define void @testGEPBitcast() {
  %arr = call ptr @arrayAlloc()
  %arr.bitcast = getelementptr [32 x i8], ptr %arr, i32 0, i32 0
  call void @runtime.trackPointer(ptr %arr.bitcast)
  %other = call ptr @runtime.alloc(i32 1, ptr null)
  call void @runtime.trackPointer(ptr %other)
  ret void
}

define void @someArbitraryFunction() {
  ret void
}

define void @earlyPopRegression() {
  %x.alloc = call ptr @runtime.alloc(i32 4, ptr null)
  call void @runtime.trackPointer(ptr %x.alloc)
  ; At this point the pass used to pop the stack chain, resulting in a potential use-after-free during allocAndSave.
  musttail call void @allocAndSave(ptr %x.alloc)
  ret void
}

define void @allocAndSave(ptr %x) {
  %y = call ptr @runtime.alloc(i32 4, ptr null)
  call void @runtime.trackPointer(ptr %y)
  store ptr %y, ptr %x
  store ptr %x, ptr @ptrGlobal
  ret void
}
