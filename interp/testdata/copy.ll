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
