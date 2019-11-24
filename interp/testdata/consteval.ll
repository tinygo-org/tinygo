target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@intToPtrResult = global i8 0
@ptrToIntResult = global i8 0

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  call void @testIntToPtr()
  call void @testPtrToInt()
  ret void
}

define internal void @testIntToPtr() {
  %nil = icmp eq i8* inttoptr (i64 1024 to i8*), null
  br i1 %nil, label %a, label %b
a:
  ; should not be reached
  store i8 1, i8* @intToPtrResult
  ret void
b:
  ; should be reached
  store i8 2, i8* @intToPtrResult
  ret void
}

define internal void @testPtrToInt() {
  %zero = icmp eq i64 ptrtoint (i8* @ptrToIntResult to i64), 0
  br i1 %zero, label %a, label %b
a:
  ; should not be reached
  store i8 1, i8* @ptrToIntResult
  ret void
b:
  ; should be reached
  store i8 2, i8* @ptrToIntResult
  ret void
}
