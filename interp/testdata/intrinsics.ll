target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@uminResult = global i32 0
@sminResult = global i32 0
@umaxResult = global i32 0
@smaxResult = global i32 0

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  call void @testUMin()
  call void @testSMin()
  call void @testUMax()
  call void @testSMax()
  ret void
}

define internal void @testUMin() {
  %umin = call i32 @llvm.umin.i32(i32 12, i32 -1)
  store i32 %umin, ptr @uminResult
  ret void
}

declare i32 @llvm.umin.i32(i32, i32)

define internal void @testSMin() {
  %smin = call i32 @llvm.smin.i32(i32 12, i32 -1)
  store i32 %smin, ptr @sminResult
  ret void
}

declare i32 @llvm.smin.i32(i32, i32)

define internal void @testUMax() {
  %umax = call i32 @llvm.umax.i32(i32 12, i32 -1)
  store i32 %umax, ptr @umaxResult
  ret void
}

declare i32 @llvm.umax.i32(i32, i32)

define internal void @testSMax() {
  %smax = call i32 @llvm.smax.i32(i32 12, i32 -1)
  store i32 %smax, ptr @smaxResult
  ret void
}

declare i32 @llvm.smax.i32(i32, i32)
