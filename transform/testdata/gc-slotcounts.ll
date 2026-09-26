target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@runtime.stackChainStart = external global ptr

declare void @runtime.trackPointer(ptr nocapture readonly)

declare noalias nonnull ptr @runtime.alloc(i32, ptr)

declare ptr @getPointer()

; The pointer entering the loop is itself a merge, but %preheader runs at most
; once per call, so its slot is stored once and never overwritten: it roots
; whichever of %left.ptr and %right.ptr the run actually selected for the rest
; of the frame, and the other was never computed on this path. Neither input
; needs a slot of its own.
define ptr @acyclicPhiIntoLoop(i1 %condition, i1 %repeat) {
entry:
  br i1 %condition, label %left, label %right

left:
  %left.ptr = call ptr @getPointer()
  br label %preheader

right:
  %right.ptr = call ptr @getPointer()
  br label %preheader

preheader:
  %entry.ptr = phi ptr [ %left.ptr, %left ], [ %right.ptr, %right ]
  br label %loop

loop:
  %merged = phi ptr [ %entry.ptr, %preheader ], [ %body.ptr, %loop ]
  call void @runtime.trackPointer(ptr %merged)
  %body.ptr = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br i1 %repeat, label %loop, label %end

end:
  ret ptr %entry.ptr
}

; Both merge blocks repeat, so both phis lose their root on the way round and
; every input needs a slot. This must not change when the expansion above is
; narrowed: it is the evidence the fix itself is still intact.
define ptr @nestedLoopPhis(i1 %repeat.inner, i1 %repeat.outer) {
entry:
  %original = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br label %outer

outer:
  %outer.ptr = phi ptr [ %original, %entry ], [ %inner.ptr, %latch ]
  br label %inner

inner:
  %inner.ptr = phi ptr [ %outer.ptr, %outer ], [ %next, %inner ]
  call void @runtime.trackPointer(ptr %inner.ptr)
  %next = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br i1 %repeat.inner, label %inner, label %latch

latch:
  br i1 %repeat.outer, label %outer, label %end

end:
  ret ptr %original
}
