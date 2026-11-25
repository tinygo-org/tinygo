target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@string = internal unnamed_addr constant [3 x i8] c"foo"
@moveDst = global [3 x i8] zeroinitializer
@copyDst = global [3 x i8] zeroinitializer

@externalSrc = external global [2 x i8]
@moveExternalDst = global [2 x i8] zeroinitializer

@moveEscapedSrc = global [4 x i8] c"abcd"
@moveEscapedDst = global [4 x i8] zeroinitializer

@volatileSrc = global [2 x i8] c"xy"
@volatileDst = global [2 x i8] zeroinitializer
@fillDst = global [4 x i8] c"ABCD"
@fillEscapedDst = global [4 x i8] c"ABCD"
@fillVolatileDst = global [4 x i8] c"ABCD"

declare void @use(ptr)

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  call void @testMove()
  call void @testCopy()
  call void @testMoveExternal()
  call void @testMoveEscaped()
  call void @testVolatileCopy()
  call void @testFill()
  call void @testFillEscaped()
  call void @testVolatileFill()
  ret void
}

; Test a simple memmove between globals.
define internal void @testMove() {
  call void @llvm.memmove.p0.p0.i64(ptr @moveDst, ptr @string, i64 3, i1 false)
  ret void
}

; Test a simple memcpy between globals.
define internal void @testCopy() {
  call void @llvm.memcpy.p0.p0.i64(ptr @copyDst, ptr @string, i64 3, i1 false)
  ret void
}

; Test a memmove from an external global.
; This should be run at runtime.
define internal void @testMoveExternal() {
  call void @llvm.memmove.p0.p0.i64(ptr @moveExternalDst, ptr @externalSrc, i64 2, i1 false)
  ret void
}

; Test a memmove from an escaped (and potentially modified) source buffer.
define internal void @testMoveEscaped() {
  call void @use(ptr @moveEscapedSrc)
  call void @llvm.memmove.p0.p0.i64(ptr @moveEscapedDst, ptr @moveEscapedSrc, i64 4, i1 false)
  ret void
}

; Test a volatile memcpy.
; This should always be run at runtime.
define internal void @testVolatileCopy() {
  call void @llvm.memcpy.p0.p0.i64(ptr @volatileDst, ptr @volatileSrc, i64 2, i1 true)
  ret void
}

declare void @llvm.memmove.p0.p0.i64(ptr, ptr, i64, i1)

declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)

define internal void @testFill() {
  %dst = getelementptr [4 x i8], ptr @fillDst, i64 0, i64 1
  call void @llvm.memset.p0.i64(ptr %dst, i8 127, i64 2, i1 false)
  call void @llvm.memset.p0.i64(ptr null, i8 0, i64 0, i1 false)
  ret void
}

define internal void @testFillEscaped() {
  call void @use(ptr @fillEscapedDst)
  %dst = getelementptr [4 x i8], ptr @fillEscapedDst, i64 0, i64 1
  call void @llvm.memset.p0.i64(ptr %dst, i8 127, i64 2, i1 false)
  ret void
}

define internal void @testVolatileFill() {
  %dst = getelementptr [4 x i8], ptr @fillVolatileDst, i64 0, i64 1
  call void @llvm.memset.p0.i64(ptr %dst, i8 127, i64 2, i1 true)
  ret void
}

declare void @llvm.memset.p0.i64(ptr, i8, i64, i1)
