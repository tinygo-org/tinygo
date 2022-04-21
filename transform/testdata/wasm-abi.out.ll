target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

; Function Attrs: nounwind
declare i64 @"externalCall$i64wrap"(i8*, i32, i64) #0

define internal i64 @testCall(i8* %ptr, i32 %len, i64 %foo) {
  %i64asptr = alloca i64, align 8
  %i64asptr1 = alloca i64, align 8
  store i64 %foo, i64* %i64asptr1, align 8
  call void @externalCall(i64* %i64asptr, i8* %ptr, i32 %len, i64* %i64asptr1)
  %retval = load i64, i64* %i64asptr, align 8
  ret i64 %retval
}

define internal i64 @testCallNonEntry(i8* %ptr, i32 %len) {
entry:
  %i64asptr = alloca i64, align 8
  %i64asptr1 = alloca i64, align 8
  br label %bb1

bb1:                                              ; preds = %entry
  store i64 3, i64* %i64asptr1, align 8
  call void @externalCall(i64* %i64asptr, i8* %ptr, i32 %len, i64* %i64asptr1)
  %retval = load i64, i64* %i64asptr, align 8
  ret i64 %retval
}

; Function Attrs: nounwind
define internal void @"exportedFunction$i64wrap"(i64 %foo) unnamed_addr #0 {
  %unused = shl i64 %foo, 1
  ret void
}

define internal void @callExportedFunction(i64 %foo) {
  call void @"exportedFunction$i64wrap"(i64 %foo)
  ret void
}

declare void @externalCall(i64*, i8*, i32, i64*)

define void @exportedFunction(i64* %0) {
entry:
  %i64 = load i64, i64* %0, align 8
  call void @"exportedFunction$i64wrap"(i64 %i64)
  ret void
}

attributes #0 = { nounwind }
