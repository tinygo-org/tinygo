target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@runtime.stackChainStart = internal global ptr null
@someGlobal = global i8 3
@ptrGlobal = global ptr null
@arrGlobal = global [8 x i8] zeroinitializer
@structGlobal = global { ptr, i32, [2 x ptr] } zeroinitializer
@ptrArrayGlobal = global [8 x ptr] zeroinitializer
@constantPtrGlobal = constant ptr @someGlobal
@runtime.gcGlobalRoots = internal constant [4 x { ptr, i32 }] [{ ptr, i32 } { ptr @ptrGlobal, i32 4 }, { ptr, i32 } { ptr @structGlobal, i32 4 }, { ptr, i32 } { ptr getelementptr (i8, ptr @structGlobal, i32 8), i32 8 }, { ptr, i32 } { ptr @ptrArrayGlobal, i32 32 }]

declare void @runtime.trackPointer(ptr nocapture readonly)

declare noalias nonnull ptr @runtime.alloc(i32, ptr)

define i32 @runtime.gcGlobalRootCount() {
entry:
  ret i32 4
}

define ptr @runtime.gcGlobalRoot(i32 %0) {
entry:
  %1 = getelementptr inbounds [4 x { ptr, i32 }], ptr @runtime.gcGlobalRoots, i32 0, i32 %0
  %2 = getelementptr inbounds nuw { ptr, i32 }, ptr %1, i32 0, i32 0
  %3 = load ptr, ptr %2, align 4
  ret ptr %3
}

define i32 @runtime.gcGlobalRootSize(i32 %0) {
entry:
  %1 = getelementptr inbounds [4 x { ptr, i32 }], ptr @runtime.gcGlobalRoots, i32 0, i32 %0
  %2 = getelementptr inbounds nuw { ptr, i32 }, ptr %1, i32 0, i32 1
  %3 = load i32, ptr %2, align 4
  ret i32 %3
}

define ptr @getPointer() {
  ret ptr @someGlobal
}

define ptr @needsStackSlots() {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %ptr = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr, ptr %3, align 4
  call void @someArbitraryFunction()
  %val = load i8, ptr @someGlobal, align 1
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %ptr
}

define ptr @needsStackSlots2() {
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr } { ptr null, i32 3, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %ptr1 = call ptr @getPointer()
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr1, ptr %3, align 4
  %ptr2 = getelementptr i8, ptr @someGlobal, i32 0
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %ptr2, ptr %4, align 4
  %unused = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %5 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %unused, ptr %5, align 4
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %ptr1
}

define ptr @noAllocatingFunction() {
  %ptr = call ptr @getPointer()
  ret ptr %ptr
}

define ptr @fibNext(ptr %x, ptr %y) {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %x.val = load i8, ptr %x, align 1
  %y.val = load i8, ptr %y, align 1
  %out.val = add i8 %x.val, %y.val
  %out.alloc = call ptr @runtime.alloc(i32 1, ptr inttoptr (i32 3 to ptr))
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %out.alloc, ptr %3, align 4
  store i8 %out.val, ptr %out.alloc, align 1
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %out.alloc
}

define ptr @allocLoop() {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr, ptr, ptr } { ptr null, i32 5, ptr null, ptr null, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %entry.x = call ptr @runtime.alloc(i32 1, ptr inttoptr (i32 3 to ptr))
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %entry.x, ptr %2, align 4
  %entry.y = call ptr @runtime.alloc(i32 1, ptr inttoptr (i32 3 to ptr))
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %entry.y, ptr %3, align 4
  store i8 1, ptr %entry.y, align 1
  br label %loop

loop:                                             ; preds = %loop, %entry
  %prev.y = phi ptr [ %entry.y, %entry ], [ %prev.x, %loop ]
  %prev.x = phi ptr [ %entry.x, %entry ], [ %next.x, %loop ]
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 5
  store ptr %prev.y, ptr %4, align 4
  %5 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %prev.x, ptr %5, align 4
  %next.x = call ptr @fibNext(ptr %prev.x, ptr %prev.y)
  %6 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 6
  store ptr %next.x, ptr %6, align 4
  %next.x.val = load i8, ptr %next.x, align 1
  %loop.done = icmp ult i8 40, %next.x.val
  br i1 %loop.done, label %end, label %loop

end:                                              ; preds = %loop
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %next.x
}

define ptr @loopPhiUntrackedInput(i1 %repeat) {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr } { ptr null, i32 3, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %loop.entry = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %loop.entry, ptr %2, align 4
  br label %loop

loop:                                             ; preds = %loop, %entry
  %loop.cur = phi ptr [ %loop.entry, %entry ], [ %loop.next, %loop ]
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %loop.cur, ptr %3, align 4
  %loop.next = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %loop.next, ptr %4, align 4
  br i1 %repeat, label %loop, label %end

end:                                              ; preds = %loop
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %loop.cur
}

