target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

declare i64 @externalCall(ptr, i32, i64)

define internal i64 @testCall(ptr %ptr, i32 %len, i64 %foo) {
  %val = call i64 @externalCall(ptr %ptr, i32 %len, i64 %foo)
  ret i64 %val
}

define internal i64 @testCallNonEntry(ptr %ptr, i32 %len) {
entry:
  br label %bb1

bb1:
  %val = call i64 @externalCall(ptr %ptr, i32 %len, i64 3)
  ret i64 %val
}

define void @exportedFunction(i64 %foo) {
  %unused = shl i64 %foo, 1
  ret void
}

define internal void @callExportedFunction(i64 %foo) {
  call void @exportedFunction(i64 %foo)
  ret void
}
