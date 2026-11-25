target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@runtime.zeroSizedAlloc = internal global i8 0, align 1

declare nonnull ptr @runtime.alloc(i32, ptr)

define void @testInt() {
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  store i32 5, ptr %stackalloc, align 4
  ret void
}

define i16 @testArray() {
  %stackalloc = alloca [6 x i8], align 2
  store [6 x i8] zeroinitializer, ptr %stackalloc, align 2
  %alloc.1 = getelementptr i16, ptr %stackalloc, i32 1
  store i16 5, ptr %alloc.1, align 2
  %alloc.2 = getelementptr i16, ptr %stackalloc, i32 2
  %val = load i16, ptr %alloc.2, align 2
  ret i16 %val
}

define void @testUnknownAlign() {
  %stackalloc4 = alloca [32 x i8], align 8
  %stackalloc3 = alloca [24 x i8], align 8
  %stackalloc2 = alloca [12 x i8], align 8
  %stackalloc1 = alloca [6 x i8], align 8
  %stackalloc = alloca [3 x i8], align 8
  store [32 x i8] zeroinitializer, ptr %stackalloc4, align 8
  store i8 5, ptr %stackalloc4, align 1
  store [24 x i8] zeroinitializer, ptr %stackalloc3, align 8
  store i16 5, ptr %stackalloc3, align 2
  store [12 x i8] zeroinitializer, ptr %stackalloc2, align 8
  store i16 5, ptr %stackalloc2, align 2
  store [6 x i8] zeroinitializer, ptr %stackalloc1, align 8
  store i16 5, ptr %stackalloc1, align 2
  store [3 x i8] zeroinitializer, ptr %stackalloc, align 8
  store i16 5, ptr %stackalloc, align 2
  ret void
}

define void @testEscapingCall() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %val = call ptr @escapeIntPtr(ptr %alloc)
  ret void
}

define void @testEscapingCall2() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %val = call ptr @escapeIntPtrSometimes(ptr %alloc, ptr %alloc)
  ret void
}

define void @testNonEscapingCall() {
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %val = call ptr @noescapeIntPtr(ptr %stackalloc)
  ret void
}

define ptr @testEscapingReturn() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  ret ptr %alloc
}

define void @testNonEscapingLoop() {
entry:
  %stackalloc = alloca [4 x i8], align 4
  br label %loop

loop:                                             ; preds = %loop, %entry
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %ptr = call ptr @noescapeIntPtr(ptr %stackalloc)
  %result = icmp eq ptr null, %ptr
  br i1 %result, label %loop, label %end

end:                                              ; preds = %loop
  ret void
}

define i32 @testNonEscapingSelect(i1 %condition, ptr %other) {
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %selected = select i1 %condition, ptr %stackalloc, ptr %other
  %value = load i32, ptr %selected, align 4
  ret i32 %value
}

define ptr @testEscapingSelect(i1 %condition, ptr %other) {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %selected = select i1 %condition, ptr %alloc, ptr %other
  ret ptr %selected
}

define i32 @testNonEscapingPhi(i1 %condition, ptr %other) {
entry:
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  br i1 %condition, label %allocated, label %existing

allocated:                                        ; preds = %entry
  br label %merge

existing:                                         ; preds = %entry
  br label %merge

merge:                                            ; preds = %existing, %allocated
  %selected = phi ptr [ %stackalloc, %allocated ], [ %other, %existing ]
  %value = load i32, ptr %selected, align 4
  ret i32 %value
}

define ptr @testEscapingPhi(i1 %condition, ptr %other) {
entry:
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br i1 %condition, label %allocated, label %existing

allocated:                                        ; preds = %entry
  br label %merge

existing:                                         ; preds = %entry
  br label %merge

merge:                                            ; preds = %existing, %allocated
  %selected = phi ptr [ %alloc, %allocated ], [ %other, %existing ]
  ret ptr %selected
}

define i32 @testLoopCarriedPhi(i32 %count, ptr %initial) {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %index = phi i32 [ 0, %entry ], [ %next, %loop ]
  %previous = phi ptr [ %initial, %entry ], [ %alloc, %loop ]
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  store i32 %index, ptr %alloc, align 4
  %value = load i32, ptr %previous, align 4
  %next = add i32 %index, 1
  %done = icmp eq i32 %next, %count
  br i1 %done, label %end, label %loop

end:                                              ; preds = %loop
  ret i32 %value
}

define i32 @testPointerPhiCycle(i1 %condition) {
entry:
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br label %loop

loop:                                             ; preds = %loop, %entry
  %selected = phi ptr [ %alloc, %entry ], [ %next, %loop ]
  %next = getelementptr i8, ptr %selected, i32 0
  br i1 %condition, label %loop, label %end

end:                                              ; preds = %loop
  %value = load i32, ptr %selected, align 4
  ret i32 %value
}

define void @testZeroSizedAlloc() {
  %ptr = call ptr @noescapeIntPtr(ptr @runtime.zeroSizedAlloc)
  ret void
}

define { ptr, ptr } @testEscapingReturnedAggregate() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  %result = call { i32, { ptr, ptr } } @wrapInAggregate(ptr %alloc)
  %err = extractvalue { i32, { ptr, ptr } } %result, 1
  ret { ptr, ptr } %err
}

define i32 @testNonEscapingReturnedAggregate() {
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %result = call { i32, { ptr, ptr } } @wrapInAggregate(ptr %stackalloc)
  %n = extractvalue { i32, { ptr, ptr } } %result, 0
  ret i32 %n
}

define { i32, { ptr, ptr } } @wrapInAggregate(ptr %p) {
  %itf = insertvalue { ptr, ptr } { ptr null, ptr undef }, ptr %p, 1
  %result = insertvalue { i32, { ptr, ptr } } { i32 1, { ptr, ptr } undef }, { ptr, ptr } %itf, 1
  ret { i32, { ptr, ptr } } %result
}

declare ptr @escapeIntPtr(ptr)

declare ptr @noescapeIntPtr(ptr captures(none))

declare ptr @escapeIntPtrSometimes(ptr captures(none), ptr)
