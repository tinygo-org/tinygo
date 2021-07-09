target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@runtime.zeroSizedAlloc = internal global i8 0, align 1

declare nonnull i8* @runtime.alloc(i32, i8*)

define void @testInt() {
  %stackalloc.alloca = alloca [1 x i32], align 4
  store [1 x i32] zeroinitializer, [1 x i32]* %stackalloc.alloca, align 4
  %stackalloc = bitcast [1 x i32]* %stackalloc.alloca to i32*
  store i32 5, i32* %stackalloc, align 4
  ret void
}

define i16 @testArray() {
  %stackalloc.alloca = alloca [2 x i32], align 4
  store [2 x i32] zeroinitializer, [2 x i32]* %stackalloc.alloca, align 4
  %stackalloc = bitcast [2 x i32]* %stackalloc.alloca to i16*
  %1 = getelementptr i16, i16* %stackalloc, i32 1
  store i16 5, i16* %1, align 2
  %2 = getelementptr i16, i16* %stackalloc, i32 2
  %3 = load i16, i16* %2, align 2
  ret i16 %3
}

define void @testEscapingCall() {
  %1 = call i8* @runtime.alloc(i32 4, i8* null)
  %2 = bitcast i8* %1 to i32*
  %3 = call i32* @escapeIntPtr(i32* %2)
  ret void
}

define void @testEscapingCall2() {
  %1 = call i8* @runtime.alloc(i32 4, i8* null)
  %2 = bitcast i8* %1 to i32*
  %3 = call i32* @escapeIntPtrSometimes(i32* %2, i32* %2)
  ret void
}

define void @testNonEscapingCall() {
  %stackalloc.alloca = alloca [1 x i32], align 4
  store [1 x i32] zeroinitializer, [1 x i32]* %stackalloc.alloca, align 4
  %stackalloc = bitcast [1 x i32]* %stackalloc.alloca to i32*
  %1 = call i32* @noescapeIntPtr(i32* %stackalloc)
  ret void
}

define i32* @testEscapingReturn() {
  %1 = call i8* @runtime.alloc(i32 4, i8* null)
  %2 = bitcast i8* %1 to i32*
  ret i32* %2
}

define void @testNonEscapingLoop() {
entry:
  %stackalloc.alloca = alloca [1 x i32], align 4
  br label %loop

loop:                                             ; preds = %loop, %entry
  store [1 x i32] zeroinitializer, [1 x i32]* %stackalloc.alloca, align 4
  %stackalloc = bitcast [1 x i32]* %stackalloc.alloca to i32*
  %0 = call i32* @noescapeIntPtr(i32* %stackalloc)
  %1 = icmp eq i32* null, %0
  br i1 %1, label %loop, label %end

end:                                              ; preds = %loop
  ret void
}

define void @testZeroSizedAlloc() {
  %1 = bitcast i8* @runtime.zeroSizedAlloc to i32*
  %2 = call i32* @noescapeIntPtr(i32* %1)
  ret void
}

declare i32* @escapeIntPtr(i32*)

declare i32* @noescapeIntPtr(i32* nocapture)

declare i32* @escapeIntPtrSometimes(i32* nocapture, i32*)