define ptr @nestedLoopPhiUntrackedInput(i1 %repeat.inner, i1 %repeat.outer) {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr, ptr } { ptr null, i32 4, ptr null, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %original = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 5
  store ptr %original, ptr %2, align 4
  br label %outer

outer:                                            ; preds = %latch, %entry
  %outer.ptr = phi ptr [ %original, %entry ], [ %inner.ptr, %latch ]
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %outer.ptr, ptr %3, align 4
  br label %inner

inner:                                            ; preds = %inner, %outer
  %inner.ptr = phi ptr [ %outer.ptr, %outer ], [ %next, %inner ]
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %inner.ptr, ptr %4, align 4
  %next = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %5 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %next, ptr %5, align 4
  br i1 %repeat.inner, label %inner, label %latch

latch:                                            ; preds = %inner
  br i1 %repeat.outer, label %outer, label %end

end:                                              ; preds = %latch
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %original
}

define ptr @duplicateAcrossBranches(i1 %condition) {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %original = call ptr @getPointer()
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %original, ptr %2, align 4
  br i1 %condition, label %left, label %right

left:                                             ; preds = %entry
  br label %join

right:                                            ; preds = %entry
  br label %join

join:                                             ; preds = %right, %left
  %unused = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %original
}

define ptr @duplicateNestedLoopPhis(i1 %repeat.inner, i1 %repeat.outer) {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr, ptr } { ptr null, i32 4, ptr null, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %original = call ptr @getPointer()
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %original, ptr %2, align 4
  br label %outer

outer:                                            ; preds = %latch, %entry
  %outer.ptr = phi ptr [ %original, %entry ], [ %inner.ptr, %latch ]
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %outer.ptr, ptr %3, align 4
  br label %inner

inner:                                            ; preds = %inner, %outer
  %inner.ptr = phi ptr [ %outer.ptr, %outer ], [ %next, %inner ]
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %inner.ptr, ptr %4, align 4
  %next = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %5 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 5
  store ptr %next, ptr %5, align 4
  br i1 %repeat.inner, label %inner, label %latch

latch:                                            ; preds = %inner
  br i1 %repeat.outer, label %outer, label %end

end:                                              ; preds = %latch
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %inner.ptr
}

define ptr @duplicateAcyclicPhi(i1 %condition) {
entry:
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %0 = load ptr, ptr @runtime.stackChainStart, align 4
  %1 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %0, ptr %1, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  br i1 %condition, label %left, label %right

left:                                             ; preds = %entry
  %left.ptr = call ptr @getPointer()
  br label %join

right:                                            ; preds = %entry
  %right.ptr = call ptr @getPointer()
  br label %join

join:                                             ; preds = %right, %left
  %merged = phi ptr [ %left.ptr, %left ], [ %right.ptr, %right ]
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %merged, ptr %2, align 4
  %unused = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  store ptr %0, ptr @runtime.stackChainStart, align 4
  ret ptr %merged
}

declare ptr @arrayAlloc()

define void @testGEPBitcast() {
  %gc.stackobject = alloca { ptr, i32, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr } { ptr null, i32 2, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %arr = call ptr @arrayAlloc()
  %arr.bitcast = getelementptr [32 x i8], ptr %arr, i32 0, i32 0
  %3 = getelementptr { ptr, i32, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %arr.bitcast, ptr %3, align 4
  %other = call ptr @runtime.alloc(i32 1, ptr inttoptr (i32 3 to ptr))
  %4 = getelementptr { ptr, i32, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %other, ptr %4, align 4
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret void
}

define void @someArbitraryFunction() {
  ret void
}

define void @earlyPopRegression() {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %x.alloc = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %x.alloc, ptr %3, align 4
  call void @allocAndSave(ptr %x.alloc)
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret void
}

define void @allocAndSave(ptr %x) {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %y = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %y, ptr %3, align 4
  store ptr %y, ptr %x, align 4
  store ptr %x, ptr @ptrGlobal, align 4
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret void
}

declare void @"(internal/task).Pause"()

define ptr @getAndPause() {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %ptr = call ptr @getPointer()
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr, ptr %3, align 4
  call void @"(internal/task).Pause"()
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %ptr
}

; Function Attrs: memory(readwrite)
declare void @externCallWithMemAttr() #0

define ptr @getAndCallWithMemAttr() {
  %gc.stackobject = alloca { ptr, i32, ptr }, align 8
  store { ptr, i32, ptr } { ptr null, i32 1, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %ptr = call ptr @getPointer()
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr, ptr %3, align 4
  call void @externCallWithMemAttr()
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %ptr
}

define { ptr, i32, i32 } @getSlice() {
  ret { ptr, i32, i32 } { ptr @someGlobal, i32 8, i32 8 }
}

define i32 @copyToSlice(ptr %src.ptr, i32 %src.len, i32 %src.cap) {
  %dst = call { ptr, i32, i32 } @getSlice()
  %dst.ptr = extractvalue { ptr, i32, i32 } %dst, 0
  %dst.len = extractvalue { ptr, i32, i32 } %dst, 1
  %minLen = call i32 @llvm.umin.i32(i32 %dst.len, i32 %src.len)
  call void @llvm.memmove.p0.p0.i32(ptr %dst.ptr, ptr %src.ptr, i32 %minLen, i1 false)
  ret i32 %minLen
}

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.umin.i32(i32, i32) #1

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memmove.p0.p0.i32(ptr nocapture writeonly, ptr nocapture readonly, i32, i1 immarg) #2

attributes #0 = { memory(readwrite) }
attributes #1 = { nocallback nofree nosync nounwind speculatable willreturn memory(none) }
attributes #2 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
