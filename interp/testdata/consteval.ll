target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@intToPtrResult = global i8 0
@ptrToIntResult = global i8 0
@icmpResult = global i8 0
@pointerTagResult = global i64 0
@someArray = internal global {i16, i8, i8} zeroinitializer
@someArrayPointer = global ptr zeroinitializer

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  call void @testIntToPtr()
  call void @testPtrToInt()
  call void @testConstGEP()
  call void @testICmp()
  call void @testPointerTag()
  ret void
}

define internal void @testIntToPtr() {
  %nil = icmp eq ptr inttoptr (i64 1024 to ptr), null
  br i1 %nil, label %a, label %b
a:
  ; should not be reached
  store i8 1, ptr @intToPtrResult
  ret void
b:
  ; should be reached
  store i8 2, ptr @intToPtrResult
  ret void
}

define internal void @testPtrToInt() {
  %zero = icmp eq i64 ptrtoint (ptr @ptrToIntResult to i64), 0
  br i1 %zero, label %a, label %b
a:
  ; should not be reached
  store i8 1, ptr @ptrToIntResult
  ret void
b:
  ; should be reached
  store i8 2, ptr @ptrToIntResult
  ret void
}

define internal void @testConstGEP() {
  store ptr getelementptr inbounds (i8, ptr @someArray, i32 2), ptr @someArrayPointer
  ret void
}

define internal void @testICmp() {
  br i1 icmp eq (i64 ptrtoint (ptr @ptrToIntResult to i64), i64 0), label %equal, label %unequal
equal:
  ; should not be reached
  store i8 1, ptr @icmpResult
  ret void
unequal:
  ; should be reached
  store i8 2, ptr @icmpResult
  ret void
  ret void
}

define internal void @testPointerTag() {
  %val = and i64 ptrtoint (ptr getelementptr inbounds (i8, ptr @someArray, i32 2) to i64), 3
  store i64 %val, ptr @pointerTagResult
  ret void
}
