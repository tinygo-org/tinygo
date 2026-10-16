target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@runtime.zeroSizedAlloc = internal global i8 0, align 1

declare nonnull ptr @runtime.alloc(i32, ptr)

define void @testInt() {
  %stackalloc = alloca [4 x i8], align 4
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  store i32 5, ptr %stackalloc, align 4
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  ret void
}

define i16 @testArray() {
  %stackalloc = alloca [6 x i8], align 2
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [6 x i8] zeroinitializer, ptr %stackalloc, align 2
  %alloc.1 = getelementptr i16, ptr %stackalloc, i32 1
  store i16 5, ptr %alloc.1, align 2
  %alloc.2 = getelementptr i16, ptr %stackalloc, i32 2
  %val = load i16, ptr %alloc.2, align 2
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  ret i16 %val
}

define void @testUnknownAlign() {
  %stackalloc4 = alloca [32 x i8], align 8
  %stackalloc3 = alloca [24 x i8], align 8
  %stackalloc2 = alloca [12 x i8], align 8
  %stackalloc1 = alloca [6 x i8], align 8
  %stackalloc = alloca [3 x i8], align 8
  call void @llvm.lifetime.start.p0(ptr %stackalloc4)
  store [32 x i8] zeroinitializer, ptr %stackalloc4, align 8
  store i8 5, ptr %stackalloc4, align 1
  call void @llvm.lifetime.end.p0(ptr %stackalloc4)
  call void @llvm.lifetime.start.p0(ptr %stackalloc3)
  store [24 x i8] zeroinitializer, ptr %stackalloc3, align 8
  store i16 5, ptr %stackalloc3, align 2
  call void @llvm.lifetime.end.p0(ptr %stackalloc3)
  call void @llvm.lifetime.start.p0(ptr %stackalloc2)
  store [12 x i8] zeroinitializer, ptr %stackalloc2, align 8
  store i16 5, ptr %stackalloc2, align 2
  call void @llvm.lifetime.end.p0(ptr %stackalloc2)
  call void @llvm.lifetime.start.p0(ptr %stackalloc1)
  store [6 x i8] zeroinitializer, ptr %stackalloc1, align 8
  store i16 5, ptr %stackalloc1, align 2
  call void @llvm.lifetime.end.p0(ptr %stackalloc1)
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [3 x i8] zeroinitializer, ptr %stackalloc, align 8
  store i16 5, ptr %stackalloc, align 2
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
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
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %val = call ptr @noescapeIntPtr(ptr %stackalloc)
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
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
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  %ptr = call ptr @noescapeIntPtr(ptr %stackalloc)
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  %result = icmp eq ptr null, %ptr
  br i1 %result, label %loop, label %end

end:                                              ; preds = %loop
  ret void
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

define void @testSequentialLifetimes() {
  %stackalloc1 = alloca [12 x i8], align 4
  %stackalloc = alloca [12 x i8], align 4
  call void @llvm.lifetime.start.p0(ptr %stackalloc1)
  store [12 x i8] zeroinitializer, ptr %stackalloc1, align 4
  %v1 = call ptr @noescapeIntPtr(ptr %stackalloc1)
  call void @llvm.lifetime.end.p0(ptr %stackalloc1)
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [12 x i8] zeroinitializer, ptr %stackalloc, align 4
  %v2 = call ptr @noescapeIntPtr(ptr %stackalloc)
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  ret void
}

define void @testCrossBlockUse(i1 %c) {
entry:
  %stackalloc = alloca [4 x i8], align 4
  store [4 x i8] zeroinitializer, ptr %stackalloc, align 4
  store i32 5, ptr %stackalloc, align 4
  br i1 %c, label %then, label %done

then:                                             ; preds = %entry
  %v = call ptr @noescapeIntPtr(ptr %stackalloc)
  br label %done

done:                                             ; preds = %then, %entry
  ret void
}

define ptr @returnsArg(ptr %p) {
  ret ptr %p
}

define void @testReturnedPointer() {
  %stackalloc = alloca [8 x i8], align 4
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [8 x i8] zeroinitializer, ptr %stackalloc, align 4
  %alias = call ptr @returnsArg(ptr %stackalloc)
  store i32 7, ptr %alias, align 4
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  ret void
}

define { ptr, i32 } @returnsArgInAggregate(ptr %p) {
  %agg = insertvalue { ptr, i32 } undef, ptr %p, 0
  %agg2 = insertvalue { ptr, i32 } %agg, i32 3, 1
  ret { ptr, i32 } %agg2
}

define void @testReturnedAggregate() {
  %stackalloc = alloca [8 x i8], align 4
  store [8 x i8] zeroinitializer, ptr %stackalloc, align 4
  %agg = call { ptr, i32 } @returnsArgInAggregate(ptr %stackalloc)
  %ptr = extractvalue { ptr, i32 } %agg, 0
  store i32 7, ptr %ptr, align 4
  ret void
}

define ptr @returnsSecondArg(ptr %unused, ptr %p) {
  ret ptr %p
}

define void @testReturnedDerivedArgument() {
  %stackalloc = alloca [8 x i8], align 4
  call void @llvm.lifetime.start.p0(ptr %stackalloc)
  store [8 x i8] zeroinitializer, ptr %stackalloc, align 4
  %derived = getelementptr i8, ptr %stackalloc, i32 0
  %alias = call ptr @returnsSecondArg(ptr %stackalloc, ptr %derived)
  store i32 9, ptr %alias, align 4
  call void @llvm.lifetime.end.p0(ptr %stackalloc)
  ret void
}

declare ptr @escapeIntPtr(ptr)

declare ptr @noescapeIntPtr(ptr nocapture)

declare ptr @escapeIntPtrSometimes(ptr nocapture, ptr)

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr nocapture) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr nocapture) #0

attributes #0 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
