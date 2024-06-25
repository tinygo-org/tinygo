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
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
  %val = call ptr @escapeIntPtr(ptr %alloc)
  ret void
}

define void @testEscapingCall2() {
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
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
  %alloc = call align 4 ptr @runtime.alloc(i32 4, ptr null)
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

define void @testZeroSizedAlloc() {
  %ptr = call ptr @noescapeIntPtr(ptr @runtime.zeroSizedAlloc)
  ret void
}

declare ptr @escapeIntPtr(ptr)

declare ptr @noescapeIntPtr(ptr nocapture)

declare ptr @escapeIntPtrSometimes(ptr nocapture, ptr)
