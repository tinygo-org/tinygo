target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@runtime.stackChainStart = internal global ptr null
@someGlobal = global i8 3
@ptrGlobal = global ptr null

declare void @runtime.trackPointer(ptr nocapture readonly)

declare noalias nonnull ptr @runtime.alloc(i32, ptr)

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
  %ptr = call ptr @runtime.alloc(i32 4, ptr null)
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr, ptr %3, align 4
  call void @someArbitraryFunction()
  %val = load i8, ptr @someGlobal, align 1
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret ptr %ptr
}

define ptr @needsStackSlots2() {
  %gc.stackobject = alloca { ptr, i32, ptr, ptr, ptr, ptr, ptr }, align 8
  store { ptr, i32, ptr, ptr, ptr, ptr, ptr } { ptr null, i32 5, ptr null, ptr null, ptr null, ptr null, ptr null }, ptr %gc.stackobject, align 4
  %1 = load ptr, ptr @runtime.stackChainStart, align 4
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 0
  store ptr %1, ptr %2, align 4
  store ptr %gc.stackobject, ptr @runtime.stackChainStart, align 4
  %ptr1 = call ptr @getPointer()
  %3 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 4
  store ptr %ptr1, ptr %3, align 4
  %4 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 3
  store ptr %ptr1, ptr %4, align 4
  %5 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %ptr1, ptr %5, align 4
  %ptr2 = getelementptr i8, ptr @someGlobal, i32 0
  %6 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 5
  store ptr %ptr2, ptr %6, align 4
  %unused = call ptr @runtime.alloc(i32 4, ptr null)
  %7 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 6
  store ptr %unused, ptr %7, align 4
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
  %out.alloc = call ptr @runtime.alloc(i32 1, ptr null)
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
  %entry.x = call ptr @runtime.alloc(i32 1, ptr null)
  %2 = getelementptr { ptr, i32, ptr, ptr, ptr, ptr, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %entry.x, ptr %2, align 4
  %entry.y = call ptr @runtime.alloc(i32 1, ptr null)
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
  %other = call ptr @runtime.alloc(i32 1, ptr null)
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
  %x.alloc = call ptr @runtime.alloc(i32 4, ptr null)
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
  %y = call ptr @runtime.alloc(i32 4, ptr null)
  %3 = getelementptr { ptr, i32, ptr }, ptr %gc.stackobject, i32 0, i32 2
  store ptr %y, ptr %3, align 4
  store ptr %y, ptr %x, align 4
  store ptr %x, ptr @ptrGlobal, align 4
  store ptr %1, ptr @runtime.stackChainStart, align 4
  ret void
}
